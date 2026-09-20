// Package azerothstore owns the community store domain: products, wallets and
// orders. Purchases are paid with points held by the gateway and delivered
// in-game through the azeroth-character delivery capability.
package azerothstore

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore/domain"
)

// Name is the stable plugin name.
const Name = "azeroth-store"

const (
	accountDirectoryService = "azeroth.account.directory"
	deliveryService         = "azeroth.character.delivery"
	identityUserDirectory   = "identity.user.directory"
	itemCatalogService      = "azeroth.item.catalog"
)

// accountDirectory is the cross-plugin capability published by azeroth-account.
type accountDirectory interface {
	LinkedAccount(ctx context.Context, userID string) (username string, accountID *int64, err error)
}

// Store is the persistence contract implemented by the repository.
type Store interface {
	Products(ctx context.Context) ([]domain.Product, error)
	ProductBySKU(ctx context.Context, sku string) (domain.Product, error)
	CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error)
	UpdateProduct(ctx context.Context, product domain.Product) (domain.Product, error)
	SetProductActive(ctx context.Context, sku string, active bool) (domain.Product, error)
	Wallet(ctx context.Context, userID uuid.UUID) (int64, error)
	Grant(ctx context.Context, userID uuid.UUID, points int64, reason string, actorID uuid.UUID) (int64, error)
	CreateOrder(ctx context.Context, userID uuid.UUID, product domain.Product, character string, accountID int64) (domain.Order, error)
	CompleteOrder(ctx context.Context, orderID uuid.UUID, output string) error
	FailOrder(ctx context.Context, orderID uuid.UUID, output string) (domain.Order, error)
	SetOrderOutput(ctx context.Context, orderID uuid.UUID, output string) error
	ReconcilePendingOrders(ctx context.Context, limit int) (int, error)
	Orders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, error)
	AdminOrders(ctx context.Context, status string, limit, offset int) ([]domain.Order, error)
}

// Config configures the azeroth-store plugin.
type Config struct {
	Store Store
	Audit audit.Recorder
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	store    Store
	audit    audit.Recorder
	accounts accountDirectory
	delivery delivery.Service
	users    userdir.Directory
	catalog  azerothdb.Catalog
}

// New creates the azeroth-store plugin.
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

	accounts, err := services.Consume[accountDirectory](reg.Services, accountDirectoryService)
	if err != nil {
		return fmt.Errorf("azeroth-store: account directory unavailable: %w", err)
	}
	deliveryService, err := services.Consume[delivery.Service](reg.Services, deliveryService)
	if err != nil {
		return fmt.Errorf("azeroth-store: delivery service unavailable: %w", err)
	}
	users, err := services.Consume[userdir.Directory](reg.Services, identityUserDirectory)
	if err != nil {
		return fmt.Errorf("azeroth-store: user directory unavailable: %w", err)
	}
	catalog, err := services.Consume[azerothdb.Catalog](reg.Services, itemCatalogService)
	if err != nil {
		return fmt.Errorf("azeroth-store: item catalog unavailable: %w", err)
	}
	p.accounts = accounts
	p.delivery = deliveryService
	p.users = users
	p.catalog = catalog

	reg.Mux.Handle("GET /api/v1/store/products",
		reg.RequirePermission(PermissionCatalogRead, http.HandlerFunc(p.handleProducts)))
	reg.Mux.Handle("GET /api/v1/store/products/{sku}",
		reg.RequirePermission(PermissionCatalogRead, http.HandlerFunc(p.handleProduct)))
	reg.Mux.Handle("POST /api/v1/store/products",
		reg.RequirePermission(PermissionAdminProducts, http.HandlerFunc(p.handleCreateProduct)))
	reg.Mux.Handle("PUT /api/v1/store/products/{sku}",
		reg.RequirePermission(PermissionAdminProducts, http.HandlerFunc(p.handleUpdateProduct)))
	reg.Mux.Handle("DELETE /api/v1/store/products/{sku}",
		reg.RequirePermission(PermissionAdminProducts, http.HandlerFunc(p.handleDeleteProduct)))
	reg.Mux.Handle("GET /api/v1/store/wallet",
		reg.RequirePermission(PermissionWalletRead, http.HandlerFunc(p.handleWallet)))
	reg.Mux.Handle("GET /api/v1/store/orders",
		reg.RequirePermission(PermissionOrdersRead, http.HandlerFunc(p.handleOrders)))
	reg.Mux.Handle("POST /api/v1/store/orders",
		reg.RequirePermission(PermissionPurchase, http.HandlerFunc(p.handlePurchase)))
	reg.Mux.Handle("POST /api/v1/store/wallets/grant",
		reg.RequirePermission(PermissionAdminWallets, http.HandlerFunc(p.handleGrant)))
	reg.Mux.Handle("GET /api/v1/admin/store/orders",
		reg.RequirePermission(PermissionAdminOrdersRead, http.HandlerFunc(p.handleAdminOrders)))
	return services.Provide[AccountView](reg.Services, AccountService, p)
}
