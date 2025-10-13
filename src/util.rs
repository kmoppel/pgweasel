use chrono::{DateTime, Local, NaiveDate, NaiveDateTime, TimeZone};
use regex::Regex;
use std::error::Error;
use std::fmt;

use crate::{Cli, ConvertedArgs};

#[derive(Debug)]
pub enum TimeParseError {
    InvalidFormat(String),
    ParseError(String),
}

impl fmt::Display for TimeParseError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            TimeParseError::InvalidFormat(msg) => write!(f, "Invalid format: {}", msg),
            TimeParseError::ParseError(msg) => write!(f, "Parse error: {}", msg),
        }
    }
}

impl Error for TimeParseError {}

/// Parses time interval input like "10min" or full timestamp strings in common formats,
/// and returns a DateTime struct or an error.
///
/// Supports:
/// - Time intervals: "10min", "2h", "30s", "1d" (relative to reference_time)
/// - Special keywords: "today"
/// - ISO timestamps: "2025-09-19 15:30:00", "2025-09-19T15:30:00Z"
/// - Date only: "2025-09-19" (uses local timezone)
pub fn time_or_interval_string_to_time(
    human_input: &str,
    reference_time: Option<DateTime<Local>>,
) -> Result<DateTime<Local>, TimeParseError> {
    if human_input.is_empty() {
        return Err(TimeParseError::InvalidFormat("Empty input".to_string()));
    }

    let reference_time = reference_time.unwrap_or_else(|| Local::now());

    // Special case for "today"
    if human_input.to_lowercase() == "today" {
        let date = reference_time.date_naive();
        return Ok(Local
            .from_local_datetime(&date.and_hms_opt(0, 0, 0).unwrap())
            .unwrap());
    }

    // Try parsing time intervals first
    if let Ok(datetime) = parse_time_interval(human_input, reference_time) {
        return Ok(datetime);
    }

    // Try parsing as full timestamp
    if let Ok(datetime) = parse_timestamp(human_input, reference_time) {
        return Ok(datetime);
    }

    Err(TimeParseError::InvalidFormat(format!(
        "Unsupported time delta / timestamp format: {}",
        human_input
    )))
}

fn parse_time_interval(
    input: &str,
    reference_time: DateTime<Local>,
) -> Result<DateTime<Local>, TimeParseError> {
    // Parse duration using a comprehensive regex approach
    let duration_regex =
        Regex::new(r"^(-?\d+)(ns|us|µs|ms|s|m|min|minutes|h|hours|d|day|days)$").unwrap();

    if let Some(captures) = duration_regex.captures(input) {
        let value: i64 = captures[1]
            .parse()
            .map_err(|e| TimeParseError::ParseError(format!("Invalid duration value: {}", e)))?;
        let unit = &captures[2];

        let duration = match unit {
            "ns" => chrono::Duration::nanoseconds(value),
            "us" | "µs" => chrono::Duration::microseconds(value),
            "ms" => chrono::Duration::milliseconds(value),
            "s" => chrono::Duration::seconds(value),
            "m" | "min" | "minutes" => chrono::Duration::minutes(value),
            "h" | "hours" => chrono::Duration::hours(value),
            "d" | "day" | "days" => chrono::Duration::hours(value * 24),
            _ => {
                return Err(TimeParseError::InvalidFormat(format!(
                    "Unknown unit: {}",
                    unit
                )));
            }
        };

        // For negative intervals (with explicit minus sign), add to reference time (future)
        // For positive intervals (without sign), subtract from reference time (past/"ago")
        let result_time = if input.starts_with('-') {
            reference_time + duration.abs()
        } else {
            reference_time - duration
        };

        return Ok(result_time);
    }

    Err(TimeParseError::InvalidFormat(format!(
        "Not a valid time interval: {}",
        input
    )))
}

fn parse_timestamp(
    input: &str,
    _reference_time: DateTime<Local>,
) -> Result<DateTime<Local>, TimeParseError> {
    // Common timestamp formats
    let formats = vec![
        "%Y-%m-%d %H:%M:%S%.3f %Z", // 2025-09-19 15:30:00.123 UTC
        "%Y-%m-%d %H:%M:%S %Z",     // 2025-09-19 15:30:00 UTC
        "%Y-%m-%dT%H:%M:%S%.3fZ",   // 2025-09-19T15:30:00.123Z
        "%Y-%m-%dT%H:%M:%SZ",       // 2025-09-19T15:30:00Z
        "%Y-%m-%d %H:%M:%S%.3f",    // 2025-09-19 15:30:00.123 (local)
        "%Y-%m-%d %H:%M:%S",        // 2025-09-19 15:30:00 (local)
        "%Y-%m-%dT%H:%M:%S%.3f",    // 2025-09-19T15:30:00.123 (local)
        "%Y-%m-%dT%H:%M:%S",        // 2025-09-19T15:30:00 (local)
    ];

    // Try parsing with timezone info first
    for format in &formats {
        if let Ok(dt) = DateTime::parse_from_str(input, format) {
            return Ok(dt.with_timezone(&Local));
        }
    }

    // Try parsing as naive datetime (local timezone)
    let naive_formats = vec![
        "%Y-%m-%d %H:%M:%S%.3f",
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%dT%H:%M:%S%.3f",
        "%Y-%m-%dT%H:%M:%S",
    ];

    for format in &naive_formats {
        if let Ok(naive_dt) = chrono::NaiveDateTime::parse_from_str(input, format) {
            if let Some(local_dt) = Local.from_local_datetime(&naive_dt).single() {
                return Ok(local_dt);
            }
        }
    }

    // Handle date-only format (YYYY-MM-DD)
    if input.len() == 10 && input.chars().nth(4) == Some('-') && input.chars().nth(7) == Some('-') {
        if let Ok(date) = NaiveDate::parse_from_str(input, "%Y-%m-%d") {
            if let Some(datetime) = Local
                .from_local_datetime(&date.and_hms_opt(0, 0, 0).unwrap())
                .single()
            {
                return Ok(datetime);
            }
        }
    }

    Err(TimeParseError::ParseError(format!(
        "Unable to parse timestamp: {}",
        input
    )))
}

pub fn parse_timestamp_from_string(input: &str) -> Result<DateTime<Local>, String> {
    let input = input.trim();

    // Common formats to try
    let formats = vec![
        "%Y-%m-%d %H:%M:%S%.3f %Z",     // 2025-08-24 00:05:48.870 CEST
        "%Y-%m-%d %H:%M:%S%.f %Z",       // with any fractional seconds
        "%Y-%m-%d %H:%M:%S %Z",          // without fractional seconds
        "%Y-%m-%d %H:%M:%S%.3f",         // without timezone
        "%Y-%m-%d %H:%M:%S%.f",          // without timezone, any fractional
        "%Y-%m-%d %H:%M:%S",             // without timezone and fractional
        "%Y-%m-%d %H:%M",                // without seconds
    ];

    // Try parsing with timezone first
    for format in formats.iter() {
        if let Ok(dt) = DateTime::parse_from_str(input, format) {
            return Ok(dt.with_timezone(&Local));
        }
    }

    // Try parsing as naive datetime and convert to local
    for format in formats.iter() {
        if let Ok(naive_dt) = NaiveDateTime::parse_from_str(input, format) {
            if let Some(local_dt) = Local.from_local_datetime(&naive_dt).single() {
                return Ok(local_dt);
            }
        }
    }

    Err(format!("Unable to parse timestamp: '{}'", input))
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::{Datelike, Local, TimeZone};

    #[test]
    fn test_today() {
        let result = time_or_interval_string_to_time("today", None).unwrap();
        let today = Local::now().date_naive();
        assert_eq!(result.date_naive(), today);
        assert_eq!(
            result.time(),
            chrono::NaiveTime::from_hms_opt(0, 0, 0).unwrap()
        );
    }

    #[test]
    fn test_time_intervals() {
        let reference = Local.with_ymd_and_hms(2025, 9, 19, 15, 30, 0).unwrap();

        // Test minutes ago
        let result = time_or_interval_string_to_time("10m", Some(reference)).unwrap();
        let expected = reference - chrono::Duration::minutes(10);
        assert_eq!(result, expected);

        // Test hours ago
        let result = time_or_interval_string_to_time("2h", Some(reference)).unwrap();
        let expected = reference - chrono::Duration::hours(2);
        assert_eq!(result, expected);

        // Test days (converted to hours)
        let result = time_or_interval_string_to_time("1d", Some(reference)).unwrap();
        let expected = reference - chrono::Duration::hours(24);
        assert_eq!(result, expected);
    }

    #[test]
    fn test_time_intervals_extended() {
        let reference = Local.with_ymd_and_hms(2025, 9, 19, 15, 30, 0).unwrap();

        // Test "min" and "minutes"
        let result = time_or_interval_string_to_time("10min", Some(reference)).unwrap();
        let expected = reference - chrono::Duration::minutes(10);
        assert_eq!(result, expected);

        let result = time_or_interval_string_to_time("5minutes", Some(reference)).unwrap();
        let expected = reference - chrono::Duration::minutes(5);
        assert_eq!(result, expected);

        // Test "hours"
        let result = time_or_interval_string_to_time("2hours", Some(reference)).unwrap();
        let expected = reference - chrono::Duration::hours(2);
        assert_eq!(result, expected);
    }

    #[test]
    fn test_negative_intervals() {
        let reference = Local.with_ymd_and_hms(2025, 9, 19, 15, 30, 0).unwrap();

        // Test negative interval (future)
        let result = time_or_interval_string_to_time("-10m", Some(reference)).unwrap();
        let expected = reference + chrono::Duration::minutes(10);
        assert_eq!(result, expected);
    }

    #[test]
    fn test_date_only() {
        let result = time_or_interval_string_to_time("2025-09-19", None).unwrap();
        assert_eq!(result.date_naive().to_string(), "2025-09-19");
        assert_eq!(
            result.time(),
            chrono::NaiveTime::from_hms_opt(0, 0, 0).unwrap()
        );
    }

    #[test]
    fn test_full_timestamp() {
        let result = time_or_interval_string_to_time("2025-09-19 15:30:00", None).unwrap();
        assert_eq!(result.date_naive().to_string(), "2025-09-19");
        assert_eq!(
            result.time(),
            chrono::NaiveTime::from_hms_opt(15, 30, 0).unwrap()
        );
    }

    #[test]
    fn test_invalid_input() {
        let result = time_or_interval_string_to_time("invalid", None);
        assert!(result.is_err());

        let result = time_or_interval_string_to_time("", None);
        assert!(result.is_err());
    }

    #[test]
    fn test_line_has_timestamp_prefix() {
        // Test various PostgreSQL log line formats that should match
        let test_cases = vec![
            "2025-05-02 12:27:52.634 EEST [2380404]",
            "2025-05-05 06:00:51 UTC:90.190.32.92(32890):postgres@postgres:[1315]:LOG:  statement: BEGIN;",
            "May 30 11:03:43 i13400f postgres[693826]: [5-1] 2025-05-30 11:03:43.622 EEST [693826] LOG:  database system is ready to accept connections",
            "2025-01-09 20:48:11.713 GMT LOG:  checkpoint starting: time",
            "2022-02-19 14:47:24 +08 [66019]: [10-1] session=6210927b.101e3,user=postgres,db=ankara,app=PostgreSQL JDBC Driver,client=localhost | LOG:  duration: 0.073 ms",
        ];

        for test_case in test_cases {
            let (has_timestamp, extracted_time) = line_has_timestamp_prefix(test_case);
            assert!(has_timestamp, "Expected timestamp to be found in: {}", test_case);
            assert!(extracted_time.is_some(), "Expected extracted time to be Some for: {}", test_case);
            println!("✓ Found timestamp in: {} -> {:?}", test_case, extracted_time);
        }

        // Test lines that should NOT match
        let negative_test_cases = vec![
            "This is just a regular log line without timestamp",
            "bla 2025-05-02 12:27:52.634 EEST [2380404]",
            "    ON CONFLICT (id) DO UPDATE SET master_time = (now() at time zone 'utc');",
        ];

        for test_case in negative_test_cases {
            let (has_timestamp, extracted_time) = line_has_timestamp_prefix(test_case);
            assert!(!has_timestamp, "Expected no timestamp in: {}", test_case);
            assert!(extracted_time.is_none(), "Expected extracted time to be None for: {}", test_case);
            println!("✓ Correctly identified no timestamp in: {}", test_case);
        }
    }

    #[test]
    fn test_parse_timestamp_from_string() {
        let result = parse_timestamp_from_string("2025-05-02 18:25:51.151 EEST").unwrap();
        assert_eq!(result.month(), 5);
    }
}

pub fn convert_args(cli: &Cli) -> Result<ConvertedArgs, Box<dyn Error>> {
    let begin = if let Some(begin_str) = &cli.begin {
        match time_or_interval_string_to_time(begin_str, None) {
            Ok(datetime) => {
                println!(
                    "Parsed begin time: {}",
                    datetime.format("%Y-%m-%d %H:%M:%S %Z")
                );
                Some(datetime)
            }
            Err(e) => {
                return Err(Box::new(e));
            }
        }
    } else {
        None
    };

    Ok(ConvertedArgs { begin })
}

pub fn line_has_timestamp_prefix(line: &str) -> (bool, Option<String>) {
    // Check if the line starts with a timestamp pattern
    let timestamp_regex = Regex::new(r"^(?<syslog>[A-Za-z]{3} [0-9]{1,2} [0-9:]{6,} .*?: \[[0-9\-]+\] )?(?P<time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})").unwrap();
    
    if let Some(captures) = timestamp_regex.captures(line) {
        let time_match = captures.name("time").map(|m| m.as_str().to_string());
        (true, time_match)
    } else {
        (false, None)
    }
}