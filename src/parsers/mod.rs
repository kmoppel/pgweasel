use chrono::{DateTime, FixedOffset, Local};

use crate::errors::{PostgresLog, Severity};

mod csv_log_parser;
mod log_log_parser;

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

#[derive(Debug)]
pub struct LogLine {
    pub timtestamp: DateTime<FixedOffset>,
    pub severity: Severity,
    pub message: String,
    pub raw: String,
}

/// Trait for all parsers: produce an iterator over filtered log lines.
pub trait LogParser {
    type Iter: Iterator<Item = Result<LogLine>>;

    fn parse(self, min_severity: i32, mask: Option<String>, begin: Option<DateTime<Local>>, end: Option<DateTime<Local>>) -> Self::Iter;
}
