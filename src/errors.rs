use crate::Cli;
use crate::ConvertedArgs;
use crate::logparser::LOG_ENTRY_START_REGEX;
use crate::logreader;
use crate::util::parse_timestamp_from_string;
use log::{debug, error};

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs) {
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

    match lines_result {
        Ok(lines) => {
            for (line_number, line_result) in lines.enumerate() {
                match line_result {
                    Ok(line) => {
                        if LOG_ENTRY_START_REGEX.is_match(&line) {
                            if let Some(caps) = LOG_ENTRY_START_REGEX.captures(&line) {
                                if &caps["log_level"] != "ERROR"
                                    && &caps["log_level"] != "FATAL"
                                    && &caps["log_level"] != "PANIC"
                                {
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
