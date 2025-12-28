use std::{any::Any, collections::HashMap};

use chrono::{DateTime, Local};

use crate::{aggregators::Aggregator, format::Format, severity::Severity};

#[derive(Clone, Default)]
pub struct ErrorFrequencyAggregator {
    // TODO: Check ablity to store u8 arrays directly to avoid UTF-8 conversion overhead
    counts: HashMap<String, u64>,
}

impl ErrorFrequencyAggregator {
    pub fn new() -> Self {
        Self {
            counts: HashMap::new(),
        }
    }
}

impl Aggregator for ErrorFrequencyAggregator {
    fn update(
        &mut self,
        record: &[u8],
        fmt: &Format,
        _severity: &Severity,
        _log_time: DateTime<Local>,
    ) {
        // TODO: Handle the case where message extraction fails
        let message = match fmt.message_from_bytes(record) {
            Some(msg) => msg,
            None => return,
        };
        let message = String::from_utf8_lossy(message).to_string();

        *self.counts.entry(message).or_insert(0) += 1;
    }

    fn merge_box(&mut self, other: &dyn Aggregator) {
        let other = other
            .as_any()
            .downcast_ref::<ErrorFrequencyAggregator>()
            .expect("Aggregator type mismatch");

        for (msg, count) in &other.counts {
            *self.counts.entry(msg.clone()).or_insert(0) += count;
        }
    }

    fn print(&mut self) {
        let mut entries: Vec<_> = self.counts.iter().collect();

        // Sort descending by frequency
        entries.sort_by(|a, b| b.1.cmp(a.1));

        println!("Most frequent error messages:");
        for (msg, count) in entries {
            println!("{:>6}  {}", count, msg);
        }
    }

    fn boxed_clone(&self) -> Box<dyn Aggregator> {
        Box::new(self.clone())
    }

    fn as_any(&self) -> &dyn Any {
        self
    }
}
