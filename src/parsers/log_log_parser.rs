use std::{
    fs::File,
    io::{BufRead, BufReader},
};

use chrono::{DateTime, Local};

use crate::{
    errors::deserialize_helper,
    parsers::{LogLine, LogParser}, severity::Severity,
};

pub struct LogLogParser {
    pub remaining_string: String,
}

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

impl LogParser for LogLogParser {
    fn parse(
        &mut self,
        file: File,
        min_severity: i32,
        mask: Option<String>,
        begin: Option<DateTime<Local>>,
        end: Option<DateTime<Local>>,
    ) -> Box<dyn Iterator<Item = Result<LogLine>> + '_> {
        let reader = BufReader::new(file);
        let iter = reader.lines().filter_map(move |lin| {
            let line = match lin {
                Ok(l) => l,
                Err(err) => return Some(Err(format!("Failed to read! Err: {err}").into())),
            };

            self.remaining_string.push_str(&line);
            if !is_pg_timestamp_start(&line) {
                return None;
            }

            let result_line = self.remaining_string.clone();
            self.remaining_string = line;

            if let Some(some_mask) = &mask {
                if !result_line.starts_with(some_mask) {
                    return None;
                };
            }

            let severity = Severity::from_log_string(&result_line);
            let log_level_num: i32 = (&severity).into();
            if log_level_num < min_severity {
                return None;
            }

            let mut parts = result_line.split_whitespace();
            // TODO: Handle unwraps
            let timestamp = deserialize_helper(&format!(
                "{} {} {}",
                parts.next().unwrap(),
                parts.next().unwrap(),
                parts.next().unwrap()
            ))
            .unwrap();

            let log_time_local = timestamp.with_timezone(&chrono::Local);
            if let Some(begin) = begin {
                if log_time_local < begin {
                    return None;
                }
            }
            if let Some(end) = end {
                if log_time_local > end {
                    return None;
                }
            }

            Some(Ok(LogLine {
                raw: result_line,
                timestamp,
                severity,
                message: self.remaining_string.clone(),
            }))
        });
        Box::new(iter)
    }
}

pub fn is_pg_timestamp_start(line: &str) -> bool {
    // Minimum length: "YYYY-MM-DD HH:MM:SS" → 19 chars
    if line.len() < 19 {
        return false;
    }

    let bytes = line.as_bytes();

    // Check fixed characters
    if bytes[4] != b'-'
        || bytes[7] != b'-'
        || bytes[10] != b' '
        || bytes[13] != b':'
        || bytes[16] != b':'
    {
        return false;
    }

    // Check digits in required places
    fn is_digit(b: u8) -> bool {
        b.is_ascii_digit()
    }

    for &idx in &[
        0, 1, 2, 3, // YYYY
        5, 6, // MM
        8, 9, // DD
        11, 12, // HH
        14, 15, // MM
        17, 18,
    ]
    // SS
    {
        if !is_digit(bytes[idx]) {
            return false;
        }
    }

    true
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_timestamp_finder() {
        let good = "2025-08-27 01:24:43.415 EEST [3863330] LOG: something";
        let bad = "ERROR: something";
        let almost = "2025/08/27 01:24:43 LOG";

        assert!(is_pg_timestamp_start(good));
        assert!(!is_pg_timestamp_start(bad));
        assert!(!is_pg_timestamp_start(almost));
    }
    //     use std::{fs::File, path::PathBuf};

    //     use crate::{errors::Severity, parsers::csv_log_parser::CsvLogParser};

    //     #[test]
    //     fn test_csv_parser() -> Result<()> {
    //         let path: PathBuf = PathBuf::from("./testdata/csvlog_pg14.csv");
    //         let file = File::open(path.clone())?;
    //         let parser = CsvLogParser::new(FileWithPath { file, path });

    //         let intseverity = (&(Severity::LOG)).into();
    //         let iter = parser.parse(
    //             intseverity,
    //             Some("2025-05-21 13:00:03.127".to_string()),
    //             None,
    //             None,
    //         );
    //         for line in iter {
    //             let line = line?;
    //             println!("{:?}", line);
    //         }

    //         Ok(())
    //     }
}
