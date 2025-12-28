mod error_frequency;
mod top_slow_query;
mod connections;

use std::any::Any;

pub use error_frequency::ErrorFrequencyAggregator;
pub use top_slow_query::TopSlowQueries;
pub use connections::ConnectionsAggregator;

use crate::{format::Format, severity::Severity};

pub trait Aggregator: Send + Sync {
    fn update(&mut self, log_line: &[u8], severity: &Severity, fmt: &Format);
    fn merge_box(&mut self, other: &dyn Aggregator);
    fn print(&mut self);
    fn boxed_clone(&self) -> Box<dyn Aggregator>;
    fn as_any(&self) -> &dyn Any;
}
