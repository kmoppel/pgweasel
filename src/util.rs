use chrono::{DateTime, Local, NaiveDate, NaiveDateTime, TimeZone};
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
            TimeParseError::InvalidFormat(msg) => write!(f, "Invalid format: {msg}"),
            TimeParseError::ParseError(msg) => write!(f, "Parse error: {msg}"),
        }
    }
}

impl Error for TimeParseError {}

/// Parses time interval input like "10min" or full timestamp strings in common formats,
/// and returns a ``DateTime`` struct or an error.
///
/// Supports:
/// - Time intervals: "10min", "2h", "30s", "1d" (relative to ``reference_time``)
/// - Special keywords: "today"
/// - ISO timestamps: "2025-09-19 15:30:00", "2025-09-19T15:30:00Z"
/// - Date only: "2025-09-19" (uses local timezone)
/// - 2006-01-02 15:04:05.000 MST
/// - 2006-01-02 15:04:05 MST
pub fn time_or_interval_string_to_time(
    human_input: &str,
    reference_time: Option<DateTime<Local>>,
) -> Result<DateTime<Local>, TimeParseError> {
    if human_input.is_empty() {
        return Err(TimeParseError::InvalidFormat("Empty input".to_string()));
    }
    let parsed_time = parse_timestamp_from_string(human_input);
    if let Ok(dt) = parsed_time {
        return Ok(dt);
    }

    let reference_time = reference_time.unwrap_or_else(Local::now);

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
        "Unsupported time delta / timestamp format: {human_input}",
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
            .map_err(|e| TimeParseError::ParseError(format!("Invalid interval value: {e}")))?;
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
                    "Unknown unit: {unit}"
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
        "Not a valid time interval: {input}"
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
        if let Ok(naive_dt) = chrono::NaiveDateTime::parse_from_str(input, format)
            && let Some(local_dt) = Local.from_local_datetime(&naive_dt).single()
        {
            return Ok(local_dt);
        }
    }

    // Handle date-only format (YYYY-MM-DD)
    if input.len() == 10
        && input.chars().nth(4) == Some('-')
        && input.chars().nth(7) == Some('-')
        && let Ok(date) = NaiveDate::parse_from_str(input, "%Y-%m-%d")
        && let Some(datetime) = Local
            .from_local_datetime(&date.and_hms_opt(0, 0, 0).unwrap())
            .single()
    {
        return Ok(datetime);
    }

    Err(TimeParseError::ParseError(format!(
        "Unable to parse timestamp: {input}",
    )))
}

/// Parses the timestamp at the start of a log record without allocating.
///
/// The timestamp occupies the first three whitespace-separated tokens
/// (date, time, timezone), e.g. `2025-08-24 00:05:48.870 CEST`. Rather than
/// re-joining those tokens into a fresh `String` for every record, we slice the
/// contiguous span of the original text covering them and parse that directly.
pub fn parse_timestamp_prefix(text: &str) -> Result<DateTime<Local>, String> {
    let mut tokens = text.split_whitespace();
    let first = tokens.next().ok_or("Missing timestamp first part")?;
    let second = tokens.next().ok_or("Missing timestamp second part")?;
    let third = tokens.next().ok_or("Missing timestamp third part")?;

    // Fast path: the Postgres prefix is a rigid `YYYY-MM-DD HH:MM:SS[.fff]`
    // followed by a timezone abbreviation. Parse the date/time bytes directly
    // instead of letting chrono re-parse the format pattern for every record.
    // The timezone token is intentionally ignored and the value is interpreted
    // as local time, matching `parse_timestamp_from_string`.
    if let Some(naive) = parse_pg_naive_datetime(first, second)
        && let Some(local) = Local.from_local_datetime(&naive).single()
    {
        return Ok(local);
    }

    // Fall back to the general (slower) parser for non-standard formats.
    // `first` and `third` are sub-slices of `text`, so their pointers fall
    // within `text`'s allocation; the offsets bound the timestamp span.
    let base = text.as_ptr() as usize;
    let start = first.as_ptr() as usize - base;
    let end = third.as_ptr() as usize - base + third.len();
    parse_timestamp_from_string(&text[start..end])
}

/// Parses the rigid Postgres date/time tokens without invoking chrono's
/// format-pattern parser. `date` must be `YYYY-MM-DD` and `time` must be
/// `HH:MM:SS` optionally followed by `.` and up to nine fractional-second
/// digits. Returns `None` for anything that does not match exactly, leaving
/// the caller to fall back to a more permissive parser.
fn parse_pg_naive_datetime(date: &str, time: &str) -> Option<NaiveDateTime> {
    let d = date.as_bytes();
    if d.len() != 10 || d[4] != b'-' || d[7] != b'-' {
        return None;
    }
    let year = parse_ascii_u32(&d[0..4])?;
    let month = parse_ascii_u32(&d[5..7])?;
    let day = parse_ascii_u32(&d[8..10])?;

    let t = time.as_bytes();
    if t.len() < 8 || t[2] != b':' || t[5] != b':' {
        return None;
    }
    let hour = parse_ascii_u32(&t[0..2])?;
    let minute = parse_ascii_u32(&t[3..5])?;
    let second = parse_ascii_u32(&t[6..8])?;

    let nanos = if t.len() > 8 {
        if t[8] != b'.' {
            return None;
        }
        let frac = &t[9..];
        if frac.is_empty() || frac.len() > 9 {
            return None;
        }
        let mut value = parse_ascii_u32(frac)?;
        // Scale the fractional part up to nanoseconds (e.g. `.870` -> 870_000_000).
        for _ in 0..(9 - frac.len()) {
            value *= 10;
        }
        value
    } else {
        0
    };

    let date = NaiveDate::from_ymd_opt(i32::try_from(year).ok()?, month, day)?;
    date.and_hms_nano_opt(hour, minute, second, nanos)
}

/// Parses an ASCII byte slice of decimal digits into a `u32`, rejecting any
/// non-digit byte. The slices passed here are short (at most nine digits) so
/// overflow is not a concern.
#[inline]
fn parse_ascii_u32(bytes: &[u8]) -> Option<u32> {
    let mut acc: u32 = 0;
    for &b in bytes {
        if !b.is_ascii_digit() {
            return None;
        }
        acc = acc * 10 + u32::from(b - b'0');
    }
    Some(acc)
}

pub fn parse_timestamp_from_string(input: &str) -> Result<DateTime<Local>, String> {
    let input = input.trim();

    // log_line_prefix formats to try
    // %t	Time stamp without milliseconds 2006-01-02 15:04:05.000 MST
    // %m	Time stamp with milliseconds 2006-01-02 15:04:05 MST
    // %n	Time stamp with milliseconds (as a Unix epoch)  TODO
    let formats = [
        "%Y-%m-%d %H:%M:%S%.3f %Z", // 2025-08-24 00:05:48.870 CEST
        "%Y-%m-%d %H:%M:%S% %Z",
    ];

    // Try parsing with timezone first
    for format in &formats {
        if let Ok(dt) = DateTime::parse_from_str(input, format) {
            return Ok(dt.with_timezone(&Local));
        }
    }

    // Try parsing as naive datetime and convert to local
    for format in &formats {
        if let Ok(naive_dt) = NaiveDateTime::parse_from_str(input, format)
            && let Some(local_dt) = Local.from_local_datetime(&naive_dt).single()
        {
            return Ok(local_dt);
        }
    }

    Err(format!("Unable to parse timestamp: '{input}'"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::{Datelike, Local, TimeZone};

    #[test]
    fn test_parse_timestamp_prefix_matches_full_line() {
        // Full log line: the prefix parse must yield the same instant as parsing
        // just the timestamp tokens, ignoring the trailing log message.
        let line = "2025-08-24 00:05:48.870 CEST [123] LOG:  database system is ready";
        let prefix = parse_timestamp_prefix(line).unwrap();
        let expected = parse_timestamp_from_string("2025-08-24 00:05:48.870 CEST").unwrap();
        assert_eq!(prefix, expected);
    }

    #[test]
    fn test_parse_timestamp_prefix_missing_parts() {
        assert!(parse_timestamp_prefix("2025-08-24 00:05:48.870").is_err());
        assert!(parse_timestamp_prefix("").is_err());
    }

    #[test]
    fn test_parse_timestamp_prefix_without_millis() {
        // The fast path must also handle `HH:MM:SS` without a fractional part
        // and agree with the general parser.
        let line = "2025-08-24 00:05:48 CEST [123] LOG:  ready";
        let prefix = parse_timestamp_prefix(line).unwrap();
        let expected = parse_timestamp_from_string("2025-08-24 00:05:48 CEST").unwrap();
        assert_eq!(prefix, expected);
    }

    #[test]
    fn test_parse_pg_naive_datetime_fast_path() {
        use chrono::Timelike;

        let naive = parse_pg_naive_datetime("2025-08-24", "00:05:48.870").unwrap();
        assert_eq!(naive.year(), 2025);
        assert_eq!(naive.month(), 8);
        assert_eq!(naive.day(), 24);
        assert_eq!(naive.hour(), 0);
        assert_eq!(naive.minute(), 5);
        assert_eq!(naive.second(), 48);
        assert_eq!(naive.nanosecond(), 870_000_000);

        // No fractional seconds.
        let naive = parse_pg_naive_datetime("2025-12-31", "23:59:59").unwrap();
        assert_eq!(naive.nanosecond(), 0);
        assert_eq!(naive.hour(), 23);

        // Malformed inputs return None so the caller can fall back.
        assert!(parse_pg_naive_datetime("2025/08/24", "00:05:48").is_none());
        assert!(parse_pg_naive_datetime("2025-08-24", "00-05-48").is_none());
        assert!(parse_pg_naive_datetime("2025-08-24", "00:05:48.").is_none());
        assert!(parse_pg_naive_datetime("2025-08-24", "0a:05:48").is_none());
        assert!(parse_pg_naive_datetime("2025-13-24", "00:05:48").is_none());
    }

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
    fn test_parse_timestamp_from_string() {
        let result = parse_timestamp_from_string("2025-05-02 18:25:51.151 EEST").unwrap();
        // println!("Result: {}", result);
        assert_eq!(result.year(), 2025);
        assert_eq!(result.month(), 5);
        let result_no_millis = parse_timestamp_from_string("2025-05-02 18:25:51 EEST").unwrap();
        // println!("Result no millis: {}", result_no_millis);
        assert_eq!(result_no_millis.year(), 2025);
        assert_eq!(result_no_millis.month(), 5);
    }
}
