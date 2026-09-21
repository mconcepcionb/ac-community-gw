package audit

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type captureRecorder struct {
	entries []Entry
	err     error
}

func (c *captureRecorder) Record(_ context.Context, entry Entry) error {
	c.entries = append(c.entries, entry)
	return c.err
}

func TestNopRecorder(t *testing.T) {
	if err := (NopRecorder{}).Record(context.Background(), Entry{}); err != nil {
		t.Fatalf("NopRecorder.Record = %v", err)
	}
}

func TestMultiRecorderFansOut(t *testing.T) {
	first := &captureRecorder{}
	second := &captureRecorder{}
	recorder := NewMultiRecorder(first, second)

	if err := recorder.Record(context.Background(), Entry{Action: "account.create"}); err != nil {
		t.Fatalf("Record = %v", err)
	}
	if len(first.entries) != 1 || len(second.entries) != 1 {
		t.Fatalf("fan-out = %d/%d", len(first.entries), len(second.entries))
	}
}

func TestMultiRecorderReturnsFirstErrorButKeepsGoing(t *testing.T) {
	firstErr := errors.New("boom")
	first := &captureRecorder{err: firstErr}
	second := &captureRecorder{}
	recorder := NewMultiRecorder(first, second)

	if err := recorder.Record(context.Background(), Entry{}); !errors.Is(err, firstErr) {
		t.Fatalf("Record = %v, want %v", err, firstErr)
	}
	if len(second.entries) != 1 {
		t.Fatal("every recorder must be called even after an error")
	}
}

func TestLogRecorderOmitsMetadata(t *testing.T) {
	var buf bytes.Buffer
	recorder := NewLogRecorder(slog.New(slog.NewJSONHandler(&buf, nil)))

	entry := Entry{
		ActorID:    uuid.New(),
		Action:     "account.create",
		Permission: "azeroth.account.create",
		TargetType: "account",
		TargetID:   "thrall",
		Result:     ResultSuccess,
		RequestID:  "req-1",
		Metadata:   map[string]any{"password": "super-secret"},
	}
	if err := recorder.Record(context.Background(), entry); err != nil {
		t.Fatalf("Record = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"account.create", "azeroth.account.create", "thrall", "success", "req-1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("log output %q missing %q", out, want)
		}
	}
	if strings.Contains(out, "super-secret") || strings.Contains(out, "password") {
		t.Fatalf("log output must not include entry metadata: %q", out)
	}
}
