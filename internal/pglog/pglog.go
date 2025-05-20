package pglog

import (
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/util"
)

type LogEntry struct {
	LogTime       string
	ErrorSeverity string
	Message       string
	Lines         []string
	CsvRecords    []string
}

var REGEX_USER_AT_DB = regexp.MustCompile(`(?s)^(?P<log_time>[\d\-:\. ]+ [A-Z]+).*?[:\s]+(?P<user_name>[A-Za-z0-9_\-]+)@(?P<database_name>[A-Za-z0-9_\-]+)[:\s]+.*?(?P<error_severity>[A-Z0-9]+)[:\s]+.*$`)

// Postgres log levels are DEBUG5, DEBUG4, DEBUG3, DEBUG2, DEBUG1, INFO, NOTICE, WARNING, ERROR, LOG, FATAL, and PANIC
// but move LOG lower as too chatty otherwise
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
