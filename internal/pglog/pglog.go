package pglog

import "time"

type LogEntry struct {
	LogTime        time.Time `json:"log_time"`
	ConnectionFrom string    `json:"connection_from"`
	UserName       string    `json:"user_name"`
	DatabaseName   string    `json:"database_name"`
	ProcessID      string    `json:"process_id"`
	ErrorSeverity  string    `json:"error_severity"`
	Message        string    `json:"message"`
}

// type PgLogEntry struct {
// 	LogTime              time.Time `json:"log_time"`
// 	UserName             string    `json:"user_name"`
// 	DatabaseName         string    `json:"database_name"`
// 	ProcessID            int       `json:"process_id"`
// 	ConnectionFrom       string    `json:"connection_from"`
// 	SessionID            string    `json:"session_id"`
// 	SessionLineNum       int64     `json:"session_line_num"`
// 	CommandTag           string    `json:"command_tag"`
// 	SessionStartTime     time.Time `json:"session_start_time"`
// 	VirtualTransactionID string    `json:"virtual_transaction_id"`
// 	TransactionID        *int64    `json:"transaction_id"`
// 	ErrorSeverity        string    `json:"error_severity"`
// 	SQLStateCode         string    `json:"sql_state_code"`
// 	Message              string    `json:"message"`
// 	Detail               string    `json:"detail"`
// 	Hint                 string    `json:"hint"`
// 	InternalQuery        string    `json:"internal_query"`
// 	InternalQueryPos     *int      `json:"internal_query_pos"`
// 	Context              string    `json:"context"`
// 	Query                string    `json:"query"`
// 	QueryPos             *int      `json:"query_pos"`
// 	Location             string    `json:"location"`
// 	ApplicationName      string    `json:"application_name"`
// 	BackendType          string    `json:"backend_type"`
// 	LeaderPID            *int      `json:"leader_pid"`
// 	QueryID              *int64    `json:"query_id"`
// }
