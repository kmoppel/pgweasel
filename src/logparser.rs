use log::debug;
use once_cell::sync::Lazy;
use regex::Regex;
/// Turns input log lines into structured log entries
use std::io::Result;

#[allow(dead_code)]
static TIMESTAMP_REGEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^(?P<time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})").unwrap()
});

#[allow(dead_code)]
static SEVERITY_REGEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^.*?[\s:\-](?P<log_level>[A-Z12345]{3,12}):  (?P<message>.*)$").unwrap()
});

pub static LOG_ENTRY_START_REGEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^(?P<time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})[\s:\-].*?[\s:\-]?(?P<log_level>[A-Z12345]{3,12}):  ").unwrap()
});

#[allow(dead_code)]
struct LogEntry {
    pub log_time: String,
    pub severity: String,
    pub message: String,
    pub lines: Vec<String>, // As-is lines for that record
}

/// Reads a PostgreSQL log file and returns an iterator over LogEntry structs
///
/// This function parses PostgreSQL log files by detecting timestamp patterns at the beginning
/// of lines and grouping all subsequent lines until the next timestamp into a single LogEntry.
///
/// # Arguments
///
/// * `filepath` - A string slice that holds the path to the PostgreSQL log file
///
/// # Returns
///
/// * `Result<Vec<LogEntry>>` - A vector of LogEntry structs representing parsed log records
///
/// # Examples
///
/// ```
/// use pgweasel_rust::logparser::getentries;
///
/// let entries = getentries("path/to/postgresql.log")?;
/// for entry in entries {
///     println!("Time: {}, Severity: {}, Message: {}",
///              entry.log_time, entry.severity, entry.message);
/// }
/// ```
#[allow(dead_code)]
fn getentries(filepath: &str) -> Result<Vec<LogEntry>> {
    let file = std::fs::File::open(filepath)?;
    let lines = std::io::BufRead::lines(std::io::BufReader::new(file));
    let mut entries = Vec::new();
    let mut current_entry_lines = Vec::new();
    let mut current_timestamp = String::new();
    let mut current_severity = String::new();

    for line_result in lines {
        let line = line_result?;
        debug!("***Processing line: {}", line);
        // Check if this line starts with a timestamp
        if let Some(captures) = TIMESTAMP_REGEX.captures(&line) {
            // If we have accumulated lines for a previous entry, save it
            if !current_entry_lines.is_empty() {
                let message = extract_message(&current_entry_lines);
                entries.push(LogEntry {
                    log_time: current_timestamp.clone(),
                    severity: current_severity.clone(),
                    message,
                    lines: current_entry_lines.clone(),
                });
            }

            // Start a new entry
            current_timestamp = captures.name("time").unwrap().as_str().to_string();
            current_severity = extract_severity(&line, &SEVERITY_REGEX);
            current_entry_lines = vec![line];
        } else {
            // This line is a continuation of the current log entry
            current_entry_lines.push(line);
        }
    }

    // Don't forget the last entry if the file doesn't end with a new timestamp
    if !current_entry_lines.is_empty() {
        let message = extract_message(&current_entry_lines);
        entries.push(LogEntry {
            log_time: current_timestamp,
            severity: current_severity,
            message,
            lines: current_entry_lines,
        });
    }

    Ok(entries)
}

/// Extracts the severity level from a log line
#[allow(dead_code)]
fn extract_severity(line: &str, severity_regex: &Regex) -> String {
    if let Some(captures) = severity_regex.captures(line) {
        captures.get(1).unwrap().as_str().to_string()
    } else {
        // Fallback: try to find common severity keywords
        let line_upper = line.to_uppercase();
        if line_upper.contains("ERROR") {
            "ERROR".to_string()
        } else if line_upper.contains("WARNING") {
            "WARNING".to_string()
        } else if line_upper.contains("FATAL") {
            "FATAL".to_string()
        } else if line_upper.contains("PANIC") {
            "PANIC".to_string()
        } else if line_upper.contains("LOG") {
            "LOG".to_string()
        } else {
            "UNKNOWN".to_string()
        }
    }
}

/// Extracts the message content from log entry lines
#[allow(dead_code)]
fn extract_message(lines: &[String]) -> String {
    if lines.is_empty() {
        return String::new();
    }

    let first_line = &lines[0];

    // Try to extract message after the timestamp and severity
    // Pattern: timestamp [severity] message
    let message_regex =
        Regex::new(r"^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d{3})? \w+ (?:\[\w+\])?\s*(.*)")
            .unwrap();

    let mut message_parts = Vec::new();

    if let Some(captures) = message_regex.captures(first_line) {
        let first_message = captures.get(1).unwrap().as_str().trim();
        if !first_message.is_empty() {
            message_parts.push(first_message.to_string());
        }
    }

    // Add any continuation lines
    for line in lines.iter().skip(1) {
        message_parts.push(line.trim().to_string());
    }

    message_parts.join(" ")
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    #[test]
    fn test_getentries_with_single_log_entry() {
        let mut temp_file = NamedTempFile::new().unwrap();
        writeln!(
            temp_file,
            "2023-09-19 10:30:45.123 UTC [LOG] This is a single line log entry"
        )
        .unwrap();

        let temp_path = temp_file.path().to_str().unwrap();
        let entries = getentries(temp_path).unwrap();

        assert_eq!(entries.len(), 1);
        assert_eq!(entries[0].log_time, "2023-09-19 10:30:45.123 UTC");
        assert_eq!(entries[0].severity, "LOG");
        assert_eq!(entries[0].message, "This is a single line log entry");
        assert_eq!(entries[0].lines.len(), 1);
    }

    #[test]
    fn test_getentries_with_multiline_log_entry() {
        let mut temp_file = NamedTempFile::new().unwrap();
        writeln!(
            temp_file,
            "2023-09-19 10:30:45.123 UTC [ERROR] Database connection failed"
        )
        .unwrap();
        writeln!(temp_file, "    Connection timeout after 30 seconds").unwrap();
        writeln!(temp_file, "    Host: localhost:5432").unwrap();
        writeln!(
            temp_file,
            "2023-09-19 10:30:46.456 UTC [LOG] Retrying connection"
        )
        .unwrap();

        let temp_path = temp_file.path().to_str().unwrap();
        let entries = getentries(temp_path).unwrap();

        assert_eq!(entries.len(), 2);

        // First entry (multiline)
        assert_eq!(entries[0].log_time, "2023-09-19 10:30:45.123 UTC");
        assert_eq!(entries[0].severity, "ERROR");
        assert!(entries[0].message.contains("Database connection failed"));
        assert!(entries[0].message.contains("Connection timeout"));
        assert_eq!(entries[0].lines.len(), 3);

        // Second entry
        assert_eq!(entries[1].log_time, "2023-09-19 10:30:46.456 UTC");
        assert_eq!(entries[1].severity, "LOG");
        assert_eq!(entries[1].message, "Retrying connection");
        assert_eq!(entries[1].lines.len(), 1);
    }

    #[test]
    fn test_getentries_simple() {
        let mut temp_file = NamedTempFile::new().unwrap();
        writeln!(
            temp_file,
            "2025-05-02 18:25:51.151 EEST [2698052] krl@postgres STATEMENT:  select dadasdas"
        )
        .unwrap();
        writeln!(temp_file, "  dasda").unwrap();
        writeln!(temp_file, "2025-05-02 18:18:26.523 EEST [2240722] LOG:  listening on IPv4 address \"0.0.0.0\", port 5432").unwrap();
        writeln!(temp_file, "2025-05-02 18:18:26.533 EEST [2240726] LOG:  database system was shut down at 2025-05-01 18:18:26 EEST").unwrap();

        let temp_path = temp_file.path().to_str().unwrap();
        let entries = getentries(temp_path).unwrap();

        assert_eq!(entries.len(), 3);
        assert_eq!(entries[0].severity, "STATEMENT");
        assert_eq!(entries[1].severity, "LOG");
        assert_eq!(entries[2].severity, "LOG");
    }

    #[test]
    fn test_extract_severity() {
        assert_eq!(
            extract_severity("2023-09-19 10:30:45 UTC [ERROR] message", &SEVERITY_REGEX),
            "ERROR"
        );
        assert_eq!(
            extract_severity("2023-09-19 10:30:45 UTC [LOG] message", &SEVERITY_REGEX),
            "LOG"
        );
        assert_eq!(
            extract_severity("2023-09-19 10:30:45 UTC ERROR: message", &SEVERITY_REGEX),
            "ERROR"
        );
        assert_eq!(
            extract_severity("2023-09-19 10:30:45 UTC warning something", &SEVERITY_REGEX),
            "WARNING"
        );
    }

    #[test]
    fn test_extract_message() {
        let lines = vec![
            "2023-09-19 10:30:45 UTC [ERROR] Database connection failed".to_string(),
            "    Additional context line 1".to_string(),
            "    Additional context line 2".to_string(),
        ];

        let message = extract_message(&lines);
        assert!(message.contains("Database connection failed"));
        assert!(message.contains("Additional context line 1"));
        assert!(message.contains("Additional context line 2"));
    }
}
