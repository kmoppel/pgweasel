use std::{any::Any, collections::HashMap, time::Duration};

use chrono::{DateTime, Local, TimeZone};

use crate::{aggregators::Aggregator, format::Format, severity::Severity};

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
    connection_attempts_by_time_bucket: HashMap<String, u16>,
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
            connection_attempts_by_time_bucket: HashMap::new(),
            bucket_interval: Duration::from_mins(10),
        }
    }
}

impl Aggregator for ConnectionsAggregator {
    fn update(
        &mut self,
        record: &[u8],
        fmt: &Format,
        severity: &Severity,
        log_time: DateTime<Local>,
    ) {
        // TODO: Handle the case where message extraction fails
        let message = match fmt.message_from_bytes(record) {
            Some(msg) => msg,
            None => return,
        };
        if (severity == &Severity::Fatal)
            && memchr::memmem::find(record, b"password authentication failed").is_some()
        {
            self.connection_failures += 1;
            return;
        };

        if severity != &Severity::Log {
            return;
        }

        if message.starts_with(b"connection received:") {
            self.total_connection_attempts += 1;
            let host = fmt.host_from_bytes(record).unwrap_or(b"unknown");
            self.connections_by_host
                .entry(String::from_utf8_lossy(host).to_string())
                .and_modify(|count| *count += 1)
                .or_insert(1);

            let bucket_time = round_floor(log_time, self.bucket_interval);
            let bucket_time_str = bucket_time.to_string();
            self.connection_attempts_by_time_bucket
                .entry(bucket_time_str)
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
            *self.connections_by_host.entry(host.clone()).or_insert(0) += count;
        }

        for (bucket, count) in &other.connection_attempts_by_time_bucket {
            *self
                .connection_attempts_by_time_bucket
                .entry(bucket.clone())
                .or_insert(0) += count;
        }
    }

    fn print(&mut self) {
        println!(
            "Total connection attempts: {}",
            self.total_connection_attempts
        );
        println!(
            "Total authenticated connections: {}",
            self.total_authenticated
        );
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
        println!("Connections by time bucket:");
        for (appname, count) in &self.connection_attempts_by_time_bucket {
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

fn duration_to_nanos(d: Duration) -> i128 {
    d.as_secs() as i128 * 1_000_000_000 + d.subsec_nanos() as i128
}

fn datetime_to_nanos(dt: DateTime<Local>) -> i128 {
    dt.timestamp() as i128 * 1_000_000_000 + dt.timestamp_subsec_nanos() as i128
}

fn nanos_to_datetime(nanos: i128) -> DateTime<Local> {
    let secs = nanos / 1_000_000_000;
    let nsecs = (nanos % 1_000_000_000) as u32;

    Local
        .timestamp_opt(secs as i64, nsecs)
        .single()
        .expect("valid timestamp")
}

pub fn round_floor(dt: DateTime<Local>, interval: Duration) -> DateTime<Local> {
    let i = duration_to_nanos(interval);
    let t = datetime_to_nanos(dt);

    nanos_to_datetime(t - (t % i))
}
