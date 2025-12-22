use std::{
    fs::File,
    io::{BufRead, BufReader, Cursor},
    iter, mem,
};

use chrono::{DateTime, Local};
use csv::ReaderBuilder;

use crate::{
    parsers::{LogLine, LogParser},
    severity::Severity,
    util::{TimeParseError, line_has_timestamp_prefix, parse_timestamp_from_string},
};

#[derive(Default)]
pub struct CsvLogParser {
    pub current: String,
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
        let mut lines = BufReader::new(file).lines();

        Box::new(iter::from_fn(move || {
            while let Some(lin) = lines.next() {
                let line = match lin {
                    Ok(l) => l,
                    Err(err) => {
                        return Some(Err(crate::Error::FailedToRead { error: err }));
                    }
                };

                let (has_ts, _) = line_has_timestamp_prefix(&line);

                if !has_ts {
                    self.current.push_str(&line);
                    self.current.push('\n');
                    continue;
                }

                // Flush previous record
                let prev = mem::replace(&mut self.current, line);
                if prev.is_empty() {
                    continue;
                }

                match process_record(prev, min_severity, &mask, begin, end) {
                    Some(Ok(log_line)) => return Some(Ok(log_line)),
                    Some(Err(e)) => return Some(Err(e)),
                    None => continue,
                };
            }

            // EOF: flush last buffered record
            if !self.current.is_empty() {
                let last = mem::take(&mut self.current);
                return process_record(last, min_severity, &mask, begin, end);
            }

            None
        }))
    }
}

fn process_record(
    line: String,
    min_severity: i32,
    mask: &Option<String>,
    begin: Option<DateTime<Local>>,
    end: Option<DateTime<Local>>,
) -> Option<Result<LogLine>> {
    if let Some(mask) = mask {
        if !line.starts_with(mask) {
            return None;
        }
    }

    let severity = Severity::from_csv_string(&line);
    let level: i32 = (&severity).into();
    if level < min_severity {
        return None;
    }

    let cursor = Cursor::new(&line);
    let mut rdr = ReaderBuilder::new().has_headers(false).from_reader(cursor);

    let record = match rdr.records().next()? {
        Ok(r) => r,
        Err(e) => return Some(Err(crate::Error::FailedToParseCsv { error: e })),
    };

    let ts = match parse_timestamp_from_string(record.get(0)?) {
        Ok(t) => t,
        Err(e) => return Some(Err(TimeParseError::ParseError(e).into())),
    };

    if begin.is_some_and(|b| ts < b) || end.is_some_and(|e| ts > e) {
        return None;
    }

    Some(Ok(LogLine {
        raw: line,
        timestamp: ts.into(),
        severity,
        message: record.get(13).unwrap_or("").to_string(),
    }))
}
