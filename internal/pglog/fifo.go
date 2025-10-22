package pglog

import "time"

// entryWithTimestamp wraps a LogEntry with its insertion timestamp
type entryWithTimestamp struct {
	entry     LogEntry
	timestamp time.Time
}

// Fifo is a fixed-size ring buffer for LogEntry items.
// When the buffer is full and a new entry is added, the oldest entry (by insertion timestamp) is replaced.
type Fifo struct {
	entries []entryWithTimestamp
	maxSize int
}

// NewFifo creates a new Fifo with the specified maximum size.
// If maxSize is <= 0, it defaults to 10.
func NewFifo(maxSize int) *Fifo {
	if maxSize <= 0 {
		maxSize = 10
	}
	return &Fifo{
		entries: make([]entryWithTimestamp, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a new LogEntry to the ring buffer.
// If the buffer is full, the oldest entry (by insertion timestamp) is replaced.
func (f *Fifo) Add(entry LogEntry) {
	now := time.Now()

	if len(f.entries) < f.maxSize {
		// Buffer not yet full, just append
		f.entries = append(f.entries, entryWithTimestamp{
			entry:     entry,
			timestamp: now,
		})
		return
	}

	// Buffer is full, find and replace the oldest entry by timestamp
	oldestIndex := 0
	oldestTime := f.entries[0].timestamp

	for i := 1; i < len(f.entries); i++ {
		if f.entries[i].timestamp.Before(oldestTime) {
			oldestIndex = i
			oldestTime = f.entries[i].timestamp
		}
	}

	// Replace the oldest entry
	f.entries[oldestIndex] = entryWithTimestamp{
		entry:     entry,
		timestamp: now,
	}
}

// GetAll returns all entries in the ring buffer sorted by insertion timestamp (oldest first).
func (f *Fifo) GetAll() []LogEntry {
	if len(f.entries) == 0 {
		return nil
	}

	// Create a copy of entries and sort by timestamp
	sorted := make([]entryWithTimestamp, len(f.entries))
	copy(sorted, f.entries)

	// Simple bubble sort by timestamp
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].timestamp.Before(sorted[i].timestamp) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Extract just the LogEntry values
	result := make([]LogEntry, len(sorted))
	for i, e := range sorted {
		result[i] = e.entry
	}
	return result
}

func (f *Fifo) GetAllByTimestampString(ts string) []LogEntry {
	var result []LogEntry
	for _, e := range f.entries {
		if e.entry.LogTime == ts {
			result = append(result, e.entry)
		}
	}
	return result
}

// Len returns the current number of entries in the ring buffer.
func (f *Fifo) Len() int {
	return len(f.entries)
}

// IsEmpty returns true if the ring buffer contains no entries.
func (f *Fifo) IsEmpty() bool {
	return len(f.entries) == 0
}

// IsFull returns true if the ring buffer is full.
func (f *Fifo) IsFull() bool {
	return len(f.entries) >= f.maxSize
}

// Clear removes all entries from the ring buffer.
func (f *Fifo) Clear() {
	f.entries = f.entries[:0]
}
