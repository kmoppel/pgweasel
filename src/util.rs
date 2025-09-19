use chrono::{DateTime, Local, NaiveDate, TimeZone};
use regex::Regex;
use std::error::Error;
use std::fmt;

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

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::{Local, TimeZone};

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
}
