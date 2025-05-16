package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHumanTimedeltaToTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Time
	}{
		{"-2h", now.Add(time.Duration(-2) * time.Hour)},
		{"-10m", now.Add(time.Duration(-10) * time.Minute)},
		{"5m", now.Add(time.Duration(5) * time.Minute)},
		{"-48h", now.Add(time.Duration(-48) * time.Hour)},
		{"-30s", now.Add(time.Duration(-30) * time.Second)},
	}

	for _, tt := range tests {
		got, err := HumanTimedeltaToTime(tt.input)
		assert.NoError(t, err, "should not error for input %s", tt.input)
		// Allow a small delta for roundings
		assert.InDelta(t, tt.expected.UnixMilli(), got.UnixMilli(), 100, "unexpected time delta for input %s", tt.input)
	}
}

func TestHumanTimedeltaToTime_InvalidInput(t *testing.T) {
	t1, err := HumanTimedeltaToTime("1d")
	assert.Error(t, err, "days not supported")
	assert.True(t, t1.IsZero(), "should return zero time for invalid input")
}

func TestHumanTimedeltaToTime_TimestampInput(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2025-05-08 12:25:47.010 UTC", time.Date(2025, 5, 8, 12, 25, 47, 10*1e6, time.FixedZone("UTC", 0))},
		{"2019-10-21 12:03:42.567 EEST", time.Date(2019, 10, 21, 12, 3, 42, 567*1e6, time.FixedZone("EEST", 3*3600))},
		{"2019-10-21 12:03:42 EEST", time.Date(2019, 10, 21, 12, 3, 42, 0, time.FixedZone("EEST", 3*3600))},
	}

	for _, tt := range tests {
		got, err := HumanTimedeltaToTime(tt.input)
		assert.NoError(t, err, "should not error for input %s", tt.input)
		// Allow a small delta for roundings
		assert.InDelta(t, tt.expected.UnixMilli(), got.UnixMilli(), 100, "unexpected time delta for input %s", tt.input)
	}
}
