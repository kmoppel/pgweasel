use std::{
    fs::File,
    io::{BufRead, BufReader, Cursor},
};

use chrono::{DateTime, Local};
use csv::ReaderBuilder;

use crate::{
    parsers::{LogLine, LogParser},
    severity::Severity,
    util::{line_has_timestamp_prefix, parse_timestamp_from_string},
};

#[derive(Default)]
pub struct CsvLogParser {
    pub remaining_string: String,
}

use crate::Result;

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
                Err(err) => return Some(Err(crate::Error::FailedToRead { error: err })),
            };
            self.remaining_string.push_str(&line);
            self.remaining_string.push('\n');
            let (has, _) = line_has_timestamp_prefix(&line);
            if !has {
                return None;
            }

            let result_line = self.remaining_string.clone();
            self.remaining_string = String::new();

            if let Some(some_mask) = &mask {
                if !result_line.starts_with(some_mask) {
                    return None;
                };
            }

            let severity = Severity::from_csv_string(&result_line);
            let log_level_num: i32 = (&severity).into();
            if log_level_num < min_severity {
                return None;
            }

            let cursor = Cursor::new(result_line.clone());
            let rdr = ReaderBuilder::new().has_headers(false).from_reader(cursor);

            let record = match rdr.into_records().next().unwrap() {
                Ok(r) => r,
                Err(err) => return Some(Err(crate::Error::FailedToParseCsv { error: err })),
            };

            let log_time_local = parse_timestamp_from_string(record.get(0).unwrap()).unwrap();
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
                timestamp: log_time_local.into(),
                severity,
                message: record.get(13).unwrap().to_string(),
            }))
        });
        Box::new(iter)
    }
}
