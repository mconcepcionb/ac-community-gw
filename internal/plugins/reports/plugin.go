// Package reports owns player-submitted reports about other players.
package reports

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/domain"
)

// Name is the stable plugin name.
const Name = "reports"

// Store is the persistence contract implemented by the repository.
type Store interface {
	Create(ctx context.Context, reporterID uuid.UUID, target, category, message string) (domain.Report, error)
	ByReporter(ctx context.Context, reporterID uuid.UUID, limit, offset int) ([]domain.Report, error)
	List(ctx context.Context, status string, limit, offset int) ([]domain.Report, error)
	Close(ctx context.Context, id uuid.UUID) (domain.Report, error)
}

// Config configures the reports plugin.
type Config struct {
	Store Store
	Audit audit.Recorder
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	store Store
	audit audit.Recorder
}

// New creates the reports plugin.
func New(cfg Config) *Plugin {
	recorder := cfg.Audit
	if recorder == nil {
		recorder = audit.NopRecorder{}
	}
	return &Plugin{store: cfg.Store, audit: recorder}
}

// Name implements plugins.Plugin.
func (p *Plugin) Name() string { return Name }

// Register implements plugins.Plugin.
func (p *Plugin) Register(_ context.Context, reg *plugins.Registry) error {
	for _, def := range permissionDefs() {
		if err := reg.Permissions.Register(def); err != nil {
			return err
		}
	}
	reg.Mux.Handle("POST /api/v1/reports",
		rateLimit(reg, reg.RequirePermission(PermissionCreate, http.HandlerFunc(p.handleCreate))))
	reg.Mux.Handle("GET /api/v1/reports/mine",
		reg.RequirePermission(PermissionCreate, http.HandlerFunc(p.handleMine)))
	reg.Mux.Handle("GET /api/v1/admin/reports",
		reg.RequirePermission(PermissionRead, http.HandlerFunc(p.handleList)))
	reg.Mux.Handle("POST /api/v1/admin/reports/{id}/close",
		reg.RequirePermission(PermissionRead, http.HandlerFunc(p.handleClose)))
	return nil
}

func rateLimit(reg *plugins.Registry, next http.Handler) http.Handler {
	if reg.RateLimit == nil {
		return next
	}
	return reg.RateLimit(next)
}
