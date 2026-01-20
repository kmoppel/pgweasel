pub fn message(record: &[u8]) -> Option<&[u8]> {
    extract_csv_field(record, 14)
}

/// Extracts nth field from CSV record
/// Field index is 1-based.
fn extract_csv_field(record: &[u8], field_index: usize) -> Option<&[u8]> {
    if field_index == 0 {
        return None;
    }

    let mut in_quotes = false;
    let mut current_field = 1;
    let mut field_start = 0;

    let mut i = 0;
    while i < record.len() {
        match record[i] {
            b'"' => {
                if in_quotes && i + 1 < record.len() && record[i + 1] == b'"' {
                    i += 1; // escaped quote
                } else {
                    in_quotes = !in_quotes;
                }
            }
            b',' if !in_quotes => {
                if current_field == field_index {
                    return Some(strip_csv_quotes(&record[field_start..i]));
                }
                current_field += 1;
                field_start = i + 1;
            }
            _ => {}
        }
        i += 1;
    }

    // Handle last field
    if current_field == field_index {
        Some(strip_csv_quotes(&record[field_start..]))
    } else {
        None
    }
}

#[inline]
fn strip_csv_quotes(field: &[u8]) -> &[u8] {
    if field.len() >= 2 && field[0] == b'"' && field[field.len() - 1] == b'"' {
        &field[1..field.len() - 1]
    } else {
        field
    }
}

#[cfg(test)]
mod test {

    use super::*;

    #[test]
    fn test_message() {
        let line = b"2025-12-01 01:56:57.080 EET,,,1637804,\"10.203.8.108:53096\",692cd9c9.18fdac,1,\"\",2025-12-01 01:56:57 EET,,0,LOG,00000,\"connection received: host=10.203.8.108 port=53096\",,,,,,,,,\"\",\"not initialized\",,0
";

        assert_eq!(
            message(line),
            Some(b"connection received: host=10.203.8.108 port=53096".as_slice())
        );
    }
}
