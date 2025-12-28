
#[inline]
pub fn message(record: &[u8]) -> Option<&[u8]> {
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

#[cfg(test)]
mod test {

    use super::*;

    #[test]
    fn test_message() {
        let line = b"2025-01-01,1,db,user,ERROR,bad thing happened";

        assert_eq!(message(line), Some(b"bad thing happened".as_slice()));
    }
}
