use aho_corasick::AhoCorasick;

use crate::{filters::Filter, format::Format};

#[derive(Clone)]
pub struct SystemFilter {
    ac: AhoCorasick,
}

impl SystemFilter {
    pub fn new() -> Self {
        static PATTERNS: &[&[u8]] = &[
            // Autovacuum / maintenance
            b"autovacuum",
            b"checkpointer",
            b"background writer",
            b"bgwriter",
            // WAL / replication
            b"wal",
            b"replication",
            b"logical replication",
            b"replication slot",
            b"walreceiver",
            b"walsender",
            b"archiver",
            // Startup / shutdown
            b"starting PostgreSQL",
            b"database system is starting",
            b"database system is ready",
            b"database system is shutting down",
            b"startup process",
            b"shut down",
            b"listening",
            // Configuration changes
            b"reloading configuration",
            b"configuration file",
            // b"parameter",
            b"SIGHUP",
            // Extensions
            b"extension",
            b"shared_preload_libraries",
            b"CREATE EXTENSION",
        ];

        let ac = AhoCorasick::builder()
            .ascii_case_insensitive(true)
            .build(PATTERNS)
            .expect("failed to build Aho-Corasick automaton");

        Self { ac }
    }
}

impl Filter for SystemFilter {
    fn matches(&self, record: &[u8], _fmt: &Format) -> bool {
        self.ac.is_match(record)
    }
}

#[cfg(test)]
mod test {

    use super::*;

    #[test]
    fn test_system_filter() {
        let filter = super::SystemFilter::new();

        let test_cases: Vec<(&[u8], bool)> = vec![
            (b"autovacuum process started", true),
            (b"Background writer is active", true),
            (b"WAL segment created", true),
            (b"Database system is starting up", true),
            (b"Reloading configuration file", true),
            (b"Creating extension pg_stat_statements", true),
            (b"listening on IPv4 address \"127.0.0.1\", port 54316", true),
            (b"This is a normal log message", false),
            (b"User logged in successfully", false),
        ];

        for (input, expected) in test_cases {
            let result = filter.matches(input, &super::Format::Plain);
            assert_eq!(
                result,
                expected,
                "Failed on input: {:?}",
                String::from_utf8_lossy(input)
            );
        }
    }
}
