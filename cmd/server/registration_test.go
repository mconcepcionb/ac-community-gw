package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/events"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothadmin"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothcharacter"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothinfo"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothitem"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/gatewayadmin"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports"
	storeplugin "github.com/mconcepcionb/ac-community-gw/internal/plugins/store"
)

type registrationExecutor struct{}

func (registrationExecutor) Execute(context.Context, string) (string, error) { return "", nil }

func newRegistrationRegistry() *plugins.Registry {
	return &plugins.Registry{
		Mux:         http.NewServeMux(),
		Commands:    commands.NewRegistry(),
		Services:    services.NewRegistry(),
		Events:      events.NewBus(),
		Permissions: permissions.NewRegistry(),
		Audit:       audit.NopRecorder{},
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
		RateLimit: func(next http.Handler) http.Handler { return next },
	}
}

// newManager wires every compiled-in plugin in the same order as cmd/server.
func newManager() *plugins.Manager {
	manager := plugins.NewManager()
	manager.Add(identitydiscord.New(identitydiscord.Config{}))
	manager.Add(azerothaccount.New(azerothaccount.Config{}))
	manager.Add(azerothcharacter.New(azerothcharacter.Config{}))
	manager.Add(azerothadmin.New(registrationExecutor{}))
	manager.Add(azerothinfo.New(registrationExecutor{}))
	manager.Add(azerothitem.New(azerothitem.Config{}))
	manager.Add(storeplugin.New(storeplugin.Config{}))
	manager.Add(reports.New(reports.Config{}))
	manager.Add(adminnotes.New(adminnotes.Config{}))
	manager.Add(apikeys.New(apikeys.Config{}))
	manager.Add(gatewayadmin.New(gatewayadmin.Config{}))
	return manager
}

func TestAllPluginsRegister(t *testing.T) {
	registry := newRegistrationRegistry()
	manager := newManager()
	if err := manager.RegisterAll(context.Background(), registry); err != nil {
		t.Fatalf("RegisterAll = %v", err)
	}

	names := manager.Names()
	if len(names) != 11 {
		t.Fatalf("plugins = %v", names)
	}
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] {
			t.Fatalf("plugin name %q is empty or duplicated", name)
		}
		seen[name] = true
	}

	for _, def := range registry.Permissions.Definitions() {
		if def.Namespace == "" || !strings.HasPrefix(string(def.Name), def.Namespace+".") {
			t.Errorf("permission %q is not namespaced %q", def.Name, def.Namespace)
		}
	}
	if len(registry.Permissions.Definitions()) == 0 {
		t.Fatal("no permissions were registered")
	}
}

func TestAllPluginsRegisterRoutes(t *testing.T) {
	registry := newRegistrationRegistry()
	if err := newManager().RegisterAll(context.Background(), registry); err != nil {
		t.Fatalf("RegisterAll = %v", err)
	}

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/me"},
		{http.MethodGet, "/api/v1/admin/roles"},
		{http.MethodGet, "/api/v1/azeroth/accounts"},
		{http.MethodGet, "/api/v1/azeroth/characters"},
		{http.MethodGet, "/api/v1/azeroth/public/leaderboards/progression"},
		{http.MethodGet, "/api/v1/azeroth/online"},
		{http.MethodGet, "/api/v1/azeroth/info/status"},
		{http.MethodGet, "/api/v1/azeroth/items"},
		{http.MethodGet, "/api/v1/store/products"},
		{http.MethodGet, "/api/v1/reports/mine"},
		{http.MethodGet, "/api/v1/admin/annotations"},
		{http.MethodGet, "/api/v1/admin/api-keys"},
		{http.MethodGet, "/api/v1/admin/users/00000000-0000-0000-0000-000000000000"},
		{http.MethodGet, "/api/v1/admin/audit"},
	}
	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		if _, pattern := registry.Mux.Handler(req); pattern == "" {
			t.Errorf("route %s %s is not registered", route.method, route.path)
		}
	}
}

func TestRegisterAllRejectsDuplicatePluginName(t *testing.T) {
	registry := newRegistrationRegistry()
	manager := plugins.NewManager()
	manager.Add(identitydiscord.New(identitydiscord.Config{}))
	manager.Add(identitydiscord.New(identitydiscord.Config{}))

	err := manager.RegisterAll(context.Background(), registry)
	if !errors.Is(err, plugins.ErrDuplicateName) {
		t.Fatalf("RegisterAll = %v, want ErrDuplicateName", err)
	}
}
