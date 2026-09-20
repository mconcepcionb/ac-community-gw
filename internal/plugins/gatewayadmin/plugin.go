// Package gatewayadmin owns the gateway-generic console aggregates: the
// community user 360 view and the audit viewer. It is a gateway plugin (ADR
// 0014): it holds no game domain logic and reads game data only through the
// core service/capability registry.
package gatewayadmin

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/storeview"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
)

// Name is the stable plugin name.
const Name = "gateway-admin"

// Capabilities consumed from other plugins.
const (
	accountDirectoryService   = "azeroth.account.directory"
	identityUserAdminService  = "identity.user.admin"
	characterDirectoryService = "azeroth.character.directory"
	storeAccountService       = "azeroth.store.account"
)

// accountDirectory is the capability published by azeroth-account.
type accountDirectory interface {
	LinkedAccount(ctx context.Context, userID string) (username string, accountID *int64, err error)
}

// userAdmin is the staff-facing capability published by identity-discord.
type userAdmin interface {
	UserByID(ctx context.Context, userID uuid.UUID) (userdir.User, error)
	Roles(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// characterDirectory is the capability published by azeroth-character.
type characterDirectory interface {
	CharactersByUser(ctx context.Context, userID string) ([]azerothdb.Character, error)
}

// storeAccount is the capability published by azeroth-store.
type storeAccount interface {
	Wallet(ctx context.Context, userID uuid.UUID) (int64, error)
	Orders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]storeview.Order, error)
}

// Config configures the gateway admin plugin.
type Config struct {
	Audit       audit.Recorder
	AuditReader audit.Reader
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	audit       audit.Recorder
	auditReader audit.Reader
	// registry resolves cross-plugin capabilities on demand so the aggregates
	// do not depend on plugin registration order.
	registry *services.Registry
}

// New creates the gateway admin plugin.
func New(cfg Config) *Plugin {
	recorder := cfg.Audit
	if recorder == nil {
		recorder = audit.NopRecorder{}
	}
	return &Plugin{audit: recorder, auditReader: cfg.AuditReader}
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
	p.registry = reg.Services

	reg.Mux.Handle("GET /api/v1/admin/users/{id}",
		reg.RequirePermission(permissionUserRead, http.HandlerFunc(p.handleUser360)))
	return nil
}
