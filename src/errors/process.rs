use std::time::Instant;

use log::debug;

use crate::convert_args::ConvertedArgs;
use crate::errors::Severity;
use crate::parsers::get_parser;

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

pub fn process_errors(converted_args: ConvertedArgs, min_severity: &Severity) -> Result<()> {
    let min_severity_num: i32 = min_severity.into();

    for file_with_path in converted_args.files {
        if converted_args.verbose {
            debug!("Processing file: {}", file_with_path.path.to_str().unwrap());
        }

        let start = Instant::now();
        let mut parser = get_parser(file_with_path.path.clone())?;

        debug!("Read data within: {:?}", start.elapsed());

        for record in parser.parse(
            file_with_path.file,
            min_severity_num,
            converted_args.mask.clone(),
            converted_args.begin,
            converted_args.end,
        ) {
            let result = record?;
            println!(
                "timestamp {}, severity: {}, message: {}, raw string {}",
                result.timestamp, result.severity, result.message, result.raw
            );
        }
        debug!("Finished in: {:?}", start.elapsed());
    }
    Ok(())
}
