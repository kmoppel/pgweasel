use crate::Cli;
use crate::ConvertedArgs;
use crate::logparser::LOG_ENTRY_START_REGEX;
use crate::logreader;
use crate::util::parse_timestamp_from_string;

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs) {
    let filename = cli.filename.as_deref();
    let verbose = cli.verbose;

    if converted_args.begin.is_some() && cli.verbose {
        println!(
            "Filtering logs from begin time: {}",
            converted_args.begin.unwrap()
        );
    }

    let lines_result = match filename {
        Some(file) => {
            if verbose {
                println!("Parsing file: {}", file);
            }
            logreader::getlines(file)
        }
        None => {
            if verbose {
                println!("Reading from stdin...");
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
                                                println!(
                                                    "Skipping log line as before begin time: {}",
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
                    Err(e) => eprintln!("Error reading line {}: {}", line_number + 1, e),
                }
            }
        }
        Err(e) => {
            if let Some(file) = filename {
                eprintln!("Error opening file '{}': {}", file, e);
            } else {
                eprintln!("Error reading from stdin: {}", e);
            }
        }
    }
}
