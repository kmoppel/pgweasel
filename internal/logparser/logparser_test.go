package logparser_test

import (
	"testing"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/stretchr/testify/assert"
)

var log1 = `2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`

func TestFileLogger(t *testing.T) {
	e, err := logparser.ParseEntryFromLogline(log1, "%m [%p] %q%u@%d ")
	assert.NoError(t, err)
	assert.Equal(t, "2025-05-02 12:27:52.634 EEST", e.LogTime)
}
