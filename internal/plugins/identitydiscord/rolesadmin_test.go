package identitydiscord

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	identitydiscordrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord/repository/generated"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

func roleRegistry(perms ...permissions.Permission) *permissions.Registry {
	registry := permissions.NewRegistry()
	for _, perm := range perms {
		_ = registry.Register(permissions.Definition{
			Name: perm, Description: string(perm), Owner: Name, Namespace: "gw",
		})
	}
	return registry
}

func TestHandleListRoles(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleListRoles(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	repo := &fakeRepo{
		rolesList:  []string{"admin"},
		roleGrants: []identitydiscordrepo.RolePermission{{Role: "admin", Permission: "gw.identity.roles.manage"}},
		mappings:   []identitydiscordrepo.DiscordRoleMapping{{DiscordRoleID: "1", Role: "admin"}},
	}
	plugin = New(Config{Repository: repo})
	plugin.permissions = roleRegistry(PermissionRolesManage)
	rec = testsupport.NewRecorder()
	plugin.handleListRoles(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)

	body := rec.Body.String()
	for _, want := range []string{"admin", "gw.identity.roles.manage", "1"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body %q missing %q", body, want)
		}
	}

	repo.listRolesErr = errors.New("db")
	rec = testsupport.NewRecorder()
	plugin.handleListRoles(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleReplaceRolePermissions(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleReplaceRolePermissions(rec, testsupport.NewRequest(http.MethodPut, "/x", `{}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	repo := &fakeRepo{}
	plugin = New(Config{Repository: repo})
	plugin.permissions = roleRegistry(PermissionRolesManage)

	// empty role
	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"permissions":[]}`), "role", "")
	rec = testsupport.NewRecorder()
	plugin.handleReplaceRolePermissions(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// unknown permission
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"permissions":["gw.identity.roles.manage","nope"]}`), "role", "admin")
	rec = testsupport.NewRecorder()
	plugin.handleReplaceRolePermissions(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// success normalizes the set
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"permissions":[" gw.identity.roles.manage ","gw.identity.roles.manage"]}`), "role", "admin")
	rec = testsupport.NewRecorder()
	plugin.handleReplaceRolePermissions(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNoContent)
	if repo.lastReplaceRole != "admin" || len(repo.lastReplacePerms) != 1 || repo.lastReplacePerms[0] != "gw.identity.roles.manage" {
		t.Fatalf("replace = %q,%v", repo.lastReplaceRole, repo.lastReplacePerms)
	}

	repo.replaceErr = errors.New("db")
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"permissions":[]}`), "role", "admin")
	rec = testsupport.NewRecorder()
	plugin.handleReplaceRolePermissions(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleDiscordMappings(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleUpsertDiscordMapping(rec, testsupport.NewRequest(http.MethodPut, "/x", `{}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	repo := &fakeRepo{}
	plugin = New(Config{Repository: repo})

	// missing role
	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"role":""}`), "discord_role_id", "1")
	rec = testsupport.NewRecorder()
	plugin.handleUpsertDiscordMapping(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// success
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"role":"admin"}`), "discord_role_id", "1")
	rec = testsupport.NewRecorder()
	plugin.handleUpsertDiscordMapping(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNoContent)
	if repo.lastMappingID != "1" || repo.lastMappingRole != "admin" {
		t.Fatalf("mapping = %q,%q", repo.lastMappingID, repo.lastMappingRole)
	}

	// upsert role error
	repo.upsertRoleErr = errors.New("db")
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"role":"admin"}`), "discord_role_id", "1")
	rec = testsupport.NewRecorder()
	plugin.handleUpsertDiscordMapping(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	repo.upsertRoleErr = nil

	// mapping error
	repo.upsertErr = errors.New("db")
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"role":"admin"}`), "discord_role_id", "1")
	rec = testsupport.NewRecorder()
	plugin.handleUpsertDiscordMapping(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	repo.upsertErr = nil

	// delete: missing id
	rec = testsupport.NewRecorder()
	plugin.handleDeleteDiscordMapping(rec, testsupport.NewRequest(http.MethodDelete, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// delete: error
	repo.deleteErr = errors.New("db")
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodDelete, "/x", ""), "discord_role_id", "1")
	rec = testsupport.NewRecorder()
	plugin.handleDeleteDiscordMapping(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// delete: success
	repo.deleteErr = nil
	rec = testsupport.NewRecorder()
	plugin.handleDeleteDiscordMapping(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNoContent)
}

func TestNormalizePermissions(t *testing.T) {
	got := normalizePermissions([]string{" b ", "a", "", "a", "c"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("normalize = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalize = %v, want %v", got, want)
		}
	}
}

func TestPermissionCatalog(t *testing.T) {
	if catalog := New(Config{}).permissionCatalog(); len(catalog) != 0 {
		t.Fatalf("catalog = %v", catalog)
	}
	plugin := New(Config{})
	plugin.permissions = roleRegistry(PermissionUserRead)
	catalog := plugin.permissionCatalog()
	if len(catalog) != 1 || catalog[0].Name != string(PermissionUserRead) {
		t.Fatalf("catalog = %+v", catalog)
	}
}

func TestEffectivePermissions(t *testing.T) {
	if perms := New(Config{}).effectivePermissions([]string{"admin"}); len(perms) != 0 {
		t.Fatalf("perms = %v", perms)
	}
	authorizer := permissions.NewAuthorizer()
	authorizer.Grant("admin", PermissionRolesManage)
	plugin := New(Config{Authorizer: authorizer})
	perms := plugin.effectivePermissions([]string{"admin"})
	if len(perms) != 1 || perms[0] != string(PermissionRolesManage) {
		t.Fatalf("perms = %v", perms)
	}
}

func TestProvision(t *testing.T) {
	repo := &fakeRepo{}
	plugin := New(Config{Repository: repo})
	req := testsupport.NewRequest(http.MethodGet, "/x", "")
	userID, created, err := plugin.provision(req, auth.DiscordUser{ID: "42"})
	if err != nil || created || userID == uuid.Nil {
		t.Fatalf("provision = %s,%v,%v", userID, created, err)
	}
}

type fakeCleaner struct {
	sessions int
	states   int
}

func (f *fakeCleaner) DeleteExpiredSessions(context.Context) error    { f.sessions++; return nil }
func (f *fakeCleaner) DeleteExpiredOAuthStates(context.Context) error { f.states++; return nil }

func TestRunCleanup(t *testing.T) {
	// A nil cleaner returns immediately.
	RunCleanup(context.Background(), time.Millisecond, nil, nil)

	cleaner := &fakeCleaner{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		RunCleanup(ctx, time.Millisecond, slog.Default(), cleaner)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for cleaner.sessions == 0 || cleaner.states == 0 {
		select {
		case <-deadline:
			cancel()
			t.Fatal("cleanup never ran")
		case <-time.After(time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunCleanup did not stop after cancel")
	}
}

func TestEventNames(t *testing.T) {
	if (UserAuthenticated{}).EventName() != EventUserAuthenticated ||
		(DiscordRolesChanged{}).EventName() != EventDiscordRolesChanged ||
		(UserProvisioned{}).EventName() != EventUserProvisioned {
		t.Fatal("event names mismatch")
	}
}

func TestRegisterWiresIdentity(t *testing.T) {
	reg := &plugins.Registry{
		Mux:         http.NewServeMux(),
		Services:    services.NewRegistry(),
		Permissions: permissions.NewRegistry(),
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
		RateLimit: func(next http.Handler) http.Handler { return next },
	}
	plugin := New(Config{
		Provider:   &fakeProvider{},
		States:     NewMemoryStateStore(),
		Repository: &fakeRepo{},
	})
	if err := plugin.Register(context.Background(), reg); err != nil {
		t.Fatalf("Register = %v", err)
	}
	if !reg.Permissions.Has(PermissionUserRead) || !reg.Permissions.Has(PermissionRolesManage) {
		t.Fatal("identity permissions not registered")
	}
	for _, name := range []string{ServiceUserDirectory, ServiceUserAdmin} {
		if !reg.Services.Has(name) {
			t.Fatalf("capability %s not published", name)
		}
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/identity/users", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}
