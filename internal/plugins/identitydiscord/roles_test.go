package identitydiscord

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
)

func TestNormalizeRoles(t *testing.T) {
	got := normalizeRoles([]string{"b", "a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("normalizeRoles = %v", got)
	}
	if normalizeRoles(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestSameRoles(t *testing.T) {
	if !sameRoles([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatal("equal slices should match")
	}
	if sameRoles([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("different lengths should not match")
	}
	if sameRoles([]string{"a"}, []string{"b"}) {
		t.Fatal("different values should not match")
	}
}

func TestSyncRolesMapsDiscordRoles(t *testing.T) {
	provider := &fakeProvider{roleIDs: []string{"role-1", "role-2"}}
	repo := &fakeRepo{mapped: []string{"member"}}
	plugin := New(Config{Provider: provider, Repository: repo, GuildID: "guild-1"})

	roles, changed, err := plugin.syncRoles(httptest.NewRequest(http.MethodGet, "/", nil), "token", uuid.New())
	if err != nil {
		t.Fatalf("syncRoles: %v", err)
	}
	if !changed || len(roles) != 1 || roles[0] != "member" {
		t.Fatalf("roles = %v changed = %v", roles, changed)
	}
}

func TestSyncRolesUnchangedDoesNotUpdate(t *testing.T) {
	provider := &fakeProvider{roleIDs: []string{"role-1"}}
	repo := &fakeRepo{mapped: []string{"member"}, roles: []string{"member"}}
	plugin := New(Config{Provider: provider, Repository: repo, GuildID: "guild-1"})

	_, changed, err := plugin.syncRoles(httptest.NewRequest(http.MethodGet, "/", nil), "token", uuid.New())
	if err != nil {
		t.Fatalf("syncRoles: %v", err)
	}
	if changed {
		t.Fatal("expected no change")
	}
}

func TestSyncRolesWithoutGuildKeepsPrevious(t *testing.T) {
	repo := &fakeRepo{roles: []string{"member"}}
	plugin := New(Config{Repository: repo})

	roles, changed, err := plugin.syncRoles(httptest.NewRequest(http.MethodGet, "/", nil), "token", uuid.New())
	if err != nil {
		t.Fatalf("syncRoles: %v", err)
	}
	if changed || len(roles) != 1 || roles[0] != "member" {
		t.Fatalf("roles = %v changed = %v", roles, changed)
	}
}

func TestSyncRolesUsesConfiguredMapping(t *testing.T) {
	provider := &fakeProvider{roleIDs: []string{"1550848038216409149"}}
	repo := &fakeRepo{}
	plugin := New(Config{
		Provider:     provider,
		Repository:   repo,
		GuildID:      "guild-1",
		RoleMappings: map[string]string{"1550848038216409149": "ac-core.admin"},
	})

	roles, changed, err := plugin.syncRoles(httptest.NewRequest(http.MethodGet, "/", nil), "token", uuid.New())
	if err != nil {
		t.Fatalf("syncRoles: %v", err)
	}
	if !changed || len(roles) != 1 || roles[0] != "ac-core.admin" {
		t.Fatalf("roles = %v changed = %v", roles, changed)
	}
}

var _ auth.DiscordProvider = (*fakeProvider)(nil)
