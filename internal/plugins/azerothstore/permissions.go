package azerothstore

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionCatalogRead allows reading the product catalog.
	PermissionCatalogRead permissions.Permission = "store.catalog.read"
	// PermissionWalletRead allows reading one's own wallet balance.
	PermissionWalletRead permissions.Permission = "store.wallet.read"
	// PermissionOrdersRead allows reading one's own orders.
	PermissionOrdersRead permissions.Permission = "store.orders.read"
	// PermissionPurchase allows buying store products.
	PermissionPurchase permissions.Permission = "store.purchase"
	// PermissionAdminWallets allows granting points to any wallet.
	PermissionAdminWallets permissions.Permission = "store.admin.wallets"
	// PermissionAdminProducts allows managing the product catalog.
	PermissionAdminProducts permissions.Permission = "store.admin.products"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionCatalogRead,
			Description: "Read the store product catalog",
			Owner:       Name,
		},
		{
			Name:        PermissionWalletRead,
			Description: "Read the authenticated user's wallet balance",
			Owner:       Name,
		},
		{
			Name:        PermissionOrdersRead,
			Description: "Read the authenticated user's orders",
			Owner:       Name,
		},
		{
			Name:        PermissionPurchase,
			Description: "Buy store products",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminWallets,
			Description: "Grant points to any wallet",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminProducts,
			Description: "Manage the store product catalog",
			Owner:       Name,
		},
	}
}
