//! # Bloom Filter
//!
//! A space-efficient probabilistic data structure for membership queries.
//!
//! A Bloom filter can definitively report that an item is *not* in the set,
//! but may return false positives — it may report an item as present when it
//! was never inserted. It never produces false negatives.
//!
//! ## Example
//!
//! ```rust
//! use bloom_filter::BloomFilter;
//!
//! // Build a filter expecting 1 000 items at a 1 % false-positive rate.
//! let mut filter = BloomFilter::with_capacity(1_000, 0.01);
//!
//! filter.insert(b"hello");
//! filter.insert(b"world");
//!
//! assert!(filter.contains(b"hello"));
//! assert!(filter.contains(b"world"));
//! assert!(!filter.contains(b"not inserted")); // very likely false
//! ```

/// A Bloom filter backed by a plain bit-vector.
///
/// Internally the bit-vector is stored as a `Vec<u64>` and hash positions are
/// derived via *double hashing*: two independent hash values `h1` and `h2` are
/// combined as `h1 + i * h2` for `i` in `0..num_hashes`, which avoids the need
/// for external crates while still providing good uniformity.
pub struct BloomFilter {
    /// The underlying bit storage (each `u64` holds 64 bits).
    bits: Vec<u64>,
    /// Total number of addressable bits (`bits.len() * 64`).
    num_bits: usize,
    /// Number of hash functions (positions) used per element.
    num_hashes: usize,
}

impl BloomFilter {
    /// Creates a new, empty `BloomFilter` with an explicit bit-array size and
    /// hash-function count.
    ///
    /// # Arguments
    ///
    /// * `num_bits`   – Total number of bits in the filter. Rounded up to the
    ///                  nearest multiple of 64 for storage efficiency.
    /// * `num_hashes` – Number of independent hash probes per element. Higher
    ///                  values reduce false positives up to a point, then
    ///                  increase them again.
    ///
    /// # Panics
    ///
    /// Panics if `num_bits` is 0 or `num_hashes` is 0.
    pub fn new(num_bits: usize, num_hashes: usize) -> Self {
        assert!(num_bits > 0, "num_bits must be greater than zero");
        assert!(num_hashes > 0, "num_hashes must be greater than zero");

        // Round up to a whole number of u64 words.
        let words = num_bits.div_ceil(64);
        let actual_bits = words * 64;

        BloomFilter {
            bits: vec![0u64; words],
            num_bits: actual_bits,
            num_hashes,
        }
    }

    /// Creates a `BloomFilter` sized for the given expected item count and
    /// target false-positive probability.
    ///
    /// Uses the standard optimal formulas:
    ///
    /// ```text
    /// m = -n * ln(p) / (ln(2))²
    /// k = (m / n) * ln(2)
    /// ```
    ///
    /// where `n` is the expected number of items, `p` is the desired
    /// false-positive rate, `m` is the number of bits, and `k` is the number
    /// of hash functions.
    ///
    /// # Arguments
    ///
    /// * `expected_items`      – Anticipated number of distinct items to insert.
    /// * `false_positive_rate` – Desired false-positive probability in `(0, 1)`.
    ///
    /// # Panics
    ///
    /// Panics if `expected_items` is 0 or `false_positive_rate` is not in the
    /// open interval `(0.0, 1.0)`.
    pub fn with_capacity(expected_items: usize, false_positive_rate: f64) -> Self {
        assert!(expected_items > 0, "expected_items must be greater than zero");
        assert!(
            false_positive_rate > 0.0 && false_positive_rate < 1.0,
            "false_positive_rate must be in (0, 1)"
        );

        let n = expected_items as f64;
        let p = false_positive_rate;

        // Optimal bit count.
        let m = (-n * p.ln() / (2.0_f64.ln().powi(2))).ceil() as usize;
        let m = m.max(1);

        // Optimal hash count.
        let k = ((m as f64 / n) * 2.0_f64.ln()).round() as usize;
        let k = k.max(1);

        Self::new(m, k)
    }

    /// Inserts `item` into the filter.
    ///
    /// After this call, [`contains`](Self::contains) is guaranteed to return
    /// `true` for the same byte slice.
    pub fn insert(&mut self, item: &[u8]) {
        let (h1, h2) = Self::hash_pair(item);
        for i in 0..self.num_hashes {
            let bit = Self::nth_bit(h1, h2, i, self.num_bits);
            self.set_bit(bit);
        }
    }

    /// Returns `true` if `item` *might* have been inserted; returns `false` if
    /// it definitely has not.
    ///
    /// False positives are possible; false negatives are not.
    pub fn contains(&self, item: &[u8]) -> bool {
        let (h1, h2) = Self::hash_pair(item);
        (0..self.num_hashes).all(|i| {
            let bit = Self::nth_bit(h1, h2, i, self.num_bits);
            self.get_bit(bit)
        })
    }

    /// Returns the number of addressable bits in the filter.
    pub fn num_bits(&self) -> usize {
        self.num_bits
    }

    /// Returns the number of hash functions used per element.
    pub fn num_hashes(&self) -> usize {
        self.num_hashes
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    /// Computes two independent 64-bit hash values for `data` using a pair of
    /// simple polynomial (FNV-inspired) hash functions.
    ///
    /// * `h1` uses the FNV-1a basis and prime.
    /// * `h2` uses a different basis so that the two values are independent
    ///   enough for double-hashing purposes.
    fn hash_pair(data: &[u8]) -> (u64, u64) {
        // FNV-1a parameters
        const FNV_BASIS: u64 = 0xcbf2_9ce4_8422_2325;
        const FNV_PRIME: u64 = 0x0000_0100_0000_01b3;

        // Second hash uses a different starting basis.
        const BASIS2: u64 = 0x517c_c1b7_2722_0a95;
        const PRIME2: u64 = 0x0000_0100_0000_0151;

        let mut h1 = FNV_BASIS;
        let mut h2 = BASIS2;

        for &byte in data {
            h1 ^= byte as u64;
            h1 = h1.wrapping_mul(FNV_PRIME);

            h2 ^= byte as u64;
            h2 = h2.wrapping_mul(PRIME2);
        }

        (h1, h2)
    }

    /// Computes the `i`-th bit position via double hashing:
    /// `position = (h1 + i * h2) mod num_bits`.
    #[inline]
    fn nth_bit(h1: u64, h2: u64, i: usize, num_bits: usize) -> usize {
        let pos = h1.wrapping_add((i as u64).wrapping_mul(h2));
        (pos % num_bits as u64) as usize
    }

    /// Sets bit `index` in the bit-vector.
    #[inline]
    fn set_bit(&mut self, index: usize) {
        self.bits[index / 64] |= 1u64 << (index % 64);
    }

    /// Returns `true` if bit `index` is set.
    #[inline]
    fn get_bit(&self, index: usize) -> bool {
        (self.bits[index / 64] >> (index % 64)) & 1 == 1
    }
}

// =============================================================================
// Tests
// =============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    // -------------------------------------------------------------------------
    // Basic functionality
    // -------------------------------------------------------------------------

    #[test]
    fn test_basic_insert_and_contains() {
        let mut bf = BloomFilter::new(1024, 4);
        bf.insert(b"hello");
        bf.insert(b"world");

        assert!(bf.contains(b"hello"), "inserted item must be found");
        assert!(bf.contains(b"world"), "inserted item must be found");
    }

    #[test]
    fn test_item_not_inserted_returns_false_most_of_the_time() {
        // For a very small set and large filter, the probability of a false
        // positive for any single probe is negligible.
        let mut bf = BloomFilter::new(65536, 7);
        bf.insert(b"present");

        // "absent" is overwhelmingly likely to be absent.
        assert!(!bf.contains(b"absent"));
    }

    // -------------------------------------------------------------------------
    // No false negatives guarantee
    // -------------------------------------------------------------------------

    #[test]
    fn test_no_false_negatives() {
        let mut bf = BloomFilter::with_capacity(500, 0.01);

        let items: Vec<Vec<u8>> = (0u32..500).map(|i| i.to_le_bytes().to_vec()).collect();

        for item in &items {
            bf.insert(item);
        }

        for item in &items {
            assert!(
                bf.contains(item),
                "false negative detected for item {:?}",
                item
            );
        }
    }

    // -------------------------------------------------------------------------
    // False positive rate
    // -------------------------------------------------------------------------

    #[test]
    fn test_false_positive_rate_within_2x_expected() {
        let n = 1_000usize;
        let target_fpr = 0.01_f64;
        let mut bf = BloomFilter::with_capacity(n, target_fpr);

        // Insert n distinct items (encoded as big-endian u32).
        for i in 0u32..(n as u32) {
            bf.insert(&i.to_be_bytes());
        }

        // Query a disjoint set of n items.
        let probes = 10_000usize;
        let offset = n as u32;
        let false_positives = (0u32..(probes as u32))
            .filter(|&j| bf.contains(&(offset + j).to_be_bytes()))
            .count();

        let actual_fpr = false_positives as f64 / probes as f64;
        let max_allowed = target_fpr * 2.0;

        assert!(
            actual_fpr <= max_allowed,
            "false positive rate {:.4} exceeds 2x target {:.4}",
            actual_fpr,
            max_allowed
        );
    }

    // -------------------------------------------------------------------------
    // Edge cases
    // -------------------------------------------------------------------------

    #[test]
    fn test_empty_filter_contains_nothing() {
        let bf = BloomFilter::new(256, 3);
        assert!(!bf.contains(b"anything"));
        assert!(!bf.contains(b""));
        assert!(!bf.contains(b"x"));
    }

    #[test]
    fn test_empty_byte_slice_can_be_inserted_and_found() {
        let mut bf = BloomFilter::new(256, 3);
        bf.insert(b"");
        assert!(bf.contains(b""), "empty slice should be findable after insert");
    }

    #[test]
    fn test_single_element() {
        let mut bf = BloomFilter::new(512, 5);
        bf.insert(b"only");

        assert!(bf.contains(b"only"));
        assert!(!bf.contains(b"other"));
    }

    #[test]
    fn test_with_capacity_accessors() {
        let bf = BloomFilter::with_capacity(1_000, 0.01);
        // Optimal m ≈ 9586 bits → rounded up to next multiple of 64.
        // Optimal k ≈ 7.
        assert!(bf.num_bits() >= 9586, "num_bits should be at least the computed m");
        assert_eq!(bf.num_hashes(), 7);
    }

    #[test]
    fn test_new_panics_on_zero_bits() {
        let result = std::panic::catch_unwind(|| BloomFilter::new(0, 3));
        assert!(result.is_err(), "should panic with num_bits = 0");
    }

    #[test]
    fn test_new_panics_on_zero_hashes() {
        let result = std::panic::catch_unwind(|| BloomFilter::new(256, 0));
        assert!(result.is_err(), "should panic with num_hashes = 0");
    }

    #[test]
    fn test_with_capacity_panics_on_zero_items() {
        let result = std::panic::catch_unwind(|| BloomFilter::with_capacity(0, 0.01));
        assert!(result.is_err(), "should panic with expected_items = 0");
    }

    #[test]
    fn test_with_capacity_panics_on_invalid_fpr() {
        let r1 = std::panic::catch_unwind(|| BloomFilter::with_capacity(100, 0.0));
        let r2 = std::panic::catch_unwind(|| BloomFilter::with_capacity(100, 1.0));
        let r3 = std::panic::catch_unwind(|| BloomFilter::with_capacity(100, 1.5));
        assert!(r1.is_err());
        assert!(r2.is_err());
        assert!(r3.is_err());
    }

    // -------------------------------------------------------------------------
    // Bit-level determinism
    // -------------------------------------------------------------------------

    #[test]
    fn test_repeated_insert_is_idempotent() {
        let mut bf = BloomFilter::new(512, 4);
        bf.insert(b"item");
        let bits_after_first = bf.bits.clone();
        bf.insert(b"item");
        assert_eq!(
            bf.bits, bits_after_first,
            "inserting the same item twice should not change the filter"
        );
    }
}
