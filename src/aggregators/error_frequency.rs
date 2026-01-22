use std::{any::Any, collections::HashMap};

use chrono::{DateTime, Local};

use crate::{aggregators::Aggregator, error::Result, format::Format, severity::Severity};

#[derive(Clone, Default)]
pub struct ErrorFrequencyAggregator {
    // TODO: Check ablity to store u8 arrays directly to avoid UTF-8 conversion overhead
    counts: HashMap<String, u64>,
    limit: usize,
}

impl ErrorFrequencyAggregator {
    pub fn new(limit: usize) -> Self {
        Self {
            counts: HashMap::new(),
            limit,
        }
    }
}

impl Aggregator for ErrorFrequencyAggregator {
    fn update(
        &mut self,
        record: &[u8],
        fmt: &Format,
        _severity: Severity,
        _log_time: DateTime<Local>,
    ) -> Result<()> {
        let message =
            fmt.message_from_bytes(record)
                .ok_or(crate::Error::NotAbleToExtractMessage {
                    record: String::from_utf8(record.to_vec()).unwrap(),
                })?;
        let message = String::from_utf8_lossy(message).to_string();

        *self.counts.entry(message).or_insert(0) += 1;
        //// This code is executed in threads, so we cannot apply here the top-N logic directly
        //// Instead, we will do it in the merge_box method
        Ok(())
    }

    fn merge_box(&mut self, other: &dyn Aggregator) {
        let other = other
            .as_any()
            .downcast_ref::<ErrorFrequencyAggregator>()
            .expect("Aggregator type mismatch");

        for (msg, count) in &other.counts {
            *self.counts.entry(msg.clone()).or_insert(0) += count;
        }

        /* Enforce top-N limit */
        while self.counts.len() > self.limit {
            if let Some((least_key, _)) = self
                .counts
                .iter()
                .min_by_key(|(_, count)| *count)
                .map(|(k, v)| (k.clone(), *v))
            {
                self.counts.remove(&least_key);
            }
        }
    }

    fn print(&mut self) {
        let mut entries: Vec<_> = self.counts.iter().collect();

        // Sort descending by frequency
        entries.sort_by(|a, b| b.1.cmp(a.1));

        println!("Most frequent error messages:");
        for (msg, count) in entries {
            println!("{count:>6}  {msg}");
        }
    }

    fn boxed_clone(&self) -> Box<dyn Aggregator> {
        Box::new(self.clone())
    }

    fn as_any(&self) -> &dyn Any {
        self
    }
}
