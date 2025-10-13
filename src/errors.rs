use crate::Cli;
use crate::ConvertedArgs;
use crate::logreader;

pub fn process_errors(cli: &Cli, converted_args: &ConvertedArgs) {
    let filename = cli.filename.as_deref();
    let verbose = cli.verbose;

    if converted_args.begin.is_some() && cli.verbose {
        println!(
            "Filtering logs from begin time: {}",
            converted_args.begin.unwrap()
        ); // TODO
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
                        if line.contains("ERROR: ") {
                            println!("{}", line);
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
