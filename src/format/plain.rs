
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
                .map(|p| start + p)
                .unwrap_or(record.len());

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
    }
}
