package pglog_test

import (
	"testing"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/stretchr/testify/assert"
)

func TestSeverityToNum(t *testing.T) {
	assert.Greater(t, pglog.SeverityToNum("info"), pglog.SeverityToNum("DEBUG1"))
	assert.Greater(t, pglog.SeverityToNum("debug1"), pglog.SeverityToNum("debug2"))
	assert.Equal(t, 5, pglog.SeverityToNum("dbg"))
}
func TestIsSystemEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    pglog.LogEntry
		expected bool
	}{
		{
			name: "CSV system entry",
			entry: pglog.LogEntry{
				CsvColumns: &pglog.CsvEntry{LogTime: "2025-05-02 18:18:26.523 EEST", UserName: ""},
			},
			expected: true,
		},
		{
			name: "CSV user entry",
			entry: pglog.LogEntry{
				CsvColumns: &pglog.CsvEntry{LogTime: "2025-05-02 18:18:26.523 EEST", UserName: "postgres"},
			},
			expected: false,
		},
		{
			name: "Plain text system entry",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-02 18:18:26.523 EEST [2240722] LOG:  listening on IPv4 address "0.0.0.0", port 5432`},
				Message: `listening on IPv4 address "0.0.0.0", port 5432`,
			},
			expected: true,
		},
		{
			name: "Plain text non-system entry",
			entry: pglog.LogEntry{
				Lines: []string{`2025-05-02 18:25:03.959 EEST [2702612] krl@postgres LOG:  statement: vacuum pgbench_branches`},
			},
			expected: false,
		},
		{
			name: "Plain text non-system entry2",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-22 15:15:09.392 EEST [3239131] krl@postgres ERROR:  new row for relation "pgbench_accounts" violates check constraint "posbal"`},
				Message: `new row for relation "pgbench_accounts" violates check constraint "posbal"`,
			},
			expected: false,
		},
		{
			name: "Plain text system entry2",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-19 09:27:35.644 EEST [3775] LOG:  database system was not properly shut down; automatic recovery in progress`},
				Message: `database system was not properly shut down; automatic recovery in progress`,
			},
			expected: true,
		},
		{
			name: "Plain text system entry3",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-18 14:43:19.424 EEST [3807] LOG:  checkpoint starting: time`},
				Message: `checkpoint starting: time`,
			},
			expected: true,
		},
		{
			name: "Plain text system entry4",
			entry: pglog.LogEntry{
				Lines:         []string{`2021-05-28 12:19:06.386 JST [8216] LOG:  database system was shut down at 2021-05-28 12:19:06 JST`},
				Message:       `database system was shut down at 2021-05-28 12:19:06 JST`,
				ErrorSeverity: "LOG",
			},
			expected: true,
		},
		{
			name: "Plain text non-system entry3",
			entry: pglog.LogEntry{
				Lines:         []string{`2021-12-09 12:40:04.921 UTC-61b1f89a.4aa20-LOG:  process 305696 still waiting for ExclusiveLock on extension of relation 16538 of database 14344 after 1000.004 ms`},
				Message:       `process 305696 still waiting for ExclusiveLock on extension of relation 16538 of database 14344 after 1000.004 ms`,
				ErrorSeverity: "LOG",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsSystemEntry()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEventBucket(t *testing.T) {
	b := pglog.EventBucket{}
	b.Init()
	b.AddEvent(pglog.LogEntry{LogTime: "2025-05-02 18:18:26.523 EEST", ErrorSeverity: "LOG", Message: "Test message 1"}, time.Duration(5*time.Minute))
	b.AddEvent(pglog.LogEntry{LogTime: "2025-05-02 18:18:26.523 EEST", ErrorSeverity: "ERROR", Message: "Test message 2"}, time.Duration(5*time.Minute))
	b.AddEvent(pglog.LogEntry{LogTime: "2025-05-02 18:18:26.523 EEST", ErrorSeverity: "STATEMENT", Message: "Test message 3"}, time.Duration(5*time.Minute))
	assert.Equal(t, 2, b.TotalEvents, "Event bucket should contain 2 events")
	assert.Equal(t, 1, b.TotalBySeverity["ERROR"], "Event bucket should contain 1 ERROR event")
}
