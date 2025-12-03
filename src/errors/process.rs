use std::time::Instant;
use std::{fs::File, io::Read};

use csv::ReaderBuilder;
use flate2::read::GzDecoder;
use log::{debug, error};

use crate::convert_args::ConvertedArgs;
use crate::errors::log_record::PostgresLog;
use crate::errors::severities::log_entry_severity_to_num;

pub fn process_errors(converted_args: &ConvertedArgs, min_severity: &str) {
    let verbose = converted_args.cli.verbose;
    let min_severity_num = log_entry_severity_to_num(min_severity);

    for filename in &converted_args.file_list {
        if verbose {
            debug!("Processing CSV file: {}", filename.to_str().unwrap());
        }

        let reader: Box<dyn Read> = if filename.ends_with(".gz") {
            match File::open(filename) {
                Ok(file) => Box::new(GzDecoder::new(file)),
                Err(e) => {
                    error!("Error opening file {}: {}", filename.to_str().unwrap(), e);
                    continue;
                }
            }
        } else {
            match File::open(filename) {
                Ok(file) => Box::new(file),
                Err(e) => {
                    error!("Error opening file {}: {}", filename.to_str().unwrap(), e);
                    continue;
                }
            }
        };
        let mut csv_reader = ReaderBuilder::new()
            .has_headers(false)
            .flexible(true) // Allow variable number of columns
            .from_reader(reader);

        let mut log_records: Vec<PostgresLog> = Vec::new();
        let start = Instant::now();
        for result in csv_reader.records() {
            let record = result.unwrap();
            let log_level_num = log_entry_severity_to_num(&record[11]);
            if log_level_num < min_severity_num {
                continue;
            }
            if let Some(timestamp_str) = &converted_args.cli.timestamp_mask {
                if !record[0].starts_with(timestamp_str) {
                    continue;
                }
            }
            let log_record: PostgresLog = match record.deserialize(None) {
                Ok(rec) => rec,
                Err(e) => {
                    error!("Error deserializing CSV record in file {}: {}", filename.to_str().unwrap(), e);
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
