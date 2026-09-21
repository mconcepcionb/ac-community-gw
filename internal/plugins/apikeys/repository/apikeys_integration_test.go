//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestAPIKeyLifecycle(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	key, raw, err := store.Create(ctx, "it-key", []string{"gw.report.read", "gw.audit.read"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = store.Revoke(context.Background(), key.ID) })
	if raw == "" || key.KeyPrefix == "" || len(key.Permissions) != 2 {
		t.Fatalf("created = %+v raw=%q", key, raw)
	}

	principal, ok := store.Authenticate(ctx, raw)
	if !ok || principal.UserID != key.ID || len(principal.Permissions) != 2 {
		t.Fatalf("Authenticate = %+v, %v", principal, ok)
	}
	if _, ok := store.Authenticate(ctx, "not-a-key"); ok {
		t.Fatal("a key without the ak_ prefix must not authenticate")
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, item := range list {
		if item.ID == key.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("created key missing from List")
	}

	_, rotated, err := store.Rotate(ctx, key.ID)
	if err != nil || rotated == raw {
		t.Fatalf("Rotate = %q, %v", rotated, err)
	}
	if _, ok := store.Authenticate(ctx, raw); ok {
		t.Fatal("the old secret must stop authenticating after rotation")
	}
	if _, ok := store.Authenticate(ctx, rotated); !ok {
		t.Fatal("the rotated secret must authenticate")
	}

	if _, _, err := store.Rotate(ctx, uuid.New()); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Rotate missing = %v, want ErrKeyNotFound", err)
	}
	if err := store.Revoke(ctx, key.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
}
