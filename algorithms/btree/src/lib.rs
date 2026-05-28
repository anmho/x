//! # B+ Tree
//!
//! A generic, order-configurable B+ tree that supports:
//! - `insert` — key/value insertion with leaf-level splits propagating to root
//! - `search` — exact O(log n) key lookup
//! - `range`  — inclusive range scan in O(log n + k) using the leaf linked list
//!
//! ## Design
//!
//! Nodes are stored in a `Vec`-backed arena and referenced by `usize` indices.
//! This avoids `Rc<RefCell<>>` indirection while still allowing the leaf linked
//! list to be expressed safely.  Index `0` is permanently reserved as the "null"
//! sentinel so that `Option<usize>` can be avoided on the hot path.

/// Arena index used as a null / "no node" sentinel.
const NULL: usize = 0;

// ── Node representation ───────────────────────────────────────────────────────

/// A node stored inside the arena.
#[derive(Debug)]
enum Node<K, V> {
    /// Placeholder for index 0 (the null sentinel).
    Null,
    /// An internal node holds separator keys and child pointers.
    Internal(Internal<K>),
    /// A leaf node holds key-value pairs and links to adjacent leaves.
    Leaf(Leaf<K, V>),
}

#[derive(Debug)]
struct Internal<K> {
    /// Separator keys: `keys[i]` is the smallest key reachable via `children[i+1]`.
    keys: Vec<K>,
    /// Child arena indices: `children.len() == keys.len() + 1`.
    children: Vec<usize>,
}

#[derive(Debug)]
struct Leaf<K, V> {
    keys: Vec<K>,
    values: Vec<V>,
    /// Arena index of the next leaf (NULL if this is the last leaf).
    next: usize,
}

// ── BPlusTree public API ──────────────────────────────────────────────────────

/// A B+ tree generic over key type `K` and value type `V`.
///
/// Internal nodes branch on at most `order` children; leaf nodes hold at most
/// `order - 1` key-value pairs.  The minimum supported `order` is 3.
///
/// All leaves are chained in key order via a singly-linked list, enabling
/// efficient range scans without touching internal nodes after the initial
/// descent.
pub struct BPlusTree<K, V> {
    /// Node arena — index 0 is always `Node::Null`.
    arena: Vec<Node<K, V>>,
    /// Arena index of the root node.
    root: usize,
    /// Maximum number of children an internal node may have (= branching factor).
    order: usize,
}

impl<K: Ord + Clone, V: Clone> BPlusTree<K, V> {
    /// Create a new, empty B+ tree with the given `order` (branching factor).
    ///
    /// `order` is clamped to a minimum of 3.
    ///
    /// # Examples
    ///
    /// ```
    /// use bplus_tree::BPlusTree;
    /// let mut tree: BPlusTree<i32, &str> = BPlusTree::new(4);
    /// ```
    pub fn new(order: usize) -> Self {
        let order = order.max(3);
        // Index 0 — null sentinel.
        // Index 1 — initial empty root leaf.
        let arena = vec![
            Node::Null,
            Node::Leaf(Leaf {
                keys: Vec::new(),
                values: Vec::new(),
                next: NULL,
            }),
        ];
        BPlusTree {
            arena,
            root: 1,
            order,
        }
    }

    // ── Helpers ───────────────────────────────────────────────────────────────

    /// Maximum keys per leaf node = order - 1.
    fn max_leaf_keys(&self) -> usize {
        self.order - 1
    }

    /// Maximum keys per internal node = order - 1  (=> max children = order).
    fn max_internal_keys(&self) -> usize {
        self.order - 1
    }

    /// Allocate a new node in the arena and return its index.
    fn alloc(&mut self, node: Node<K, V>) -> usize {
        let idx = self.arena.len();
        self.arena.push(node);
        idx
    }

    // ── Search ────────────────────────────────────────────────────────────────

    /// Search for `key` and return a reference to its associated value, or
    /// `None` if the key is not present.
    ///
    /// # Examples
    ///
    /// ```
    /// use bplus_tree::BPlusTree;
    /// let mut tree = BPlusTree::new(4);
    /// tree.insert(1u32, "one");
    /// assert_eq!(tree.search(&1), Some(&"one"));
    /// assert_eq!(tree.search(&99), None);
    /// ```
    pub fn search(&self, key: &K) -> Option<&V> {
        let leaf_idx = self.find_leaf(self.root, key);
        match &self.arena[leaf_idx] {
            Node::Leaf(leaf) => {
                match leaf.keys.binary_search(key) {
                    Ok(pos) => Some(&leaf.values[pos]),
                    Err(_) => None,
                }
            }
            _ => None,
        }
    }

    /// Descend from `node` to the leaf that should contain `key`.
    fn find_leaf(&self, mut node: usize, key: &K) -> usize {
        loop {
            match &self.arena[node] {
                Node::Leaf(_) => return node,
                Node::Internal(internal) => {
                    // Find child index: the first separator key > key.
                    let child_pos = internal.keys.partition_point(|k| k <= key);
                    node = internal.children[child_pos];
                }
                Node::Null => unreachable!("traversed into null node"),
            }
        }
    }

    // ── Range scan ────────────────────────────────────────────────────────────

    /// Return all `(key, value)` pairs where `from <= key <= to`, in ascending
    /// key order.
    ///
    /// Uses the leaf linked list for O(log n + k) performance where k is the
    /// number of results.
    ///
    /// # Examples
    ///
    /// ```
    /// use bplus_tree::BPlusTree;
    /// let mut tree = BPlusTree::new(4);
    /// for i in 0u32..10 { tree.insert(i, i * 2); }
    /// let result = tree.range(&3, &6);
    /// assert_eq!(result, vec![(&3, &6), (&4, &8), (&5, &10), (&6, &12)]);
    /// ```
    pub fn range(&self, from: &K, to: &K) -> Vec<(&K, &V)> {
        if from > to {
            return Vec::new();
        }
        let mut results = Vec::new();
        let mut node_idx = self.find_leaf(self.root, from);
        loop {
            match &self.arena[node_idx] {
                Node::Leaf(leaf) => {
                    for (k, v) in leaf.keys.iter().zip(leaf.values.iter()) {
                        if k > to {
                            return results;
                        }
                        if k >= from {
                            results.push((k, v));
                        }
                    }
                    if leaf.next == NULL {
                        break;
                    }
                    node_idx = leaf.next;
                }
                _ => break,
            }
        }
        results
    }

    // ── Insert ────────────────────────────────────────────────────────────────

    /// Insert `key` with `value`.  If `key` already exists its value is
    /// overwritten.
    ///
    /// # Examples
    ///
    /// ```
    /// use bplus_tree::BPlusTree;
    /// let mut tree = BPlusTree::new(4);
    /// tree.insert(42u32, "hello");
    /// tree.insert(42u32, "world"); // overwrites
    /// assert_eq!(tree.search(&42), Some(&"world"));
    /// ```
    pub fn insert(&mut self, key: K, value: V) {
        if let Some(split) = self.insert_recursive(self.root, key, value) {
            // The root was split — create a new root.
            let new_root = self.alloc(Node::Internal(Internal {
                keys: vec![split.separator],
                children: vec![self.root, split.new_node],
            }));
            self.root = new_root;
        }
    }

    /// Recursively insert into the subtree rooted at `node`.
    ///
    /// Returns `Some(Split)` if the node overflowed and was split, otherwise
    /// `None`.
    fn insert_recursive(&mut self, node: usize, key: K, value: V) -> Option<Split<K>> {
        // Determine node kind without holding a reference into arena.
        let is_leaf = matches!(&self.arena[node], Node::Leaf(_));

        if is_leaf {
            self.insert_into_leaf(node, key, value)
        } else {
            // Find child to recurse into.
            let child_pos = match &self.arena[node] {
                Node::Internal(internal) => internal.keys.partition_point(|k| k <= &key),
                _ => unreachable!(),
            };
            let child_idx = match &self.arena[node] {
                Node::Internal(internal) => internal.children[child_pos],
                _ => unreachable!(),
            };

            let maybe_split = self.insert_recursive(child_idx, key, value);

            if let Some(split) = maybe_split {
                self.insert_into_internal(node, child_pos, split)
            } else {
                None
            }
        }
    }

    /// Insert a key/value pair into a leaf node.
    ///
    /// Returns a `Split` if the leaf overflowed.
    fn insert_into_leaf(&mut self, node: usize, key: K, value: V) -> Option<Split<K>> {
        let leaf = match &mut self.arena[node] {
            Node::Leaf(l) => l,
            _ => unreachable!(),
        };

        match leaf.keys.binary_search(&key) {
            Ok(pos) => {
                // Overwrite existing value.
                leaf.values[pos] = value;
                return None;
            }
            Err(pos) => {
                leaf.keys.insert(pos, key);
                leaf.values.insert(pos, value);
            }
        }

        // Check overflow.
        if leaf.keys.len() <= self.max_leaf_keys() {
            return None;
        }

        // Split: right half goes to a new sibling.
        let mid = leaf.keys.len() / 2;
        let right_keys = leaf.keys.split_off(mid);
        let right_values = leaf.values.split_off(mid);
        let old_next = leaf.next;

        let separator = right_keys[0].clone();

        let new_leaf_idx = self.alloc(Node::Leaf(Leaf {
            keys: right_keys,
            values: right_values,
            next: old_next,
        }));

        // Update the original leaf's `next` pointer.
        match &mut self.arena[node] {
            Node::Leaf(l) => l.next = new_leaf_idx,
            _ => unreachable!(),
        }

        Some(Split {
            separator,
            new_node: new_leaf_idx,
        })
    }

    /// Push a split result up into an internal node.
    ///
    /// `child_pos` is the position of the child that just split.
    /// Returns a `Split` if the internal node itself overflowed.
    fn insert_into_internal(
        &mut self,
        node: usize,
        child_pos: usize,
        split: Split<K>,
    ) -> Option<Split<K>> {
        let internal = match &mut self.arena[node] {
            Node::Internal(i) => i,
            _ => unreachable!(),
        };

        // Insert separator key and new right child pointer.
        internal.keys.insert(child_pos, split.separator);
        internal.children.insert(child_pos + 1, split.new_node);

        if internal.keys.len() <= self.max_internal_keys() {
            return None;
        }

        // Split the internal node at the median key.
        let mid = internal.keys.len() / 2;
        let separator = internal.keys[mid].clone();

        // Right half of keys (excluding the median, which is pushed up).
        let right_keys = internal.keys.split_off(mid + 1);
        internal.keys.pop(); // remove the promoted separator from left node

        // Right half of children.
        let right_children = internal.children.split_off(mid + 1);

        let new_internal_idx = self.alloc(Node::Internal(Internal {
            keys: right_keys,
            children: right_children,
        }));

        Some(Split {
            separator,
            new_node: new_internal_idx,
        })
    }
}

/// Represents the outcome of splitting a node during insertion.
struct Split<K> {
    /// The smallest key in the new right node (promoted to the parent).
    separator: K,
    /// Arena index of the newly-created right node.
    new_node: usize,
}

// ── Tests ─────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    // ── Basic correctness ─────────────────────────────────────────────────────

    #[test]
    fn empty_tree_search_returns_none() {
        let tree: BPlusTree<i32, i32> = BPlusTree::new(4);
        assert_eq!(tree.search(&0), None);
        assert_eq!(tree.search(&-1), None);
        assert_eq!(tree.search(&100), None);
    }

    #[test]
    fn single_element_insert_and_search() {
        let mut tree = BPlusTree::new(4);
        tree.insert(42u32, "answer");
        assert_eq!(tree.search(&42), Some(&"answer"));
        assert_eq!(tree.search(&41), None);
        assert_eq!(tree.search(&43), None);
    }

    #[test]
    fn duplicate_key_overwrites_value() {
        let mut tree = BPlusTree::new(4);
        tree.insert(1u32, "first");
        tree.insert(1u32, "second");
        assert_eq!(tree.search(&1), Some(&"second"));
    }

    // ── Many inserts triggering multiple splits ────────────────────────────────

    #[test]
    fn insert_and_search_100_keys_ascending() {
        let mut tree = BPlusTree::new(4);
        for i in 0u32..100 {
            tree.insert(i, i * 10);
        }
        for i in 0u32..100 {
            assert_eq!(tree.search(&i), Some(&(i * 10)), "missing key {i}");
        }
        assert_eq!(tree.search(&100), None);
    }

    #[test]
    fn insert_and_search_100_keys_descending() {
        let mut tree = BPlusTree::new(5);
        for i in (0u32..100).rev() {
            tree.insert(i, i * 3);
        }
        for i in 0u32..100 {
            assert_eq!(tree.search(&i), Some(&(i * 3)), "missing key {i}");
        }
    }

    #[test]
    fn insert_and_search_200_keys_interleaved() {
        let mut tree = BPlusTree::new(6);
        // Insert even keys then odd keys to exercise varied split patterns.
        for i in (0u32..200).step_by(2) {
            tree.insert(i, i);
        }
        for i in (1u32..200).step_by(2) {
            tree.insert(i, i);
        }
        for i in 0u32..200 {
            assert_eq!(tree.search(&i), Some(&i), "missing key {i}");
        }
    }

    #[test]
    fn insert_1000_keys_order_3() {
        let mut tree = BPlusTree::new(3);
        for i in 0u32..1000 {
            tree.insert(i, i);
        }
        for i in 0u32..1000 {
            assert_eq!(tree.search(&i), Some(&i), "missing key {i}");
        }
    }

    #[test]
    fn insert_1000_keys_large_order() {
        let mut tree = BPlusTree::new(32);
        for i in 0u32..1000 {
            tree.insert(i, i * 2);
        }
        for i in 0u32..1000 {
            assert_eq!(tree.search(&i), Some(&(i * 2)), "missing key {i}");
        }
    }

    // ── Range scan ────────────────────────────────────────────────────────────

    #[test]
    fn range_scan_basic() {
        let mut tree = BPlusTree::new(4);
        for i in 0u32..10 {
            tree.insert(i, i * 2);
        }
        let result = tree.range(&3, &6);
        let expected: Vec<(u32, u32)> = (3u32..=6).map(|i| (i, i * 2)).collect();
        let got: Vec<(u32, u32)> = result.iter().map(|(&k, &v)| (k, v)).collect();
        assert_eq!(got, expected);
    }

    #[test]
    fn range_scan_full_tree() {
        let mut tree = BPlusTree::new(4);
        for i in 0u32..50 {
            tree.insert(i, i);
        }
        let result = tree.range(&0, &49);
        assert_eq!(result.len(), 50);
        for (idx, (&k, &v)) in result.iter().enumerate() {
            assert_eq!(k as usize, idx);
            assert_eq!(v, k);
        }
    }

    #[test]
    fn range_scan_beyond_bounds() {
        let mut tree = BPlusTree::new(4);
        for i in 5u32..15 {
            tree.insert(i, i);
        }
        // from < min key in tree
        let result = tree.range(&0, &7);
        let got: Vec<u32> = result.iter().map(|(&k, _)| k).collect();
        assert_eq!(got, vec![5, 6, 7]);

        // to > max key in tree
        let result = tree.range(&12, &100);
        let got: Vec<u32> = result.iter().map(|(&k, _)| k).collect();
        assert_eq!(got, vec![12, 13, 14]);
    }

    #[test]
    fn range_scan_single_element_result() {
        let mut tree = BPlusTree::new(4);
        for i in 0u32..20 {
            tree.insert(i, i);
        }
        let result = tree.range(&10, &10);
        assert_eq!(result.len(), 1);
        assert_eq!(*result[0].0, 10u32);
    }

    #[test]
    fn range_scan_empty_result_inverted_bounds() {
        let mut tree = BPlusTree::new(4);
        for i in 0u32..10 {
            tree.insert(i, i);
        }
        let result = tree.range(&7, &3);
        assert!(result.is_empty());
    }

    #[test]
    fn range_scan_empty_tree() {
        let tree: BPlusTree<u32, u32> = BPlusTree::new(4);
        assert!(tree.range(&0, &100).is_empty());
    }

    #[test]
    fn range_scan_no_matching_keys() {
        let mut tree = BPlusTree::new(4);
        tree.insert(1u32, 1);
        tree.insert(10u32, 10);
        let result = tree.range(&3, &7);
        assert!(result.is_empty());
    }

    #[test]
    fn range_scan_large_tree() {
        let mut tree = BPlusTree::new(5);
        for i in 0u32..500 {
            tree.insert(i, i);
        }
        let result = tree.range(&100, &199);
        assert_eq!(result.len(), 100);
        for (idx, (&k, &v)) in result.iter().enumerate() {
            assert_eq!(k, 100 + idx as u32);
            assert_eq!(v, k);
        }
    }

    // ── Order boundary (min order = 3) ────────────────────────────────────────

    #[test]
    fn order_below_minimum_is_clamped_to_3() {
        // Passing order=1 or order=2 should be clamped to 3 and still work correctly.
        let mut tree = BPlusTree::new(1);
        for i in 0u32..50 {
            tree.insert(i, i);
        }
        for i in 0u32..50 {
            assert_eq!(tree.search(&i), Some(&i));
        }
    }

    // ── String keys ───────────────────────────────────────────────────────────

    #[test]
    fn string_keys() {
        let mut tree: BPlusTree<String, u32> = BPlusTree::new(4);
        let words = vec!["banana", "apple", "cherry", "date", "elderberry", "fig"];
        for (i, w) in words.iter().enumerate() {
            tree.insert(w.to_string(), i as u32);
        }
        assert_eq!(tree.search(&"apple".to_string()), Some(&1));
        assert_eq!(tree.search(&"cherry".to_string()), Some(&2));
        assert_eq!(tree.search(&"grape".to_string()), None);
    }

    // ── Duplicate overwrites across splits ────────────────────────────────────

    #[test]
    fn overwrite_after_many_inserts() {
        let mut tree = BPlusTree::new(4);
        for i in 0u32..100 {
            tree.insert(i, i);
        }
        // Overwrite a key that may have ended up in a split node.
        tree.insert(50u32, 9999);
        assert_eq!(tree.search(&50), Some(&9999));
        // Make sure neighbors are unaffected.
        assert_eq!(tree.search(&49), Some(&49));
        assert_eq!(tree.search(&51), Some(&51));
    }

    // ── Leaf linked list integrity ────────────────────────────────────────────

    #[test]
    fn range_scan_respects_leaf_order_after_splits() {
        // Insert in reverse order; the linked list must still be sorted.
        let mut tree = BPlusTree::new(4);
        for i in (0u32..100).rev() {
            tree.insert(i, i);
        }
        let result = tree.range(&0, &99);
        assert_eq!(result.len(), 100);
        for (i, (&k, _)) in result.iter().enumerate() {
            assert_eq!(k, i as u32, "leaf list out of order at position {i}");
        }
    }
}
