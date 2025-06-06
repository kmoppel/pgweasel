package pglog

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/beorn7/perks/quantile"
	"github.com/icza/gox/gox"
	"github.com/kmoppel/pgweasel/internal/util"
)

var ERROR_SEVERITIES = []string{"WARNING", "ERROR", "FATAL", "PANIC"}
var ALL_SEVERITIES = []string{"DEBUG5", "DEBUG4", "DEBUG3", "DEBUG2", "DEBUG1", "INFO", "NOTICE", "WARNING", "ERROR", "LOG", "FATAL", "PANIC"}
var ALL_SEVERITIES_MAP = map[string]bool{
	"DEBUG5":  true,
	"DEBUG4":  true,
	"DEBUG3":  true,
	"DEBUG2":  true,
	"DEBUG1":  true,
	"INFO":    true,
	"NOTICE":  true,
	"WARNING": true,
	"ERROR":   true,
	"LOG":     true,
	"FATAL":   true,
	"PANIC":   true,
}

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

// For "peaks"
type EventBucket struct {
	BucketsBySeverity    map[string]map[time.Time]int
	TotalEvents          int
	TotalBySeverity      map[string]int
	LockEvents           map[time.Time]int
	ConnectEvents        map[time.Time]int
	BucketActualLogTimes map[time.Time]string // map with first actual LogTime string for each time bucket for display purposes
}

type StatsAggregator struct {
	TotalEvents                   int
	TotalEventsBySeverity         map[string]int
	FirstEventTime                time.Time
	LastEventTime                 time.Time
	Connections                   int
	Disconnections                int
	SlowQueries                   int
	QueryTimesHistogram           *quantile.Stream
	CheckpointsTimed              int
	CheckpointsForced             int
	LongestCheckpointSeconds      float64
	Autovacuums                   int
	AutovacuumMaxDurationSeconds  float64
	AutovacuumMaxDurationTable    string
	Autoanalyzes                  int
	AutoanalyzeMaxDurationSeconds float64
	AutoanalyzeMaxDurationTable   string
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

func (e LogEntry) String() string {
	if e.CsvColumns != nil {
		return e.CsvColumns.String()
	} else {
		return strings.Join(e.Lines, "\n")
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
	"cache lookup ",
	"function ",
	"requested ",
	"unrecognized ",
	"internal error",
	"restartpoint ",
	"was at log time",
	"recovery ",
}

// Case sensitive
var POSTGRES_SYSTEM_MESSAGES_IDENT_CONTAINS = []string{
	" XID",
	" corruption ",
	" wraparound ",
	" postmaster ",
	" configuration file ",
	" relfrozenxid ",
	"multixact",
	"MultiXact",
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

var NON_SYSTEM_REGEXES = []*regexp.Regexp{
	regexp.MustCompile(`^cannot execute \w+ in a read-only transaction`),
}

var POSTGRES_SYS_FATAL_PREFIXES_TO_IGNORE = []string{
	"password authentication failed ",
	"connection to client lost",
}

// Case sensitive
var LOCKING_RELATED_MESSAGE_CONTAINS_LIST = []string{
	" conflicts ",
	" conflicting ",
	" still waiting for ",
	"Wait queue:",
	"while locking tuple",
	"while updating tuple",
	"conflict detected",
	"deadlock detected",
	"buffer deadlock",
	"blocked by process ",
	"recovery conflict ",
	" concurrent update",
	"could not serialize",
	"could not obtain ",
	"lock on relation ",
	"cannot lock rows",
	" semaphore:",
}

var LOCKING_RELATED_MESSAGE_REGEXES = []*regexp.Regexp{
	regexp.MustCompile(`^process [0-9]+ acquired`),
}

func (e LogEntry) IsSystemEntry() bool {
	if e.ErrorSeverity == "PANIC" {
		return true
	}

	if e.CsvColumns != nil {
		if strings.HasPrefix(e.CsvColumns.Message, "connection ") {
			return false
		}
		return e.CsvColumns.UserName == "" // TODO re-check that this assumption is correct
	}

	if e.ErrorSeverity == "FATAL" {
		for _, prefix := range POSTGRES_SYS_FATAL_PREFIXES_TO_IGNORE {
			if strings.HasPrefix(e.Message, prefix) {
				return false
			}
		}
		return true
	}

	for _, regex := range NON_SYSTEM_REGEXES {
		if regex.MatchString(e.Message) {
			return false
		}
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

func (e LogEntry) IsLockingRelatedEntry() bool {
	for _, ident := range LOCKING_RELATED_MESSAGE_CONTAINS_LIST {
		if strings.Contains(e.Message, ident) {
			return true
		}
	}

	for _, regex := range LOCKING_RELATED_MESSAGE_REGEXES {
		if regex.MatchString(e.Message) {
			return true
		}
	}

	return false
}

func (b *EventBucket) Init() {
	b.BucketsBySeverity = make(map[string]map[time.Time]int)

	// Initialize map entry for each severity level
	for _, severity := range ALL_SEVERITIES {
		b.BucketsBySeverity[severity] = make(map[time.Time]int)
	}
	b.TotalBySeverity = make(map[string]int)
	b.LockEvents = make(map[time.Time]int)
	b.ConnectEvents = make(map[time.Time]int)
	b.BucketActualLogTimes = make(map[time.Time]string) // initialize
}

func (b *EventBucket) AddEvent(e LogEntry, bucketInterval time.Duration) {
	if b.BucketsBySeverity == nil {
		panic("BucketsBySeverity is nil, call Init()")
	}
	// Extra context not considered a separate event here
	if !ALL_SEVERITIES_MAP[e.ErrorSeverity] {
		return
	}

	bucketTime := e.GetTime().Truncate(bucketInterval)

	// Record the first LogTime string for this bucketTime
	if _, exists := b.BucketActualLogTimes[bucketTime]; !exists {
		b.BucketActualLogTimes[bucketTime] = e.LogTime
	}

	b.BucketsBySeverity[e.ErrorSeverity][bucketTime]++
	b.TotalBySeverity[e.ErrorSeverity]++
	b.TotalEvents++
	if e.IsLockingRelatedEntry() {
		b.LockEvents[bucketTime]++
	}
	if strings.HasPrefix(e.Message, "connection received") {
		b.ConnectEvents[bucketTime]++
	}

}

func (b *EventBucket) GetTopBucketsBySeverity() map[string]map[time.Time]int {
	ret := make(map[string]map[time.Time]int)

	for severity, bucket := range b.BucketsBySeverity {
		if len(bucket) == 0 {
			continue
		}

		var maxCount int
		var maxTime time.Time

		// Find the highest count for this severity
		for bucketTime, eventCount := range bucket {
			if eventCount > maxCount {
				maxCount = eventCount
				maxTime = bucketTime
			}
		}

		// Only include non-zero entries
		if maxCount > 0 {
			ret[severity] = map[time.Time]int{maxTime: maxCount}
			// log.Debug().Msgf("Top bucket for %s: %s with %d events", severity, maxTime.Format(time.RFC3339), maxCount)
		}
	}
	return ret
}

// Returns the top LockEvents period
func (b *EventBucket) GetTopLockingPeriod() (time.Time, int, string) {
	if len(b.LockEvents) == 0 {
		return time.Time{}, 0, ""
	}

	var maxTime time.Time
	var maxCount int

	for bucketTime, count := range b.LockEvents {
		if count > maxCount {
			maxCount = count
			maxTime = bucketTime
		}
	}

	return maxTime, maxCount, b.BucketActualLogTimes[maxTime]
}

func (b *EventBucket) GetFirstRealTimeStringForBucket(bucket time.Time) string {
	return b.BucketActualLogTimes[bucket]
}

func (b *EventBucket) GetTopConnectPeriod() (time.Time, int, string) {
	if len(b.ConnectEvents) == 0 {
		return time.Time{}, 0, ""
	}

	var maxTime time.Time
	var maxCount int

	for bucketTime, count := range b.ConnectEvents {
		if count > maxCount {
			maxCount = count
			maxTime = bucketTime
		}
	}

	return maxTime, maxCount, b.BucketActualLogTimes[maxTime]
}

func (sa *StatsAggregator) Init() {
	sa.TotalEventsBySeverity = make(map[string]int)
	sa.QueryTimesHistogram = quantile.NewTargeted(map[float64]float64{
		0.50: 0.001,
		0.90: 0.001,
		0.99: 0.001,
	})
}

func (sa *StatsAggregator) AddEvent(e LogEntry) {
	if sa.TotalEventsBySeverity == nil {
		panic("Call Init() first")
	}
	// Extra context not considered a separate event here
	if !ALL_SEVERITIES_MAP[e.ErrorSeverity] {
		return
	}

	et := e.GetTime()
	if sa.FirstEventTime.IsZero() || et.Before(sa.FirstEventTime) {
		sa.FirstEventTime = et
	}
	if sa.LastEventTime.IsZero() || et.After(sa.LastEventTime) {
		sa.LastEventTime = et
	}

	sa.TotalEventsBySeverity[e.ErrorSeverity]++
	sa.TotalEvents++

	if strings.HasPrefix(e.Message, "connection received") {
		sa.Connections++
	}
	if strings.HasPrefix(e.Message, "disconnection:") {
		sa.Disconnections++
	}
	if e.ErrorSeverity == "LOG" {
		// Query durations / quantiles
		if strings.Contains(e.Message, "duration: ") && strings.Contains(e.Message, " ms") && !(strings.Contains(e.Message, " bind ") || strings.Contains(e.Message, " parse ")) {
			durMs := util.ExtractDurationMillisFromLogMessage(e.Message)
			if durMs > 0 {
				sa.QueryTimesHistogram.Insert(durMs)
				sa.SlowQueries++
			}
		}
		// Checkpoints
		if strings.HasPrefix(e.Message, "checkpoint starting: ") { // include also restartpoints?
			if strings.Contains(e.Message, "starting: time") {
				sa.CheckpointsTimed++
			} else {
				sa.CheckpointsForced++
			}
		}
		if strings.HasPrefix(e.Message, "checkpoint complete: ") {
			durSeconds := util.ExtractCheckpointDurationSecondsFromLogMessage(e.Message)
			if durSeconds > sa.LongestCheckpointSeconds {
				sa.LongestCheckpointSeconds = durSeconds
			}
		}
		// Autovacuum and autoanalyze
		if strings.HasPrefix(e.Message, "automatic analyze") {
			sa.Autoanalyzes++
			durSeconds, tbl := util.ExtractAutovacuumOrAnalyzeDurationSecondsFromLogMessage(e.Message)
			if durSeconds > sa.LongestCheckpointSeconds {
				sa.AutoanalyzeMaxDurationSeconds = durSeconds
				sa.AutoanalyzeMaxDurationTable = tbl
			}
		}
		if strings.HasPrefix(e.Message, "automatic vacuum") {
			sa.Autovacuums++
			durSeconds, tbl := util.ExtractAutovacuumOrAnalyzeDurationSecondsFromLogMessage(e.Message)
			if durSeconds > sa.LongestCheckpointSeconds {
				sa.AutovacuumMaxDurationSeconds = durSeconds
				sa.AutovacuumMaxDurationTable = tbl
			}
		}
	}
}

func (sa *StatsAggregator) ShowStats() {
	eventsPerMinute := float64(sa.TotalEvents) / (sa.LastEventTime.Sub(sa.FirstEventTime).Minutes())
	fmt.Println("Total events:", sa.TotalEvents, fmt.Sprintf("(%.2f events/minute)", eventsPerMinute))
	for severity, count := range sa.TotalEventsBySeverity {
		if count > 0 {
			fmt.Printf("%s events: %d (%.1f%%)\n", severity, count, float64(count)/float64(sa.TotalEvents)*100)
		} else {
			fmt.Printf("%s events: %d\n", severity, count)
		}
	}
	fmt.Println("First event time:", sa.FirstEventTime)
	fmt.Println("Last event time:", sa.LastEventTime)

	var connectsPerMinute float64
	timeDiffMinutes := sa.LastEventTime.Sub(sa.FirstEventTime).Minutes()
	if timeDiffMinutes > 0 {
		connectsPerMinute = float64(sa.Connections) / timeDiffMinutes
	}
	fmt.Println("Total connections:", sa.Connections, fmt.Sprintf("(%.2f connections/minute)", connectsPerMinute))
	fmt.Println("Total disconnections:", sa.Disconnections)
	if sa.QueryTimesHistogram != nil {
		fmt.Println("Query times histogram:")
		for _, q := range []float64{0.50, 0.90, 0.99} {
			value := sa.QueryTimesHistogram.Query(q)
			fmt.Printf("  %.2f quantile: %.2f ms\n", q, value)
		}
	} else {
		fmt.Println("No query times histogram available")
	}
	fmt.Println("Query durations records:", sa.SlowQueries)
	fmt.Println("Checkpoints timed:", sa.CheckpointsTimed)
	fmt.Println("Checkpoints forced:", sa.CheckpointsForced)
	fmt.Printf("Longest checkpoint duration: %.3f s\n", sa.LongestCheckpointSeconds) // TODO show duration "as is" ?
	fmt.Println("Autovacuums:", sa.Autovacuums)
	fmt.Printf("Longest autovacuum duration: %.3f s (on \"%s\")\n", sa.AutovacuumMaxDurationSeconds, sa.AutovacuumMaxDurationTable)
	fmt.Println("Autoanalyzes:", sa.Autoanalyzes)
	fmt.Printf("Longest autoanalyze duration: %.3f s (on \"%s\")\n", sa.AutoanalyzeMaxDurationSeconds, sa.AutoanalyzeMaxDurationTable)
}
