package logparser_test

import (
	"testing"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/stretchr/testify/assert"
)

const DEBIAN_DEFAULT_LOG_LINE_PREFIX = "%m [%p] %q%u@%d "

var log1 = `2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`

func TestFileLogger(t *testing.T) {
	r := logparser.CompileRegexForLogLinePrefix(DEBIAN_DEFAULT_LOG_LINE_PREFIX)
	e, err := logparser.ParseEntryFromLogline(log1, r)
	assert.NoError(t, err)
	ts, err := logparser.TimestringToTime("2025-05-02 12:27:52.634 EEST")
	assert.NoError(t, err)
	assert.Equal(t, ts, e.LogTime)
}

func TestHasTimestampPrefix(t *testing.T) {
	assert.True(t, logparser.HasTimestampPrefix("2025-05-02 12:27:52.634 EEST [2380404]"))
	assert.False(t, logparser.HasTimestampPrefix("bla 2025-05-02 12:27:52.634 EEST [2380404]"))
}
