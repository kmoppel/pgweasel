use std::{any::Any, collections::HashMap, time::Duration};

use crate::{aggregators::Aggregator, severity::Severity};

#[derive(Clone, Debug, Default)]
pub struct ConnectionsAggregator {
    total_connection_attempts: u16,
    total_authenticated: u16,
    total_authenticated_ssl: u16,
    connection_failures: u16,
    connections_by_host: HashMap<String, u16>,
    connections_by_database: HashMap<String, u16>,
    connections_by_user: HashMap<String, u16>,
    connections_by_appname: HashMap<String, u16>,
    // TODO: implement time-bucketed connection attempts when clear how it is gathered
    // connection_attempts_by_time_bucket: map[time.Time]int
    bucket_interval: Duration,
}

impl ConnectionsAggregator {
    pub fn new() -> Self {
        ConnectionsAggregator {
            total_connection_attempts: 0,
            total_authenticated: 0,
            total_authenticated_ssl: 0,
            connection_failures: 0,
            connections_by_host: HashMap::new(),
            connections_by_database: HashMap::new(),
            connections_by_user: HashMap::new(),
            connections_by_appname: HashMap::new(),
            bucket_interval: Duration::from_mins(10),
        }
    }
}

impl Aggregator for ConnectionsAggregator {
    fn update(&mut self, log_line: &[u8], severity: &Severity, fmt: &crate::format::Format) {
        // TODO: Handle the case where message extraction fails
        let message = match fmt.message_from_bytes(log_line) {
            Some(msg) => msg,
            None => return,
        };
        if (severity == &Severity::Fatal)
            && memchr::memmem::find(log_line, b"password authentication failed").is_some()
        {
            self.connection_failures += 1;
            return;
        };
        if severity == &Severity::Log {
            return;
        }
        if message.starts_with(b"connection received:") {
            self.total_connection_attempts += 1;
            let host = fmt.host_from_bytes(log_line).unwrap_or(b"unknown");
            self.connections_by_host
                .entry(String::from_utf8_lossy(host).to_string())
                .and_modify(|count| *count += 1)
                .or_insert(1);
        }
    }

    fn merge_box(&mut self, other: &dyn Aggregator) {
        let other = other
            .as_any()
            .downcast_ref::<ConnectionsAggregator>()
            .expect("Aggregator type mismatch");

        self.total_connection_attempts += other.total_connection_attempts;
        self.total_authenticated += other.total_authenticated;
        self.total_authenticated_ssl += other.total_authenticated_ssl;
        self.connection_failures += other.connection_failures;

        for (host, count) in &other.connections_by_host {
            *self
                .connections_by_host
                .entry(host.clone())
                .or_insert(0) += count;
        }
    }

    fn print(&mut self) {
        println!("Total connection attempts: {}", self.total_connection_attempts);
        println!("Total authenticated connections: {}", self.total_authenticated);
        println!(
            "Total authenticated SSL connections: {}",
            self.total_authenticated_ssl
        );
        println!("Total connection failures: {}", self.connection_failures);
        println!("Connections by host:");
        for (host, count) in &self.connections_by_host {
            println!("  {:>6}  {}", count, host);
        }
        println!("Connections by database:");
        for (db, count) in &self.connections_by_database {
            println!("  {:>6}  {}", count, db);
        }
        println!("Connections by user:");
        for (user, count) in &self.connections_by_user {
            println!("  {:>6}  {}", count, user);
        }
        println!("Connections by application name:");
        for (appname, count) in &self.connections_by_appname {
            println!("  {:>6}  {}", count, appname);
        }
    }

    fn boxed_clone(&self) -> Box<dyn Aggregator> {
        Box::new(self.clone())
    }

    fn as_any(&self) -> &dyn Any {
        self
    }
}
