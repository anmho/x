//! A simplified in-memory Log-Structured Merge (LSM) tree.
//!
//! An LSM tree is a data structure with performance characteristics that make it
//! attractive for providing indexed access to files with high insert volume.
//! Writes go to a mutable in-memory buffer (the *memtable*); once the memtable
//! reaches a configured capacity it is flushed to an immutable sorted run on
//! "disk" (here represented as a `Vec` of sorted key-value pairs called an
//! SSTable). Periodic compaction merges all SSTables into a single run,
//! discarding overwritten values and tombstoned (deleted) keys.

use std::collections::BTreeMap;

/// Sentinel value stored in place of a real value to mark a key as deleted.
const TOMBSTONE: &str = "\x00__TOMBSTONE__\x00";

/// A single SSTable — an immutable, sorted run of key-value (or key-tombstone)
/// entries produced by flushing the memtable.
///
/// Entries are sorted by key in ascending order.
pub type SSTable = Vec<(String, String)>;

/// A simplified in-memory LSM tree.
///
/// # Overview
///
/// Writes are accumulated in a mutable `BTreeMap` called the *memtable*.
/// When the memtable reaches `memtable_capacity` entries it is flushed to a
/// new [`SSTable`] and appended to `sstables`.  Older SSTables are at lower
/// indices; the most recently flushed SSTable is at the highest index.
///
/// Reads check the memtable first, then walk `sstables` from newest to oldest,
/// returning the first match (or `None` when the latest record is a tombstone).
///
/// [`compact`](LsmTree::compact) merges all SSTables into one, resolving
/// conflicts and dropping tombstoned keys.
pub struct LsmTree {
    /// Mutable write buffer.  Entries here shadow any older SSTable entries.
    memtable: BTreeMap<String, String>,

    /// An optional immutable snapshot of the memtable that is in the process
    /// of being written out.  In this simplified implementation it is used as
    /// an intermediate holding area during [`flush`](LsmTree::flush).
    immutable_memtable: Option<BTreeMap<String, String>>,

    /// Sorted runs produced by previous flushes.  Index 0 is the oldest run;
    /// the last element is the most recent.
    sstables: Vec<SSTable>,

    /// Maximum number of entries the memtable may hold before an automatic
    /// flush is triggered.
    memtable_capacity: usize,
}

impl LsmTree {
    /// Create a new, empty `LsmTree`.
    ///
    /// # Arguments
    ///
    /// * `memtable_capacity` — number of entries the memtable may hold before
    ///   an automatic flush to an SSTable is triggered.  Must be at least 1.
    ///
    /// # Panics
    ///
    /// Panics if `memtable_capacity` is 0.
    pub fn new(memtable_capacity: usize) -> Self {
        assert!(memtable_capacity > 0, "memtable_capacity must be at least 1");
        LsmTree {
            memtable: BTreeMap::new(),
            immutable_memtable: None,
            sstables: Vec::new(),
            memtable_capacity,
        }
    }

    /// Insert or update a key-value pair.
    ///
    /// The entry is written to the mutable memtable.  If the memtable has
    /// reached `memtable_capacity` *after* the write, it is automatically
    /// flushed to a new SSTable via [`flush`](LsmTree::flush).
    pub fn put(&mut self, key: String, value: String) {
        self.memtable.insert(key, value);
        if self.memtable.len() >= self.memtable_capacity {
            self.flush();
        }
    }

    /// Delete a key by writing a *tombstone* marker.
    ///
    /// The tombstone is written to the memtable and will shadow any older
    /// value for the same key in the SSTables.  The key will be fully removed
    /// after the next [`compact`](LsmTree::compact).
    pub fn delete(&mut self, key: &str) {
        self.memtable.insert(key.to_string(), TOMBSTONE.to_string());
        if self.memtable.len() >= self.memtable_capacity {
            self.flush();
        }
    }

    /// Look up a key, returning `Some(&str)` with the associated value, or
    /// `None` if the key does not exist or has been deleted.
    ///
    /// The search order is:
    /// 1. Mutable memtable (most recent writes).
    /// 2. Immutable memtable (if a flush is in progress).
    /// 3. SSTables from newest to oldest.
    ///
    /// The first match wins.  If the matched record is a tombstone the method
    /// returns `None`, indicating the key has been deleted.
    pub fn get(&self, key: &str) -> Option<&str> {
        // 1. Check mutable memtable.
        if let Some(value) = self.memtable.get(key) {
            return if value == TOMBSTONE { None } else { Some(value.as_str()) };
        }

        // 2. Check immutable memtable (snapshot being flushed).
        if let Some(imm) = &self.immutable_memtable {
            if let Some(value) = imm.get(key) {
                return if value == TOMBSTONE { None } else { Some(value.as_str()) };
            }
        }

        // 3. Walk SSTables newest-first.
        for sstable in self.sstables.iter().rev() {
            if let Ok(idx) = sstable.binary_search_by_key(&key, |(k, _)| k.as_str()) {
                let value = &sstable[idx].1;
                return if value == TOMBSTONE { None } else { Some(value.as_str()) };
            }
        }

        None
    }

    /// Flush the current mutable memtable to a new SSTable.
    ///
    /// The memtable's entries are sorted by key (they are already ordered in
    /// a `BTreeMap`) and appended as a new level to `sstables`.  The memtable
    /// is then cleared.  If the memtable is empty this is a no-op.
    pub fn flush(&mut self) {
        if self.memtable.is_empty() {
            return;
        }

        // Move the memtable into the immutable slot to simulate the two-phase
        // flush found in production LSM implementations.
        let snapshot = std::mem::replace(&mut self.memtable, BTreeMap::new());
        self.immutable_memtable = Some(snapshot);

        // "Write" the immutable memtable out to an SSTable.
        if let Some(imm) = self.immutable_memtable.take() {
            let sstable: SSTable = imm.into_iter().collect();
            self.sstables.push(sstable);
        }
    }

    /// Merge all SSTables into a single SSTable, discarding tombstoned keys
    /// and retaining only the most-recent value for each key.
    ///
    /// After compaction `sstables` contains at most one entry.  The mutable
    /// memtable is left untouched — it will shadow the compacted SSTable as
    /// usual.
    pub fn compact(&mut self) {
        if self.sstables.len() <= 1 {
            return;
        }

        // Walk all SSTables newest-first so that the first time we encounter a
        // key it is the most recent version.
        let mut seen: BTreeMap<String, String> = BTreeMap::new();

        for sstable in self.sstables.iter().rev() {
            for (key, value) in sstable {
                // Only record the first (newest) occurrence of each key.
                seen.entry(key.clone()).or_insert_with(|| value.clone());
            }
        }

        // Build one compacted SSTable, dropping tombstones.
        let compacted: SSTable = seen
            .into_iter()
            .filter(|(_, v)| v != TOMBSTONE)
            .collect();

        self.sstables = if compacted.is_empty() {
            Vec::new()
        } else {
            vec![compacted]
        };
    }

    /// Return the number of SSTables currently held by this tree.
    ///
    /// Useful for asserting flush and compaction behaviour in tests.
    pub fn sstable_count(&self) -> usize {
        self.sstables.len()
    }

    /// Return the number of entries currently in the mutable memtable.
    pub fn memtable_len(&self) -> usize {
        self.memtable.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // -----------------------------------------------------------------------
    // Basic put / get
    // -----------------------------------------------------------------------

    #[test]
    fn test_put_and_get_single_entry() {
        let mut tree = LsmTree::new(10);
        tree.put("alpha".to_string(), "1".to_string());
        assert_eq!(tree.get("alpha"), Some("1"));
    }

    #[test]
    fn test_put_overwrites_existing_key() {
        let mut tree = LsmTree::new(10);
        tree.put("k".to_string(), "old".to_string());
        tree.put("k".to_string(), "new".to_string());
        assert_eq!(tree.get("k"), Some("new"));
    }

    #[test]
    fn test_get_missing_key_returns_none() {
        let tree = LsmTree::new(10);
        assert_eq!(tree.get("missing"), None);
    }

    // -----------------------------------------------------------------------
    // Delete / tombstone
    // -----------------------------------------------------------------------

    #[test]
    fn test_delete_makes_get_return_none() {
        let mut tree = LsmTree::new(10);
        tree.put("x".to_string(), "hello".to_string());
        tree.delete("x");
        assert_eq!(tree.get("x"), None);
    }

    #[test]
    fn test_delete_non_existent_key_returns_none() {
        let mut tree = LsmTree::new(10);
        tree.delete("ghost");
        assert_eq!(tree.get("ghost"), None);
    }

    #[test]
    fn test_put_after_delete_restores_value() {
        let mut tree = LsmTree::new(10);
        tree.put("y".to_string(), "value".to_string());
        tree.delete("y");
        tree.put("y".to_string(), "restored".to_string());
        assert_eq!(tree.get("y"), Some("restored"));
    }

    // -----------------------------------------------------------------------
    // Flush
    // -----------------------------------------------------------------------

    #[test]
    fn test_manual_flush_moves_entries_to_sstable() {
        let mut tree = LsmTree::new(100);
        tree.put("a".to_string(), "1".to_string());
        tree.put("b".to_string(), "2".to_string());
        assert_eq!(tree.sstable_count(), 0);

        tree.flush();

        assert_eq!(tree.sstable_count(), 1);
        assert_eq!(tree.memtable_len(), 0);
        // Entries are still readable after flush.
        assert_eq!(tree.get("a"), Some("1"));
        assert_eq!(tree.get("b"), Some("2"));
    }

    #[test]
    fn test_auto_flush_triggers_at_capacity() {
        // capacity = 3, so after the 3rd put the memtable should be flushed.
        let mut tree = LsmTree::new(3);
        tree.put("p".to_string(), "1".to_string());
        tree.put("q".to_string(), "2".to_string());
        assert_eq!(tree.sstable_count(), 0, "should not flush before capacity");

        tree.put("r".to_string(), "3".to_string()); // triggers flush
        assert_eq!(tree.sstable_count(), 1, "flush should have fired");
        assert_eq!(tree.memtable_len(), 0);
    }

    #[test]
    fn test_flush_on_empty_memtable_is_noop() {
        let mut tree = LsmTree::new(10);
        tree.flush();
        assert_eq!(tree.sstable_count(), 0);
    }

    #[test]
    fn test_multiple_flushes_produce_multiple_sstables() {
        let mut tree = LsmTree::new(100);
        tree.put("a".to_string(), "1".to_string());
        tree.flush();
        tree.put("b".to_string(), "2".to_string());
        tree.flush();
        assert_eq!(tree.sstable_count(), 2);
    }

    // -----------------------------------------------------------------------
    // Compact
    // -----------------------------------------------------------------------

    #[test]
    fn test_compact_merges_sstables_into_one() {
        let mut tree = LsmTree::new(100);
        tree.put("a".to_string(), "1".to_string());
        tree.flush();
        tree.put("b".to_string(), "2".to_string());
        tree.flush();
        tree.put("c".to_string(), "3".to_string());
        tree.flush();

        assert_eq!(tree.sstable_count(), 3);
        tree.compact();
        assert_eq!(tree.sstable_count(), 1);

        // All values still accessible.
        assert_eq!(tree.get("a"), Some("1"));
        assert_eq!(tree.get("b"), Some("2"));
        assert_eq!(tree.get("c"), Some("3"));
    }

    #[test]
    fn test_compact_removes_tombstoned_keys() {
        let mut tree = LsmTree::new(100);
        tree.put("gone".to_string(), "bye".to_string());
        tree.flush();
        tree.delete("gone");
        tree.flush();

        tree.compact();

        // After compaction the key should be gone entirely.
        assert_eq!(tree.get("gone"), None);
        // Compacted result should be empty — no sstable with live entries.
        assert_eq!(tree.sstable_count(), 0);
    }

    #[test]
    fn test_compact_keeps_newest_value_for_key() {
        let mut tree = LsmTree::new(100);
        tree.put("k".to_string(), "old".to_string());
        tree.flush();
        tree.put("k".to_string(), "new".to_string());
        tree.flush();

        tree.compact();

        assert_eq!(tree.sstable_count(), 1);
        assert_eq!(tree.get("k"), Some("new"));
    }

    #[test]
    fn test_compact_single_sstable_is_noop() {
        let mut tree = LsmTree::new(100);
        tree.put("m".to_string(), "v".to_string());
        tree.flush();

        tree.compact();
        assert_eq!(tree.sstable_count(), 1);
        assert_eq!(tree.get("m"), Some("v"));
    }

    // -----------------------------------------------------------------------
    // Reads across memtable + multiple SSTable levels
    // -----------------------------------------------------------------------

    #[test]
    fn test_memtable_shadows_sstable() {
        let mut tree = LsmTree::new(100);
        tree.put("key".to_string(), "sstable_val".to_string());
        tree.flush();
        // Overwrite in memtable — should shadow the SSTable entry.
        tree.put("key".to_string(), "memtable_val".to_string());

        assert_eq!(tree.get("key"), Some("memtable_val"));
    }

    #[test]
    fn test_newer_sstable_shadows_older_sstable() {
        let mut tree = LsmTree::new(100);
        tree.put("key".to_string(), "v1".to_string());
        tree.flush(); // SSTable 0: key -> v1
        tree.put("key".to_string(), "v2".to_string());
        tree.flush(); // SSTable 1: key -> v2

        // Newest SSTable should win.
        assert_eq!(tree.get("key"), Some("v2"));
    }

    #[test]
    fn test_delete_in_memtable_hides_sstable_value() {
        let mut tree = LsmTree::new(100);
        tree.put("hidden".to_string(), "there".to_string());
        tree.flush();
        // Tombstone lives in the memtable; value is in the SSTable.
        tree.delete("hidden");

        assert_eq!(tree.get("hidden"), None);
    }

    #[test]
    fn test_reads_across_multiple_levels() {
        let mut tree = LsmTree::new(100);

        // Level 0 (oldest SSTable).
        tree.put("a".to_string(), "a0".to_string());
        tree.put("b".to_string(), "b0".to_string());
        tree.flush();

        // Level 1: overwrite "a", add "c".
        tree.put("a".to_string(), "a1".to_string());
        tree.put("c".to_string(), "c1".to_string());
        tree.flush();

        // Memtable: overwrite "b", delete "c".
        tree.put("b".to_string(), "b_mem".to_string());
        tree.delete("c");

        assert_eq!(tree.get("a"), Some("a1"),  "newest SSTable wins for 'a'");
        assert_eq!(tree.get("b"), Some("b_mem"), "memtable wins for 'b'");
        assert_eq!(tree.get("c"), None,          "tombstone in memtable hides SSTable 'c'");
    }
}
