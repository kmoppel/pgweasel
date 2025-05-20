package pglog_test

import (
	"testing"

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
				CsvRecords: []string{"2025-05-02 18:18:26.523 EEST", ""},
			},
			expected: true,
		},
		{
			name: "CSV user entry",
			entry: pglog.LogEntry{
				CsvRecords: []string{"2025-05-02 18:18:26.523 EEST", "postgres"},
			},
			expected: false,
		},
		{
			name: "Plain text system entry",
			entry: pglog.LogEntry{
				Lines: []string{`2025-05-02 18:18:26.523 EEST [2240722] LOG:  listening on IPv4 address "0.0.0.0", port 5432`},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsSystemEntry()
			assert.Equal(t, tt.expected, result)
		})
	}
}
