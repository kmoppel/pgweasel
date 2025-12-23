mod slow_query;

pub use slow_query::SlowQueryAggregator;

pub trait Aggregator {
    fn add(&mut self, log_line: &str);
    fn print(&mut self);
}
