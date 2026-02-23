mod filter_contains;
mod filter_contains_ci;
mod filter_slow;
mod locking_filter;
mod system_filter;

pub use filter_contains::FilterContains;
pub use filter_contains_ci::FilterContainsCi;
pub use filter_slow::FilterSlow;
pub use locking_filter::LockingFilter;
pub use system_filter::SystemFilter;

use crate::format::Format;

pub trait Filter: Sync {
    fn matches(&self, record: &[u8], fmt: &Format) -> bool;
}
