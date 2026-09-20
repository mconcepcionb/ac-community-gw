// Package adminnotes owns staff annotations attached to gateway entities.
package adminnotes

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/domain"
)

// Name is the stable plugin name.
const Name = "admin-notes"

// Store is the persistence contract implemented by the repository.
type Store interface {
	Create(ctx context.Context, targetType, targetID string, authorID uuid.UUID, body string) (domain.Annotation, error)
	ByTarget(ctx context.Context, targetType, targetID string, limit, offset int) ([]domain.Annotation, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Annotation, error)
	Update(ctx context.Context, id uuid.UUID, body string) (domain.Annotation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Config configures the admin notes plugin.
type Config struct {
	Store      Store
	Audit      audit.Recorder
	Authorizer *permissions.Authorizer
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	store      Store
	audit      audit.Recorder
	authorizer *permissions.Authorizer
}

// New creates the admin notes plugin.
func New(cfg Config) *Plugin {
	recorder := cfg.Audit
	if recorder == nil {
		recorder = audit.NopRecorder{}
	}
	return &Plugin{store: cfg.Store, audit: recorder, authorizer: cfg.Authorizer}
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
	reg.Mux.Handle("GET /api/v1/admin/annotations",
		reg.RequirePermission(PermissionRead, http.HandlerFunc(p.handleList)))
	reg.Mux.Handle("POST /api/v1/admin/annotations",
		reg.RequirePermission(PermissionWrite, http.HandlerFunc(p.handleCreate)))
	reg.Mux.Handle("PATCH /api/v1/admin/annotations/{id}",
		reg.RequirePermission(PermissionWrite, http.HandlerFunc(p.handleUpdate)))
	reg.Mux.Handle("DELETE /api/v1/admin/annotations/{id}",
		reg.RequirePermission(PermissionWrite, http.HandlerFunc(p.handleDelete)))
	return nil
}

// canManage reports whether the roles grant the moderation permission.
func (p *Plugin) canManage(roles []string) bool {
	if p.authorizer == nil {
		return false
	}
	permRoles := make([]permissions.Role, 0, len(roles))
	for _, role := range roles {
		permRoles = append(permRoles, permissions.Role(role))
	}
	return p.authorizer.Can(permRoles, PermissionManage)
}
