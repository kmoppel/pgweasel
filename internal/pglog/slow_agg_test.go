package pglog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlowLogAggregator(t *testing.T) {
	aggregator := SlowLogStatsAggregator{
		StmtTagDurations: make(map[string][]SlowLogDurEntry),
	}

	// Test entries with durations - using realistic PostgreSQL log message patterns
	entry1 := LogEntry{
		ErrorSeverity: "LOG",
		Message:       "duration: 123.45 ms  execute <unnamed>: SELECT * FROM table",
	}

	entry2 := LogEntry{
		ErrorSeverity: "LOG",
		Message:       "duration: 456.78 ms  execute P_1: UPDATE table SET col = value",
	}

	entry3 := LogEntry{
		ErrorSeverity: "LOG",
		Message:       "statement: INSERT INTO table VALUES (1, 2, 3)",
	}

	// Add entries to aggregator
	aggregator.Add(entry1)
	aggregator.Add(entry2)
	aggregator.Add(entry3) // This should be skipped (no duration)

	// Check that SELECT was recorded
	assert.Contains(t, aggregator.StmtTagDurations, "SELECT")
	assert.Len(t, aggregator.StmtTagDurations["SELECT"], 1)
	assert.InDelta(t, 123.45, aggregator.StmtTagDurations["SELECT"][0].Duration, 0.01)

	// Check that UPDATE was recorded
	assert.Contains(t, aggregator.StmtTagDurations, "UPDATE")
	assert.Len(t, aggregator.StmtTagDurations["UPDATE"], 1)
	assert.InDelta(t, 456.78, aggregator.StmtTagDurations["UPDATE"][0].Duration, 0.01)

	// Check that INSERT was not recorded (no duration)
	assert.NotContains(t, aggregator.StmtTagDurations, "INSERT")

	// Test ShowStats (should not panic)
	aggregator.ShowStats()
}
