use crate::Cli;
use crate::ConvertedArgs;
use crate::files;
use crate::logparser::get_log_records_from_line_stream;
use crate::postgres::VALID_SEVERITIES;
use crate::util::parse_timestamp_from_string;
use csv::ReaderBuilder;
use flate2::read::GzDecoder;
use log::{debug, error};
use std::fs::File;
use std::io::{BufReader, Read};

/// Validate that a severity level string is valid
pub fn validate_severity(severity: &str) -> Result<(), String> {
    VALID_SEVERITIES
        .contains(&severity.to_uppercase().as_str())
        .then_some(())
        .ok_or_else(|| {
            format!(
                "Invalid severity level: '{}'. Valid values are: {}",
                severity,
                VALID_SEVERITIES.join(", ")
            )
        })
}

/// Convert PostgreSQL log severity level to a numeric priority
fn log_entry_severity_to_num(severity: &str) -> i32 {
    match severity.to_uppercase().as_str() {
        "DEBUG5" => 0,
        "DEBUG4" => 1,
        "DEBUG3" => 2,
        "DEBUG2" => 3,
        "DEBUG1" => 4,
        "LOG" => 5,
        "INFO" => 5,
        "NOTICE" => 6,
        "WARNING" => 7,
        "ERROR" => 8,
        "FATAL" => 9,
        "PANIC" => 10,
        _ => 5, // Default to LOG level
    }
}

/// Check if any input file has a CSV extension
fn has_csv_files(input_files: &[String]) -> bool {
    input_files
        .iter()
        .any(|f| f.ends_with(".csv") || f.ends_with(".csv.gz"))
}

/// Process CSV format log entries
fn process_csv_errors(cli: &Cli, converted_args: &ConvertedArgs, min_severity: &str) {
    let verbose = cli.verbose;
    let min_severity_num = log_entry_severity_to_num(min_severity);

    if converted_args.begin.is_some() && verbose {
        debug!(
            "Filtering logs from begin time: {}",
            converted_args.begin.unwrap()
        );
    }

    if converted_args.end.is_some() && verbose {
        debug!(
            "Filtering logs until end time: {}",
            converted_args.end.unwrap()
        );
    }

    // Process each input file separately as CSV
    for filename in &cli.input_files {
        if verbose {
            debug!("Processing CSV file: {}", filename);
        }

        // Create a reader for the file
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

        // Create CSV reader
        let mut csv_reader = ReaderBuilder::new()
            .has_headers(false)
            .flexible(true) // Allow variable number of columns
            .from_reader(BufReader::new(reader));

        // Process CSV records
        for (record_number, result) in csv_reader.records().enumerate() {
            match result {
                Ok(record) => {
                    let num_fields = record.len();

                    // Validate column count (23, 24, or 26 columns)
                    if num_fields != 23 && num_fields != 24 && num_fields != 26 {
                        if verbose {
                            debug!(
                                "Skipping CSV record {} with unexpected column count: {} (expected 23, 24, or 26)",
                                record_number + 1,
                                num_fields
                            );
                        }
                        continue;
                    }

                    // Extract fields (all as strings)
                    // Column indices based on the provided field list:
                    // 0: LogTime, 11: ErrorSeverity
                    let log_time = record.get(0).unwrap_or("");
                    let error_severity = record.get(11).unwrap_or("");

                    // Filter by severity level
                    let log_level_num = log_entry_severity_to_num(error_severity);
                    if log_level_num < min_severity_num {
                        continue;
                    }

                    // Filter by begin time
                    if cli.begin.is_some() {
                        if let Ok(tz) = parse_timestamp_from_string(log_time) {
                            if tz < converted_args.begin.unwrap() {
                                if verbose {
                                    debug!(
                                        "Skipping CSV record as before begin time: {}",
                                        log_time
                                    );
                                }
                                continue;
                            }
                        }
                    }

                    // Filter by end time
                    if cli.end.is_some() {
                        if let Ok(tz) = parse_timestamp_from_string(log_time) {
                            if tz > converted_args.end.unwrap() {
                                if verbose {
                                    debug!("Skipping CSV record as after end time: {}", log_time);
                                }
                                continue;
                            }
                        }
                    }

                    // Print the CSV record
                    // Reconstruct the CSV line with proper quoting
                    let fields: Vec<String> = record
                        .iter()
                        .map(|field| {
                            // Quote field if it contains comma, newline, or quote
                            if field.contains(',') || field.contains('\n') || field.contains('"') {
                                format!("\"{}\"", field.replace('"', "\"\""))
                            } else {
                                field.to_string()
                            }
                        })
                        .collect();
                    println!("{}", fields.join(","));
                }
                Err(e) => {
                    error!(
                        "Error parsing CSV record {} in {}: {}",
                        record_number + 1,
                        filename,
                        e
                    );
                }
            }
        }
    }
}

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs, min_severity: &str) {
    // Check if we're processing CSV files
    if has_csv_files(&cli.input_files) {
        process_csv_errors(cli, converted_args, min_severity);
        return;
    }

    // Text log processing using log record iterator
    let verbose = cli.verbose;

    if converted_args.begin.is_some() && verbose {
        debug!(
            "Filtering logs from begin time: {}",
            converted_args.begin.unwrap()
        );
    }

    if converted_args.end.is_some() && verbose {
        debug!(
            "Filtering logs until end time: {}",
            converted_args.end.unwrap()
        );
    }

    let lines_result = files::get_lines_from_source(&cli.input_files, verbose);

    let min_severity_num = log_entry_severity_to_num(min_severity);

    match lines_result {
        Ok(lines) => {
            // Use the new iterator to get structured log entries
            let log_entries = get_log_records_from_line_stream(lines);

            for entry_result in log_entries {
                match entry_result {
                    Ok(entry) => {
                        // Filter by severity level
                        let log_level_num = log_entry_severity_to_num(&entry.error_severity);

                        if log_level_num < min_severity_num {
                            continue;
                        }

                        // Filter by begin time
                        if cli.begin.is_some() {
                            if let Ok(tz) = parse_timestamp_from_string(&entry.log_time) {
                                if tz < converted_args.begin.unwrap() {
                                    if verbose {
                                        debug!(
                                            "Skipping log entry as before begin time: {}",
                                            &entry.log_time
                                        );
                                    }
                                    continue;
                                }
                            }
                        }

                        // Filter by end time
                        if cli.end.is_some() {
                            if let Ok(tz) = parse_timestamp_from_string(&entry.log_time) {
                                if tz > converted_args.end.unwrap() {
                                    if verbose {
                                        debug!(
                                            "Skipping log entry as after end time: {}",
                                            &entry.log_time
                                        );
                                    }
                                    continue;
                                }
                            }
                        }

                        // Print all lines of the log entry
                        for line in &entry.lines {
                            println!("{}", line);
                        }
                    }
                    Err(e) => error!("Error reading log entry: {}", e),
                }
            }
        }
        Err(e) => {
            if cli.input_files.is_empty() {
                error!("Error reading from stdin: {}", e);
            } else {
                error!("Error processing input files: {}", e);
            }
        }
    }
}
