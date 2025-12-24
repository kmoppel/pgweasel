use std::time::Duration;

use regex::Regex;

use crate::util::TimeParseError;

use memchr::memmem;

pub fn extract_duration(record: &[u8]) -> Option<Duration> {
    let start = memmem::find(record, b"duration:")?;
    let mut i = start + b"duration:".len();

    // skip whitespace
    while i < record.len() && record[i] == b' ' {
        i += 1;
    }

    let num_start = i;

    // parse number
    while i < record.len() && (record[i].is_ascii_digit() || record[i] == b'.') {
        i += 1;
    }

    if i == num_start {
        return None;
    }

    let value = std::str::from_utf8(&record[num_start..i]).ok()?;

    // skip whitespace
    while i < record.len() && record[i] == b' ' {
        i += 1;
    }

    // parse unit
    let unit_start = i;
    while i < record.len() && record[i].is_ascii_alphabetic() {
        i += 1;
    }

    let unit = &record[unit_start..i];

    parse_duration_bytes(value, unit)
}

fn parse_duration_bytes(value: &str, unit: &[u8]) -> Option<Duration> {
    let v: f64 = value.parse().ok()?;

    match unit {
        b"ns" => Some(Duration::from_nanos(v as u64)),
        b"us" => Some(Duration::from_micros(v as u64)),
        b"ms" => Some(Duration::from_secs_f64(v / 1_000.0)),
        b"s" => Some(Duration::from_secs_f64(v)),
        b"m" | b"min" | b"minutes" => Some(Duration::from_secs_f64(v * 60.0)),
        _ => None,
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn simpleduration_extract_from_csv() {
        let log = b"Big text and duration: 121.997 ms more text";

        assert_eq!(extract_duration(log), Some(Duration::from_micros(121_997)));
    }

    #[test]
    fn simple_duration_extract_from_log() {
        let log = b"2025-05-21 11:00:40.296 UTC [675]: [3-1] db=postgres,user=cloudsqladmin,host=127.0.0.1 LOG:  duration: 3.032 ms  statement: SELECT extname, current_timestamp FROM pg_catalog.pg_extension UNION SELECT plugin, current_timestamp FROM pg_catalog.pg_replication_slots WHERE slot_type = 'logical' AND database = current_database();";

        assert_eq!(extract_duration(log), Some(Duration::from_micros(3_032)));
    }
}
