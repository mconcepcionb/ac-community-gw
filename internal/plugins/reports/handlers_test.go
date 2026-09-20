package reports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/domain"
)

type fakeStore struct {
	closedErr error
}

func (f *fakeStore) Create(_ context.Context, reporterID uuid.UUID, target, category, message string) (domain.Report, error) {
	return domain.Report{
		ID:         uuid.New(),
		ReporterID: reporterID,
		Target:     target,
		Category:   category,
		Message:    message,
		Status:     domain.StatusOpen,
	}, nil
}

func (f *fakeStore) ByReporter(context.Context, uuid.UUID, int, int) ([]domain.Report, error) {
	return nil, nil
}

func (f *fakeStore) List(context.Context, string, int, int) ([]domain.Report, error) {
	return nil, nil
}

func (f *fakeStore) Close(context.Context, uuid.UUID) (domain.Report, error) {
	return domain.Report{}, f.closedErr
}

func TestHandleCreateRejectsMissingFields(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports",
		strings.NewReader(`{"target":"","category":"","message":""}`))
	rec := httptest.NewRecorder()
	plugin.handleCreate(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleCloseNotFound(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{closedErr: domain.ErrReportNotFound}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/reports/"+uuid.New().String()+"/close", nil)
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()
	plugin.handleClose(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}
