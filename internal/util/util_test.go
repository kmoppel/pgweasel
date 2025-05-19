package util_test

import (
	"testing"
	"time"

	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestHumanTimedeltaToTime(t *testing.T) {
	now := time.Now()

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
}
