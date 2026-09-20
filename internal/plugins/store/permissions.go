package store

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionCatalogRead allows reading the product catalog.
	PermissionCatalogRead permissions.Permission = "gw.store.catalog.read"
	// PermissionWalletRead allows reading one's own wallet balance.
	PermissionWalletRead permissions.Permission = "gw.store.wallet.read"
	// PermissionOrdersRead allows reading one's own orders.
	PermissionOrdersRead permissions.Permission = "gw.store.orders.read"
	// PermissionPurchase allows buying store products.
	PermissionPurchase permissions.Permission = "gw.store.purchase"
	// PermissionAdminWallets allows granting points to any wallet.
	PermissionAdminWallets permissions.Permission = "gw.store.admin.wallets"
	// PermissionAdminProducts allows managing the product catalog.
	PermissionAdminProducts permissions.Permission = "gw.store.admin.products"
	// PermissionAdminOrdersRead allows reading every order.
	PermissionAdminOrdersRead permissions.Permission = "gw.store.admin.orders.read"
	// PermissionAdminOrdersResolve allows refunding or retrying stuck orders.
	PermissionAdminOrdersResolve permissions.Permission = "gw.store.admin.orders.resolve"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionCatalogRead,
			Description: "Read the store product catalog",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionWalletRead,
			Description: "Read the authenticated user's wallet balance",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionOrdersRead,
			Description: "Read the authenticated user's orders",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionPurchase,
			Description: "Buy store products",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionAdminWallets,
			Description: "Grant points to any wallet",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionAdminProducts,
			Description: "Manage the store product catalog",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionAdminOrdersRead,
			Description: "Read every store order",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionAdminOrdersResolve,
			Description: "Refund or retry stuck store orders",
			Owner:       Name,
			Namespace:   "gw",
		},
	}
}
