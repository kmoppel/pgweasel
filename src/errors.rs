use crate::logreader;

pub fn process_errors(filename: &str, verbose: bool) {
    if verbose {
        println!("Parsing file: {}", filename);
    }

    // Use the logreader module to read the file line by line
    match logreader::getlines(filename) {
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
        Err(e) => eprintln!("Error opening file '{}': {}", filename, e),
    }
}
