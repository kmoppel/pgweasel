use std::{
    fs::File,
    io::{BufRead, BufReader},
};

use chrono::{DateTime, Local};

use crate::{
    convert_args::FileWithPath,
    parsers::{LogLine, LogParser},
};

pub struct LogLogParser {
    reader: BufReader<File>,
}

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

impl LogLogParser {
    pub fn new(file_with_path: FileWithPath) -> Self {
        Self {
            reader: BufReader::new(file_with_path.file),
        }
    }
}

impl LogParser for LogLogParser {
    type Iter = Box<dyn Iterator<Item = Result<LogLine>>>;

    fn parse(
        mut self,
        min_severity: i32,
        mask: Option<String>,
        begin: Option<DateTime<Local>>,
        end: Option<DateTime<Local>>,
    ) -> Self::Iter {
        let iter = self.reader.lines().filter_map(move |lin| {
            let line = match lin {
                Ok(l) => l,
                Err(err) => return Some(Err(format!("Failed to read! Err: {err}").into())),
            };
            if let Some(some_mask) = &mask {
                if !line.starts_with(some_mask) {
                    return None;
                };
            }
            Some(Ok(LogLine { raw: line, timtestamp: todo!(), severity: todo!(), message: todo!() }))
        });
        Box::new(iter)
    }
}

#[cfg(test)]
mod tests {
    use std::{fs::File, path::PathBuf};

    use crate::{errors::Severity, parsers::csv_log_parser::CsvLogParser};

    use super::*;

    #[test]
    fn test_csv_parser() -> Result<()> {
        let path: PathBuf = PathBuf::from("./testdata/csvlog_pg14.csv");
        let file = File::open(path.clone())?;
        let parser = CsvLogParser::new(FileWithPath { file, path });

        let intseverity = (&(Severity::LOG)).into();
        let iter = parser.parse(
            intseverity,
            Some("2025-05-21 13:00:03.127".to_string()),
            None,
            None,
        );
        for line in iter {
            let line = line?;
            println!("{:?}", line);
        }

        Ok(())
    }
}
