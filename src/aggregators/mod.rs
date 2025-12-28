mod connections;
mod error_frequency;
mod top_slow_query;

use std::any::Any;

use chrono::{DateTime, Local};
pub use connections::ConnectionsAggregator;
pub use error_frequency::ErrorFrequencyAggregator;
pub use top_slow_query::TopSlowQueries;

use crate::{error::Result, format::Format, severity::Severity};

pub trait Aggregator: Send + Sync {
    fn update(
        &mut self,
        record: &[u8],
        fmt: &Format,
        severity: &Severity,
        log_time: DateTime<Local>,
    ) -> Result<()>;
    fn merge_box(&mut self, other: &dyn Aggregator);
    fn print(&mut self);
    fn boxed_clone(&self) -> Box<dyn Aggregator>;
    fn as_any(&self) -> &dyn Any;
}
