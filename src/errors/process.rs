use std::io::BufReader;
use std::time::Instant;

use csv::ReaderBuilder;
use log::{debug, error};

use crate::convert_args::ConvertedArgs;
use crate::errors::Severity;
use crate::errors::log_record::PostgresLog;

pub fn process_errors(converted_args: &ConvertedArgs, min_severity: &Severity) {
    let min_severity_num: i32 = min_severity.into();

    for file_with_path in &converted_args.files {
        if converted_args.verbose {
            debug!(
                "Processing CSV file: {}",
                file_with_path.path.to_str().unwrap()
            );
        }

        let reader = BufReader::new(&file_with_path.file);
        let mut csv_reader = ReaderBuilder::new()
            .has_headers(false)
            .flexible(true) // Allow variable number of columns
            .from_reader(reader);

        let mut log_records: Vec<PostgresLog> = Vec::new();
        let start = Instant::now();
        for result in csv_reader.records() {
            let record = result.unwrap();
            let level: Severity = record[11].to_string().into();
            let log_level_num: i32 = (&level).into();
            if log_level_num < min_severity_num {
                continue;
            }
            if let Some(timestamp_str) = converted_args.matches.get_one::<String>("mask") {
                if !record[0].starts_with(timestamp_str) {
                    continue;
                }
            }
            let log_record: PostgresLog = match record.deserialize(None) {
                Ok(rec) => rec,
                Err(e) => {
                    error!(
                        "Error deserializing CSV record in file {}: {}",
                        file_with_path.path.to_str().unwrap(),
                        e
                    );
                    continue;
                }
            };
            if let Some(log_time) = log_record.log_time {
                let log_time_local = log_time.with_timezone(&chrono::Local);
                if let Some(begin) = converted_args.begin {
                    if log_time_local < begin {
                        continue;
                    }
                }
                if let Some(end) = converted_args.end {
                    if log_time_local > end {
                        continue;
                    }
                }
            }

            log_records.push(log_record);
        }
        debug!("Read data within: {:?}", start.elapsed());

        for record in &log_records {
            println!("{:?}", record);
        }
        debug!("Finished in: {:?}", start.elapsed());
    }
}
