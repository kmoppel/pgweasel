package logparser_test

import (
	"log"
	"testing"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/stretchr/testify/assert"
)

var log1 = `2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`

func TestFileLogger(t *testing.T) {
	e, err := logparser.EventLinesToPgLogEntry(log1, logparser.DEFAULT_REGEX)
	assert.NoError(t, err)
	ts, err := logparser.TimestringToTime("2025-05-02 12:27:52.634 EEST")
	assert.NoError(t, err)
	assert.Equal(t, ts, e.LogTime)
}

func TestHasTimestampPrefix(t *testing.T) {
	assert.True(t, logparser.HasTimestampPrefix("2025-05-02 12:27:52.634 EEST [2380404]"))
	assert.False(t, logparser.HasTimestampPrefix("bla 2025-05-02 12:27:52.634 EEST [2380404]"))
	assert.True(t, logparser.HasTimestampPrefix("2025-05-05 06:00:51 UTC:90.190.32.92(32890):postgres@postgres:[1315]:LOG:  statement: BEGIN;"))
}

func TestFallbackRegex(t *testing.T) {
	regex := logparser.DEFAULT_REGEX
	expectedGroups := []string{"log_time", "error_severity", "message"}

	testLines := []string{
		"2024-05-07 10:22:13 UTC [12345]: [1-1] user=admin,db=exampledb,app=psql LOG:  connection received: host=203.0.113.45 port=52344",
		"2025-05-07 12:34:56.789 UTC ERROR: some error message",
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
