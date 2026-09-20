package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
	"github.com/mconcepcionb/ac-community-gw/internal/core/metrics"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/persistence"
)

// Dependencies are the collaborators required to build the HTTP server.
type Dependencies struct {
	Config      *config.Config
	Logger      *slog.Logger
	Sessions    *auth.Manager
	Permissions *permissions.Registry
	Authorizer  *permissions.Authorizer
	Audit       audit.Recorder
	Readiness   *persistence.Registry
	Metrics     *metrics.Registry
}

// Server owns the HTTP mux and lifecycle.
type Server struct {
	cfg         *config.Config
	logger      *slog.Logger
	sessions    *auth.Manager
	permissions *permissions.Registry
	authorizer  *permissions.Authorizer
	audit       audit.Recorder
	readiness   *persistence.Registry
	metrics     *metrics.Registry
	mux         *http.ServeMux
}

// New creates the server and registers the operational endpoints.
func New(deps Dependencies) *Server {
	metricRegistry := deps.Metrics
	if metricRegistry == nil {
		metricRegistry = metrics.NewRegistry()
	}
	s := &Server{
		cfg:         deps.Config,
		logger:      deps.Logger,
		sessions:    deps.Sessions,
		permissions: deps.Permissions,
		authorizer:  deps.Authorizer,
		audit:       deps.Audit,
		readiness:   deps.Readiness,
		metrics:     metricRegistry,
		mux:         http.NewServeMux(),
	}
	if s.audit == nil {
		s.audit = audit.NopRecorder{}
	}
	if s.readiness == nil {
		s.readiness = persistence.NewRegistry()
	}
	s.routes()
	return s
}

// Mux exposes the mux so plugins can mount their domain routes.
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

// RequireAuth resolves the session and stores the principal in the context.
func (s *Server) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := s.sessions.TokenFromRequest(r)
		principal, err := s.sessions.Resolve(r.Context(), token)
		if err != nil {
			WriteError(w, r, ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
	})
}

// RequirePermission authenticates and then checks a permission.
func (s *Server) RequirePermission(permission permissions.Permission, next http.Handler) http.Handler {
	return s.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			WriteError(w, r, ErrUnauthorized)
			return
		}
		if !s.authorizer.Can(toRoles(principal.Roles), permission) {
			WriteError(w, r, ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

// Handler returns the fully wrapped HTTP handler. Access logging wraps the
// recovery middleware so panicking requests are still logged.
func (s *Server) Handler() http.Handler {
	return chain(s.mux,
		requestIDMiddleware,
		func(next http.Handler) http.Handler { return accessLogMiddleware(s.logger, next) },
		func(next http.Handler) http.Handler { return recoveryMiddleware(s.logger, next) },
	)
}

// Run serves HTTP until the context is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("http: listening", slog.String("addr", s.cfg.HTTPAddr))
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		s.logger.Info("http: shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	if s.cfg.Metrics.Enabled {
		s.mux.HandleFunc("GET /metrics", s.handleMetrics)
	}
	// The gateway serves the API and operational endpoints only. The SPA is a
	// separate artifact served by a reverse proxy in front of the gateway (see
	// docs/ADR/0012). Unknown paths keep the JSON error envelope so API clients
	// never receive HTML.
	s.mux.HandleFunc("/", handleNotFound)
}

// HealthResponse is the JSON body returned by GET /healthz.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
} // @name HealthResponse

// ReadyResponse is the JSON body returned by GET /readyz.
type ReadyResponse struct {
	Status string            `json:"status" example:"ready"`
	Checks map[string]string `json:"checks"`
} // @name ReadyResponse

// handleHealthz reports process liveness.
//
//	@Summary		Liveness probe
//	@Description	Reports that the process is alive. It does not check downstream dependencies.
//	@Tags			operations
//	@ID				operations.health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/healthz [get]
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// handleReadyz evaluates every registered readiness check.
//
//	@Summary		Readiness probe
//	@Description	Evaluates every registered readiness check and reports per-check status.
//	@Tags			operations
//	@ID				operations.ready
//	@Produce		json
//	@Success		200	{object}	ReadyResponse
//	@Failure		503	{object}	ReadyResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/readyz [get]
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	results := s.readiness.Evaluate(r.Context())
	checks := make(map[string]string, len(results))
	ready := true
	for _, result := range results {
		if result.Err != nil {
			ready = false
			checks[result.Name] = "unavailable"
			s.logger.Warn("readiness check failed",
				slog.String("check", result.Name),
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("error", result.Err.Error()),
			)
			continue
		}
		checks[result.Name] = "ok"
	}
	status := http.StatusOK
	state := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		state = "unavailable"
	}
	WriteJSON(w, status, ReadyResponse{Status: state, Checks: checks})
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, ErrNotFound)
}

func toRoles(roles []string) []permissions.Role {
	out := make([]permissions.Role, 0, len(roles))
	for _, role := range roles {
		out = append(out, permissions.Role(role))
	}
	return out
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
