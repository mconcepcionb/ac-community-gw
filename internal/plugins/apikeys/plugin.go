// Package apikeys owns service credentials: scoped, revocable API keys.
package apikeys

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys/domain"
)

// Name is the stable plugin name.
const Name = "apikeys"

// Store is the persistence contract implemented by the repository.
type Store interface {
	Create(ctx context.Context, name string, permissions []string) (domain.APIKey, string, error)
	List(ctx context.Context) ([]domain.APIKey, error)
	Rotate(ctx context.Context, id uuid.UUID) (domain.APIKey, string, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

// Config configures the apikeys plugin.
type Config struct {
	Store Store
	Audit audit.Recorder
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	store    Store
	audit    audit.Recorder
	registry *plugins.Registry
}

// New creates the apikeys plugin.
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
	p.registry = reg
	reg.Mux.Handle("GET /api/v1/admin/permissions",
		reg.RequirePermission(PermissionManage, http.HandlerFunc(p.handlePermissions)))
	reg.Mux.Handle("GET /api/v1/admin/api-keys",
		reg.RequirePermission(PermissionManage, http.HandlerFunc(p.handleList)))
	reg.Mux.Handle("POST /api/v1/admin/api-keys",
		reg.RequirePermission(PermissionManage, http.HandlerFunc(p.handleCreate)))
	reg.Mux.Handle("POST /api/v1/admin/api-keys/{id}/rotate",
		reg.RequirePermission(PermissionManage, http.HandlerFunc(p.handleRotate)))
	reg.Mux.Handle("DELETE /api/v1/admin/api-keys/{id}",
		reg.RequirePermission(PermissionManage, http.HandlerFunc(p.handleRevoke)))
	return nil
}
