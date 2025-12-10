use std::{fs::File, path::PathBuf};

use chrono::{DateTime, FixedOffset, Local};

use crate::{
    convert_args::{ConvertedArgs, FileWithPath},
    errors::Severity,
};

mod csv_log_parser;
mod log_log_parser;

pub use csv_log_parser::CsvLogParser;
pub use log_log_parser::LogLogParser;

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
    fn parse(
        &mut self,
        file: File,
        min_severity: i32,
        mask: Option<String>,
        begin: Option<DateTime<Local>>,
        end: Option<DateTime<Local>>,
    ) -> Box<dyn Iterator<Item = Result<LogLine>>>;
}

pub fn get_parser(path: PathBuf) -> Result<Box<dyn LogParser>> {
    match path.extension() {
        Some(ext) if ext == "csv" => Ok(Box::new(CsvLogParser {})),
        Some(ext) if ext == "log" => Ok(Box::new(LogLogParser {})),
        Some(ext) => Err(format!("File extension: {:?} not supported", ext).into()),
        None => Err("No Extension".into()),
    }
}
