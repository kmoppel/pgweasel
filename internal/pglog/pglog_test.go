package pglog_test

import (
	"testing"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/stretchr/testify/assert"
)

func TestSeverityToNum(t *testing.T) {
	assert.Greater(t, pglog.SeverityToNum("info"), pglog.SeverityToNum("DEBUG"))
	assert.Greater(t, pglog.SeverityToNum("debug1"), pglog.SeverityToNum("debug2"))
	assert.Equal(t, -1, pglog.SeverityToNum("dbg"))
}
