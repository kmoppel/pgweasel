use aho_corasick::AhoCorasick;

use crate::{filters::Filter, format::Format};

#[derive(Clone)]
pub struct LockingFilter {
    ac: AhoCorasick,
}

impl LockingFilter {
    pub fn new() -> Self {
        static PATTERNS: &[&[u8]] = &[
            b" conflicts ",
            b" conflicting ",
            b" still waiting for ",
            b"Wait queue:",
            b"while locking tuple",
            b"while updating tuple",
            b"conflict detected",
            b"deadlock detected",
            b"buffer deadlock",
            b"blocked by process ",
            b"recovery conflict ",
            b" concurrent update",
            b"could not serialize",
            b"could not obtain ",
            b"lock on relation ",
            b"cannot lock rows",
            b" semaphore:",
        ];

        let ac = AhoCorasick::builder()
            .ascii_case_insensitive(true)
            .build(PATTERNS)
            .expect("failed to build Aho-Corasick automaton");

        Self { ac }
    }
}

impl Filter for LockingFilter {
    fn matches(&self, record: &[u8], _fmt: &Format) -> bool {
        if self.ac.is_match(record) {
            return true;
        }

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
