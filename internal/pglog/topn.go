package pglog

import "container/heap"

type TopNSlowLogEntry struct {
	Rec        *LogEntry
	DurationMs int
}
type SlowEntryHeap []TopNSlowLogEntry

// https://pkg.go.dev/container/heap#example-package-IntHeap
func (h SlowEntryHeap) Len() int           { return len(h) }
func (h SlowEntryHeap) Less(i, j int) bool { return h[i].DurationMs < h[j].DurationMs }
func (h SlowEntryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *SlowEntryHeap) Push(x any) {
	*h = append(*h, x.(TopNSlowLogEntry))
}

func (h *SlowEntryHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type TopN struct {
	N    int
	Heap *SlowEntryHeap
}

func NewTopN(n int) *TopN {
	h := &SlowEntryHeap{}
	heap.Init(h)
	return &TopN{N: n, Heap: h}
}

func (t *TopN) Add(val TopNSlowLogEntry) {
	if t.Heap.Len() < t.N {
		heap.Push(t.Heap, val)
	} else if val.DurationMs > (*t.Heap)[0].DurationMs {
		heap.Pop(t.Heap)
		heap.Push(t.Heap, val)
	}
}

func (t *TopN) Values() []TopNSlowLogEntry {
	hCopy := make([]TopNSlowLogEntry, len(*t.Heap))
	copy(hCopy, *t.Heap)
	return hCopy
}
