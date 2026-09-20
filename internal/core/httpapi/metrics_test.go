package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
	"github.com/mconcepcionb/ac-community-gw/internal/core/metrics"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/persistence"
)

func newMetricsServer(t *testing.T, token string) *Server {
	t.Helper()
	registry := metrics.NewRegistry()
	registry.Inc("identity_login_started_total")
	sessions := auth.NewManager(auth.ManagerOptions{Store: auth.NewMemoryStore(), TTL: time.Hour})
	return New(Dependencies{
		Config: &config.Config{
			HTTPAddr:        ":0",
			ShutdownTimeout: time.Second,
			Metrics:         config.Metrics{Enabled: true, Token: token},
		},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Sessions:    sessions,
		Permissions: permissions.NewRegistry(),
		Authorizer:  permissions.NewAuthorizer(),
		Audit:       audit.NopRecorder{},
		Readiness:   persistence.NewRegistry(),
		Metrics:     registry,
	})
}

func TestMetricsDisabledByDefault(t *testing.T) {
	server, _, _ := testServer(t)
	rec := request(server.Handler(), http.MethodGet, "/metrics", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestMetricsRequiresConfiguredToken(t *testing.T) {
	server := newMetricsServer(t, "secret")

	anonymous := request(server.Handler(), http.MethodGet, "/metrics", nil)
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d", anonymous.Code)
	}

	wrongReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	wrongReq.Header.Set("Authorization", "Bearer nope")
	wrong := httptest.NewRecorder()
	server.Handler().ServeHTTP(wrong, wrongReq)
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token status = %d", wrong.Code)
	}

	authorizedReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	authorizedReq.Header.Set("Authorization", "Bearer secret")
	authorized := httptest.NewRecorder()
	server.Handler().ServeHTTP(authorized, authorizedReq)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d", authorized.Code)
	}
	if !strings.Contains(authorized.Body.String(), "identity_login_started_total 1") {
		t.Fatalf("body = %q", authorized.Body.String())
	}
}
