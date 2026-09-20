package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/persistence"
)

type fakeCheck struct {
	name string
	err  error
}

func (f fakeCheck) Name() string                  { return f.name }
func (f fakeCheck) Check(_ context.Context) error { return f.err }

func testServer(t *testing.T, checks ...persistence.Check) (*Server, *auth.Manager, *permissions.Authorizer) {
	t.Helper()
	readiness := persistence.NewRegistry()
	for _, check := range checks {
		readiness.Add(check)
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:      auth.NewMemoryStore(),
		TTL:        time.Hour,
		Secure:     false,
		CookieName: "acgw_session",
	})
	authorizer := permissions.NewAuthorizer()
	server := New(Dependencies{
		Config:      &config.Config{HTTPAddr: ":0", ShutdownTimeout: time.Second},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Sessions:    sessions,
		Permissions: permissions.NewRegistry(),
		Authorizer:  authorizer,
		Audit:       audit.NopRecorder{},
		Readiness:   readiness,
	})
	return server, sessions, authorizer
}

func request(handler http.Handler, method, target string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	server, _, _ := testServer(t)
	rec := request(server.Handler(), http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReadyzReady(t *testing.T) {
	server, _, _ := testServer(t, fakeCheck{name: "postgres"})
	rec := request(server.Handler(), http.MethodGet, "/readyz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReadyzUnavailable(t *testing.T) {
	server, _, _ := testServer(t, fakeCheck{
		name: "postgres",
		err:  errors.New("dial tcp 10.0.0.5:5432: connection refused"),
	})
	rec := request(server.Handler(), http.MethodGet, "/readyz", nil)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "connection refused") || strings.Contains(body, "10.0.0.5") {
		t.Fatalf("internal error leaked: %s", body)
	}
	if !strings.Contains(body, "unavailable") {
		t.Fatalf("body = %s", body)
	}
}

func TestRequireAuthRejectsAnonymous(t *testing.T) {
	server, _, _ := testServer(t)
	server.Mux().Handle("GET /api/v1/me", server.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})))

	rec := request(server.Handler(), http.MethodGet, "/api/v1/me", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequirePermissionForbidden(t *testing.T) {
	server, sessions, _ := testServer(t)
	server.Mux().Handle("GET /api/v1/protected", server.RequirePermission("test.permission", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})))

	token, err := sessions.Create(context.Background(), auth.Principal{
		UserID:    uuid.New(),
		DiscordID: "123",
		Roles:     []string{"user"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	rec := request(server.Handler(), http.MethodGet, "/api/v1/protected", &http.Cookie{Name: "acgw_session", Value: token})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequirePermissionAllowed(t *testing.T) {
	server, sessions, authorizer := testServer(t)
	authorizer.Grant("admin", "test.permission")
	server.Mux().Handle("GET /api/v1/protected", server.RequirePermission("test.permission", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})))

	token, err := sessions.Create(context.Background(), auth.Principal{
		UserID:    uuid.New(),
		DiscordID: "123",
		Roles:     []string{"admin"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	rec := request(server.Handler(), http.MethodGet, "/api/v1/protected", &http.Cookie{Name: "acgw_session", Value: token})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUnknownRouteReturnsJSONError(t *testing.T) {
	server, _, _ := testServer(t)
	rec := request(server.Handler(), http.MethodGet, "/api/v1/unknown", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Error.Code != "not_found" {
		t.Fatalf("code = %q", payload.Error.Code)
	}
	if payload.RequestID == "" {
		t.Fatal("request id missing")
	}
}
