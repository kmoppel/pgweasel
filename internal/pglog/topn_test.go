package pglog_test

import (
	"testing"

	. "github.com/kmoppel/pgweasel/internal/pglog"
)

func TestSlowEntryHeap_TopN(t *testing.T) {
	topN := 2
	h := NewTopN(topN)

	entries := []TopNSlowLogEntry{
		{Rec: &LogEntry{LogTime: "t1"}, DurationMs: 10},
		{Rec: &LogEntry{LogTime: "t2"}, DurationMs: 50},
		{Rec: &LogEntry{LogTime: "t3"}, DurationMs: 30},
		{Rec: &LogEntry{LogTime: "t4"}, DurationMs: 70},
		{Rec: &LogEntry{LogTime: "t5"}, DurationMs: 20},
	}

	for _, entry := range entries {
		h.Add(entry)
	}

	// Should contain the two largest DurationMs: 70 and 50
	found := map[int]bool{}
	for _, e := range *h.Heap {
		found[e.DurationMs] = true
	}
	if !found[70] || !found[50] {
		t.Errorf("expected top durations 70 and 50, got %+v", *h.Heap)
	}
}
