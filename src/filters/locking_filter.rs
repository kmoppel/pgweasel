use crate::{filters::Filter, format::Format};

#[derive(Clone)]
pub struct LockingFilter;

impl Filter for LockingFilter {
    fn matches(&self, record: &[u8], _fmt: &Format) -> bool {
        if memchr::memmem::find(record, b" conflicts ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b" conflicting ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b" still waiting for ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"Wait queue:").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"while locking tuple").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"while updating tuple").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"conflict detected").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"deadlock detected").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"buffer deadlock").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"blocked by process ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"recovery conflict ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b" concurrent update").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"could not serialize").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"could not obtain ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"lock on relation ").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b"cannot lock rows").is_some() {
            return true;
        };
        if memchr::memmem::find(record, b" semaphore:").is_some() {
            return true;
        };

        matches_process_acquired(record)
    }
}

pub fn matches_process_acquired(record: &[u8]) -> bool {
    const PREFIX: &[u8] = b"process ";
    const SUFFIX: &[u8] = b" acquired";

    let mut i = 0;

    while i + PREFIX.len() <= record.len() {
        // Look for "process "
        if record[i..].starts_with(PREFIX) {
            let mut j = i + PREFIX.len();

            // Must have at least one digit
            if j >= record.len() || !record[j].is_ascii_digit() {
                i += 1;
                continue;
            }

            // Consume [0-9]+
            while j < record.len() && record[j].is_ascii_digit() {
                j += 1;
            }

            // Must be followed by " acquired"
            if j + SUFFIX.len() <= record.len() && record[j..].starts_with(SUFFIX) {
                return true;
            }
        }

        i += 1;
    }

    false
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_matches_process_acquired() {
        assert!(matches_process_acquired(b"process 123 acquired"));
        assert!(matches_process_acquired(b"foo process 9 acquired bar"));
        assert!(matches_process_acquired(b"xprocess 1 acquired"));
        assert!(!matches_process_acquired(b"process acquired"));
        assert!(!matches_process_acquired(b"process  acquired"));
    }
}
