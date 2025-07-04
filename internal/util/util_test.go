package util_test

import (
	"log"
	"testing"
	"time"

	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestHumanTimedeltaToTime(t *testing.T) {
	now := time.Now()
	year, month, day := now.Date()

	tests := []struct {
		input    string
		expected time.Time
	}{
		{"-2h", now.Add(time.Duration(-2) * time.Hour)},
		{"2h", now.Add(time.Duration(-2) * time.Hour)},
		{"-10m", now.Add(time.Duration(-10) * time.Minute)},
		{"5m", now.Add(time.Duration(-5) * time.Minute)},
		{"-48h", now.Add(time.Duration(-48) * time.Hour)},
		{"-30s", now.Add(time.Duration(-30) * time.Second)},
		{"1 hour ago", now.Add(time.Duration(-1) * time.Hour)},
		{"1d", now.Add(time.Duration(-24) * time.Hour)},
		{"today", time.Date(year, month, day, 0, 0, 0, 0, now.Location())},
	}

	for _, tt := range tests {
		got, err := util.HumanTimeOrDeltaStringToTime(tt.input, now)
		assert.NoError(t, err, "should not error for input %s", tt.input)
		// Allow a small delta for roundings
		assert.InDelta(t, tt.expected.UnixMilli(), got.UnixMilli(), 100, "unexpected time delta for input %s", tt.input)
	}
}

func TestHumanTimedeltaToTime_TimestampInput(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Time
	}{
		// Added a test case for date-only input "2025-05-08"
		{"6 July 2025", time.Date(2025, 07, 06, 0, 0, 0, 0, time.Local)},
		{"2025-05-08", time.Date(2025, 5, 8, 0, 0, 0, 0, time.Local)},
		{"2025-05-08 12:25:47.010 UTC", time.Date(2025, 5, 8, 12, 25, 47, 10*1e6, time.FixedZone("UTC", 0))},
		{"2019-10-21 12:03:42.567 EEST", time.Date(2019, 10, 21, 12, 3, 42, 567*1e6, time.FixedZone("EEST", 3*3600))},
		{"2019-10-21 12:03:42 EEST", time.Date(2019, 10, 21, 12, 3, 42, 0, time.FixedZone("EEST", 3*3600))},
	}

	for _, tt := range tests {
		got, err := util.HumanTimeOrDeltaStringToTime(tt.input, now)
		assert.NoError(t, err, "should not error for input %s", tt.input)
		// Allow a small delta for roundings
		assert.InDelta(t, tt.expected.UnixMilli(), got.UnixMilli(), 100, "unexpected time delta for input %s", tt.input)
	}
}
func TestIntervalToMillis(t *testing.T) {
	tests := []struct {
		input       string
		expected    int
		expectError bool
	}{
		{"1s", 1000, false},
		{"2m", 2 * 60 * 1000, false},
		{"3h", 3 * 60 * 60 * 1000, false},
		{"500ms", 500, false},
		{"1min", 1 * 60 * 1000, false},
		{"2mins", 2 * 60 * 1000, false},
		{"10sec", 10000, false},
		{"5secs", 5000, false},
		{"1hr", 3600000, false},
		{"2hrs", 7200000, false},
		{"100", 100, false},
		{"abc", 0, true},
		{"1d", 0, true}, // "d" is not supported by time.ParseDuration
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := util.IntervalToMillis(tt.input)
		if tt.expectError {
			assert.Error(t, err, "expected error for input %s", tt.input)
		} else {
			assert.NoError(t, err, "unexpected error for input %s", tt.input)
			assert.Equal(t, tt.expected, got, "unexpected millis for input %s", tt.input)
		}
	}
}

func TestTimestringToTime(t *testing.T) {
	ts := util.TimestringToTime("2025-05-02 12:27:52.634 EEST")
	assert.Equal(t, ts.Year(), 2025)
	assert.Equal(t, ts.Month(), time.Month(5))
	ts2 := util.TimestringToTime("1748867052.047")
	assert.Equal(t, ts2.Year(), 2025)
	assert.Equal(t, ts2.Month(), time.Month(6))
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
		millis := util.ExtractDurationMillisFromLogMessage(tt.message)
		assert.InDelta(t, tt.expectedMillis, millis, 0.1, "Duration should be within 0.1ms of expected value")
	}
}
func TestExtractCheckpointDurationSecondsFromLogMessage(t *testing.T) {
	tests := []struct {
		message          string
		expectedDuration float64
	}{
		{
			message:          "checkpoint complete: wrote 66 buffers (0.4%); 0 WAL file(s) added, 0 removed, 0 recycled; write=6.468 s, sync=0.036 s, total=6.517 s; sync files=48, longest=0.009 s, average=0.001 s; distance=152 kB, estimate=152 kB",
			expectedDuration: 6.517,
		},
		{
			message:          "checkpoint complete: wrote 1524 buffers (9.3%); 0 WAL file(s) added, 0 removed, 1 recycled; write=0.091 s, sync=0.008 s, total=0.184 s; sync files=12, longest=0.003 s, average=0.001 s; distance=32823 kB, estimate=32823 kB",
			expectedDuration: 0.184,
		},
		{
			message:          "checkpoint starting: time",
			expectedDuration: 0,
		},
	}

	for _, tt := range tests {
		duration := util.ExtractCheckpointDurationSecondsFromLogMessage(tt.message)
		assert.InDelta(t, tt.expectedDuration, duration, 0.001, "Duration should be within 0.001s of expected value for message: %s", tt.message)
	}
}

func TestExtractAutovacuumOrAnalyzeDurationSecondsFromLogMessage(t *testing.T) {
	msgVacuum := `automatic vacuum of table "rbcc-postgres.public.event_log": index scans: 0
	      pages: 165509 removed, 4251301 remain, 1468607 scanned (33.25% of total)
	      tuples: 11539400 removed, 272957578 remain, 0 are dead but not yet removable
	      removable cutoff: 129452, which was 1 XIDs old when operation ended
	      frozen: 107 pages from table (0.00% of total) had 3631 tuples frozen
	      index scan not needed: 166775 pages from table (3.78% of total) had 11539400 dead item identifiers removed
	      avg read rate: 5.492 MB/s, avg write rate: 4.932 MB/s
	      buffer usage: 1468979 hits, 1635272 misses, 1468726 dirtied
	      WAL usage: 1802239 records, 166829 full page images, 178819332 bytes
	      system usage: CPU: user: 2.59 s, system: 4.46 s, elapsed: 2326.38 s`
	msgAnalyze := `automatic analyze of table "rbcc-postgres.public.event_log"
          avg read rate: 14.355 MB/s, avg write rate: 0.004 MB/s
          buffer usage: 671 hits, 30049 misses, 9 dirtied
          system usage: CPU: user: 0.26 s, system: 0.35 s, elapsed: 16.35 s`

	durVacuum, tblName := util.ExtractAutovacuumOrAnalyzeDurationSecondsFromLogMessage(msgVacuum)
	assert.InDelta(t, 2326.38, durVacuum, 0.01, "Vacuum elapsed duration should be parsed correctly")
	assert.Equal(t, tblName, "rbcc-postgres.public.event_log", "Vacuum table name should be parsed correctly")

	durAnalyze, tblName := util.ExtractAutovacuumOrAnalyzeDurationSecondsFromLogMessage(msgAnalyze)
	log.Println("Analyze duration:", durAnalyze, "Table name:", tblName)
	assert.InDelta(t, 16.35, durAnalyze, 0.01, "Analyze elapsed duration should be parsed correctly")
	assert.Equal(t, tblName, "rbcc-postgres.public.event_log", "Vacuum table name should be parsed correctly")
}
func TestHumanTimeOrDeltaStringToTime_Days(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Time
	}{
		{"-1d", now.Add(time.Duration(-24) * time.Hour)},
		{"2d", now.Add(time.Duration(-48) * time.Hour)},
		{"3days", now.Add(time.Duration(-72) * time.Hour)},
		{"-2days", now.Add(time.Duration(-48) * time.Hour)},
		{"1day", now.Add(time.Duration(-24) * time.Hour)},
		{"14d", now.Add(time.Duration(-336) * time.Hour)},
	}

	for _, tt := range tests {
		got, err := util.HumanTimeOrDeltaStringToTime(tt.input, now)
		assert.NoError(t, err, "should not error for input %s", tt.input)
		// Allow a small delta for roundings
		assert.InDelta(t, tt.expected.UnixMilli(), got.UnixMilli(), 100, "unexpected time delta for input %s", tt.input)
	}
}
