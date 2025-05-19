package pglog

import (
	"strings"
	"time"
)

type LogEntry struct {
	LogTime       time.Time
	ErrorSeverity string
	Message       string
	Lines         []string
}

// Postgres log levels are DEBUG5, DEBUG4, DEBUG3, DEBUG2, DEBUG1, INFO, NOTICE, WARNING, ERROR, LOG, FATAL, and PANIC
// but move LOG lower as too chatty otherwise
func (e *LogEntry) SeverityNum() int {
	switch strings.ToUpper(e.ErrorSeverity) {
	case "DEBUG5":
		return 0
	case "DEBUG4":
		return 1
	case "DEBUG3":
		return 2
	case "DEBUG2":
		return 3
	case "DEBUG1":
		return 4
	case "LOG":
		return 5
	case "INFO":
		return 5
	case "NOTICE":
		return 6
	case "WARNING":
		return 7
	case "ERROR":
		return 8
	case "FATAL":
		return 9
	case "PANIC":
		return 10
	default:
		return -1
	}
}

func SeverityToNum(severity string) int {
	severity = strings.ToUpper(severity)
	switch severity {
	case "DEBUG5":
		return 0
	case "DEBUG4":
		return 1
	case "DEBUG3":
		return 2
	case "DEBUG2":
		return 3
	case "DEBUG1":
		return 4
	case "LOG":
		return 5
	case "INFO":
		return 5
	case "NOTICE":
		return 6
	case "WARNING":
		return 7
	case "ERROR":
		return 8
	case "FATAL":
		return 9
	case "PANIC":
		return 10
	default:
		return -1
	}
}

type CsvEntry struct {
	LogTime              string // Column 1
	UserName             string // Column 2
	DatabaseName         string // Column 3
	ProcessID            string // Column 4
	ConnectionFrom       string // Column 5
	SessionID            string // Column 6
	SessionLineNum       string // Column 7
	CommandTag           string // Column 8
	SessionStartTime     string // Column 9
	VirtualTransactionID string // Column 10
	TransactionID        string // Column 11
	ErrorSeverity        string // Column 12
	SQLStateCode         string // Column 13
	Message              string // Column 14
	Detail               string // Column 15
	Hint                 string // Column 16
	InternalQuery        string // Column 17
	InternalQueryPos     string // Column 18
	Context              string // Column 19
	Query                string // Column 20
	QueryPos             string // Column 21
	Location             string // Column 22
	ApplicationName      string // Column 23
	BackendType          string // Column 24
	//   leader_pid integer,
	//   query_id bigint
	Lines []string // All lines for quick joining
}

func (e *CsvEntry) SeverityNum() int {
	switch strings.ToUpper(e.ErrorSeverity) {
	case "DEBUG5":
		return 0
	case "DEBUG4":
		return 1
	case "DEBUG3":
		return 2
	case "DEBUG2":
		return 3
	case "DEBUG1":
		return 4
	case "LOG":
		return 5
	case "INFO":
		return 5
	case "NOTICE":
		return 6
	case "WARNING":
		return 7
	case "ERROR":
		return 8
	case "FATAL":
		return 9
	case "PANIC":
		return 10
	default:
		return -1
	}
}
