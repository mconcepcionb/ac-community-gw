//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestClaimLifecycle(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM account_claims WHERE user_id = $1`, userID) })

	expires := time.Now().Add(time.Hour).UTC()
	if _, err := store.UpsertClaim(ctx, userID, "ITACCOUNT", "hash-1", expires); err != nil {
		t.Fatalf("UpsertClaim: %v", err)
	}
	got, err := store.GetClaim(ctx, userID)
	if err != nil || got.CodeHash != "hash-1" || got.AccountUsername != "ITACCOUNT" {
		t.Fatalf("GetClaim = %+v, %v", got, err)
	}
	incremented, err := store.IncrementClaimAttempts(ctx, userID)
	if err != nil || incremented.Attempts < 1 {
		t.Fatalf("IncrementClaimAttempts = %+v, %v", incremented, err)
	}
	claims, err := store.ListClaims(ctx, 50, 0)
	if err != nil || !containsClaim(claims, userID) {
		t.Fatalf("ListClaims = %+v, %v", claims, err)
	}
	if err := store.DeleteClaim(ctx, userID); err != nil {
		t.Fatalf("DeleteClaim: %v", err)
	}
	if _, err := store.GetClaim(ctx, userID); !errors.Is(err, domain.ErrClaimNotFound) {
		t.Fatalf("GetClaim after delete = %v, want ErrClaimNotFound", err)
	}
	if err := store.DeleteClaim(ctx, uuid.New()); err != nil {
		t.Fatalf("DeleteClaim unknown = %v, want nil", err)
	}
}

func TestLinkListDeleteAndConflict(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	otherID := uuid.New()
	username := "IT" + uuid.New().String()[:8]
	accountID := int64(9)
	for _, id := range []uuid.UUID{userID, otherID} {
		if _, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
			VALUES ($1, $2, 'it-link')`, id, "itlink-"+id.String()); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM azeroth_account_links WHERE user_id IN ($1, $2)`, userID, otherID)
		_, _ = db.Exec(`DELETE FROM community_users WHERE id IN ($1, $2)`, userID, otherID)
	})

	if _, err := store.Upsert(ctx, userID, username, &accountID); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	links, err := store.List(ctx)
	if err != nil || !containsLink(links, userID) {
		t.Fatalf("List = %+v, %v", links, err)
	}
	if _, err := store.Upsert(ctx, otherID, username, nil); !errors.Is(err, domain.ErrAccountAlreadyLinked) {
		t.Fatalf("duplicate username = %v, want ErrAccountAlreadyLinked", err)
	}
	if err := store.Delete(ctx, userID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, userID); !errors.Is(err, domain.ErrLinkNotFound) {
		t.Fatalf("Get after delete = %v, want ErrLinkNotFound", err)
	}
	if err := store.Delete(ctx, uuid.New()); err != nil {
		t.Fatalf("Delete unknown = %v, want nil", err)
	}
}

func containsClaim(claims []domain.Claim, userID uuid.UUID) bool {
	for _, claim := range claims {
		if claim.UserID == userID {
			return true
		}
	}
	return false
}

func containsLink(links []domain.Link, userID uuid.UUID) bool {
	for _, link := range links {
		if link.UserID == userID {
			return true
		}
	}
	return false
}
