package logparser_test

import (
	"log"
	"testing"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/stretchr/testify/assert"
)

var log1 = []string{`2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`}
var log2 = []string{`2025-05-02 18:25:51.151 EEST [2698052] krl@postgres STATEMENT:  select dadasdas
	dasda
	adsdas;`}
var log3 = []string{`2025-05-02 18:18:26.523 EEST [2240722] LOG:  listening on IPv4 address "0.0.0.0", port 5432`}
var log4 = []string{`2025-05-02 18:18:26.533 EEST [2240726] LOG:  database system was shut down at 2025-05-01 18:18:26 EEST`}

func TestFileLogger(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log1, logparser.DEFAULT_REGEX, "")
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 12:27:52.634 EEST")
	assert.Equal(t, e.ErrorSeverity, "ERROR")
	assert.Equal(t, e.Message, `column "asdasd" does not exist at character 8`)
}

func TestFileLogger2(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log2, logparser.DEFAULT_REGEX, "")
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 18:25:51.151 EEST")
	assert.Equal(t, e.ErrorSeverity, "STATEMENT")
	assert.Equal(t, e.Message, `select dadasdas
	dasda
	adsdas;`)
}

func TestFileLogger3(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log3, logparser.DEFAULT_REGEX, "")
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 18:18:26.523 EEST")
	assert.Equal(t, e.ErrorSeverity, "LOG")
	assert.Equal(t, e.Message, `listening on IPv4 address "0.0.0.0", port 5432`)
}

func TestFileLogger4(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log4, logparser.DEFAULT_REGEX, "")
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 18:18:26.533 EEST")
	assert.Equal(t, e.ErrorSeverity, "LOG")
	assert.Equal(t, e.Message, `database system was shut down at 2025-05-01 18:18:26 EEST`)
}

func TestHasTimestampPrefix(t *testing.T) {
	assert.True(t, logparser.HasTimestampPrefix("2025-05-02 12:27:52.634 EEST [2380404]"))
	assert.False(t, logparser.HasTimestampPrefix("bla 2025-05-02 12:27:52.634 EEST [2380404]"))
	assert.True(t, logparser.HasTimestampPrefix("2025-05-05 06:00:51 UTC:90.190.32.92(32890):postgres@postgres:[1315]:LOG:  statement: BEGIN;"))
	assert.False(t, logparser.HasTimestampPrefix("    ON CONFLICT (id) DO UPDATE SET master_time = (now() at time zone 'utc');"))
	assert.True(t, logparser.HasTimestampPrefix("May 30 11:03:43 i13400f postgres[693826]: [5-1] 2025-05-30 11:03:43.622 EEST [693826] LOG:  database system is ready to accept connections"))
	assert.True(t, logparser.HasTimestampPrefix("2025-01-09 20:48:11.713 GMT LOG:  checkpoint starting: time"))
	assert.True(t, logparser.HasTimestampPrefix("2022-02-19 14:47:24 +08 [66019]: [10-1] session=6210927b.101e3,user=postgres,db=ankara,app=PostgreSQL JDBC Driver,client=localhost | LOG:  duration: 0.073 ms"))
	assert.True(t, logparser.HasTimestampPrefix("1748867052.047 [2995904] LOG:  database system is ready to accept connections"))
}

func TestDefaultRegex(t *testing.T) {
	regex := logparser.DEFAULT_REGEX
	expectedGroups := []string{"log_time", "error_severity", "message"}

	testLines := []string{
		"2024-05-07 10:22:13 UTC [12345]: [1-1] user=admin,db=exampledb,app=psql LOG:  connection received: host=203.0.113.45 port=52344",
		"2025-05-21 15:09:59.648 EEST [3284734] STATEMENT:  asdasd",
		"2025-05-02 18:25:03.976 EEST [2702613] krl@postgres LOG:  statement: BEGIN;",
	}

	for _, logLine := range testLines {
		log.Println("Testing log line:", logLine)
		match := regex.FindStringSubmatch(logLine)
		if match == nil {
			t.Fatalf("FALLBACK_REGEX did not match the log line: %s", logLine)
		}

		result := make(map[string]string)
		for i, name := range regex.SubexpNames() {
			if i > 0 && name != "" {
				result[name] = match[i]
			}
		}

		for _, eg := range expectedGroups {
			grp, ok := result[eg]
			log.Println("grp", eg, ":", grp)
			if !ok || grp == "" {
				t.Errorf("Empty group %s for line: %s", eg, logLine)
			}
		}
	}
}
func TestExtractDurationMillisFromLogMessage(t *testing.T) {
	tests := []struct {
		message        string
		expectedMillis float64
	}{
		{
			message:        "duration: 123 ms statement: SELECT * FROM table",
			expectedMillis: 123,
		},
		{
			message:        "2025-05-16 14:26:01.872 UTC [3076] LOG:  duration: 18.237 ms",
			expectedMillis: 18.237,
		},
		{
			message:        "LOG: statement executed, duration: 5.678 ms",
			expectedMillis: 5.678,
		},
		{
			message:        "LOG: statement executed without timing info",
			expectedMillis: 0,
		},
	}

	for _, tt := range tests {
		millis := logparser.ExtractDurationMillisFromLogMessage(tt.message)
		assert.InDelta(t, tt.expectedMillis, millis, 0.1, "Duration should be within 0.1ms of expected value")
	}
}
