use std::time::Instant;

use log::debug;

use crate::Severity;
use crate::aggregators::Aggregator;
use crate::convert_args::ConvertedArgs;
use crate::parsers::get_parser;

use crate::Result;

pub fn output_results(
    converted_args: ConvertedArgs,
    min_severity: &Severity,
    agragators: &mut Vec<Box<dyn Aggregator>>,
) -> Result<()> {
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
            debug!("Log line received: {result:?}");
            let mut tmp = agragators.into_iter();
            while let Some(agregator) = tmp.next() {
                agregator.add(result.clone());
            }
        }
        let mut tmp = agragators.into_iter();
        while let Some(agregator) = tmp.next() {
            agregator.print();
        }
        debug!("Finished in: {:?}", start.elapsed());
    }
    Ok(())
}
