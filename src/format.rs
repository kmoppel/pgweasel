use crate::severity::Severity;

pub enum Format {
    Csv,
    Plain,
}

impl Format {
    pub fn from_file_extension(file_name: &str) -> Self {
        if file_name.ends_with(".csv") {
            Format::Csv
        } else {
            Format::Plain
        }
    }

    pub fn severity_from_string(&self, text: &str) -> Severity {
        match self {
            Format::Csv => Severity::from_csv_string(text),
            Format::Plain => Severity::from_log_string(text),
        }
    }

    pub fn message_from_bytes<'a>(&self, record: &'a [u8]) -> Option<&'a [u8]> {
        match self {
            Format::Plain => Self::message_plain(record),
            Format::Csv => Self::message_csv(record),
        }
    }

    fn message_plain(record: &[u8]) -> Option<&[u8]> {
        let mut i = 0;
        while i + 1 < record.len() {
            if record[i] == b':' && record[i + 1] == b' ' {
                let start = i + 2;

                // Find newline and stop there
                let end = record[start..]
                    .iter()
                    .position(|&b| b == b'\n')
                    .map(|p| start + p)
                    .unwrap_or(record.len());

                return Some(&record[start..end]);
            }
            i += 1;
        }
        None
    }

    fn message_csv(record: &[u8]) -> Option<&[u8]> {
        // Find last comma → message is final field
        let mut i = record.len();
        while i > 0 {
            i -= 1;
            if record[i] == b',' {
                return record.get(i + 1..);
            }
        }
        None
    }
}

#[cfg(test)]
mod test {
    use super::Format;

    #[test]
    fn plain_message() {
        let line = b"2025-01-01 UTC [1] ERROR: bad thing happened\nError details...";
        let fmt = Format::Plain;

        assert_eq!(
            fmt.message_from_bytes(line),
            Some(b"bad thing happened".as_slice())
        );
    }

    #[test]
    fn csv_message() {
        let line = b"2025-01-01,1,db,user,ERROR,bad thing happened";
        let fmt = Format::Csv;

        assert_eq!(
            fmt.message_from_bytes(line),
            Some(b"bad thing happened".as_slice())
        );
    }
}
