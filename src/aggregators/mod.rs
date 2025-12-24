mod top_slow_query;

use std::any::Any;

pub use top_slow_query::TopSlowQueries;

pub trait Aggregator: Send + Sync {
    fn update(&mut self, log_line: &[u8]);
    fn merge_box(&mut self, other: &dyn Aggregator);
    fn print(&mut self);
    fn boxed_clone(&self) -> Box<dyn Aggregator>;
    fn as_any(&self) -> &dyn Any;
}
