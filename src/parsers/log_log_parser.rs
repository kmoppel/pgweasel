use std::{
    fs::File,
    io::{BufRead, BufReader},
};

use chrono::{DateTime, FixedOffset, Local, Utc};

use crate::{
    convert_args::{ConvertedArgs, FileWithPath},
    parsers::{LogLine, LogParser},
};

pub struct LogLogParser;

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
    ) -> Box<dyn Iterator<Item = Result<LogLine>> + 'static> {
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

            let now_utc: DateTime<Utc> = Utc::now();
            let offset = FixedOffset::east_opt(2 * 3600).unwrap();
            Some(Ok(LogLine {
                raw: line,
                timtestamp: now_utc.with_timezone(&offset),
                severity: crate::errors::Severity::DEBUG2,
                message: "Not yet implemented".to_string(),
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
// }
