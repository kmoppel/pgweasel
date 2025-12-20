use chrono::{DateTime, FixedOffset};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PostgresLog {
    #[serde(with = "crate::parsers::date_serializer")]
    pub log_time: Option<DateTime<FixedOffset>>,
    pub user_name: Option<String>,
    pub database_name: Option<String>,
    pub process_id: Option<i32>,
    pub connection_from: Option<String>,
    pub session_id: String,
    pub session_line_num: i64,
    pub command_tag: Option<String>,
    #[serde(with = "crate::parsers::date_serializer")]
    pub session_start_time: Option<DateTime<FixedOffset>>,
    pub virtual_transaction_id: Option<String>,
    pub transaction_id: Option<i64>,
    pub error_severity: String,
    pub sql_state_code: Option<String>,
    pub message: Option<String>,
    pub detail: Option<String>,
    pub hint: Option<String>,
    pub internal_query: Option<String>,
    pub internal_query_pos: Option<i32>,
    pub context: Option<String>,
    pub query: Option<String>,
    pub query_pos: Option<i32>,
    pub location: Option<String>,
    pub application_name: Option<String>,
    pub backend_type: Option<String>,
    pub leader_pid: Option<i32>,
    pub query_id: Option<i64>,
}
