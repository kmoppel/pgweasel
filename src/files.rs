use crate::logreader;
use log::debug;
use std::fs;
use std::io::{Error, ErrorKind, Result};
use std::path::Path;

/// Get lines from either a file, a folder (all .log files), or stdin
pub fn get_lines_from_source(
    filename: Option<&str>,
    verbose: bool,
) -> Result<Box<dyn Iterator<Item = Result<String>>>> {
    match filename {
        Some(file) => {
            let path = Path::new(file);
            
            // Check if path is a directory
            if path.is_dir() {
                if verbose {
                    debug!("Processing directory: {}", file);
                }
                
                // Read all entries in the directory
                let mut log_files: Vec<String> = fs::read_dir(path)?
                    .filter_map(|entry| {
                        entry.ok().and_then(|e| {
                            let path = e.path();
                            if path.is_file() && 
                               path.extension().and_then(|s| s.to_str()) == Some("log") {
                                path.to_str().map(|s| s.to_string())
                            } else {
                                None
                            }
                        })
                    })
                    .collect();
                
                // Sort files by name
                log_files.sort();
                
                if verbose {
                    debug!("Found {} .log files", log_files.len());
                }
                
                if log_files.is_empty() {
                    return Err(Error::new(
                        ErrorKind::NotFound,
                        format!("No .log files found in directory: {}", file)
                    ));
                }
                
                // Create a chained iterator over all log files
                let mut iter: Box<dyn Iterator<Item = Result<String>>> = 
                    Box::new(std::iter::empty());
                
                for log_file in log_files {
                    if verbose {
                        debug!("Adding file to processing queue: {}", log_file);
                    }
                    let lines = logreader::getlines(&log_file)?;
                    iter = Box::new(iter.chain(lines));
                }
                
                Ok(iter)
            } else {
                // It's a file
                if verbose {
                    debug!("Parsing file: {}", file);
                }
                logreader::getlines(file)
            }
        }
        None => {
            if verbose {
                debug!("Reading from stdin...");
            }
            logreader::getlines_from_stdin()
        }
    }
}
