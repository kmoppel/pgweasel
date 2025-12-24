use std::time::Duration;

use regex::Regex;

use crate::util::TimeParseError;

use memchr::memmem;

pub fn extract_duration(record: &[u8]) -> Option<Duration> {
    let start = memmem::find(record, b"duration:")?;
    let mut i = start + b"duration:".len();

    // skip whitespace
    while i < record.len() && record[i] == b' ' {
        i += 1;
    }

    let num_start = i;

    // parse number
    while i < record.len() && (record[i].is_ascii_digit() || record[i] == b'.') {
        i += 1;
    }

    if i == num_start {
        return None;
    }

    let value = std::str::from_utf8(&record[num_start..i]).ok()?;

    // skip whitespace
    while i < record.len() && record[i] == b' ' {
        i += 1;
    }

    // parse unit
    let unit_start = i;
    while i < record.len() && record[i].is_ascii_alphabetic() {
        i += 1;
    }

    let unit = &record[unit_start..i];

    parse_duration_bytes(value, unit)
}

fn parse_duration_bytes(value: &str, unit: &[u8]) -> Option<Duration> {
    let v: f64 = value.parse().ok()?;

    match unit {
        b"ns" => Some(Duration::from_nanos(v as u64)),
        b"us" => Some(Duration::from_micros(v as u64)),
        b"ms" => Some(Duration::from_secs_f64(v / 1_000.0)),
        b"s" => Some(Duration::from_secs_f64(v)),
        b"m" | b"min" | b"minutes" => Some(Duration::from_secs_f64(v * 60.0)),
        _ => None,
    }
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
    fn simpleduration_extract_from_csv() {
        let log = b"Big text and duration: 121.997 ms more text";

        assert_eq!(extract_duration(log), Some(Duration::from_micros(121_997)));
    }

    #[test]
    fn simple_duration_extract_from_log() {
        let log = b"2025-05-21 11:00:40.296 UTC [675]: [3-1] db=postgres,user=cloudsqladmin,host=127.0.0.1 LOG:  duration: 3.032 ms  statement: SELECT extname, current_timestamp FROM pg_catalog.pg_extension UNION SELECT plugin, current_timestamp FROM pg_catalog.pg_replication_slots WHERE slot_type = 'logical' AND database = current_database();";

        assert_eq!(extract_duration(log), Some(Duration::from_micros(3_032)));
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
