mod filter_contains;
mod filter_slow;

pub use filter_contains::FilterContains;
pub use filter_slow::FilterSlow;

pub trait Filter: Sync {
    fn matches(&self, record: &[u8]) -> bool;
}
