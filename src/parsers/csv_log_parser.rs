use std::{
    fs::File,
    io::{BufRead, BufReader, Cursor},
};

use chrono::{DateTime, Local};
use csv::ReaderBuilder;

use crate::{
    errors::PostgresLog,
    parsers::{LogLine, LogParser}, severity::Severity,
};

pub struct CsvLogParser;

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

impl LogParser for CsvLogParser {
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
            if let Some(some_mask) = &mask {
                if !line.starts_with(some_mask) {
                    return None;
                };
            }

            let severity = Severity::from_csv_string(&line);
            let log_level_num: i32 = (&severity).into();
            if log_level_num < min_severity {
                return None;
            }

            let cursor = Cursor::new(line.clone());
            let rdr = ReaderBuilder::new()
                .has_headers(false)
                .from_reader(cursor);

            let record = match rdr.into_records().next().unwrap() {
                Ok(r) => r,
                Err(err) => return Some(Err(format!("Failed to parse! Err: {err}").into())),
            };

            let log_record: PostgresLog = match record.deserialize(None) {
                Ok(rec) => rec,
                Err(err) => {
                    return Some(Err(
                        format!("Failed to parse to postgres log! Err: {err}").into()
                    ));
                }
            };

            if let Some(log_time) = log_record.log_time {
                let log_time_local = log_time.with_timezone(&chrono::Local);
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
            }

            Some(Ok(LogLine {
                raw: line,
                timestamp: log_record.log_time.unwrap(),
                severity: log_record.error_severity.into(),
                message: log_record.message.unwrap(),
            }))
        });
        Box::new(iter)
    }
}

// #[cfg(test)]
// mod tests {
//     use std::{fs::File, path::PathBuf};

//     use crate::{errors::Severity, parsers::csv_log_parser::CsvLogParser};

//     use super::*;

//     #[test]
//     fn test_csv_parser() -> Result<()> {
//         let path: PathBuf = PathBuf::from("./testdata/csvlog_pg14.csv");
//         let file = File::open(path.clone())?;
//         let parser = Box::new(CsvLogParser {});

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
// }
