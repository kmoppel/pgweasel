use std::time::Duration;

use regex::Regex;

use crate::{aggregators::Aggregator, parsers::LogLine, util::parse_duration};

pub struct SlowQueryAggregator {
    // treshold to consider query slow, in miliseconds
    treshold: Duration,
    slow_queries: Vec<(LogLine, Duration)>,
}

impl SlowQueryAggregator {
    pub fn new(treshold: Duration) -> Self {
        SlowQueryAggregator {
            treshold,
            slow_queries: vec![],
        }
    }
}

impl Aggregator for SlowQueryAggregator {
    fn add(&mut self, log_line: LogLine) {
        if let Some(duration) = extract_duration(&log_line.message) {
            if duration > self.treshold {
                self.slow_queries.push((log_line, duration));
            }
        }
    }

    fn print(&mut self) {
        self.slow_queries.sort_by(|a, b| b.1.cmp(&a.1));
        for (log_line, duration) in &self.slow_queries {
            println!("{:?} | {}", duration, log_line.message);
        }
    }
}

fn extract_duration(log: &str) -> Option<Duration> {
    let re = Regex::new(r"duration:\s+([\d.]+\s ?(ns|us|µs|ms|s|m|min|minutes))").ok()?;
    let caps = re.captures(log)?;

    let ms = caps.get(1)?.as_str();
    parse_duration(ms).ok()
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn log_extract_test() {
        let log = "Big text and duration: 121.997 ms more text";

        assert_eq!(extract_duration(log), Some(Duration::from_millis(122)));
    }
}
