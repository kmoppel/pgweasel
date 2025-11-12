use std::{fs::File, io::Read};

use csv::ReaderBuilder;
use flate2::read::GzDecoder;
use log::{debug, error};

use crate::{Cli, ConvertedArgs, errors::{csv_reader::log_record::PostgresLog, log_entry_severity_to_num}};

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs, min_severity: &str) {
    let verbose = cli.verbose;
    let min_severity_num = log_entry_severity_to_num(min_severity);

    for filename in &cli.input_files {
        if verbose {
            debug!("Processing CSV file: {}", filename);
        }

        let reader: Box<dyn Read> = if filename.ends_with(".gz") {
            match File::open(filename) {
                Ok(file) => Box::new(GzDecoder::new(file)),
                Err(e) => {
                    error!("Error opening file {}: {}", filename, e);
                    continue;
                }
            }
        } else {
            match File::open(filename) {
                Ok(file) => Box::new(file),
                Err(e) => {
                    error!("Error opening file {}: {}", filename, e);
                    continue;
                }
            }
        };
        let mut csv_reader = ReaderBuilder::new()
            .has_headers(false)
            .flexible(true) // Allow variable number of columns
            .from_reader(reader);

        let mut log_records: Vec<PostgresLog> = Vec::new();
        for result in csv_reader.deserialize::<PostgresLog>() {
            match result {
                Ok(rec) => {
                    let log_level_num = log_entry_severity_to_num(&rec.error_severity);
                    if log_level_num < min_severity_num {
                        continue;
                    }
                    if let Some(log_time) = rec.log_time {
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
                    log_records.push(rec);
                },
                Err(e) => {
                    error!("Error reading CSV record in file {}: {}", filename, e);
                }
            };
        }
        for record in &log_records {
            println!("{:?}", record);
        }
    }
}
