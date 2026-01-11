use std::{any::Any, collections::BTreeMap, time::Duration};

use chrono::{DateTime, Local, TimeZone};

use crate::{aggregators::Aggregator, error::Result, format::Format, severity::Severity};

#[derive(Clone, Default)]
pub struct ErrorHistogramAggregator {
    bucket_width: Duration,
    buckets: BTreeMap<i64, i64>,
}

impl ErrorHistogramAggregator {
    pub fn new(bucket_width: Duration) -> Self {
        Self {
            bucket_width,
            buckets: BTreeMap::new(),
        }
    }

    fn bucket(&self, log_time: DateTime<Local>) -> Result<i64> {
        let ts = log_time.timestamp();
        let width: i64 = self.bucket_width.as_secs().try_into().map_err(|_| crate::error::Error::TimestampBeforeEpoch { timestamp: log_time.to_rfc2822() })?;

        Ok((ts / width) * width)
    }
}

impl Aggregator for ErrorHistogramAggregator {
    fn update(
        &mut self,
        _record: &[u8],
        _fmt: &Format,
        _severity: Severity,
        log_time: DateTime<Local>,
    ) -> Result<()> {
        let bucket = self.bucket(log_time)?;
        *self.buckets.entry(bucket).or_insert(0) += 1;

        Ok(())
    }

    fn merge_box(&mut self, other: &dyn Aggregator) {
        let other = other
            .as_any()
            .downcast_ref::<ErrorHistogramAggregator>()
            .expect("Aggregator type mismatch");

        for (&bucket, &count) in &other.buckets {
            *self.buckets.entry(bucket).or_insert(0) += count;
        }
    }

    fn print(&mut self) {
        const BAR_WIDTH: usize = 50;

        let max_count = self.buckets.values().copied().max().unwrap_or(0);
        if max_count == 0 {
            return;
        }

        for (&bucket, &count) in &self.buckets {
            let filled = ((count as f64 / max_count as f64) * BAR_WIDTH as f64)
                .round()
                .clamp(0.0, BAR_WIDTH as f64) as usize;

            let empty = BAR_WIDTH - filled;

            let time = Local.timestamp_opt(bucket, 0).single();
            if let Some(time) = time {
                println!(
                    "[{}] {}{} {}",
                    time.format("%Y-%m-%d %H:%M:%S"),
                    "#".repeat(filled),
                    "-".repeat(empty),
                    count
                );
            }
        }
    }

    fn boxed_clone(&self) -> Box<dyn Aggregator> {
        Box::new(self.clone())
    }

    fn as_any(&self) -> &dyn Any {
        self
    }
}
