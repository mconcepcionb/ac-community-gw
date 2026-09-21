package adminnotes

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

func TestHandleCreateSuccess(t *testing.T) {
	store := &fakeStore{}
	plugin := New(Config{Store: store})
	author := uuid.New()
	rec := testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.WithPrincipal(
		testsupport.NewRequest(http.MethodPost, "/x", `{"target_type":"account","target_id":"ADMIN","body":"note"}`), author))

	testsupport.AssertStatus(t, rec, http.StatusCreated)
	if store.created.AuthorID != author || store.created.TargetID != "ADMIN" {
		t.Fatalf("created = %+v", store.created)
	}
}

func TestHandleCreateErrors(t *testing.T) {
	rec := testsupport.NewRecorder()
	New(Config{}).handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/x", `{}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store := &fakeStore{}
	plugin := New(Config{Store: store})
	rec = testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/x",
		`{"target_type":"account","target_id":"x","body":"`+strings.Repeat("a", maxBodyLen+1)+`"}`))
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	store.createErr = errors.New("db")
	rec = testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"target_type":"account","target_id":"x","body":"n"}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleList(t *testing.T) {
	rec := testsupport.NewRecorder()
	New(Config{}).handleList(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store := &fakeStore{annotations: []domain.Annotation{{ID: uuid.New(), TargetID: "ADMIN", Body: "note"}}}
	plugin := New(Config{Store: store})
	rec = testsupport.NewRecorder()
	plugin.handleList(rec, testsupport.NewRequest(http.MethodGet, "/x?target_type=account&target_id=ADMIN", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), "ADMIN") {
		t.Fatalf("body = %s", rec.Body.String())
	}

	store.byTargetErr = errors.New("db")
	rec = testsupport.NewRecorder()
	plugin.handleList(rec, testsupport.NewRequest(http.MethodGet, "/x?target_type=account&target_id=ADMIN", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleUpdateErrors(t *testing.T) {
	rec := testsupport.NewRecorder()
	New(Config{}).handleUpdate(rec, testsupport.NewRequest(http.MethodPatch, "/x", `{}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store := &fakeStore{}
	plugin := New(Config{Store: store})

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPatch, "/x", `{"body":"b"}`), "id", "bad")
	rec = testsupport.NewRecorder()
	plugin.handleUpdate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPatch, "/x", `{"body":""}`), "id", uuid.NewString())
	rec = testsupport.NewRecorder()
	plugin.handleUpdate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	store.getErr = domain.ErrAnnotationNotFound
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPatch, "/x", `{"body":"b"}`), "id", uuid.NewString())
	rec = testsupport.NewRecorder()
	plugin.handleUpdate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	store.getErr = errors.New("db")
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPatch, "/x", `{"body":"b"}`), "id", uuid.NewString())
	rec = testsupport.NewRecorder()
	plugin.handleUpdate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store.getErr = nil
	store.updateErr = domain.ErrAnnotationNotFound
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPatch, "/x", `{"body":"b"}`), "id", uuid.NewString())
	rec = testsupport.NewRecorder()
	plugin.handleUpdate(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)
}

func TestHandleDelete(t *testing.T) {
	rec := testsupport.NewRecorder()
	New(Config{}).handleDelete(rec, testsupport.NewRequest(http.MethodDelete, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	author := uuid.New()
	annotationID := uuid.New()
	store := &fakeStore{annotation: domain.Annotation{ID: annotationID, AuthorID: author, TargetID: "ADMIN"}}
	plugin := New(Config{Store: store})

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodDelete, "/x", ""), "id", "bad")
	rec = testsupport.NewRecorder()
	plugin.handleDelete(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	store.getErr = domain.ErrAnnotationNotFound
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodDelete, "/x", ""), "id", annotationID.String())
	rec = testsupport.NewRecorder()
	plugin.handleDelete(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	store.getErr = nil
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodDelete, "/x", ""), "id", annotationID.String())
	rec = testsupport.NewRecorder()
	plugin.handleDelete(rec, testsupport.WithPrincipal(req, uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusForbidden)

	rec = testsupport.NewRecorder()
	plugin.handleDelete(rec, testsupport.WithPrincipal(req, author))
	testsupport.AssertStatus(t, rec, http.StatusNoContent)
	if store.deleted != annotationID {
		t.Fatal("annotation was not deleted")
	}

	store.deleteErr = errors.New("db")
	rec = testsupport.NewRecorder()
	plugin.handleDelete(rec, testsupport.WithPrincipal(req, author))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestCanManage(t *testing.T) {
	if New(Config{}).canManage([]string{"admin"}) {
		t.Fatal("canManage must be false without an authorizer")
	}
	authorizer := permissions.NewAuthorizer()
	authorizer.Grant("admin", PermissionManage)
	plugin := New(Config{Authorizer: authorizer})
	if !plugin.canManage([]string{"admin"}) || plugin.canManage([]string{"other"}) {
		t.Fatal("canManage mismatch")
	}
}

func TestPermissionDefsAreNamespaced(t *testing.T) {
	for _, def := range permissionDefs() {
		if def.Owner != Name || def.Namespace != "gw" || !strings.HasPrefix(string(def.Name), "gw.notes.") {
			t.Fatalf("definition not namespaced: %+v", def)
		}
	}
}

func TestRegisterWiresNotes(t *testing.T) {
	reg := &plugins.Registry{
		Mux:         http.NewServeMux(),
		Permissions: permissions.NewRegistry(),
		Audit:       audit.NopRecorder{},
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
	}
	plugin := New(Config{Store: &fakeStore{}})
	if err := plugin.Register(context.Background(), reg); err != nil {
		t.Fatalf("Register = %v", err)
	}
	for _, perm := range []permissions.Permission{PermissionRead, PermissionWrite, PermissionManage} {
		if !reg.Permissions.Has(perm) {
			t.Fatalf("permission %s not registered", perm)
		}
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/annotations?target_type=account&target_id=ADMIN", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}
