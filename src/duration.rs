use std::time::Duration;

use regex::Regex;

use crate::util::TimeParseError;

pub fn extract_duration(log: &str) -> Option<Duration> {
    let re = Regex::new(r"duration:\s+([\d.]+\s ?(ns|us|µs|ms|s|m|min|minutes))").ok()?;
    let caps = re.captures(log)?;

    let ms = caps.get(1)?.as_str();
    parse_duration(ms).ok()
}

/// Parses time interval input like "10min", "10 min", "10.5s"
/// and returns a Duration struct or an error.
///
/// Supports:
/// - Time intervals: "10min", "2h", "30s", "1d"
/// - Units: ns, us, µs, ms, s, m, min, minutes
/// - Decimal values: "1.5s" get rounded to nearest integer
pub fn parse_duration(input: &str) -> Result<Duration, TimeParseError> {
    let duration_regex = Regex::new(r"^([\d.]+) ?(ns|us|µs|ms|s|m|min|minutes)$").unwrap();

    if let Some(captures) = duration_regex.captures(input) {
        let f_value: f64 = captures[1]
            .parse()
            .map_err(|e| TimeParseError::ParseError(format!("Invalid duration value: {}", e)))?;
        let unit = &captures[2];
        let value = f_value.round() as u64;

        return match unit {
            "ns" => Ok(Duration::from_nanos(value)),
            "us" | "µs" => Ok(Duration::from_micros(value)),
            "ms" => Ok(Duration::from_millis(value)),
            "s" => Ok(Duration::from_secs(value)),
            "m" | "min" | "minutes" => Ok(Duration::from_secs(value * 60)),
            _ => {
                return Err(TimeParseError::InvalidFormat(format!(
                    "Unknown unit: {}",
                    unit
                )));
            }
        };
    }

    Err(TimeParseError::InvalidFormat(format!(
        "Not a valid time interval: {}",
        input
    )))
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn csv_log_extract_test() {
        let log = "Big text and duration: 121.997 ms more text";

        assert_eq!(extract_duration(log), Some(Duration::from_millis(122)));
    }

    #[test]
    fn simple_log_extract_test() {
        let log = "2025-05-21 11:00:40.296 UTC [675]: [3-1] db=postgres,user=cloudsqladmin,host=127.0.0.1 LOG:  duration: 3.032 ms  statement: SELECT extname, current_timestamp FROM pg_catalog.pg_extension UNION SELECT plugin, current_timestamp FROM pg_catalog.pg_replication_slots WHERE slot_type = 'logical' AND database = current_database();";

        assert_eq!(extract_duration(log), Some(Duration::from_millis(3)));
    }

    #[test]
    fn test_parse_duration() {
        let cases = vec![
            ("10ms", Duration::from_millis(10)),
            ("5s", Duration::from_secs(5)),
            ("5s", Duration::from_secs(5)),
            ("2 m", Duration::from_secs(120)),
            ("100us", Duration::from_micros(100)),
            ("200ns", Duration::from_nanos(200)),
            ("15min", Duration::from_secs(900)),
            ("20minutes", Duration::from_secs(1200)),
        ];

        for (input, expected) in cases {
            let result = parse_duration(input).unwrap();
            assert_eq!(result, expected, "Failed on input: {}", input);
        }

        let invalid_cases = vec!["10xyz", "abc", ""];

        for input in invalid_cases {
            let result = parse_duration(input);
            assert!(result.is_err(), "Expected error for input: {}", input);
        }
    }
}
