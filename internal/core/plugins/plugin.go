// Package plugins defines the contract between the core and the compiled-in
// plugins, plus the registry of core facilities available during registration.
package plugins

import (
	"context"
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/events"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
)

// Plugin is a compiled-in module owning a slice of domain behaviour.
type Plugin interface {
	// Name returns the stable plugin name (for example "azeroth-account").
	Name() string
	// Register wires routes, permissions, commands, services and event
	// handlers into the core registries.
	Register(ctx context.Context, reg *Registry) error
}

// Registry aggregates the core facilities plugins may use. It is passed to
// every plugin during registration and is the only coupling point between a
// plugin and the rest of the system.
type Registry struct {
	Mux         *http.ServeMux
	Commands    *commands.Registry
	Services    *services.Registry
	Events      *events.Bus
	Permissions *permissions.Registry
	Audit       audit.Recorder

	// RequireAuth and RequirePermission are HTTP middleware provided by the
	// core. Plugins use them when mounting protected routes.
	RequireAuth       func(http.Handler) http.Handler
	RequirePermission func(permissions.Permission, http.Handler) http.Handler
	// RateLimit is HTTP middleware that rejects abusive clients. Plugins use
	// it to protect sensitive endpoints such as login.
	RateLimit func(http.Handler) http.Handler
}
