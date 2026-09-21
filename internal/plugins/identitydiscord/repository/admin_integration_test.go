//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestRolesPermissionsAndMappings(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	ctx := context.Background()
	store := New(db)

	defs := []permissions.Definition{{Name: "gw.it.read", Description: "read", Owner: "identity-discord"}}
	if err := store.SyncPermissions(ctx, defs); err != nil {
		t.Fatalf("SyncPermissions: %v", err)
	}
	if err := store.UpsertRole(ctx, "it-admin"); err != nil {
		t.Fatalf("UpsertRole: %v", err)
	}
	if err := store.GrantRolePermission(ctx, "it-admin", "gw.it.read"); err != nil {
		t.Fatalf("GrantRolePermission: %v", err)
	}
	grants, err := store.ListRolePermissions(ctx)
	if err != nil || len(grants) == 0 || grants[0].Role != "it-admin" {
		t.Fatalf("ListRolePermissions = %+v, %v", grants, err)
	}
	roles, err := store.ListRoles(ctx)
	if err != nil || !containsString(roles, "it-admin") {
		t.Fatalf("ListRoles = %+v, %v", roles, err)
	}
	if err := store.ReplaceRolePermissions(ctx, "it-admin", []string{"gw.it.read"}); err != nil {
		t.Fatalf("ReplaceRolePermissions: %v", err)
	}
	if err := store.RevokeRolePermission(ctx, "it-admin", "gw.it.read"); err != nil {
		t.Fatalf("RevokeRolePermission: %v", err)
	}

	if err := store.UpsertDiscordRoleMapping(ctx, "disc-1", "it-admin"); err != nil {
		t.Fatalf("UpsertDiscordRoleMapping: %v", err)
	}
	mappings, err := store.ListDiscordRoleMappings(ctx)
	if err != nil || len(mappings) == 0 || mappings[0].DiscordRoleID != "disc-1" {
		t.Fatalf("ListDiscordRoleMappings = %+v, %v", mappings, err)
	}
	mapped, err := store.MappedRoles(ctx, []string{"disc-1"})
	if err != nil || !containsString(mapped, "it-admin") {
		t.Fatalf("MappedRoles = %+v, %v", mapped, err)
	}
	if empty, err := store.MappedRoles(ctx, nil); err != nil || empty != nil {
		t.Fatalf("MappedRoles nil = %+v, %v", empty, err)
	}
	if err := store.DeleteDiscordRoleMapping(ctx, "disc-1"); err != nil {
		t.Fatalf("DeleteDiscordRoleMapping: %v", err)
	}
}

func TestProvisionProfileAndUserRoles(t *testing.T) {
	db := integration.Postgres(t)
	reset(t, db)
	ctx := context.Background()
	store := New(db)

	userID, created, err := store.Provision(ctx, auth.DiscordUser{ID: "it-disc", Username: "ituser", GlobalName: "It User"})
	if err != nil || !created || userID == uuid.Nil {
		t.Fatalf("Provision = %s, %v, %v", userID, created, err)
	}
	again, createdAgain, err := store.Provision(ctx, auth.DiscordUser{ID: "it-disc", Username: "ituser"})
	if err != nil || again != userID || createdAgain {
		t.Fatalf("second Provision = %s, %v, %v", again, createdAgain, err)
	}

	if id, found, err := store.UserIDByDiscordID(ctx, "it-disc"); err != nil || !found || id != userID {
		t.Fatalf("UserIDByDiscordID = %s, %v, %v", id, found, err)
	}
	if id, found, _ := store.UserIDByDiscordID(ctx, "missing-disc"); found || id != uuid.Nil {
		t.Fatal("unknown discord id should not resolve")
	}

	profile, err := store.UserProfile(ctx, userID)
	if err != nil || profile.Username == "" {
		t.Fatalf("UserProfile = %+v, %v", profile, err)
	}

	users, err := store.ResolveUserByName(ctx, "ituser")
	if err != nil || len(users) == 0 {
		t.Fatalf("ResolveUserByName = %+v, %v", users, err)
	}

	if err := store.UpdateUserRoles(ctx, userID, []string{"it-admin"}); err != nil {
		t.Fatalf("UpdateUserRoles: %v", err)
	}
	userRoles, err := store.UserRoles(ctx, userID)
	if err != nil || !containsString(userRoles, "it-admin") {
		t.Fatalf("UserRoles = %+v, %v", userRoles, err)
	}
	if missing, err := store.UserRoles(ctx, uuid.New()); err != nil || missing != nil {
		t.Fatalf("UserRoles unknown = %+v, %v", missing, err)
	}

	if err := store.DeleteExpiredSessions(ctx); err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	if err := store.DeleteExpiredOAuthStates(ctx); err != nil {
		t.Fatalf("DeleteExpiredOAuthStates: %v", err)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
