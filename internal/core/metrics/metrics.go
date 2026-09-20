// Package metrics provides a minimal in-process counter registry.
//
// It exists so the gateway can expose operational counters without pulling in
// a full metrics dependency. Counters are process-local and reset on restart.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
)

// Registry is a concurrency-safe set of named counters.
type Registry struct {
	mu       sync.Mutex
	counters map[string]int64
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{counters: make(map[string]int64)}
}

// Inc increments a counter by one.
func (r *Registry) Inc(name string) {
	r.Add(name, 1)
}

// Add increments a counter by delta.
func (r *Registry) Add(name string, delta int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] += delta
}

// Snapshot returns a copy of the current counter values.
func (r *Registry) Snapshot() map[string]int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int64, len(r.counters))
	for name, value := range r.counters {
		out[name] = value
	}
	return out
}

// WritePrometheus writes every counter in the Prometheus text exposition
// format.
func (r *Registry) WritePrometheus(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	snapshot := r.Snapshot()
	names := make([]string, 0, len(snapshot))
	for name := range snapshot {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		_, _ = fmt.Fprintf(w, "# TYPE %s counter\n%s %d\n", name, name, snapshot[name])
	}
}
