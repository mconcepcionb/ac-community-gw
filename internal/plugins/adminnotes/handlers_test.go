package adminnotes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/domain"
)

type fakeStore struct {
	annotation domain.Annotation
	getErr     error
}

func (f *fakeStore) Create(
	_ context.Context,
	targetType, targetID string,
	authorID uuid.UUID,
	body string,
) (domain.Annotation, error) {
	return domain.Annotation{
		ID:         uuid.New(),
		TargetType: targetType,
		TargetID:   targetID,
		AuthorID:   authorID,
		Body:       body,
	}, nil
}

func (f *fakeStore) ByTarget(context.Context, string, string, int, int) ([]domain.Annotation, error) {
	return nil, nil
}

func (f *fakeStore) Get(context.Context, uuid.UUID) (domain.Annotation, error) {
	return f.annotation, f.getErr
}

func (f *fakeStore) Update(_ context.Context, id uuid.UUID, body string) (domain.Annotation, error) {
	annotation := f.annotation
	annotation.ID = id
	annotation.Body = body
	return annotation, nil
}

func (f *fakeStore) Delete(context.Context, uuid.UUID) error { return nil }

func TestHandleCreateRejectsMissingFields(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/annotations",
		strings.NewReader(`{"target_type":"","target_id":"","body":""}`))
	rec := httptest.NewRecorder()
	plugin.handleCreate(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestHandleCreateRejectsUnknownTargetType(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/annotations",
		strings.NewReader(`{"target_type":"realm","target_id":"x","body":"note"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreate(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestHandleListRejectsMissingTarget(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/annotations", nil)
	rec := httptest.NewRecorder()
	plugin.handleList(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestHandleUpdateForbiddenForOtherAuthor(t *testing.T) {
	author := uuid.New()
	plugin := New(Config{
		Store:      &fakeStore{annotation: domain.Annotation{ID: uuid.New(), AuthorID: author}},
		Authorizer: permissions.NewAuthorizer(),
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/annotations/"+uuid.New().String(),
		strings.NewReader(`{"body":"updated"}`))
	req.SetPathValue("id", uuid.New().String())
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: uuid.New()}))
	rec := httptest.NewRecorder()
	plugin.handleUpdate(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestHandleUpdateAllowsAuthor(t *testing.T) {
	author := uuid.New()
	annotationID := uuid.New()
	plugin := New(Config{
		Store: &fakeStore{annotation: domain.Annotation{ID: annotationID, AuthorID: author}},
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/annotations/"+annotationID.String(),
		strings.NewReader(`{"body":"updated"}`))
	req.SetPathValue("id", annotationID.String())
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: author}))
	rec := httptest.NewRecorder()
	plugin.handleUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHandleUpdateAllowsModerator(t *testing.T) {
	author := uuid.New()
	annotationID := uuid.New()
	authorizer := permissions.NewAuthorizer()
	authorizer.Grant("moderator", PermissionManage)
	plugin := New(Config{
		Store:      &fakeStore{annotation: domain.Annotation{ID: annotationID, AuthorID: author}},
		Authorizer: authorizer,
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/annotations/"+annotationID.String(),
		strings.NewReader(`{"body":"moderated"}`))
	req.SetPathValue("id", annotationID.String())
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID: uuid.New(),
		Roles:  []string{"moderator"},
	}))
	rec := httptest.NewRecorder()
	plugin.handleUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
