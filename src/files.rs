use crate::logreader;
use log::debug;
use std::io::Result;

/// Get lines from either a file or stdin
pub fn get_lines_from_source(
    filename: Option<&str>,
    verbose: bool,
) -> Result<Box<dyn Iterator<Item = Result<String>>>> {
    match filename {
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
    }
}
