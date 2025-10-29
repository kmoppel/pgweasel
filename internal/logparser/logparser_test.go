package logparser_test

import (
	"testing"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/stretchr/testify/assert"
)

var log1 = []string{`2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`}
var log2 = []string{`2025-05-02 18:25:51.151 EEST [2698052] krl@postgres STATEMENT:  select dadasdas`, `dasda`, `adsdas;`}
var log3 = []string{`2025-05-02 18:18:26.523 EEST [2240722] LOG:  listening on IPv4 address "0.0.0.0", port 5432`}
var log4 = []string{`2025-05-02 18:18:26.533 EEST [2240726] LOG:  database system was shut down at 2025-05-01 18:18:26 EEST`}

func TestFileLogger(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log1, "")
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 12:27:52.634 EEST")
	assert.Equal(t, e.ErrorSeverity, "ERROR")
	assert.Equal(t, e.Message, `column "asdasd" does not exist at character 8`)
}

func TestFileLogger2(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log2, "")
	// fmt.Printf("LogTime: %s, ErrorSeverity: %s, Message: %s\n", e.LogTime, e.ErrorSeverity, e.Message)
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 18:25:51.151 EEST")
	assert.Equal(t, e.ErrorSeverity, "STATEMENT")
	assert.Equal(t, e.Message, "select dadasdas\ndasda\nadsdas;")
}

func TestFileLogger3(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log3, "")
	assert.NoError(t, err)
	assert.Equal(t, e.LogTime, "2025-05-02 18:18:26.523 EEST")
	assert.Equal(t, e.ErrorSeverity, "LOG")
	assert.Equal(t, e.Message, `listening on IPv4 address "0.0.0.0", port 5432`)
}

func TestFileLogger4(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log4, "")
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
	assert.True(t, logparser.HasTimestampPrefix("2025-01-09 20:48:11.713 GMT LOG:  checkpoint starting: time"))
	assert.True(t, logparser.HasTimestampPrefix("2022-02-19 14:47:24 +08 [66019]: [10-1] session=6210927b.101e3,user=postgres,db=ankara,app=PostgreSQL JDBC Driver,client=localhost | LOG:  duration: 0.073 ms"))
	assert.True(t, logparser.HasTimestampPrefix("1748867052.047 [2995904] LOG:  database system is ready to accept connections"))
}

func TestPeekRecord(t *testing.T) {
	// Test with a CSV file
	csvFile := "../../testdata/csvlog1.csv"
	entry, err := logparser.PeekRecordFromFile(csvFile, nil, true)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.LogTime)
	assert.NotEmpty(t, entry.ErrorSeverity)
	assert.NotNil(t, entry.CsvColumns)

	// Test with a plain text log file
	logFile := "../../testdata/debian_default.log"
	entry, err = logparser.PeekRecordFromFile(logFile, nil, false)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.LogTime)
	assert.NotEmpty(t, entry.ErrorSeverity)
	assert.NotEmpty(t, entry.Lines)

	// Test with non-existent file
	entry, err = logparser.PeekRecordFromFile("non-existent-file.log", nil, false)
	assert.Error(t, err)
	assert.Nil(t, entry)
}

func TestEventLinesToPgLogEntryTimestampSeverityMessage(t *testing.T) {
	// Test LOG entry
	logLines1 := []string{"2025-05-02 18:18:26.544 EEST [2240722] LOG:  database system is ready to accept connections"}
	entry1, err := logparser.EventLinesToPgLogEntry(logLines1, "")
	assert.NoError(t, err)
	assert.Equal(t, "2025-05-02 18:18:26.544 EEST", entry1.LogTime)
	assert.Equal(t, "LOG", entry1.ErrorSeverity)
	assert.Equal(t, "database system is ready to accept connections", entry1.Message)
	assert.Equal(t, logLines1, entry1.Lines)

	// Test ERROR entry
	logLines2 := []string{"2025-05-02 18:18:26.555 EEST [2698052] krl@postgres ERROR:  column \"xxxx\" does not exist at character 8"}
	entry2, err := logparser.EventLinesToPgLogEntry(logLines2, "")
	assert.NoError(t, err)
	assert.Equal(t, "2025-05-02 18:18:26.555 EEST", entry2.LogTime)
	assert.Equal(t, "ERROR", entry2.ErrorSeverity)
	assert.Equal(t, "column \"xxxx\" does not exist at character 8", entry2.Message)
	assert.Equal(t, logLines2, entry2.Lines)
}
