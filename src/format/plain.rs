#[inline]
pub fn message(record: &[u8]) -> Option<&[u8]> {
    let mut start = 0;
    while start + 1 < record.len() {
        if record[start] == b':' && record[start + 1] == b' ' {
            start += 1;
            // Skip spaces after colon
            while start < record.len() && record[start] == b' ' {
                start += 1;
            }

            // Find newline and stop there
            let end = record[start..]
                .iter()
                .position(|&b| b == b'\n')
                .map_or(record.len(), |p| start + p);

            return Some(&record[start..end]);
        }
        start += 1;
    }
    None
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn plain_message() {
        let line = b"2025-01-01 UTC [1] ERROR: bad thing happened\nError details...";
        assert_eq!(Some(b"bad thing happened".as_slice()), message(line));

        let line = b"2025-08-27 17:35:28.619 EEST [275518] sitt@postgres FATAL:  password authentication failed for user \"sitt\"";
        assert_eq!(
            Some(b"password authentication failed for user \"sitt\"".as_slice()),
            message(line)
        );

        let line = b"2025-05-21 11:01:20 UTC-682db26c.535-LOG:  disconnection: session time: 0:00:20.034 user=azuresu database=azure_maintenance host=127.0.0.1 port=55304";
        assert_eq!(
            Some(b"disconnection: session time: 0:00:20.034 user=azuresu database=azure_maintenance host=127.0.0.1 port=55304".as_slice()),
            message(line)
        );
    }
}
