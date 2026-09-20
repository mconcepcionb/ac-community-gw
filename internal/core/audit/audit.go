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

// NopRecorder discards audit entries.
type NopRecorder struct{}

// Record implements Recorder.
func (NopRecorder) Record(context.Context, Entry) error { return nil }

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
