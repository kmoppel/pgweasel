use regex::Regex;

use crate::{aggregators::Aggregator, parsers::LogLine};

pub struct SlowQueryAggregator {
    // treshold to consider query slow, in miliseconds
    treshold: u64,
    slow_queries: Vec<(LogLine, u64)>,
}

impl SlowQueryAggregator {
    pub fn new(treshold: u64) -> Self {
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
            println!("{} ms | {}", duration, log_line.message);
        }
    }
}

fn extract_duration(log: &str) -> Option<u64> {
    let re = Regex::new(r"duration:\s+([\d.]+)\s+ms").ok()?;
    let caps = re.captures(log)?;

    let ms: f64 = caps.get(1)?.as_str().parse().ok()?;
    Some(ms.round() as u64)
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn log_extract_test() {
        let log = "duration: 121.397 ms";

        assert_eq!(extract_duration(log), Some(121));
    }
}
