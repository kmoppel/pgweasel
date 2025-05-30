package pglog

import (
	"regexp"
	"strings"
	"time"

	"github.com/icza/gox/gox"
	"github.com/kmoppel/pgweasel/internal/util"
)

var ERROR_SEVERITIES = []string{"WARNING", "ERROR", "FATAL", "PANIC"}

type CsvEntry struct {
	CsvColumnCount       int    // <v13=23, v14=24,v15+=26
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
	CsvColumns    *CsvEntry
}

var REGEX_USER_AT_DB = regexp.MustCompile(`(?s)^(?P<log_time>[\d\-:\. ]{19,23} [A-Z]{2,5})[:\s\-]+.*?(?P<user_name>[A-Za-z0-9_\-]+)@(?P<database_name>[A-Za-z0-9_\-]+)[:\s\-]+.*?(?P<error_severity>[A-Z12345]{3,12})[:\s]+.*$`)

func (e CsvEntry) String() string {
	if e.CsvColumnCount == 23 {
		return strings.Join([]string{
			e.LogTime,
			gox.If(e.UserName != "", `"`+strings.ReplaceAll(e.UserName, `"`, `""`)+`"`, ""),
			gox.If(e.DatabaseName != "", `"`+strings.ReplaceAll(e.DatabaseName, `"`, `""`)+`"`, ""),
			e.ProcessID,
			gox.If(e.ConnectionFrom != "", `"`+strings.ReplaceAll(e.ConnectionFrom, `"`, `""`)+`"`, ""),
			e.SessionID,
			e.SessionLineNum,
			gox.If(e.CommandTag != "", `"`+strings.ReplaceAll(e.CommandTag, `"`, `""`)+`"`, ""),
			e.SessionStartTime,
			e.VirtualTransactionID,
			e.TransactionID,
			e.ErrorSeverity,
			e.SQLStateCode,
			gox.If(e.Message != "", `"`+strings.ReplaceAll(e.Message, `"`, `""`)+`"`, ""),
			gox.If(e.Detail != "", `"`+strings.ReplaceAll(e.Detail, `"`, `""`)+`"`, ""),
			gox.If(e.Hint != "", `"`+strings.ReplaceAll(e.Hint, `"`, `""`)+`"`, ""),
			e.InternalQuery,
			e.InternalQueryPos,
			gox.If(e.Context != "", `"`+strings.ReplaceAll(e.Context, `"`, `""`)+`"`, ""),
			gox.If(e.Query != "", `"`+strings.ReplaceAll(e.Query, `"`, `""`)+`"`, ""),
			e.QueryPos,
			e.Location,
			gox.If(e.ApplicationName != "" || (e.SQLStateCode == "00000" && e.UserName == ""), `"`+strings.ReplaceAll(e.ApplicationName, `"`, `""`)+`"`, ""),
		}, ",")
	}
	if e.CsvColumnCount == 24 {
		return strings.Join([]string{
			e.LogTime,
			gox.If(e.UserName != "", `"`+strings.ReplaceAll(e.UserName, `"`, `""`)+`"`, ""),
			gox.If(e.DatabaseName != "", `"`+strings.ReplaceAll(e.DatabaseName, `"`, `""`)+`"`, ""),
			e.ProcessID,
			gox.If(e.ConnectionFrom != "", `"`+strings.ReplaceAll(e.ConnectionFrom, `"`, `""`)+`"`, ""),
			e.SessionID,
			e.SessionLineNum,
			gox.If(e.CommandTag != "", `"`+strings.ReplaceAll(e.CommandTag, `"`, `""`)+`"`, ""),
			e.SessionStartTime,
			e.VirtualTransactionID,
			e.TransactionID,
			e.ErrorSeverity,
			e.SQLStateCode,
			gox.If(e.Message != "", `"`+strings.ReplaceAll(e.Message, `"`, `""`)+`"`, ""),
			gox.If(e.Detail != "", `"`+strings.ReplaceAll(e.Detail, `"`, `""`)+`"`, ""),
			gox.If(e.Hint != "", `"`+strings.ReplaceAll(e.Hint, `"`, `""`)+`"`, ""),
			e.InternalQuery,
			e.InternalQueryPos,
			gox.If(e.Context != "", `"`+strings.ReplaceAll(e.Context, `"`, `""`)+`"`, ""),
			gox.If(e.Query != "", `"`+strings.ReplaceAll(e.Query, `"`, `""`)+`"`, ""),
			e.QueryPos,
			e.Location,
			gox.If(e.ApplicationName != "" || (e.SQLStateCode == "00000" && e.UserName == ""), `"`+strings.ReplaceAll(e.ApplicationName, `"`, `""`)+`"`, ""),
			gox.If(e.BackendType != "", `"`+strings.ReplaceAll(e.BackendType, `"`, `""`)+`"`, ""),
		}, ",")
	}
	return strings.Join([]string{
		e.LogTime,
		gox.If(e.UserName != "", `"`+strings.ReplaceAll(e.UserName, `"`, `""`)+`"`, ""),
		gox.If(e.DatabaseName != "", `"`+strings.ReplaceAll(e.DatabaseName, `"`, `""`)+`"`, ""),
		e.ProcessID,
		gox.If(e.ConnectionFrom != "", `"`+strings.ReplaceAll(e.ConnectionFrom, `"`, `""`)+`"`, ""),
		e.SessionID,
		e.SessionLineNum,
		gox.If(e.CommandTag != "", `"`+strings.ReplaceAll(e.CommandTag, `"`, `""`)+`"`, ""),
		e.SessionStartTime,
		e.VirtualTransactionID,
		e.TransactionID,
		e.ErrorSeverity,
		e.SQLStateCode,
		gox.If(e.Message != "", `"`+strings.ReplaceAll(e.Message, `"`, `""`)+`"`, ""),
		gox.If(e.Detail != "", `"`+strings.ReplaceAll(e.Detail, `"`, `""`)+`"`, ""),
		gox.If(e.Hint != "", `"`+strings.ReplaceAll(e.Hint, `"`, `""`)+`"`, ""),
		e.InternalQuery,
		e.InternalQueryPos,
		gox.If(e.Context != "", `"`+strings.ReplaceAll(e.Context, `"`, `""`)+`"`, ""),
		gox.If(e.Query != "", `"`+strings.ReplaceAll(e.Query, `"`, `""`)+`"`, ""),
		e.QueryPos,
		e.Location,
		gox.If(e.ApplicationName != "" || (e.SQLStateCode == "00000" && e.UserName == ""), `"`+strings.ReplaceAll(e.ApplicationName, `"`, `""`)+`"`, ""),
		gox.If(e.BackendType != "", `"`+strings.ReplaceAll(e.BackendType, `"`, `""`)+`"`, ""),
		e.LeaderPid,
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
	switch strings.ToUpper(severity) {
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

// Simplistic approach. Adding severity could help a bit
var POSTGRES_SYSTEM_MESSAGES_IDENT_PREXIFES = []string{
	"invalid value ",
	"configuration file ",
	"starting ",
	"listening on ",
	"database system ",
	"received ",
	"parameter ",
	"automatic ", // vacuum / analyze
	"autovacuum: ",
	"checkpoint ",
	"Checkpoint ",
	"sending ",
	"TimescaleDB ",
	"redo ",
	"invalid ",
	"archive ",
	"selected ",
	"consistent recovery ",
	"entering ",
	"shutting  ",
	"background worker ",
	"aborting ",
	"The failed archive command was",
	"archiving write-ahead log ",
	"Failed ",
	"out of memory",
	"terminating ",
	"server process ",
	"could not create ",
	"could not write ",
	"could not attach ",
	"could not fsync ",
	"could not access ",
	"Could not open ",
	"cannot ",
	"database ",
	"WAL redo ",
	"replication ",
	"Replication ",
}

// Case sensitive
var POSTGRES_SYSTEM_MESSAGES_IDENT_CONTAINS = []string{
	" XID",
	" corruption ",
	" wraparound ",
	" postmaster ",
}

var POSTGRES_LOG_LVL_NON_SYSTEM_MESSAGES_IDENT_PREXIFES = []string{
	"duration: ",
	"statement: ",
	"connection authorized: ",
	"connection authenticated: ",
	"connection received: ",
	"disconnection: ",
	"could not receive data from client: ",
	"could not send data to client: ",
	"AUDIT: ",
	"unexpected EOF ",
}

var POSTGRES_LOG_LVL_NON_SYSTEM_REGEXES = []*regexp.Regexp{
	regexp.MustCompile(`^process [0-9]+ acquired`),
	regexp.MustCompile(`^process [0-9]+ still waiting`),
}

var POSTGRES_SYS_FATAL_PREFIXES_TO_IGNORE = []string{
	"password authentication failed ",
	"connection to client lost",
}

func (e LogEntry) IsSystemEntry() bool {
	if e.CsvColumns != nil {
		return e.CsvColumns.UserName == ""
	}

	if e.ErrorSeverity == "FATAL" || e.ErrorSeverity == "PANIC" {
		for _, prefix := range POSTGRES_SYS_FATAL_PREFIXES_TO_IGNORE {
			if strings.HasPrefix(e.Message, prefix) {
				return false
			}
		}
		return true
	}

	for _, prefix := range POSTGRES_SYSTEM_MESSAGES_IDENT_PREXIFES {
		if strings.HasPrefix(e.Message, prefix) {
			return true
		}
	}

	for _, ident := range POSTGRES_SYSTEM_MESSAGES_IDENT_CONTAINS {
		if strings.Contains(e.Message, ident) {
			return true
		}
	}

	// TODO With plain text logs very hard to detect actually without log_line_prefix so need to use that as well
	// let's assume for now user@db
	if REGEX_USER_AT_DB.MatchString(strings.Join(e.Lines, "\n")) {
		return false
	}

	// Everything with level LOG minus "slow queries" and pgaudit
	if e.ErrorSeverity == "LOG" {
		// Check if message matches any of the regexes indicating non-system messages
		for _, regex := range POSTGRES_LOG_LVL_NON_SYSTEM_REGEXES {
			if regex.MatchString(e.Message) {
				return false
			}
		}

		for _, prefix := range POSTGRES_LOG_LVL_NON_SYSTEM_MESSAGES_IDENT_PREXIFES {
			if strings.HasPrefix(e.Message, prefix) {
				return false
			}
		}
		return true
	}

	return false
}
