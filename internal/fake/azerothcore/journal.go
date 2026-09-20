package fakeazerothcore

import (
	"sync"
	"time"
)

// CommandEntry is one SOAP command handled by the double.
type CommandEntry struct {
	Seq        int64     `json:"seq"`
	At         time.Time `json:"at"`
	RequestID  string    `json:"request_id,omitempty"`
	Command    string    `json:"command"`
	Result     string    `json:"result"`
	User       string    `json:"user,omitempty"`
	Remote     string    `json:"remote,omitempty"`
	DurationMS int64     `json:"duration_ms"`
}

// journal is a bounded, concurrency-safe command log with live subscribers.
type journal struct {
	mu       sync.Mutex
	entries  []CommandEntry
	capacity int
	seq      int64
	subs     map[int]chan CommandEntry
	nextSub  int
}

func newJournal(capacity int) *journal {
	if capacity <= 0 {
		capacity = 200
	}
	return &journal{capacity: capacity, subs: make(map[int]chan CommandEntry)}
}

func (j *journal) append(entry CommandEntry) CommandEntry {
	j.mu.Lock()
	j.seq++
	entry.Seq = j.seq
	if entry.At.IsZero() {
		entry.At = time.Now()
	}
	j.entries = append(j.entries, entry)
	if len(j.entries) > j.capacity {
		j.entries = j.entries[len(j.entries)-j.capacity:]
	}
	subscribers := make([]chan CommandEntry, 0, len(j.subs))
	for _, ch := range j.subs {
		subscribers = append(subscribers, ch)
	}
	j.mu.Unlock()

	for _, ch := range subscribers {
		select {
		case ch <- entry:
		default: // slow subscriber: drop rather than block the server
		}
	}
	return entry
}

// list returns entries newest first, capped at limit.
func (j *journal) list(limit int) []CommandEntry {
	j.mu.Lock()
	defer j.mu.Unlock()
	if limit <= 0 || limit > len(j.entries) {
		limit = len(j.entries)
	}
	out := make([]CommandEntry, 0, limit)
	for i := len(j.entries) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, j.entries[i])
	}
	return out
}

func (j *journal) subscribe() (int, <-chan CommandEntry) {
	j.mu.Lock()
	defer j.mu.Unlock()
	id := j.nextSub
	j.nextSub++
	ch := make(chan CommandEntry, 16)
	j.subs[id] = ch
	return id, ch
}

func (j *journal) unsubscribe(id int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.subs, id)
}

func (j *journal) clear() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.entries = nil
}
