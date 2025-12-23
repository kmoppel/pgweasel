use std::time::Instant;

use chrono::{DateTime, Local};
use log::debug;
use memmap2::MmapOptions;

use crate::Severity;
use crate::aggregators::Aggregator;
use crate::convert_args::ConvertedArgs;
use crate::util::parse_timestamp_from_string;
use rayon::prelude::*;

use crate::Result;

pub fn output_results(
    converted_args: ConvertedArgs,
    min_severity: &Severity,
    _agragators: &mut Vec<Box<dyn Aggregator>>,
) -> Result<()> {
    let min_severity_num: i32 = min_severity.into();

    for file_with_path in converted_args.files {
        if converted_args.verbose {
            debug!("Processing file: {}", file_with_path.path.to_str().unwrap());
        }

        let mut filters = Filters {
            contains: vec![],
            starts_with: vec![],
            min_severity: min_severity_num,
            begin: converted_args.begin,
            end: converted_args.end,
            format: Format::from_file_extension(&file_with_path.path.to_string_lossy()),
        };

        let timing = Instant::now();

        let mmap = unsafe { MmapOptions::new().map(&file_with_path.file)? };
        let bytes: &[u8] = &mmap;

        let num_threads = rayon::current_num_threads();
        let chunk_size = bytes.len() / num_threads;

        let mut ranges = Vec::new();
        let mut start = 0;

        if let Some(mask) = &converted_args.mask {
            filters.contains.push(mask);
        };

        while start < bytes.len() {
            let mut end = (start + chunk_size).min(bytes.len());

            // Move end forward until a timestamp-starting line
            if end < bytes.len() {
                while end < bytes.len() {
                    if bytes[end] == b'\n' {
                        let next = end + 1;
                        if next < bytes.len() {
                            let line_end = bytes[next..]
                                .iter()
                                .position(|&b| b == b'\n')
                                .map(|p| next + p)
                                .unwrap_or(bytes.len());

                            let line =
                                unsafe { std::str::from_utf8_unchecked(&bytes[next..line_end]) };

                            if is_record_start(line) {
                                break;
                            }
                        }
                    }
                    end += 1;
                }
            }

            ranges.push(start..end);
            start = end + 1;
        }

        ranges.par_iter().try_for_each(|range| -> Result<()> {
            let slice = &bytes[range.clone()];
            let text = unsafe { std::str::from_utf8_unchecked(slice) };

            let mut record_start = 0;
            let mut offset = 0;

            for line in text.lines() {
                let line_len = line.len() + 1; // include '\n'

                if is_record_start(line) && offset != 0 {
                    let record = &slice[record_start..offset];
                    filter_record(record, &filters)?;
                    record_start = offset;
                }

                offset += line_len;
            }

            // last record in chunk
            if record_start < slice.len() {
                filter_record(&slice[record_start..slice.len()], &filters)?;
            };
            Ok(())
        })?;

        debug!("Finished in: {:?}", timing.elapsed());
    }
    Ok(())
}

enum Format {
    Csv,
    Plain,
}

impl Format {
    fn from_file_extension(file_name: &str) -> Self {
        if file_name.ends_with(".csv") {
            Format::Csv
        } else {
            Format::Plain
        }
    }
}

struct Filters<'a> {
    contains: Vec<&'a str>,
    starts_with: Vec<&'a str>,
    min_severity: i32,
    begin: Option<DateTime<Local>>,
    end: Option<DateTime<Local>>,
    format: Format,
}

#[inline]
fn filter_record(record: &[u8], filters: &Filters) -> Result<()> {
    for prefix in &filters.starts_with {
        if !record.starts_with(prefix.as_bytes()) {
            return Ok(());
        }
    }
    for substring in &filters.contains {
        if memchr::memmem::find(record, substring.as_bytes()).is_none() {
            return Ok(());
        }
    }
    let text = unsafe { std::str::from_utf8_unchecked(record) };
    let severity = match filters.format {
        Format::Csv => Severity::from_csv_string(text),
        Format::Plain => Severity::from_log_string(text),
    };
    let level: i32 = (&severity).into();
    if level < filters.min_severity {
        return Ok(());
    };

    let mut parts = text.split_whitespace();
    let ts_str = format!(
        "{} {} {}",
        parts.next().ok_or("Missing timestamp first part")?,
        parts.next().ok_or("Missing timestamp second part")?,
        parts.next().ok_or("Missing timestamp third part")?
    );

    let log_time_local = parse_timestamp_from_string(ts_str.as_str())?;
    if filters.begin.is_some_and(|b| log_time_local < b) {
        return Ok(());
    }
    if filters.end.is_some_and(|e| log_time_local > e) {
        return Ok(());
    }

    println!("{text}");
    Ok(())
}

#[inline]
fn is_record_start(line: &str) -> bool {
    let b = line.as_bytes();
    b.len() >= 23
        && b[4] == b'-'
        && b[7] == b'-'
        && b[10] == b' '
        && b[13] == b':'
        && b[16] == b':'
        && b[19] == b'.'
}
