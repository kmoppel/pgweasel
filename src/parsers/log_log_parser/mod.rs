use std::{
    fs::File,
    io::{BufRead, BufReader},
};

use chrono::{DateTime, Local};

use crate::{
    parsers::{LogLine, LogParser, date_serializer::deserialize_helper}, severity::Severity, util::line_has_timestamp_prefix,
};

#[derive(Default)]
pub struct LogLogParser {
    pub remaining_string: String,
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
        let reader = BufReader::new(file);
        let iter = reader.lines().filter_map(move |lin| {
            let line = match lin {
                Ok(l) => l,
                Err(err) => return Some(Err(crate::Error::FailedToRead { error: err })),
            };

            self.remaining_string.push_str(&line);
            let (has, _) = line_has_timestamp_prefix(&line);
            if !has {
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

#[cfg(test)]
mod tests {
    use crate::util::line_has_timestamp_prefix;

    #[test]
    fn test_timestamp_finder() {
        let good = "2025-08-27 01:24:43.415 EEST [3863330] LOG: something";
        let bad = "ERROR: something";
        let almost = "2025/08/27 01:24:43 LOG";

        assert_eq!(line_has_timestamp_prefix(good), (true, "2025-08-27 01:24:43.415 EEST".to_string().into()));
        assert_eq!(line_has_timestamp_prefix(bad), (false, None));
        assert_eq!(line_has_timestamp_prefix(almost), (false, None));
    }
}
