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

func TestHumanTimeOrDeltaStringToTime_DateInput(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Time
	}{
		// Added a test case for date-only input "2025-05-08"
		{"6 July 2025", time.Date(2025, 07, 06, 0, 0, 0, 0, time.Local)},
		{"2025-05-08", time.Date(2025, 5, 8, 0, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		got, err := util.HumanTimeOrDeltaStringToTime(tt.input, now)
		assert.NoError(t, err, "should not error for input %s", tt.input)
		// Allow +/-1h due to summer time changes
		assert.InDelta(t, tt.expected.UnixMilli(), got.UnixMilli(), 3600000+100, "unexpected time delta for input %s", tt.input)
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
		expectedStr    string
	}{
		{
			message:        "duration: 123 ms statement: SELECT * FROM table",
			expectedMillis: 123,
			expectedStr:    "123",
		},
		{
			message:        "2025-05-16 14:26:01.872 UTC [3076] LOG:  duration: 18.237 ms",
			expectedMillis: 18.237,
			expectedStr:    "18.237",
		},
		{
			message:        "LOG: statement executed, duration: 5.678 ms",
			expectedMillis: 5.678,
			expectedStr:    "5.678",
		},
		{
			message:        "LOG: statement executed without timing info",
			expectedMillis: 0,
			expectedStr:    "",
		},
	}

	for _, tt := range tests {
		millis, durationStr := util.ExtractDurationMillisFromLogMessage(tt.message)
		assert.InDelta(t, tt.expectedMillis, millis, 0.1, "Duration should be within 0.1ms of expected value")
		assert.Equal(t, tt.expectedStr, durationStr, "Duration string should match expected value")
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

func TestExtractAutovacuumReadWriteRatesFromLogMessage(t *testing.T) {
	tests := []struct {
		message   string
		readRate  float64
		writeRate float64
	}{
		{
			"avg read rate: 5.492 MB/s, avg write rate: 4.932 MB/s",
			5.492, 4.932,
		},
		{
			"avg read rate: 14.355 MB/s, avg write rate: 0.004 MB/s",
			14.355, 0.004,
		},
		{
			"avg read rate: 107.365 MB/s, avg write rate: 0.011 MB/s",
			107.365, 0.011,
		},
		{
			"avg read rate: 0.000 MB/s, avg write rate: 0.000 MB/s",
			0.000, 0.000,
		},
		{
			"some other log message without rates",
			0.0, 0.0,
		},
		{
			"partial match: avg read rate: 5.5 MB/s",
			0.0, 0.0,
		},
	}

	for _, tt := range tests {
		readRate, writeRate := util.ExtractAutovacuumReadWriteRatesFromLogMessage(tt.message)
		assert.Equal(t, tt.readRate, readRate, "read rate should match for message: %s", tt.message)
		assert.Equal(t, tt.writeRate, writeRate, "write rate should match for message: %s", tt.message)
	}
}
func TestExtractConnectHostFromLogMessage(t *testing.T) {
	tests := []struct {
		message      string
		expectedHost string
	}{
		{
			message:      "connection received: host=127.0.0.1 port=44410",
			expectedHost: "127.0.0.1",
		},
		{
			message:      "connection received: host=[local]",
			expectedHost: "local",
		},
		{
			message:      "connection received: host=192.168.1.100 port=5432",
			expectedHost: "192.168.1.100",
		},
		{
			message:      "connection received: host=localhost port=5432",
			expectedHost: "localhost",
		},
		{
			message:      "connection received: host=example.com port=5432",
			expectedHost: "example.com",
		},
		{
			message:      "connection received: host=10.0.0.1",
			expectedHost: "10.0.0.1",
		},
		{
			message:      "some other log message without host",
			expectedHost: "",
		},
		{
			message:      "connection received: port=5432",
			expectedHost: "",
		},
		{
			message:      "",
			expectedHost: "",
		},
	}

	for _, tt := range tests {
		host := util.ExtractConnectHostFromLogMessage(tt.message)
		assert.Equal(t, tt.expectedHost, host, "host should match for message: %s", tt.message)
	}
}
func TestExtractConnectUserDbAppnameSslFromLogMessage(t *testing.T) {
	tests := []struct {
		message         string
		expectedUser    string
		expectedDb      string
		expectedAppname string
		expectedSsl     bool
	}{
		{
			message:         "connection authorized: user=krl database=postgres application_name=psql",
			expectedUser:    "krl",
			expectedDb:      "postgres",
			expectedAppname: "psql",
			expectedSsl:     false,
		},
		{
			message:         "connection authorized: user=monitor database=bench SSL enabled (protocol=TLSv1.3, cipher=TLS_AES_256_GCM_SHA384, bits=256)",
			expectedUser:    "monitor",
			expectedDb:      "bench",
			expectedAppname: "",
			expectedSsl:     true,
		},
		{
			message:         "connection authorized: user=admin database=testdb application_name=pgAdmin SSL enabled",
			expectedUser:    "admin",
			expectedDb:      "testdb",
			expectedAppname: "pgAdmin",
			expectedSsl:     true,
		},
		{
			message:         "some other log message without connection info",
			expectedUser:    "",
			expectedDb:      "",
			expectedAppname: "",
			expectedSsl:     false,
		},
	}

	for _, tt := range tests {
		user, db, appname, ssl := util.ExtractConnectUserDbAppnameSslFromLogMessage(tt.message)
		assert.Equal(t, tt.expectedUser, user, "user should match for message: %s", tt.message)
		assert.Equal(t, tt.expectedDb, db, "database should match for message: %s", tt.message)
		assert.Equal(t, tt.expectedAppname, appname, "application name should match for message: %s", tt.message)
		assert.Equal(t, tt.expectedSsl, ssl, "SSL flag should match for message: %s", tt.message)
	}
}

func TestCalculatePercentile(t *testing.T) {
	tests := []struct {
		name       string
		data       []float64
		percentile float64
		expected   float64
	}{
		{
			name:       "Empty slice",
			data:       []float64{},
			percentile: 50,
			expected:   0,
		},
		{
			name:       "Single element",
			data:       []float64{42.5},
			percentile: 50,
			expected:   42.5,
		},
		{
			name:       "Two elements - 50th percentile",
			data:       []float64{10, 20},
			percentile: 50,
			expected:   15,
		},
		{
			name:       "Simple case - 25th percentile",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 25,
			expected:   2,
		},
		{
			name:       "Simple case - 75th percentile",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 75,
			expected:   4,
		},
		{
			name:       "Larger dataset - 95th percentile",
			data:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentile: 95,
			expected:   9.55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.CalculatePercentile(tt.data, tt.percentile)
			assert.InDelta(t, tt.expected, result, 0.01, "percentile calculation should be accurate")
		})
	}
}
