//go:build integration

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func reset(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE sessions, oauth_states, discord_role_mappings,
		community_users, discord_identities, role_permissions, permissions, roles
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func TestConsumeOAuthStateRejectsExpired(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	if err := store.CreateOAuthState(ctx, "s1", "verifier-1", "/store", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("create: %v", err)
	}
	verifier, returnTo, err := store.ConsumeOAuthState(ctx, "s1")
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if verifier != "verifier-1" || returnTo != "/store" {
		t.Fatalf("got %q %q", verifier, returnTo)
	}
	if _, _, err := store.ConsumeOAuthState(ctx, "s1"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("replay error = %v, want ErrStateNotFound", err)
	}

	if err := store.CreateOAuthState(ctx, "s2", "v", "", time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("create expired: %v", err)
	}
	if _, _, err := store.ConsumeOAuthState(ctx, "s2"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expired error = %v, want ErrStateNotFound", err)
	}
}

func TestResolveUserByNameFindsExactMatchBeyondFuzzyPage(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		if _, _, err := store.Provision(ctx, auth.DiscordUser{
			ID:       fmt.Sprintf("fuzz-%d", i),
			Username: fmt.Sprintf("match-%d", i),
		}); err != nil {
			t.Fatalf("provision fuzzy %d: %v", i, err)
		}
	}
	exactID, _, err := store.Provision(ctx, auth.DiscordUser{ID: "exact", Username: "match", GlobalName: "match"})
	if err != nil {
		t.Fatalf("provision exact: %v", err)
	}

	users, err := store.ResolveUserByName(ctx, "match")
	if err != nil {
		t.Fatalf("ResolveUserByName: %v", err)
	}
	found := false
	for _, user := range users {
		if user.ID == exactID {
			found = true
		}
	}
	if !found {
		t.Fatalf("exact match not returned among %d results", len(users))
	}
}

func TestProvisionIsIdempotent(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	userID, created, err := store.Provision(ctx, auth.DiscordUser{ID: "42", Username: "user", GlobalName: "User"})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if !created {
		t.Fatal("first provision should report created")
	}

	userID2, created2, err := store.Provision(ctx, auth.DiscordUser{ID: "42", Username: "renamed", GlobalName: "Renamed"})
	if err != nil {
		t.Fatalf("Provision again: %v", err)
	}
	if created2 {
		t.Fatal("second provision must not report created")
	}
	if userID2 != userID {
		t.Fatalf("user id changed: %s -> %s", userID, userID2)
	}

	var count int
	if err := db.QueryRow(`SELECT count(*) FROM community_users`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("community_users rows = %d", count)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	userID, _, err := store.Provision(ctx, auth.DiscordUser{ID: "42", Username: "user"})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	session := auth.Session{
		ID:        "hash-abc",
		UserID:    userID,
		Roles:     []string{"member", "moderator"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := store.Create(ctx, session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := store.Get(ctx, "hash-abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DiscordID != "42" || len(got.Roles) != 2 {
		t.Fatalf("session = %+v", got)
	}

	if err := store.Delete(ctx, "hash-abc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, "hash-abc"); err != auth.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionWithEmptyRoles(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	userID, _, err := store.Provision(ctx, auth.DiscordUser{ID: "42", Username: "user"})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	for name, roles := range map[string][]string{"nil": nil, "empty": {}} {
		sessionID := "hash-" + name
		err := store.Create(ctx, auth.Session{
			ID:        sessionID,
			UserID:    userID,
			Roles:     roles,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		})
		if err != nil {
			t.Fatalf("%s roles: Create: %v", name, err)
		}
		got, err := store.Get(ctx, sessionID)
		if err != nil {
			t.Fatalf("%s roles: Get: %v", name, err)
		}
		if len(got.Roles) != 0 {
			t.Fatalf("%s roles: got %v", name, got.Roles)
		}
	}

	if err := store.UpdateUserRoles(ctx, userID, nil); err != nil {
		t.Fatalf("UpdateUserRoles nil: %v", err)
	}
}

func TestOAuthStateIsSingleUse(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	if err := store.CreateOAuthState(ctx, "state-1", "verifier-1", "/dashboard", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatalf("CreateOAuthState: %v", err)
	}
	verifier, returnTo, err := store.ConsumeOAuthState(ctx, "state-1")
	if err != nil {
		t.Fatalf("ConsumeOAuthState: %v", err)
	}
	if verifier != "verifier-1" || returnTo != "/dashboard" {
		t.Fatalf("verifier = %q returnTo = %q", verifier, returnTo)
	}
	if _, _, err := store.ConsumeOAuthState(ctx, "state-1"); err != ErrStateNotFound {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestMappedRoles(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	if _, err := db.Exec(`INSERT INTO roles (name) VALUES ('member'), ('moderator')`); err != nil {
		t.Fatalf("seed roles: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO discord_role_mappings (discord_role_id, role)
		VALUES ('role-1', 'member'), ('role-2', 'moderator')`); err != nil {
		t.Fatalf("seed mappings: %v", err)
	}

	roles, err := store.MappedRoles(ctx, []string{"role-1", "role-2", "role-unmapped"})
	if err != nil {
		t.Fatalf("MappedRoles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("roles = %v", roles)
	}

	none, err := store.MappedRoles(ctx, []string{"role-unmapped"})
	if err != nil {
		t.Fatalf("MappedRoles: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("unmapped roles = %v", none)
	}
}

func TestListCommunityUsers(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	store := New(db)
	ctx := context.Background()

	userID, _, err := store.Provision(ctx, auth.DiscordUser{ID: "42", Username: "alice", GlobalName: "Alice"})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	users, err := store.ListUsers(ctx, "ali", 10, 0)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 1 || users[0].ID != userID || users[0].Username != "alice" || users[0].GlobalName != "Alice" {
		t.Fatalf("users = %+v", users)
	}

	byID, err := store.ListUsers(ctx, "42", 10, 0)
	if err != nil {
		t.Fatalf("ListUsers by id: %v", err)
	}
	if len(byID) != 1 || byID[0].ID != userID {
		t.Fatalf("byID = %+v", byID)
	}

	all, err := store.ListUsers(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListUsers all: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("all = %d", len(all))
	}
}
