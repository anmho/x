//! Consistent hashing ring implementation.
//!
//! Consistent hashing assigns keys to nodes in a way that minimises remapping
//! when nodes are added or removed. Each physical node is represented by
//! `virtual_nodes` positions on the ring; more virtual nodes give a more
//! uniform key distribution.
//!
//! # Example
//!
//! ```rust
//! use consistent_hashing::ConsistentHashRing;
//!
//! let mut ring = ConsistentHashRing::new(150);
//! ring.add_node("server-a");
//! ring.add_node("server-b");
//! ring.add_node("server-c");
//!
//! assert!(ring.get_node("my-key").is_some());
//! ```

use std::collections::BTreeMap;

// ---------------------------------------------------------------------------
// Hash function
// ---------------------------------------------------------------------------

/// FNV-1a 64-bit hash — deterministic, dependency-free.
fn fnv1a(data: &[u8]) -> u64 {
    const OFFSET_BASIS: u64 = 14_695_981_039_346_656_037;
    const FNV_PRIME: u64 = 1_099_511_628_211;

    let mut hash = OFFSET_BASIS;
    for &byte in data {
        hash ^= byte as u64;
        hash = hash.wrapping_mul(FNV_PRIME);
    }
    hash
}

/// Compute the ring position for a virtual-node label `"{node}#{replica}"`.
fn virtual_key(node: &str, replica: usize) -> u64 {
    let label = format!("{node}#{replica}");
    fnv1a(label.as_bytes())
}

/// Compute the ring position for an arbitrary string key.
fn key_position(key: &str) -> u64 {
    fnv1a(key.as_bytes())
}

// ---------------------------------------------------------------------------
// ConsistentHashRing
// ---------------------------------------------------------------------------

/// A consistent hashing ring.
///
/// Nodes are placed at `virtual_nodes` positions each around a 2^64 ring.
/// Key lookup finds the next clockwise node from the key's position, wrapping
/// around to the smallest position when the key sits beyond the largest node.
pub struct ConsistentHashRing {
    /// Number of virtual positions (replicas) placed per physical node.
    virtual_nodes: usize,
    /// Map from ring position → physical node name.
    ring: BTreeMap<u64, String>,
}

impl ConsistentHashRing {
    /// Create a new, empty ring.
    ///
    /// `virtual_nodes` controls how many positions each physical node occupies.
    /// Higher values yield more uniform key distribution at the cost of slightly
    /// more memory. A value in the range 100–300 is typical.
    pub fn new(virtual_nodes: usize) -> Self {
        Self {
            virtual_nodes,
            ring: BTreeMap::new(),
        }
    }

    /// Add a physical node to the ring.
    ///
    /// Inserts `virtual_nodes` positions for `node`. If positions collide with
    /// existing entries (extremely rare with FNV-1a on distinct labels) the new
    /// entry overwrites the old one.
    pub fn add_node(&mut self, node: &str) {
        for replica in 0..self.virtual_nodes {
            let pos = virtual_key(node, replica);
            self.ring.insert(pos, node.to_owned());
        }
    }

    /// Remove a physical node from the ring.
    ///
    /// Deletes all virtual positions that belong to `node`. Keys previously
    /// routed to `node` will be redistributed among the remaining nodes.
    pub fn remove_node(&mut self, node: &str) {
        for replica in 0..self.virtual_nodes {
            let pos = virtual_key(node, replica);
            // Only remove the entry if it actually maps to this node (guards
            // against the vanishingly rare hash collision case).
            if self.ring.get(&pos).map(|n| n.as_str()) == Some(node) {
                self.ring.remove(&pos);
            }
        }
    }

    /// Return the node responsible for `key`.
    ///
    /// Finds the first node whose ring position is ≥ the key's position
    /// (next clockwise), wrapping around to the node with the smallest position
    /// when the key sits beyond the last node. Returns `None` if the ring is
    /// empty.
    pub fn get_node(&self, key: &str) -> Option<&str> {
        if self.ring.is_empty() {
            return None;
        }
        let pos = key_position(key);
        // First node with position >= pos, or wrap around to the minimum.
        self.ring
            .range(pos..)
            .next()
            .or_else(|| self.ring.iter().next())
            .map(|(_, node)| node.as_str())
    }

    /// Return the distinct physical nodes currently in the ring, in
    /// unspecified order.
    pub fn nodes(&self) -> Vec<&str> {
        let mut seen: Vec<&str> = Vec::new();
        for node in self.ring.values() {
            if !seen.contains(&node.as_str()) {
                seen.push(node.as_str());
            }
        }
        seen
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::{HashMap, HashSet};

    fn make_ring(nodes: &[&str]) -> ConsistentHashRing {
        let mut ring = ConsistentHashRing::new(150);
        for &n in nodes {
            ring.add_node(n);
        }
        ring
    }

    // ------------------------------------------------------------------
    // Basic structural tests
    // ------------------------------------------------------------------

    #[test]
    fn empty_ring_returns_none() {
        let ring = ConsistentHashRing::new(10);
        assert_eq!(ring.get_node("anything"), None);
        assert!(ring.nodes().is_empty());
    }

    #[test]
    fn single_node_gets_all_keys() {
        let mut ring = ConsistentHashRing::new(50);
        ring.add_node("only-node");

        for key in &["a", "b", "hello", "world", "12345", ""] {
            assert_eq!(ring.get_node(key), Some("only-node"));
        }
    }

    #[test]
    fn nodes_returns_all_added_nodes() {
        let ring = make_ring(&["alpha", "beta", "gamma"]);
        let mut got: Vec<&str> = ring.nodes();
        got.sort_unstable();
        assert_eq!(got, vec!["alpha", "beta", "gamma"]);
    }

    #[test]
    fn nodes_deduplicates() {
        let mut ring = ConsistentHashRing::new(10);
        ring.add_node("x");
        ring.add_node("x"); // duplicate add
        assert_eq!(ring.nodes().len(), 1);
    }

    // ------------------------------------------------------------------
    // Add / remove nodes
    // ------------------------------------------------------------------

    #[test]
    fn add_node_increases_virtual_count() {
        let vn = 20;
        let mut ring = ConsistentHashRing::new(vn);
        assert_eq!(ring.ring.len(), 0);

        ring.add_node("a");
        assert_eq!(ring.ring.len(), vn);

        ring.add_node("b");
        assert_eq!(ring.ring.len(), vn * 2);
    }

    #[test]
    fn remove_node_empties_ring_when_sole_node() {
        let mut ring = ConsistentHashRing::new(30);
        ring.add_node("solo");
        ring.remove_node("solo");

        assert!(ring.ring.is_empty());
        assert_eq!(ring.get_node("key"), None);
        assert!(ring.nodes().is_empty());
    }

    #[test]
    fn remove_nonexistent_node_is_a_noop() {
        let mut ring = make_ring(&["a", "b"]);
        let before = ring.ring.len();
        ring.remove_node("ghost");
        assert_eq!(ring.ring.len(), before);
    }

    #[test]
    fn remove_node_leaves_others_intact() {
        let mut ring = make_ring(&["a", "b", "c"]);
        ring.remove_node("b");

        let mut got: Vec<&str> = ring.nodes();
        got.sort_unstable();
        assert_eq!(got, vec!["a", "c"]);

        // All keys route to either a or c — never b.
        for key in &["k1", "k2", "k3", "k4", "k5"] {
            let node = ring.get_node(key).unwrap();
            assert!(node == "a" || node == "c", "unexpected node: {node}");
        }
    }

    // ------------------------------------------------------------------
    // Consistency: routing is stable when unrelated nodes are touched
    // ------------------------------------------------------------------

    #[test]
    fn routing_stable_after_unrelated_add() {
        let keys: Vec<String> = (0..200).map(|i| format!("key-{i}")).collect();

        let mut ring = make_ring(&["node1", "node2", "node3"]);

        // Record initial assignments.
        let initial: HashMap<&str, &str> = keys
            .iter()
            .map(|k| (k.as_str(), ring.get_node(k).unwrap()))
            .collect();

        // Add an unrelated node.
        ring.add_node("node4");

        // Keys that still hash to the same three nodes must route identically.
        let mut changed = 0usize;
        for key in &keys {
            let new_node = ring.get_node(key).unwrap();
            if new_node != *initial.get(key.as_str()).unwrap() {
                changed += 1;
            }
        }
        // Expect roughly 1/4 of keys to migrate to node4; allow up to 40%.
        assert!(
            changed <= keys.len() * 40 / 100,
            "too many keys remapped: {changed}/{} after adding one of four nodes",
            keys.len()
        );
    }

    #[test]
    fn routing_stable_after_node_removal() {
        let keys: Vec<String> = (0..500).map(|i| format!("item-{i}")).collect();

        let mut ring = make_ring(&["n1", "n2", "n3", "n4"]);

        // Record which keys are on n1, n2, n4 (the nodes we'll keep).
        let before: HashMap<String, String> = keys
            .iter()
            .filter_map(|k| {
                let node = ring.get_node(k).unwrap();
                if node != "n3" {
                    Some((k.clone(), node.to_owned()))
                } else {
                    None
                }
            })
            .collect();

        ring.remove_node("n3");

        // Keys that were NOT on n3 should still route to the same node.
        let mut moved = 0usize;
        for (key, old_node) in &before {
            let new_node = ring.get_node(key).unwrap();
            if new_node != old_node.as_str() {
                moved += 1;
            }
        }
        // Allow at most 5% churn on the "untouched" keys.
        let threshold = before.len() * 5 / 100;
        assert!(
            moved <= threshold,
            "too many untouched keys remapped: {moved}/{} (threshold {threshold})",
            before.len()
        );
    }

    // ------------------------------------------------------------------
    // Key distribution
    // ------------------------------------------------------------------

    #[test]
    fn keys_distributed_across_all_nodes() {
        let ring = make_ring(&["s1", "s2", "s3", "s4", "s5"]);
        let mut counts: HashMap<&str, usize> = HashMap::new();

        for i in 0..1000 {
            let key = format!("key-{i}");
            let node = ring.get_node(&key).unwrap();
            *counts.entry(node).or_insert(0) += 1;
        }

        // Every node should receive at least 5% of 1000 keys.
        for (node, count) in &counts {
            assert!(
                *count >= 50,
                "node {node} received only {count} keys — distribution too uneven"
            );
        }
        // All five nodes should appear.
        assert_eq!(counts.len(), 5);
    }

    #[test]
    fn deterministic_hash_same_result_twice() {
        let ring = make_ring(&["alpha", "beta", "gamma"]);

        let result1: Vec<Option<&str>> = (0..50).map(|i| ring.get_node(&format!("k{i}"))).collect();
        let result2: Vec<Option<&str>> = (0..50).map(|i| ring.get_node(&format!("k{i}"))).collect();

        assert_eq!(result1, result2);
    }

    // ------------------------------------------------------------------
    // Edge cases
    // ------------------------------------------------------------------

    #[test]
    fn two_nodes_both_receive_keys() {
        let ring = make_ring(&["east", "west"]);
        let mut seen: HashSet<&str> = HashSet::new();

        for i in 0..100 {
            let key = format!("{i}");
            seen.insert(ring.get_node(&key).unwrap());
        }
        assert_eq!(seen.len(), 2, "both nodes should receive at least one key");
    }

    #[test]
    fn zero_virtual_nodes_ring_is_always_empty() {
        let mut ring = ConsistentHashRing::new(0);
        ring.add_node("a");
        // No virtual positions inserted.
        assert_eq!(ring.get_node("key"), None);
    }
}
