//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestAnnotationLifecycle(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	targetID := "it-annot-" + uuid.New().String()[:8]
	author := uuid.New()
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM admin_annotations WHERE target_id = $1`, targetID) })

	created, err := store.Create(ctx, "account", targetID, author, "note")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Body != "note" || created.AuthorID != author || created.ID == uuid.Nil {
		t.Fatalf("created = %+v", created)
	}

	list, err := store.ByTarget(ctx, "account", targetID, 10, 0)
	if err != nil || len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("ByTarget = %+v, %v", list, err)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil || got.Body != "note" {
		t.Fatalf("Get = %+v, %v", got, err)
	}

	updated, err := store.Update(ctx, created.ID, "edited")
	if err != nil || updated.Body != "edited" {
		t.Fatalf("Update = %+v, %v", updated, err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, created.ID); !errors.Is(err, domain.ErrAnnotationNotFound) {
		t.Fatalf("Get after delete = %v, want ErrAnnotationNotFound", err)
	}
	if _, err := store.Update(ctx, uuid.New(), "x"); !errors.Is(err, domain.ErrAnnotationNotFound) {
		t.Fatalf("Update missing = %v, want ErrAnnotationNotFound", err)
	}
	if err := store.Delete(ctx, uuid.New()); !errors.Is(err, domain.ErrAnnotationNotFound) {
		t.Fatalf("Delete missing = %v, want ErrAnnotationNotFound", err)
	}
}
