//go:build integration

package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("ACGW_DATABASE_URL")
	if url == "" {
		t.Skip("ACGW_DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedUser(t *testing.T, db *sql.DB, discordID string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
		VALUES ($1, $2, $3) ON CONFLICT (discord_id) DO NOTHING`, id, discordID, "it-link")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestAccountLinkRoundTrip(t *testing.T) {
	db := openTestDB(t)
	_, _ = db.Exec(`DELETE FROM community_users WHERE discord_id LIKE 'itlink-%'`)

	store := New(db)
	ctx := context.Background()

	userID := seedUser(t, db, "itlink-1")
	accountID := int64(7)
	link, err := store.Upsert(ctx, userID, "ITLINK", &accountID)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if link.AccountUsername != "ITLINK" || link.AccountID == nil || *link.AccountID != 7 {
		t.Fatalf("link = %+v", link)
	}

	got, err := store.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.AccountUsername != "ITLINK" {
		t.Fatalf("got = %+v", got)
	}

	otherUser := seedUser(t, db, "itlink-2")
	if _, err := store.Upsert(ctx, otherUser, "ITLINK", nil); !errors.Is(err, domain.ErrAccountAlreadyLinked) {
		t.Fatalf("expected ErrAccountAlreadyLinked, got %v", err)
	}

	links, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("links = %d", len(links))
	}

	if err := store.Delete(ctx, userID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, userID); !errors.Is(err, domain.ErrLinkNotFound) {
		t.Fatalf("expected ErrLinkNotFound, got %v", err)
	}

	_, _ = db.Exec(`DELETE FROM community_users WHERE discord_id LIKE 'itlink-%'`)
}

func TestAccountLinkUsernameIsCaseInsensitive(t *testing.T) {
	db := openTestDB(t)
	_, _ = db.Exec(`DELETE FROM community_users WHERE discord_id LIKE 'itcase-%'`)
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM community_users WHERE discord_id LIKE 'itcase-%'`) })

	store := New(db)
	ctx := context.Background()

	first := seedUser(t, db, "itcase-1")
	if _, err := store.Upsert(ctx, first, "ITCASEUSER", nil); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	second := seedUser(t, db, "itcase-2")
	if _, err := store.Upsert(ctx, second, "itcaseuser", nil); !errors.Is(err, domain.ErrAccountAlreadyLinked) {
		t.Fatalf("expected ErrAccountAlreadyLinked for different casing, got %v", err)
	}
}
