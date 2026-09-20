// Package azerothitem exposes the AzerothCore item catalog (world database
// item_template) as a read-only API.
package azerothitem

import (
	"context"
	"errors"
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
)

// Name is the stable plugin name.
const Name = "azeroth-item"

// CatalogService is the cross-plugin item lookup capability.
const CatalogService = "azeroth.item.catalog"

// Config configures the azeroth-item plugin.
type Config struct {
	Items azerothdb.ItemReader
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	items azerothdb.ItemReader
}

// New creates the azeroth-item plugin.
func New(cfg Config) *Plugin {
	return &Plugin{items: cfg.Items}
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
	reg.Mux.Handle("GET /api/v1/azeroth/items",
		reg.RequirePermission(PermissionItemList, http.HandlerFunc(p.handleListItems)))
	reg.Mux.Handle("GET /api/v1/azeroth/items/{entry}",
		reg.RequirePermission(PermissionItemList, http.HandlerFunc(p.handleGetItem)))
	return services.Provide[azerothdb.Catalog](reg.Services, CatalogService, p)
}

// LookupItem implements the azerothdb.Catalog capability.
func (p *Plugin) LookupItem(ctx context.Context, entry int64) (azerothdb.Item, bool, error) {
	if p.items == nil {
		return azerothdb.Item{}, false, errItemDBNotConfigured
	}
	item, err := p.items.FindItem(ctx, entry)
	if errors.Is(err, azerothdb.ErrItemNotFound) {
		return azerothdb.Item{}, false, nil
	}
	if err != nil {
		return azerothdb.Item{}, false, err
	}
	return item, true, nil
}
