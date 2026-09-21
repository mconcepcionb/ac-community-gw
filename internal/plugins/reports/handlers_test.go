package reports

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
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

type fakeStore struct {
	closedErr     error
	createErr     error
	byReporterErr error
	listErr       error

	reports       []domain.Report
	createdTarget string
	listStatus    string
	lastLimit     int
	lastOffset    int
}

func (f *fakeStore) Create(_ context.Context, reporterID uuid.UUID, target, category, message string) (domain.Report, error) {
	f.createdTarget = target
	if f.createErr != nil {
		return domain.Report{}, f.createErr
	}
	return domain.Report{
		ID:         uuid.New(),
		ReporterID: reporterID,
		Target:     target,
		Category:   category,
		Message:    message,
		Status:     domain.StatusOpen,
		CreatedAt:  time.Now(),
	}, nil
}

func (f *fakeStore) ByReporter(_ context.Context, _ uuid.UUID, limit, offset int) ([]domain.Report, error) {
	f.lastLimit, f.lastOffset = limit, offset
	return f.reports, f.byReporterErr
}

func (f *fakeStore) List(_ context.Context, status string, limit, offset int) ([]domain.Report, error) {
	f.listStatus = status
	f.lastLimit, f.lastOffset = limit, offset
	return f.reports, f.listErr
}

func (f *fakeStore) Close(context.Context, uuid.UUID) (domain.Report, error) {
	if f.closedErr != nil {
		return domain.Report{}, f.closedErr
	}
	return domain.Report{ID: uuid.New(), Status: domain.StatusClosed, CreatedAt: time.Now()}, nil
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

func TestHandleCreateRejectsMissingFields(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})
	rec := testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/api/v1/reports",
		`{"target":"","category":"","message":""}`))

	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)
}

func TestHandleCreateRejectsOverlongMessage(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{}})
	body := `{"target":"thrall","category":"abuse","message":"` + strings.Repeat("x", maxMessageLen+1) + `"}`
	rec := testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/api/v1/reports", body))

	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)
}

func TestHandleCreateSuccess(t *testing.T) {
	store := &fakeStore{}
	plugin := New(Config{Store: store})
	rec := testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.WithPrincipal(
		testsupport.NewRequest(http.MethodPost, "/api/v1/reports", `{"target":" thrall ","category":"abuse","message":" spam "}`),
		uuid.New()))

	testsupport.AssertStatus(t, rec, http.StatusCreated)
	if store.createdTarget != "thrall" {
		t.Fatalf("target = %q, want trimmed", store.createdTarget)
	}

	var body Report
	testsupport.DecodeJSON(t, rec, &body)
	if body.Target != "thrall" || body.Status != domain.StatusOpen {
		t.Fatalf("report = %+v", body)
	}
}

func TestHandleMineParsesPagination(t *testing.T) {
	store := &fakeStore{reports: []domain.Report{{ID: uuid.New(), Target: "thrall", Status: domain.StatusOpen}}}
	plugin := New(Config{Store: store})

	rec := testsupport.NewRecorder()
	plugin.handleMine(rec, testsupport.WithPrincipal(
		testsupport.NewRequest(http.MethodGet, "/api/v1/reports/mine?limit=5&offset=2", ""), uuid.New()))

	testsupport.AssertStatus(t, rec, http.StatusOK)
	if store.lastLimit != 5 || store.lastOffset != 2 {
		t.Fatalf("pagination = %d/%d, want 5/2", store.lastLimit, store.lastOffset)
	}

	var body ReportsResponse
	testsupport.DecodeJSON(t, rec, &body)
	if len(body.Reports) != 1 {
		t.Fatalf("reports = %+v", body.Reports)
	}
}

func TestHandleListPassesStatus(t *testing.T) {
	store := &fakeStore{}
	plugin := New(Config{Store: store})

	rec := testsupport.NewRecorder()
	plugin.handleList(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/reports?status=open", ""))

	testsupport.AssertStatus(t, rec, http.StatusOK)
	if store.listStatus != "open" {
		t.Fatalf("status = %q", store.listStatus)
	}
}

func TestHandleClose(t *testing.T) {
	plugin := New(Config{Store: &fakeStore{closedErr: domain.ErrReportNotFound}})
	req := testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodPost, "/api/v1/admin/reports/x/close", ""), "id", uuid.New().String())
	rec := testsupport.NewRecorder()
	plugin.handleClose(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	plugin = New(Config{Store: &fakeStore{}})
	req = testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodPost, "/api/v1/admin/reports/x/close", ""), "id", uuid.New().String())
	rec = testsupport.NewRecorder()
	plugin.handleClose(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)
}

func TestHandlersUnavailableWithoutStore(t *testing.T) {
	plugin := New(Config{})

	rec := testsupport.NewRecorder()
	plugin.handleCreate(rec, testsupport.NewRequest(http.MethodPost, "/api/v1/reports", `{}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	rec = testsupport.NewRecorder()
	plugin.handleMine(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/reports/mine", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	rec = testsupport.NewRecorder()
	plugin.handleList(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/reports", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	rec = testsupport.NewRecorder()
	plugin.handleClose(rec, testsupport.NewRequest(http.MethodPost, "/api/v1/admin/reports/x/close", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestPermissionDefsAreNamespaced(t *testing.T) {
	defs := permissionDefs()
	if len(defs) != 2 {
		t.Fatalf("defs = %d, want 2", len(defs))
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
	if !reg.Permissions.Has(PermissionCreate) || !reg.Permissions.Has(PermissionRead) {
		t.Fatal("both report permissions must be registered")
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/admin/reports", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}
