package pglog_test

import (
	"testing"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestSeverityToNum(t *testing.T) {
	assert.Greater(t, pglog.SeverityToNum("info"), pglog.SeverityToNum("DEBUG1"))
	assert.Greater(t, pglog.SeverityToNum("debug1"), pglog.SeverityToNum("debug2"))
	assert.Equal(t, -1, pglog.SeverityToNum("dbg"))
	assert.Equal(t, -1, pglog.SeverityToNum("HINT"))
}
func TestIsSystemEntry(t *testing.T) {
	tests := []struct {
		name                      string
		entry                     pglog.LogEntry
		systemIncludeCheckpointer bool
		expected                  bool
	}{
		{
			name: "CSV system entry",
			entry: pglog.LogEntry{
				CsvColumns: &pglog.CsvEntry{LogTime: "2025-05-02 18:18:26.523 EEST", UserName: ""},
			},
			systemIncludeCheckpointer: true,
			expected:                  true,
		},
		{
			name: "CSV user entry",
			entry: pglog.LogEntry{
				CsvColumns: &pglog.CsvEntry{LogTime: "2025-05-02 18:18:26.523 EEST", UserName: "postgres"},
			},
			systemIncludeCheckpointer: true,
			expected:                  false,
		},
		{
			name: "Plain text system entry",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-02 18:18:26.523 EEST [2240722] LOG:  listening on IPv4 address "0.0.0.0", port 5432`},
				Message: `listening on IPv4 address "0.0.0.0", port 5432`,
			},
			systemIncludeCheckpointer: true,
			expected:                  true,
		},
		{
			name: "Plain text non-system entry",
			entry: pglog.LogEntry{
				Lines: []string{`2025-05-02 18:25:03.959 EEST [2702612] krl@postgres LOG:  statement: vacuum pgbench_branches`},
			},
			systemIncludeCheckpointer: true,
			expected:                  false,
		},
		{
			name: "Plain text non-system entry2",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-22 15:15:09.392 EEST [3239131] krl@postgres ERROR:  new row for relation "pgbench_accounts" violates check constraint "posbal"`},
				Message: `new row for relation "pgbench_accounts" violates check constraint "posbal"`,
			},
			systemIncludeCheckpointer: true,
			expected:                  false,
		},
		{
			name: "Plain text system entry2",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-19 09:27:35.644 EEST [3775] LOG:  database system was not properly shut down; automatic recovery in progress`},
				Message: `database system was not properly shut down; automatic recovery in progress`,
			},
			systemIncludeCheckpointer: true,
			expected:                  true,
		},
		{
			name: "Plain text system entry3",
			entry: pglog.LogEntry{
				Lines:   []string{`2025-05-18 14:43:19.424 EEST [3807] LOG:  checkpoint starting: time`},
				Message: `checkpoint starting: time`,
			},
			systemIncludeCheckpointer: true,
			expected:                  true,
		},
		{
			name: "Plain text system entry4",
			entry: pglog.LogEntry{
				Lines:         []string{`2021-05-28 12:19:06.386 JST [8216] LOG:  database system was shut down at 2021-05-28 12:19:06 JST`},
				Message:       `database system was shut down at 2021-05-28 12:19:06 JST`,
				ErrorSeverity: "LOG",
			},
			systemIncludeCheckpointer: true,
			expected:                  true,
		},
		{
			name: "Plain text non-system entry3",
			entry: pglog.LogEntry{
				Lines:         []string{`2021-12-09 12:40:04.921 UTC-61b1f89a.4aa20-LOG:  process 305696 still waiting for ExclusiveLock on extension of relation 16538 of database 14344 after 1000.004 ms`},
				Message:       `process 305696 still waiting for ExclusiveLock on extension of relation 16538 of database 14344 after 1000.004 ms`,
				ErrorSeverity: "LOG",
			},
			systemIncludeCheckpointer: true,
			expected:                  false,
		},
		{
			name: "Plain text non-system entry4",
			entry: pglog.LogEntry{
				Lines:         []string{`2022-03-11 09:42:32.449 UTC [17504] ERROR:  cannot execute UPDATE in a read-only transaction`},
				Message:       `cannot execute UPDATE in a read-only transaction`,
				ErrorSeverity: "ERROR",
			},
			systemIncludeCheckpointer: true,
			expected:                  false,
		},
		{
			name: "Checkpoint entry with systemIncludeCheckpointer false",
			entry: pglog.LogEntry{
				Lines:         []string{`2025-05-21 10:39:58.024 UTC [39]: [1-1] db=,user=,host= LOG:  checkpoint starting: time`},
				Message:       `checkpoint starting: time`,
				ErrorSeverity: "LOG",
			},
			systemIncludeCheckpointer: false,
			expected:                  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsSystemEntry(tt.systemIncludeCheckpointer)
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
func TestIsLockingRelatedEntry(t *testing.T) {
	tests := []pglog.LogEntry{
		pglog.LogEntry{
			Message: "process 3634152 acquired ShareLock on transaction 280767 after 5.016 ms",
		},
		pglog.LogEntry{
			Message: "ERROR:  deadlock detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Message, func(t *testing.T) {
			assert.True(t, tt.IsLockingRelatedEntry())
		})
	}
}
func TestHistogramBucketGetSortedBuckets(t *testing.T) {
	tests := []struct {
		name             string
		buckets          map[time.Time]int
		bucketWith       time.Duration
		expectedCount    int
		expectedFirstVal int
		expectedLastVal  int
	}{
		{
			name:          "Empty buckets",
			buckets:       map[time.Time]int{},
			bucketWith:    time.Minute,
			expectedCount: 0,
		},
		{
			name: "Single bucket",
			buckets: map[time.Time]int{
				time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC): 5,
			},
			bucketWith:       time.Minute,
			expectedCount:    1,
			expectedFirstVal: 5,
			expectedLastVal:  5,
		},
		{
			name: "Sequential buckets",
			buckets: map[time.Time]int{
				time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC): 5,
				time.Date(2025, 5, 2, 10, 1, 0, 0, time.UTC): 10,
				time.Date(2025, 5, 2, 10, 2, 0, 0, time.UTC): 15,
			},
			bucketWith:       time.Minute,
			expectedCount:    3,
			expectedFirstVal: 5,
			expectedLastVal:  15,
		},
		{
			name: "Sparse buckets",
			buckets: map[time.Time]int{
				time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC): 5,
				time.Date(2025, 5, 2, 10, 5, 0, 0, time.UTC): 10,
			},
			bucketWith:       time.Minute,
			expectedCount:    6,
			expectedFirstVal: 5,
			expectedLastVal:  10,
		},
		{
			name: "Hourly buckets",
			buckets: map[time.Time]int{
				time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC): 100,
				time.Date(2025, 5, 2, 11, 0, 0, 0, time.UTC): 200,
				time.Date(2025, 5, 2, 12, 0, 0, 0, time.UTC): 300,
			},
			bucketWith:       time.Hour,
			expectedCount:    3,
			expectedFirstVal: 100,
			expectedLastVal:  300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := pglog.HistogramBucket{}
			h.Init(tt.bucketWith)
			h.CountBuckets = tt.buckets

			result := h.GetSortedBuckets()

			if tt.expectedCount == 0 {
				assert.Nil(t, result, "Expected nil result for empty buckets")
				return
			}

			assert.Equal(t, tt.expectedCount, len(result), "Unexpected number of buckets")
			assert.Equal(t, tt.expectedFirstVal, result[0].Count, "First bucket has unexpected count")
			assert.Equal(t, tt.expectedLastVal, result[len(result)-1].Count, "Last bucket has unexpected count")

			// Check time ordering
			for i := 1; i < len(result); i++ {
				assert.True(t, result[i-1].Time.Before(result[i].Time) || result[i-1].Time.Equal(result[i].Time),
					"Buckets are not in chronological order")
			}

			// Check if time difference between consecutive buckets equals bucketWith
			if len(result) > 1 {
				timeDiff := result[1].Time.Sub(result[0].Time)
				assert.Equal(t, tt.bucketWith, timeDiff, "Bucket time intervals are not consistent")
			}
		})
	}
}
func TestHistogramBucketAddMinErrorLevel(t *testing.T) {
	testData := []pglog.LogEntry{
		pglog.LogEntry{LogTime: "2025-05-02 10:00:00 UTC", ErrorSeverity: "DEBUG1", Message: "Debug message"},
		pglog.LogEntry{LogTime: "2025-05-02 10:01:00 UTC", ErrorSeverity: "INFO", Message: "Info message"},
		pglog.LogEntry{LogTime: "2025-05-02 10:02:00 UTC", ErrorSeverity: "NOTICE", Message: "Notice message"},
		pglog.LogEntry{LogTime: "2025-05-02 10:03:00 UTC", ErrorSeverity: "WARNING", Message: "Warning message"},
		pglog.LogEntry{LogTime: "2025-05-02 10:04:00 UTC", ErrorSeverity: "ERROR", Message: "Error message"},
	}
	bucketInterval := time.Minute
	minErrLvlSeverityNum := pglog.SeverityToNum("WARNING")
	h := pglog.HistogramBucket{}
	h.Init(bucketInterval)

	for _, entry := range testData {
		h.Add(entry, bucketInterval, minErrLvlSeverityNum)
	}

	assert.Equal(t, 2, h.TotalEvents, "TotalEvents count doesn't match expected value")
}
func TestExtractConnectUserDbAppnameSslFromLogMessage(t *testing.T) {
	msg := "connection authorized: user=postgres database=postgres application_name=x\nERR:"
	user, db, app, ssl := util.ExtractConnectUserDbAppnameSslFromLogMessage(msg)
	assert.Equal(t, "postgres", user, "User mismatch")
	assert.Equal(t, "postgres", db, "Database mismatch")
	assert.Equal(t, "x", app, "Application name mismatch")
	assert.Equal(t, false, ssl, "SSL flag mismatch")
}

func TestGetCommandTag(t *testing.T) {
	tests := []struct {
		name     string
		entry    pglog.LogEntry
		expected string
	}{
		{
			name: "CSV entry with CommandTag",
			entry: pglog.LogEntry{
				CsvColumns: &pglog.CsvEntry{
					CommandTag: "SELECT",
				},
			},
			expected: "SELECT",
		},
		{
			name: "Statement command pattern",
			entry: pglog.LogEntry{
				Message: "statement: UPDATE pgbench_accounts SET balance = 123",
			},
			expected: "UPDATE",
		},
		{
			name: "Duration execute command pattern",
			entry: pglog.LogEntry{
				Message: "duration: 41147.417 ms execute <unnamed>: SELECT id, name FROM users",
			},
			expected: "SELECT",
		},
		{
			name: "Execute command pattern",
			entry: pglog.LogEntry{
				Message: "execute P_1: UPDATE pgbench_accounts SET balance = 456",
			},
			expected: "UPDATE",
		},
		{
			name: "Multi-line with Query Text on next line",
			entry: pglog.LogEntry{
				Message: "duration: 7621.082 ms  plan:",
				Lines: []string{
					"duration: 7621.082 ms  plan:",
					"\tQuery Text: SELECT xxx",
				},
			},
			expected: "SELECT",
		},
		{
			name: "No command found",
			entry: pglog.LogEntry{
				Message: "some random log message",
			},
			expected: "",
		},
		{
			name: "ANALYZE entry",
			entry: pglog.LogEntry{
				Message: "duration: 113351.741 ms  statement: ANALYZE VERBOSE",
			},
			expected: "ANALYZE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.GetCommandTag()
			assert.Equal(t, tt.expected, result)
		})
	}
}
