use std::{
    fs::File,
    io::{BufRead, BufReader},
    iter, mem,
};

use chrono::{DateTime, Local};

use crate::{
    parsers::{LogLine, LogParser, date_serializer::deserialize_helper},
    severity::Severity,
    util::line_has_timestamp_prefix,
};

#[derive(Default)]
pub struct LogLogParser {
    pub current: String,
}

use crate::Result;

impl LogParser for LogLogParser {
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

                // New timestamp → flush previous record
                let prev = mem::replace(&mut self.current, line);
                if prev.is_empty() {
                    continue;
                }

                match process_simple_log_record(prev, min_severity, &mask, begin, end) {
                    Some(Ok(log_line)) => return Some(Ok(log_line)),
                    Some(Err(e)) => return Some(Err(e)),
                    None => continue,
                };
            }

            // EOF: flush last buffered record
            if !self.current.is_empty() {
                let last = mem::take(&mut self.current);
                return process_simple_log_record(last, min_severity, &mask, begin, end);
            }

            None
        }))
    }
}

fn process_simple_log_record(
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

    let severity = Severity::from_log_string(&line);
    let level: i32 = (&severity).into();
    if level < min_severity {
        return None;
    }

    let mut parts = line.split_whitespace();
    let ts_str = format!("{} {} {}", parts.next()?, parts.next()?, parts.next()?);

    let timestamp = match deserialize_helper(&ts_str) {
        Ok(ts) => ts,
        Err(e) => return Some(Err(e)),
    };

    let log_time_local = timestamp.with_timezone(&Local);
    if begin.is_some_and(|b| log_time_local < b) {
        return None;
    }
    if end.is_some_and(|e| log_time_local > e) {
        return None;
    }

    Some(Ok(LogLine {
        raw: line.clone(),
        timestamp,
        severity,
        message: line,
    }))
}

#[cfg(test)]
mod tests {
    use crate::util::line_has_timestamp_prefix;

    #[test]
    fn test_timestamp_finder() {
        let good = "2025-08-27 01:24:43.415 EEST [3863330] LOG: something";
        let bad = "ERROR: something";
        let almost = "2025/08/27 01:24:43 LOG";

        assert_eq!(
            line_has_timestamp_prefix(good),
            (true, "2025-08-27 01:24:43.415 EEST".to_string().into())
        );
        assert_eq!(line_has_timestamp_prefix(bad), (false, None));
        assert_eq!(line_has_timestamp_prefix(almost), (false, None));
    }
}
