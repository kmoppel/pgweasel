mod top_slow_query;

pub use top_slow_query::TopSlowQueryAggregator;

pub trait Aggregator: Send {
    fn update(&mut self, log_line: &[u8]);
    fn merge_box(&mut self, other: &dyn Aggregator);
    fn print(&mut self);
    fn boxed_clone(&self) -> Box<dyn Aggregator>;
}
