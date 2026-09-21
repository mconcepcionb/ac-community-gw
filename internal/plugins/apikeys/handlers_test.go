package apikeys

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

type fakeStore struct {
	keys      []domain.APIKey
	listErr   error
	createErr error
	rotateErr error
	revokeErr error

	createdName  string
	createdPerms []string
	revoked      uuid.UUID
}

func (f *fakeStore) List(context.Context) ([]domain.APIKey, error) {
	return f.keys, f.listErr
}

func (f *fakeStore) Create(_ context.Context, name string, perms []string) (domain.APIKey, string, error) {
	f.createdName = name
	f.createdPerms = perms
	if f.createErr != nil {
		return domain.APIKey{}, "", f.createErr
	}
	return domain.APIKey{
		ID: uuid.New(), Name: name, KeyPrefix: "abc123", Permissions: perms, CreatedAt: time.Now(),
	}, "one-time-secret", nil
}

func (f *fakeStore) Rotate(_ context.Context, id uuid.UUID) (domain.APIKey, string, error) {
	if f.rotateErr != nil {
		return domain.APIKey{}, "", f.rotateErr
	}
	return domain.APIKey{ID: id, Name: "rotated", KeyPrefix: "abc123", CreatedAt: time.Now()}, "rotated-secret", nil
}

func (f *fakeStore) Revoke(_ context.Context, id uuid.UUID) error {
	f.revoked = id
	return f.revokeErr
}

func newTestRegistry() *plugins.Registry {
	return &plugins.Registry{
		Mux:         http.NewServeMux(),
		Permissions: permissions.NewRegistry(),
		Audit:       audit.NopRecorder{},
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
		RateLimit: func(next http.Handler) http.Handler { return next },
	}
}

func TestHandleListReturnsKeysWithoutSecrets(t *testing.T) {
	lastUsed := time.Now()
	store := &fakeStore{keys: []domain.APIKey{
		{ID: uuid.New(), Name: "ci", KeyPrefix: "abc123", Permissions: []string{"gw.report.read"}, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "bot", KeyPrefix: "def456", CreatedAt: time.Now(), LastUsedAt: &lastUsed},
	}}
	plugin := New(Config{Store: store})

	rec := testsupport.NewRecorder()
	plugin.handleList(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/api-keys", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)

	var body APIKeysResponse
	testsupport.DecodeJSON(t, rec, &body)
	if len(body.Keys) != 2 || body.Keys[0].KeyPrefix != "abc123" {
		t.Fatalf("keys = %+v", body.Keys)
	}
	if body.Keys[1].LastUsedAt == "" {
		t.Fatal("last_used_at should be rendered when present")
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("list response must not contain a secret: %s", rec.Body.String())
	}
}

func TestHandleListUnavailable(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleList(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/api-keys", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleCreateValidatesAndCleansPermissions(t *testing.T) {
	store := &fakeStore{}
	plugin := New(Config{Store: store})

	rec := testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/api/v1/admin/api-keys", `{"name":"","permissions":["gw.report.read"]}`))
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	rec = testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/api/v1/admin/api-keys", `{"name":"ci","permissions":[" ","gw.report.read","gw.report.read","gw.audit.read"]}`))
	testsupport.AssertStatus(t, rec, http.StatusCreated)

	want := []string{"gw.report.read", "gw.audit.read"}
	if len(store.createdPerms) != len(want) {
		t.Fatalf("created permissions = %v, want %v", store.createdPerms, want)
	}
	for i, perm := range want {
		if store.createdPerms[i] != perm {
			t.Fatalf("created permissions = %v, want %v", store.createdPerms, want)
		}
	}

	var body APIKeySecretResponse
	testsupport.DecodeJSON(t, rec, &body)
	if body.Secret != "one-time-secret" {
		t.Fatalf("secret = %q", body.Secret)
	}
}

func TestHandleRotate(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})

	req := testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodPost, "/api/v1/admin/api-keys/x/rotate", ""), "id", "not-a-uuid")
	rec := testsupport.NewRecorder()
	plugin.handleRotate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	req = testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodPost, "/api/v1/admin/api-keys/x/rotate", ""), "id", uuid.New().String())
	rec = testsupport.NewRecorder()
	plugin.handleRotate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)
}

func TestHandleRotateNotFound(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{rotateErr: domain.ErrKeyNotFound}})
	req := testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodPost, "/api/v1/admin/api-keys/x/rotate", ""), "id", uuid.New().String())

	rec := testsupport.NewRecorder()
	plugin.handleRotate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)
}

func TestHandleRevoke(t *testing.T) {
	store := &fakeStore{}
	plugin := New(Config{Store: store})

	req := testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodDelete, "/api/v1/admin/api-keys/x", ""), "id", "bad")
	rec := testsupport.NewRecorder()
	plugin.handleRevoke(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	id := uuid.New()
	req = testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodDelete, "/api/v1/admin/api-keys/x", ""), "id", id.String())
	rec = testsupport.NewRecorder()
	plugin.handleRevoke(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNoContent)
	if store.revoked != id {
		t.Fatalf("revoked = %s, want %s", store.revoked, id)
	}
}

func TestPermissionDefsAreNamespaced(t *testing.T) {
	defs := permissionDefs()
	if len(defs) == 0 {
		t.Fatal("expected at least one permission definition")
	}
	for _, def := range defs {
		if def.Owner != Name || def.Namespace != "gw" || !strings.HasPrefix(string(def.Name), "gw.") {
			t.Fatalf("definition not namespaced: %+v", def)
		}
	}
}

func TestRegisterWiresRoutesAndPermissions(t *testing.T) {
	reg := newTestRegistry()
	plugin := New(Config{Store: &fakeStore{}})
	if err := plugin.Register(context.Background(), reg); err != nil {
		t.Fatalf("Register = %v", err)
	}

	if !reg.Permissions.Has(PermissionManage) {
		t.Fatal("PermissionManage must be registered")
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/api-keys", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}
