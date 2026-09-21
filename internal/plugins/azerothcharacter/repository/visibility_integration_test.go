//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestVisibilityLifecycle(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	name := "Itchar" + uuid.New().String()[:8]
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM character_visibility WHERE character_name = $1`, name) })

	if public, err := store.SetVisibility(ctx, name, userID, true); err != nil || !public {
		t.Fatalf("SetVisibility = %v, %v", public, err)
	}
	if public, err := store.GetVisibility(ctx, name); err != nil || !public {
		t.Fatalf("GetVisibility = %v, %v", public, err)
	}

	flags, err := store.ListByUser(ctx, userID)
	if err != nil || !flags[name] {
		t.Fatalf("ListByUser = %+v, %v", flags, err)
	}

	names, err := store.PublicNames(ctx)
	if err != nil {
		t.Fatalf("PublicNames: %v", err)
	}
	found := false
	for _, candidate := range names {
		if candidate == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("public name %q missing", name)
	}

	if public, err := store.SetVisibility(ctx, name, userID, false); err != nil || public {
		t.Fatalf("SetVisibility false = %v, %v", public, err)
	}
	if _, err := store.GetVisibility(ctx, "missing-"+uuid.New().String()[:8]); !errors.Is(err, ErrVisibilityNotFound) {
		t.Fatalf("GetVisibility missing = %v, want ErrVisibilityNotFound", err)
	}
}
