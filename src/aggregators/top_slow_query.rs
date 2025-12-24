use std::time::Duration;

use regex::Regex;

use crate::{aggregators::Aggregator};

pub struct TopSlowQueryAggregator<'a> {
    // treshold to consider query slow, in miliseconds
    ten_top_slow_queries: Vec<(&'a [u8], Duration)>,
}

impl TopSlowQueryAggregator<'_> {
    pub fn new(treshold: Duration) -> Self {
        TopSlowQueryAggregator {
            ten_top_slow_queries: vec![],
        }
    }
}

impl Aggregator for TopSlowQueryAggregator<'_> {
    fn update(&mut self, log_line: &[u8]) {
        todo!()
    }

    fn merge_box(&mut self, other: &dyn Aggregator) {
        todo!()
    }

    fn print(&mut self) {
        todo!()
    }

    fn boxed_clone(&self) -> Box<dyn Aggregator> {
        todo!()
    }
    // fn add(&mut self, log_line: &str) {
    //     if let Some(duration) = extract_duration(&log_line) {
    //         // if duration > self.treshold {
    //         //     self.slow_queries.push((&log_line, duration));
    //         // }
    //     }
    // }

    // fn print(&mut self) {
    //     self.slow_queries.sort_by(|a, b| b.1.cmp(&a.1));
    //     for (log_line, duration) in &self.slow_queries {
    //         println!("{:?} | {}", duration, log_line);
    //     }
    // }
}

