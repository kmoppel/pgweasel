/// Valid PostgreSQL log severity levels
pub const VALID_SEVERITIES: &[&str] = &[
    "DEBUG5", "DEBUG4", "DEBUG3", "DEBUG2", "DEBUG1", "LOG", "INFO", "NOTICE", "WARNING", "ERROR",
    "FATAL", "PANIC",
];

/// CSV log entry structure for PostgreSQL CSV logs
/// CSV column count varies by version: <v13=23, v14=24, v15+=26
#[derive(Debug, Clone)]
pub struct CsvEntry {
    pub csv_column_count: i32,       // <v13=23, v14=24,v15+=26
    pub log_time: String,             // Column 1
    pub user_name: String,            // Column 2
    pub database_name: String,        // Column 3
    pub process_id: String,           // Column 4
    pub connection_from: String,      // Column 5
    pub session_id: String,           // Column 6
    pub session_line_num: String,     // Column 7
    pub command_tag: String,          // Column 8
    pub session_start_time: String,   // Column 9
    pub virtual_transaction_id: String, // Column 10
    pub transaction_id: String,       // Column 11
    pub error_severity: String,       // Column 12
    pub sql_state_code: String,       // Column 13
    pub message: String,              // Column 14
    pub detail: String,               // Column 15
    pub hint: String,                 // Column 16
    pub internal_query: String,       // Column 17
    pub internal_query_pos: String,   // Column 18
    pub context: String,              // Column 19
    pub query: String,                // Column 20
    pub query_pos: String,            // Column 21
    pub location: String,             // Column 22
    pub application_name: String,     // Column 23
    pub backend_type: String,         // Column 24
    pub leader_pid: String,           // Column 25
    pub query_id: String,             // Column 26
}

/// PostgreSQL log entry structure
#[derive(Debug, Clone)]
pub struct LogEntry {
    pub log_time: String,
    pub error_severity: String,
    pub message: String,
    pub lines: Vec<String>,           // For plain text logs
    pub csv_columns: Option<CsvEntry>, // Optional CSV entry data
}

