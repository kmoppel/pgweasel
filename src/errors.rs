use crate::Cli;
use crate::ConvertedArgs;
use crate::logparser::LOG_ENTRY_START_REGEX;
use crate::logreader;
use crate::postgres::VALID_SEVERITIES;
use crate::util::parse_timestamp_from_string;
use log::{debug, error};

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

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs, min_severity: &str) {
    let filename = cli.filename.as_deref();
    let verbose = cli.verbose;

    if converted_args.begin.is_some() && cli.verbose {
        debug!(
            "Filtering logs from begin time: {}",
            converted_args.begin.unwrap()
        );
    }

    if converted_args.end.is_some() && cli.verbose {
        debug!(
            "Filtering logs until end time: {}",
            converted_args.end.unwrap()
        );
    }

    let lines_result = match filename {
        Some(file) => {
            if verbose {
                debug!("Parsing file: {}", file);
            }
            logreader::getlines(file)
        }
        None => {
            if verbose {
                debug!("Reading from stdin...");
            }
            logreader::getlines_from_stdin()
        }
    };
    
    let min_severity_num = log_entry_severity_to_num(min_severity);
    
    match lines_result {
        Ok(lines) => {
            for (line_number, line_result) in lines.enumerate() {
                match line_result {
                    Ok(line) => {
                        if LOG_ENTRY_START_REGEX.is_match(&line) {
                            if let Some(caps) = LOG_ENTRY_START_REGEX.captures(&line) {
                                // Filter by severity level
                                let log_level = &caps["log_level"];
                                let log_level_num = log_entry_severity_to_num(log_level);

                                if log_level_num < min_severity_num {
                                    continue;
                                }

                                if cli.begin.is_some() {
                                    if let Ok(tz) = parse_timestamp_from_string(&caps["time"]) {
                                        if tz < converted_args.begin.unwrap() {
                                            if verbose {
                                                debug!(
                                                    "Skipping log line as before begin time: {}",
                                                    &caps["time"]
                                                );
                                            }
                                            continue;
                                        }
                                    }
                                }
                                if cli.end.is_some() {
                                    if let Ok(tz) = parse_timestamp_from_string(&caps["time"]) {
                                        if tz > converted_args.end.unwrap() {
                                            if verbose {
                                                debug!(
                                                    "Skipping log line as after end time: {}",
                                                    &caps["time"]
                                                );
                                            }
                                            continue;
                                        }
                                    }
                                }
                            }
                            println!("{}", line); // TODO is println sync / flushed i.e. slow ?
                        }
                    }
                    Err(e) => error!("Error reading line {}: {}", line_number + 1, e),
                }
            }
        }
        Err(e) => {
            if let Some(file) = filename {
                error!("Error opening file '{}': {}", file, e);
            } else {
                error!("Error reading from stdin: {}", e);
            }
        }
    }
}
