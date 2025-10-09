use std::fs::File;
use std::io::{self, BufRead, BufReader, Result};

/// Reads a file line by line and returns an iterator over the lines
///
/// # Arguments
///
/// * `filepath` - A string slice that holds the path to the file
///
/// # Returns
///
/// * `Result<impl Iterator<Item = Result<String>>>` - An iterator that yields each line as a Result<String>
///
/// # Examples
///
/// ```
/// use pgweasel_rust::logreader::getlines;
///
/// for line_result in getlines("path/to/logfile.log")? {
///     match line_result {
///         Ok(line) => println!("{}", line),
///         Err(e) => eprintln!("Error reading line: {}", e),
///     }
/// }
/// ```
pub fn getlines(filepath: &str) -> Result<impl Iterator<Item = Result<String>>> {
    let file = File::open(filepath)?;
    let reader = BufReader::new(file);
    Ok(reader.lines())
}

/// Reads from stdin line by line and returns an iterator over the lines
///
/// # Returns
///
/// * `impl Iterator<Item = Result<String>>` - An iterator that yields each line as a Result<String>
///
/// # Examples
///
/// ```
/// use pgweasel_rust::logreader::getlines_from_stdin;
///
/// for line_result in getlines_from_stdin() {
///     match line_result {
///         Ok(line) => println!("{}", line),
///         Err(e) => eprintln!("Error reading line: {}", e),
///     }
/// }
/// ```
pub fn getlines_from_stdin() -> impl Iterator<Item = Result<String>> {
    let stdin = io::stdin();
    stdin.lock().lines()
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    #[test]
    fn test_getlines_with_valid_file() {
        // Create a temporary file with test content
        let mut temp_file = NamedTempFile::new().unwrap();
        writeln!(temp_file, "Line 1").unwrap();
        writeln!(temp_file, "Line 2").unwrap();
        writeln!(temp_file, "Line 3").unwrap();

        let temp_path = temp_file.path().to_str().unwrap();

        // Read lines using our function
        let lines: Result<Vec<String>> = getlines(temp_path).unwrap().collect();

        let lines = lines.unwrap();
        assert_eq!(lines.len(), 3);
        assert_eq!(lines[0], "Line 1");
        assert_eq!(lines[1], "Line 2");
        assert_eq!(lines[2], "Line 3");
    }

    #[test]
    fn test_getlines_with_nonexistent_file() {
        let result = getlines("nonexistent_file.txt");
        assert!(result.is_err());
    }

    #[test]
    fn test_getlines_with_empty_file() {
        let temp_file = NamedTempFile::new().unwrap();
        let temp_path = temp_file.path().to_str().unwrap();

        let lines: Result<Vec<String>> = getlines(temp_path).unwrap().collect();

        let lines = lines.unwrap();
        assert_eq!(lines.len(), 0);
    }
}
