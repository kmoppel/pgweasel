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

    match filename {
        Some(file) => {
            if verbose {
                println!("Parsing file: {}", file);
            }

            // Use the logreader module to read the file line by line
            match logreader::getlines(file) {
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
                Err(e) => eprintln!("Error opening file '{}': {}", file, e),
            }
        }
        None => {
            if verbose {
                println!("Reading from stdin...");
            }

            // Read from stdin
            let lines = logreader::getlines_from_stdin();
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
    }
}
