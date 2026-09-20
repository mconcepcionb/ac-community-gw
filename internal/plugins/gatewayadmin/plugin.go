// Package gatewayadmin owns the gateway-generic console aggregates: the
// community user 360 view and the audit viewer. It is a gateway plugin (ADR
// 0014): it holds no game domain logic and reads game data only through the
// core service/capability registry.
package gatewayadmin

import (
	"context"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
)

// Name is the stable plugin name.
const Name = "gateway-admin"

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

// Register implements plugins.Plugin. Routes arrive in later tickets.
func (p *Plugin) Register(_ context.Context, reg *plugins.Registry) error {
	for _, def := range permissionDefs() {
		if err := reg.Permissions.Register(def); err != nil {
			return err
		}
	}
	p.registry = reg.Services
	return nil
}
