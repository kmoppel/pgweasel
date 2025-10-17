use crate::postgres::LogEntry;
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

/// Iterator that converts a stream of log lines into structured LogEntry items
pub struct LogRecordIterator<I>
where
    I: Iterator<Item = Result<String>>,
{
    lines: I,
    current_entry: Option<LogEntry>,
}

impl<I> LogRecordIterator<I>
where
    I: Iterator<Item = Result<String>>,
{
    fn new(lines: I) -> Self {
        LogRecordIterator {
            lines,
            current_entry: None,
        }
    }
}

impl<I> Iterator for LogRecordIterator<I>
where
    I: Iterator<Item = Result<String>>,
{
    type Item = Result<LogEntry>;

    fn next(&mut self) -> Option<Self::Item> {
        loop {
            match self.lines.next() {
                Some(Ok(line)) => {
                    // Check if this line starts a new log entry
                    if let Some(caps) = LOG_ENTRY_START_REGEX.captures(&line) {
                        // Extract timestamp, severity, and message
                        let log_time = caps["time"].to_string();
                        let error_severity = caps["log_level"].to_string();
                        
                        // Extract the message part (everything after the severity and ":  ")
                        let message = if let Some(pos) = line.find(&format!("{}:  ", error_severity)) {
                            line[pos + error_severity.len() + 3..].to_string()
                        } else {
                            String::new()
                        };

                        // If we already have a current entry being built, return it
                        let completed_entry = self.current_entry.take();

                        // Start a new entry
                        self.current_entry = Some(LogEntry {
                            log_time,
                            error_severity,
                            message,
                            lines: vec![line],
                            csv_columns: None,
                        });

                        // Return the completed entry if we had one
                        if let Some(entry) = completed_entry {
                            return Some(Ok(entry));
                        }
                        // Otherwise continue to accumulate more lines
                    } else {
                        // This is a continuation line - add it to the current entry
                        if let Some(ref mut entry) = self.current_entry {
                            entry.lines.push(line);
                        }
                        // If we don't have a current entry, skip this line (orphaned continuation)
                    }
                }
                Some(Err(e)) => {
                    // Error reading line - return the error
                    return Some(Err(e));
                }
                None => {
                    // End of input - return any remaining entry
                    return self.current_entry.take().map(Ok);
                }
            }
        }
    }
}

/// Get an iterator of LogEntry items from a stream of log lines
///
/// This function takes an iterator of log lines (as returned by files::get_lines_from_source)
/// and returns an iterator of LogEntry items. Each LogEntry represents a complete log record,
/// which may span multiple lines.
///
/// # Arguments
///
/// * `lines` - An iterator that yields Result<String> representing individual log lines
///
/// # Returns
///
/// An iterator that yields Result<LogEntry> representing complete log records
///
/// # Example
///
/// ```no_run
/// use pgweasel::files;
/// use pgweasel::logparser::get_log_records_from_line_stream;
///
/// let lines = files::get_lines_from_source(&vec!["logfile.log".to_string()], false)?;
/// let log_entries = get_log_records_from_line_stream(lines);
///
/// for entry_result in log_entries {
///     match entry_result {
///         Ok(entry) => println!("Log: {} - {}", entry.error_severity, entry.message),
///         Err(e) => eprintln!("Error: {}", e),
///     }
/// }
/// ```
pub fn get_log_records_from_line_stream(
    lines: Box<dyn Iterator<Item = Result<String>>>,
) -> impl Iterator<Item = Result<LogEntry>> {
    LogRecordIterator::new(lines)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_log_records_from_line_stream_single_line_entries() {
        // Create test input with single-line log entries
        let lines = vec![
            Ok("2024-01-15 10:30:45 UTC LOG:  database system is ready to accept connections".to_string()),
            Ok("2024-01-15 10:30:46 UTC ERROR:  relation \"users\" does not exist".to_string()),
            Ok("2024-01-15 10:30:47 UTC WARNING:  could not open statistics file".to_string()),
        ];
        
        let iter: Box<dyn Iterator<Item = Result<String>>> = Box::new(lines.into_iter());
        let log_entries: Vec<_> = get_log_records_from_line_stream(iter)
            .collect::<std::io::Result<Vec<_>>>()
            .unwrap();

        assert_eq!(log_entries.len(), 3);
        assert_eq!(log_entries[0].error_severity, "LOG");
        assert_eq!(log_entries[0].message, "database system is ready to accept connections");
        assert_eq!(log_entries[0].lines.len(), 1);
        
        assert_eq!(log_entries[1].error_severity, "ERROR");
        assert_eq!(log_entries[1].message, "relation \"users\" does not exist");
        
        assert_eq!(log_entries[2].error_severity, "WARNING");
        assert_eq!(log_entries[2].message, "could not open statistics file");
    }

    #[test]
    fn test_get_log_records_from_line_stream_multi_line_entries() {
        // Create test input with multi-line log entries
        let lines = vec![
            Ok("2024-01-15 10:30:45 UTC ERROR:  syntax error at or near \"SELCT\"".to_string()),
            Ok("DETAIL:  Invalid SQL syntax".to_string()),
            Ok("HINT:  Did you mean SELECT?".to_string()),
            Ok("2024-01-15 10:30:46 UTC LOG:  checkpoint complete".to_string()),
            Ok("2024-01-15 10:30:47 UTC FATAL:  connection terminated".to_string()),
            Ok("CONTEXT:  while processing query".to_string()),
        ];
        
        let iter: Box<dyn Iterator<Item = Result<String>>> = Box::new(lines.into_iter());
        let log_entries: Vec<_> = get_log_records_from_line_stream(iter)
            .collect::<std::io::Result<Vec<_>>>()
            .unwrap();

        assert_eq!(log_entries.len(), 3);
        
        // First entry with 3 lines
        assert_eq!(log_entries[0].error_severity, "ERROR");
        assert_eq!(log_entries[0].message, "syntax error at or near \"SELCT\"");
        assert_eq!(log_entries[0].lines.len(), 3);
        assert!(log_entries[0].lines[1].contains("DETAIL:"));
        assert!(log_entries[0].lines[2].contains("HINT:"));
        
        // Second entry with 1 line
        assert_eq!(log_entries[1].error_severity, "LOG");
        assert_eq!(log_entries[1].lines.len(), 1);
        
        // Third entry with 2 lines
        assert_eq!(log_entries[2].error_severity, "FATAL");
        assert_eq!(log_entries[2].lines.len(), 2);
        assert!(log_entries[2].lines[1].contains("CONTEXT:"));
    }

    #[test]
    fn test_get_log_records_from_line_stream_empty_input() {
        let lines: Vec<Result<String>> = vec![];
        let iter: Box<dyn Iterator<Item = Result<String>>> = Box::new(lines.into_iter());
        let log_entries: Vec<_> = get_log_records_from_line_stream(iter)
            .collect::<std::io::Result<Vec<_>>>()
            .unwrap();

        assert_eq!(log_entries.len(), 0);
    }

    #[test]
    fn test_get_log_records_from_line_stream_with_orphaned_lines() {
        // Test with continuation lines that have no preceding log entry start
        let lines = vec![
            Ok("DETAIL:  Orphaned line".to_string()),
            Ok("2024-01-15 10:30:45 UTC LOG:  valid log entry".to_string()),
            Ok("CONTEXT:  continuation of valid entry".to_string()),
        ];
        
        let iter: Box<dyn Iterator<Item = Result<String>>> = Box::new(lines.into_iter());
        let log_entries: Vec<_> = get_log_records_from_line_stream(iter)
            .collect::<std::io::Result<Vec<_>>>()
            .unwrap();

        // Should have 1 valid entry (orphaned line is skipped)
        assert_eq!(log_entries.len(), 1);
        assert_eq!(log_entries[0].error_severity, "LOG");
        assert_eq!(log_entries[0].lines.len(), 2);
    }
}
