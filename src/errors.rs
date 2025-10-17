use crate::Cli;
use crate::ConvertedArgs;
use crate::files;
use crate::logparser::get_log_records_from_line_stream;
use crate::postgres::{CsvEntry, LogEntry, VALID_SEVERITIES};
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

/// Convert CSV records to LogEntry items
fn get_csv_log_entries(
    input_files: &[String],
    verbose: bool,
) -> Box<dyn Iterator<Item = std::io::Result<LogEntry>>> {
    let mut all_entries: Vec<std::io::Result<LogEntry>> = Vec::new();

    for filename in input_files {
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

                    // Extract and build CsvEntry
                    let csv_entry = CsvEntry {
                        csv_column_count: num_fields as i32,
                        log_time: record.get(0).unwrap_or("").to_string(),
                        user_name: record.get(1).unwrap_or("").to_string(),
                        database_name: record.get(2).unwrap_or("").to_string(),
                        process_id: record.get(3).unwrap_or("").to_string(),
                        connection_from: record.get(4).unwrap_or("").to_string(),
                        session_id: record.get(5).unwrap_or("").to_string(),
                        session_line_num: record.get(6).unwrap_or("").to_string(),
                        command_tag: record.get(7).unwrap_or("").to_string(),
                        session_start_time: record.get(8).unwrap_or("").to_string(),
                        virtual_transaction_id: record.get(9).unwrap_or("").to_string(),
                        transaction_id: record.get(10).unwrap_or("").to_string(),
                        error_severity: record.get(11).unwrap_or("").to_string(),
                        sql_state_code: record.get(12).unwrap_or("").to_string(),
                        message: record.get(13).unwrap_or("").to_string(),
                        detail: record.get(14).unwrap_or("").to_string(),
                        hint: record.get(15).unwrap_or("").to_string(),
                        internal_query: record.get(16).unwrap_or("").to_string(),
                        internal_query_pos: record.get(17).unwrap_or("").to_string(),
                        context: record.get(18).unwrap_or("").to_string(),
                        query: record.get(19).unwrap_or("").to_string(),
                        query_pos: record.get(20).unwrap_or("").to_string(),
                        location: record.get(21).unwrap_or("").to_string(),
                        application_name: record.get(22).unwrap_or("").to_string(),
                        backend_type: record.get(23).unwrap_or("").to_string(),
                        leader_pid: record.get(24).unwrap_or("").to_string(),
                        query_id: record.get(25).unwrap_or("").to_string(),
                    };

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
                    let csv_line = fields.join(",");

                    // Create LogEntry from CSV record
                    let log_entry = LogEntry {
                        log_time: csv_entry.log_time.clone(),
                        error_severity: csv_entry.error_severity.clone(),
                        message: csv_entry.message.clone(),
                        lines: vec![csv_line],
                        csv_columns: Some(csv_entry),
                    };

                    all_entries.push(Ok(log_entry));
                }
                Err(e) => {
                    all_entries.push(Err(std::io::Error::new(
                        std::io::ErrorKind::InvalidData,
                        format!(
                            "Error parsing CSV record {} in {}: {}",
                            record_number + 1,
                            filename,
                            e
                        ),
                    )));
                }
            }
        }
    }

    Box::new(all_entries.into_iter())
}

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs, min_severity: &str) {
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

    let min_severity_num = log_entry_severity_to_num(min_severity);

    // Get log entries - either from CSV files or text logs
    let log_entries: Box<dyn Iterator<Item = std::io::Result<LogEntry>>> =
        if has_csv_files(&cli.input_files) {
            get_csv_log_entries(&cli.input_files, verbose)
        } else {
            match files::get_lines_from_source(&cli.input_files, verbose) {
                Ok(lines) => Box::new(get_log_records_from_line_stream(lines)),
                Err(e) => {
                    if cli.input_files.is_empty() {
                        error!("Error reading from stdin: {}", e);
                    } else {
                        error!("Error processing input files: {}", e);
                    }
                    return;
                }
            }
        };

    // Process log entries (same logic for both CSV and text logs)
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
                                debug!("Skipping log entry as after end time: {}", &entry.log_time);
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
