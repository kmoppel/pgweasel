mod filter_contains;

pub use filter_contains::FilterContains;

pub trait Filter: Sync {
    fn matches(&self, record: &[u8]) -> bool;
}
