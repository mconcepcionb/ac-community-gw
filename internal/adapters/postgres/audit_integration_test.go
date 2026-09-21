//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestAuditStoreRoundTrip(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := NewAuditStore(db)

	actor := uuid.New()
	targetID := "IT" + uuid.New().String()[:8]
	requestID := "it-audit-" + uuid.New().String()[:8]
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM audit_log WHERE request_id = $1`, requestID)
		_, _ = db.Exec(`DELETE FROM audit_log WHERE target_id = $1`, targetID)
	})

	entry := audit.Entry{
		Timestamp:      time.Now().UTC(),
		ActorID:        actor,
		ActorDiscordID: "42",
		Action:         "it.audit",
		Permission:     "gw.it.read",
		TargetType:     "account",
		TargetID:       targetID,
		Result:         audit.ResultSuccess,
		RequestID:      requestID,
		Metadata:       map[string]any{"k": "v"},
	}
	if err := store.Record(ctx, entry); err != nil {
		t.Fatalf("Record: %v", err)
	}
	// A nil actor and no metadata exercise the null-actor branch.
	second := audit.Entry{
		Timestamp: time.Now().UTC(), Action: "it.audit2",
		Result: audit.ResultFailure, RequestID: requestID + "-2", TargetID: targetID,
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM audit_log WHERE request_id = $1`, second.RequestID) })
	if err := store.Record(ctx, second); err != nil {
		t.Fatalf("Record nil actor: %v", err)
	}

	entries, err := store.List(ctx, audit.ListFilter{ActorID: actor.String()}, 10, 0)
	if err != nil {
		t.Fatalf("List by actor: %v", err)
	}
	found := false
	for _, got := range entries {
		if got.RequestID == requestID {
			found = true
			if got.Metadata["k"] != "v" || got.ActorID != actor {
				t.Fatalf("entry = %+v", got)
			}
		}
	}
	if !found {
		t.Fatal("recorded entry missing from List")
	}

	// Exercise every filter and the limit/offset clamping.
	filters := []audit.ListFilter{
		{Target: targetID},
		{TargetType: "account"},
		{TargetID: targetID},
		{Action: "it.audit"},
		{Since: time.Now().Add(-time.Hour)},
		{Until: time.Now().Add(time.Hour)},
	}
	for _, filter := range filters {
		if _, err := store.List(ctx, filter, 0, -1); err != nil {
			t.Fatalf("List %+v: %v", filter, err)
		}
	}
	if _, err := store.List(ctx, audit.ListFilter{}, 500, 0); err != nil {
		t.Fatalf("List clamped limit: %v", err)
	}
}

func TestMigrateRuns(t *testing.T) {
	url := integration.PostgresURL(t)
	opened, err := Open(context.Background(), config.Database{URL: url})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = opened.Close() })

	if err := Migrate(context.Background(), opened, "version"); err != nil {
		t.Fatalf("Migrate version: %v", err)
	}
	if err := Migrate(context.Background(), opened, "not-a-command"); err == nil {
		t.Fatal("expected an error for an invalid goose command")
	}
}
