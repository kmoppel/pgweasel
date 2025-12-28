use std::time::Instant;

use chrono::{DateTime, Local};
use log::debug;
use memmap2::MmapOptions;

use crate::Severity;
use crate::aggregators::Aggregator;
use crate::convert_args::ConvertedArgs;
use crate::filters::{Filter, FilterContains};
use crate::format::Format;
use crate::util::parse_timestamp_from_string;
use rayon::prelude::*;

use crate::Result;

pub fn output_results(
    converted_args: ConvertedArgs,
    min_severity: &Severity,
    aggregators: &mut Vec<Box<dyn Aggregator>>,
    filters: &Vec<Box<dyn Filter>>,
) -> Result<()> {
    let min_severity_num: i32 = min_severity.into();

    for file_with_path in converted_args.files {
        if converted_args.verbose {
            debug!("Processing file: {}", file_with_path.path.to_str().unwrap());
        }

        let mut filter_container = FilterContainer {
            filters: vec![],
            custom_filters: filters,
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
            let mask_filter = Box::new(FilterContains::new(mask.clone()));
            filter_container.filters.push(mask_filter);
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
                                .position(|&b| (b == b'\n') && (b == b'\r'))
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

        debug!("File did read in: {:?}", timing.elapsed());

        let partials: Result<Vec<Vec<Box<dyn Aggregator>>>> = ranges
            .par_iter()
            .map(|range| -> Result<Vec<Box<dyn Aggregator>>> {
                let mut local_aggregators: Vec<Box<dyn Aggregator>> =
                    aggregators.iter().map(|a| a.boxed_clone()).collect();

                let slice = &bytes[range.clone()];
                // TODO: iterate slice without converting to str and back to record bytes
                let text = unsafe { std::str::from_utf8_unchecked(slice) };

                let mut record_start = 0;
                let mut offset = 0;

                for line in text.lines() {
                    let line_len = line.len() + 1; // include '\n'

                    if is_record_start(line) && offset != 0 {
                        let record = &slice[record_start..offset];
                        // debug!("Processing record: {:?} start {} offset {} line_len {}", std::str::from_utf8(record), record_start, offset, line_len);
                        filter_record(
                            record,
                            &filter_container,
                            &mut local_aggregators,
                            converted_args.print_details,
                        )?;
                        record_start = offset;
                    }

                    offset += line_len;
                    // sometimes calculated offset is not at line end due to multibyte utf8 chars

                    while (offset < slice.len()) && (&slice[offset - 1..offset] != b"\n") {
                        offset += 1;
                    }
                }

                // last record in chunk
                if record_start < slice.len() {
                    filter_record(
                        &slice[record_start..slice.len()],
                        &filter_container,
                        &mut local_aggregators,
                        converted_args.print_details,
                    )?;
                };
                Ok(local_aggregators)
            })
            .collect();

        debug!("Finished output in: {:?}", timing.elapsed());
        let partials = partials?;
        for partial in partials {
            for (i, aggregator) in partial.into_iter().enumerate() {
                aggregators[i].merge_box(aggregator.as_ref());
            }
        }
        for agg in &mut *aggregators {
            agg.print();
        }
        debug!("Finished aggregating in: {:?}", timing.elapsed());
    }
    Ok(())
}

struct FilterContainer<'a> {
    custom_filters: &'a Vec<Box<dyn Filter + 'a>>,
    filters: Vec<Box<dyn Filter>>,
    min_severity: i32,
    begin: Option<DateTime<Local>>,
    end: Option<DateTime<Local>>,
    format: Format,
}

#[inline]
fn filter_record(
    record: &[u8],
    filters: &FilterContainer,
    local_aggregators: &mut Vec<Box<dyn Aggregator>>,
    print: bool,
) -> Result<()> {
    for filter in &filters.filters {
        if !filter.matches(record, &filters.format) {
            return Ok(());
        }
    }

    // Next code is not written as filters to avoid multiple string parsing and degradation of performance
    let text = unsafe { std::str::from_utf8_unchecked(record) };
    let severity = filters.format.severity_from_string(text);
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

    for custom_filter in filters.custom_filters {
        if !custom_filter.matches(record, &filters.format) {
            return Ok(());
        }
    }

    aggragate_record(
        local_aggregators,
        record,
        &filters.format,
        &severity,
        log_time_local,
    )?;

    if print {
        println!("{text}");
    }
    Ok(())
}

#[inline]
fn aggragate_record(
    local_aggregators: &mut Vec<Box<dyn Aggregator>>,
    record: &[u8],
    fmt: &Format,
    severity: &Severity,
    log_time: DateTime<Local>,
) -> Result<()> {
    for aggregator in local_aggregators.iter_mut() {
        aggregator.update(record, fmt, severity, log_time)?;
    }
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
        && (b[19] == b'.' || b[19] == b' ')
}

#[cfg(test)]
mod test {

    use super::*;

    #[test]
    fn test_record_start() {
        let line = "2025-05-21 11:01:20 UTC-682db26c.535-LOG:  disconnection: session time: 0:00:20.034 user=azuresu database=azure_maintenance host=127.0.0.1 port=55304";
        assert!(is_record_start(line));
    }
}
