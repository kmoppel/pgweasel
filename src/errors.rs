use crate::logreader;

pub fn process_errors(filename: Option<&str>, verbose: bool) {
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
