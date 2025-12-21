use crate::parsers::LogLine;

mod zero;
mod slow_query;

pub use zero::ZeroAgregator;
pub use slow_query::SlowQueryAggregator;

pub trait Aggregator {
    fn add(&mut self, log_line: LogLine);
    fn print(&mut self);
}
