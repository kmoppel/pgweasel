package pglog

import (
	"regexp"
	"strings"
	"time"

	"github.com/icza/gox/gox"
	"github.com/kmoppel/pgweasel/internal/util"
)

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
	LeaderPid            string // Column 25
	QueryId              string // Column 26
}

type LogEntry struct {
	LogTime       string
	ErrorSeverity string
	Message       string
	Lines         []string // For plain text logs
	CsvRecords    []string // For CSV logs
	CsvColumns    CsvEntry
}

var REGEX_USER_AT_DB = regexp.MustCompile(`(?s)^(?P<log_time>[\d\-:\. ]+ [A-Z]+).*?[:\s]+(?P<user_name>[A-Za-z0-9_\-]+)@(?P<database_name>[A-Za-z0-9_\-]+)[:\s]+.*?(?P<error_severity>[A-Z0-9]+)[:\s]+.*$`)

func (e CsvEntry) String() string {
	return strings.Join([]string{
		e.LogTime,
		gox.If(e.UserName != "", `"`+e.UserName+`"`, ""),
		gox.If(e.DatabaseName != "", `"`+e.DatabaseName+`"`, ""),
		e.ProcessID,
		gox.If(e.ConnectionFrom != "", `"`+e.ConnectionFrom+`"`, ""),
		e.SessionID,
		e.SessionLineNum,
		gox.If(e.CommandTag != "", `"`+e.CommandTag+`"`, ""),
		e.SessionStartTime,
		e.VirtualTransactionID,
		e.TransactionID,
		e.ErrorSeverity,
		e.SQLStateCode,
		gox.If(e.Message != "", `"`+e.Message+`"`, ""),
		gox.If(e.Detail != "", `"`+e.Detail+`"`, ""),
		gox.If(e.Hint != "", `"`+e.Hint+`"`, ""),
		e.InternalQuery,
		e.InternalQueryPos,
		gox.If(e.Context != "", `"`+e.Context+`"`, ""),
		gox.If(e.Query != "", `"`+e.Query+`"`, ""),
		e.QueryPos,
		e.Location,
		gox.If(e.ApplicationName != "", `"`+e.ApplicationName+`"`, ""),
		gox.If(e.BackendType != "", `"`+e.BackendType+`"`, ""),
		e.QueryId,
	}, ",")
}

// Postgres log levels are DEBUG5, DEBUG4, DEBUG3, DEBUG2, DEBUG1, INFO, NOTICE, WARNING, ERROR, LOG, FATAL, and PANIC
// but move LOG lower as too chatty otherwise (connections received, slow queries, etc)
func (e LogEntry) SeverityNum() int {
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
		return 5
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
		return 5
	}
}

// Idea here is to delay time parsing as might not be needed
// for example if we are only looking for errors and have no time range set by the user
func (e LogEntry) GetTime() time.Time {
	return util.TimestringToTime(e.LogTime)
}

func (e LogEntry) IsSystemEntry() bool {
	if e.CsvRecords != nil {
		return e.CsvRecords[1] == ""
	} else { // TODO With plain text logs very hard to detect actually without log_line_prefix so need to use that as well
		// let's assume for now user@db
		if REGEX_USER_AT_DB.MatchString(strings.Join(e.Lines, "\n")) {
			return false
		}
	}
	return true
}
