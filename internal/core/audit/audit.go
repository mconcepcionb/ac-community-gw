// Package audit records security-relevant actions.
//
// The recorder is intentionally a small boundary: no secrets, tokens or
// credentials may ever be passed in an Entry.
package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Result describes the outcome of an audited action.
type Result string

const (
	// ResultSuccess means the action completed.
	ResultSuccess Result = "success"
	// ResultFailure means the action was rejected or failed.
	ResultFailure Result = "failure"
)

// Entry is a single audit record.
type Entry struct {
	Timestamp      time.Time
	ActorID        uuid.UUID
	ActorDiscordID string
	Action         string
	Permission     string
	TargetType     string
	TargetID       string
	Result         Result
	RequestID      string
	Metadata       map[string]any
}

// Recorder persists audit entries.
type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}

// ListFilter filters audit entries. Zero values mean "no filter".
type ListFilter struct {
	// ActorID matches the actor's community user id exactly.
	ActorID string
	// Target matches target_id (case-insensitive substring).
	Target string
	// TargetType matches target_type exactly (for example "account").
	TargetType string
	// TargetID matches target_id exactly. Pair it with TargetType for the
	// per-entity history query.
	TargetID string
	// Action matches the action (case-insensitive substring).
	Action string
	// Since and Until bound occurred_at (inclusive); zero means unbounded.
	Since time.Time
	Until time.Time
}

// Reader reads audit entries, newest first.
type Reader interface {
	List(ctx context.Context, filter ListFilter, limit, offset int) ([]Entry, error)
}

// NopRecorder discards audit entries.
type NopRecorder struct{}

// Record implements Recorder.
func (NopRecorder) Record(context.Context, Entry) error { return nil }

// MultiRecorder records every entry to each underlying recorder.
type MultiRecorder struct {
	recorders []Recorder
}

// NewMultiRecorder fans out audit entries to every recorder.
func NewMultiRecorder(recorders ...Recorder) *MultiRecorder {
	return &MultiRecorder{recorders: recorders}
}

// Record implements Recorder. It reports the first error but always tries every
// recorder.
func (m *MultiRecorder) Record(ctx context.Context, entry Entry) error {
	var firstErr error
	for _, recorder := range m.recorders {
		if err := recorder.Record(ctx, entry); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// LogRecorder writes audit entries to the structured logger.
type LogRecorder struct {
	logger *slog.Logger
}

// NewLogRecorder creates a log-backed recorder.
func NewLogRecorder(logger *slog.Logger) *LogRecorder {
	return &LogRecorder{logger: logger}
}

// Record implements Recorder.
func (r *LogRecorder) Record(_ context.Context, entry Entry) error {
	r.logger.Info("audit",
		slog.String("action", entry.Action),
		slog.String("permission", entry.Permission),
		slog.String("target_type", entry.TargetType),
		slog.String("target_id", entry.TargetID),
		slog.String("result", string(entry.Result)),
		slog.String("request_id", entry.RequestID),
		slog.String("actor_id", entry.ActorID.String()),
	)
	return nil
}
