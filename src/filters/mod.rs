mod filter_contains;
mod filter_slow;
mod locking_filter;

pub use filter_contains::FilterContains;
pub use filter_slow::FilterSlow;
pub use locking_filter::LockingFilter;

pub trait Filter: Sync {
    fn matches(&self, record: &[u8]) -> bool;
}
