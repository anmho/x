//! Deterministic in-memory gossip protocol simulator.
//!
//! This crate simulates epidemic (gossip) information dissemination between a
//! set of nodes in a single process — no networking involved. Each [`GossipNode`]
//! maintains a state map of `(node_id -> heartbeat_counter)` pairs. During every
//! gossip *tick* a node increments its own heartbeat and pushes its full state to
//! a small, deterministically-chosen subset of its peers. Receiving nodes merge
//! by keeping the maximum heartbeat seen for each key.
//!
//! The primary entry-point is [`GossipCluster`].

use std::collections::hash_map::DefaultHasher;
use std::collections::HashMap;
use std::hash::{Hash, Hasher};

// ---------------------------------------------------------------------------
// GossipNode
// ---------------------------------------------------------------------------

/// A single participant in the gossip cluster.
///
/// The node tracks:
/// - its own string identifier,
/// - the set of peer ids it is directly connected to,
/// - a state map (`node_id -> heartbeat`) holding the latest heartbeat counter
///   it has seen for every node in the cluster,
/// - a *fanout* — the maximum number of peers it contacts per gossip round.
#[derive(Debug, Clone)]
pub struct GossipNode {
    /// Unique identifier for this node.
    pub id: String,
    /// Ids of nodes this node knows about and can gossip to.
    pub known_peers: Vec<String>,
    /// Latest heartbeat counter seen for each node id.
    pub state: HashMap<String, u64>,
    /// How many peers to contact per gossip round.
    pub fanout: usize,
}

impl GossipNode {
    /// Create a new node with the given id and fanout. The state map is
    /// initialised with the node's own heartbeat at `0`.
    pub fn new(id: &str, fanout: usize) -> Self {
        let mut state = HashMap::new();
        state.insert(id.to_string(), 0);
        GossipNode {
            id: id.to_string(),
            known_peers: Vec::new(),
            state,
            fanout,
        }
    }

    /// Merge a received state map into this node's state, keeping the maximum
    /// heartbeat for each key.
    pub fn merge(&mut self, incoming: &HashMap<String, u64>) {
        for (key, &value) in incoming {
            let entry = self.state.entry(key.clone()).or_insert(0);
            if value > *entry {
                *entry = value;
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Peer-selection helper
// ---------------------------------------------------------------------------

/// Deterministically select up to `fanout` peer ids from `peers` using the
/// node's own id concatenated with the current round number as a seed.
///
/// The selection uses [`DefaultHasher`] so it is reproducible across runs on
/// the same platform / Rust toolchain version (good enough for tests; not
/// suitable for cross-binary stability).
fn select_peers<'a>(node_id: &str, round: u64, peers: &'a [String], fanout: usize) -> Vec<&'a String> {
    if peers.is_empty() || fanout == 0 {
        return Vec::new();
    }

    // Build a deterministic ordering by hashing (node_id, round, index).
    let mut indexed: Vec<(u64, &String)> = peers
        .iter()
        .enumerate()
        .map(|(i, peer)| {
            let mut h = DefaultHasher::new();
            node_id.hash(&mut h);
            round.hash(&mut h);
            (i as u64).hash(&mut h);
            (h.finish(), peer)
        })
        .collect();

    // Sort by hash value to get a stable shuffle.
    indexed.sort_by_key(|(hash, _)| *hash);

    indexed.into_iter().take(fanout).map(|(_, peer)| peer).collect()
}

// ---------------------------------------------------------------------------
// GossipCluster
// ---------------------------------------------------------------------------

/// A collection of [`GossipNode`]s that can exchange state via the gossip
/// protocol.
///
/// # Example
/// ```
/// use gossip::GossipCluster;
///
/// let mut cluster = GossipCluster::new();
/// cluster.add_node("a", 2);
/// cluster.add_node("b", 2);
/// cluster.add_node("c", 2);
/// cluster.connect("a", "b");
/// cluster.connect("b", "c");
/// cluster.connect("a", "c");
///
/// let rounds = cluster.rounds_to_converge(100);
/// assert!(rounds.is_some());
/// ```
#[derive(Debug, Default)]
pub struct GossipCluster {
    /// All nodes keyed by their id.
    pub nodes: HashMap<String, GossipNode>,
    /// Internal round counter used to seed peer selection.
    round: u64,
}

impl GossipCluster {
    /// Create an empty cluster.
    pub fn new() -> Self {
        GossipCluster {
            nodes: HashMap::new(),
            round: 0,
        }
    }

    /// Add a new node to the cluster. The node starts with no known peers.
    ///
    /// # Panics
    /// Panics if a node with `id` already exists.
    pub fn add_node(&mut self, id: &str, fanout: usize) {
        assert!(
            !self.nodes.contains_key(id),
            "node '{}' already exists",
            id
        );
        self.nodes.insert(id.to_string(), GossipNode::new(id, fanout));
    }

    /// Make nodes `a` and `b` mutual peers (bidirectional edge).
    ///
    /// # Panics
    /// Panics if either node id does not exist in the cluster.
    pub fn connect(&mut self, a: &str, b: &str) {
        assert!(self.nodes.contains_key(a), "node '{}' not found", a);
        assert!(self.nodes.contains_key(b), "node '{}' not found", b);

        let node_a = self.nodes.get_mut(a).unwrap();
        if !node_a.known_peers.contains(&b.to_string()) {
            node_a.known_peers.push(b.to_string());
        }

        let node_b = self.nodes.get_mut(b).unwrap();
        if !node_b.known_peers.contains(&a.to_string()) {
            node_b.known_peers.push(a.to_string());
        }
    }

    /// Execute one gossip round for `node_id`:
    /// 1. Increment the node's own heartbeat counter.
    /// 2. Select up to `fanout` peers deterministically.
    /// 3. Push the node's full state map to each selected peer (merge on
    ///    receipt).
    ///
    /// # Panics
    /// Panics if `node_id` does not exist.
    pub fn tick(&mut self, node_id: &str) {
        let node = self.nodes.get_mut(node_id).expect("node not found");

        // 1. Increment own heartbeat.
        let hb = node.state.entry(node_id.to_string()).or_insert(0);
        *hb += 1;

        // 2. Deterministically select peers.
        let peers_snapshot: Vec<String> = node.known_peers.clone();
        let fanout = node.fanout;
        let state_snapshot: HashMap<String, u64> = node.state.clone();

        let selected: Vec<String> = select_peers(node_id, self.round, &peers_snapshot, fanout)
            .into_iter()
            .cloned()
            .collect();

        // 3. Merge state into each selected peer.
        for peer_id in &selected {
            if let Some(peer) = self.nodes.get_mut(peer_id) {
                peer.merge(&state_snapshot);
            }
        }
    }

    /// Returns `true` if every node in the cluster has an identical state map.
    ///
    /// An empty cluster is considered converged.
    pub fn converged(&self) -> bool {
        let mut iter = self.nodes.values();
        let first = match iter.next() {
            Some(n) => &n.state,
            None => return true,
        };
        iter.all(|n| &n.state == first)
    }

    /// Run gossip rounds in round-robin order (sorted node ids for
    /// determinism) until [`GossipCluster::converged`] returns `true` or
    /// `max_rounds` full passes have been completed.
    ///
    /// Returns `Some(rounds)` where `rounds` is the number of full passes
    /// (each pass = one tick per node) needed, or `None` if convergence was
    /// not achieved within `max_rounds`.
    pub fn rounds_to_converge(&mut self, max_rounds: usize) -> Option<usize> {
        // Collect node ids sorted for a deterministic iteration order.
        let mut node_ids: Vec<String> = self.nodes.keys().cloned().collect();
        node_ids.sort();

        for pass in 1..=max_rounds {
            self.round = pass as u64;
            for id in &node_ids.clone() {
                self.tick(id);
            }
            if self.converged() {
                return Some(pass);
            }
        }
        None
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    /// Helper: build a fully-connected cluster of `n` nodes each with the
    /// given fanout.
    fn fully_connected(n: usize, fanout: usize) -> GossipCluster {
        let mut cluster = GossipCluster::new();
        let ids: Vec<String> = (0..n).map(|i| format!("node{}", i)).collect();
        for id in &ids {
            cluster.add_node(id, fanout);
        }
        for i in 0..ids.len() {
            for j in (i + 1)..ids.len() {
                cluster.connect(&ids[i], &ids[j]);
            }
        }
        cluster
    }

    /// Helper: build a ring topology (each node knows its left and right
    /// neighbour only).
    fn ring(n: usize, fanout: usize) -> GossipCluster {
        let mut cluster = GossipCluster::new();
        let ids: Vec<String> = (0..n).map(|i| format!("node{}", i)).collect();
        for id in &ids {
            cluster.add_node(id, fanout);
        }
        for i in 0..n {
            let next = (i + 1) % n;
            cluster.connect(&ids[i], &ids[next]);
        }
        cluster
    }

    /// A 3-node fully-connected cluster should converge quickly (well within
    /// 10 rounds).
    #[test]
    fn three_node_fully_connected_converges() {
        let mut cluster = fully_connected(3, 2);
        let rounds = cluster.rounds_to_converge(10);
        assert!(
            rounds.is_some(),
            "3-node fully-connected cluster did not converge within 10 rounds"
        );
        assert!(cluster.converged());
    }

    /// A 10-node ring topology should converge within 50 rounds.
    #[test]
    fn ten_node_ring_converges() {
        let mut cluster = ring(10, 2);
        let rounds = cluster.rounds_to_converge(50);
        assert!(
            rounds.is_some(),
            "10-node ring did not converge within 50 rounds (got None)"
        );
        assert!(cluster.converged());
    }

    /// An isolated node (no peers) means the cluster cannot converge because
    /// the isolated node's state will never reach the others (and vice-versa).
    #[test]
    fn isolated_node_prevents_convergence() {
        let mut cluster = GossipCluster::new();
        cluster.add_node("a", 2);
        cluster.add_node("b", 2);
        cluster.add_node("isolated", 2);
        // Only connect a <-> b; "isolated" has no peers.
        cluster.connect("a", "b");

        let result = cluster.rounds_to_converge(20);
        // With 3 nodes but one isolated, the cluster state maps will never be
        // identical, so convergence should not happen.
        assert!(
            result.is_none(),
            "cluster unexpectedly converged despite isolated node"
        );
    }

    /// A node that joins late and is subsequently connected should allow the
    /// cluster to re-converge.
    #[test]
    fn late_joining_node_reconverges() {
        // Start with a 3-node fully-connected cluster and let it converge.
        let mut cluster = fully_connected(3, 2);
        let initial = cluster.rounds_to_converge(20);
        assert!(initial.is_some(), "initial 3-node cluster did not converge");

        // Add a 4th node with no connections yet — convergence should be lost.
        cluster.add_node("late", 2);
        assert!(
            !cluster.converged(),
            "cluster should not be converged after adding an isolated late node"
        );

        // Connect the late node to an existing node.
        cluster.connect("late", "node0");

        // Now the cluster should be able to re-converge.
        let re_converge = cluster.rounds_to_converge(50);
        assert!(
            re_converge.is_some(),
            "cluster did not re-converge after late node joined"
        );
        assert!(cluster.converged());
    }
}
