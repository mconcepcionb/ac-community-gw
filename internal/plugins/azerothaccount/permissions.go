package azerothaccount

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionAccountRead allows reading linked account information.
	PermissionAccountRead permissions.Permission = "azeroth.account.read"
	// PermissionAccountManage allows creating and modifying accounts.
	PermissionAccountManage permissions.Permission = "azeroth.account.manage"
	// PermissionAccountList allows listing AzerothCore login accounts.
	PermissionAccountList permissions.Permission = "azeroth.account.list"
	// PermissionAccountLink allows creating and removing user/account links.
	PermissionAccountLink permissions.Permission = "azeroth.account.link"
	// PermissionAccountSelf allows a user to create and link their own account.
	PermissionAccountSelf permissions.Permission = "azeroth.account.self"
	// PermissionAdminClaimsRead allows staff to list pending account claims.
	PermissionAdminClaimsRead permissions.Permission = "azeroth.admin.claims.read"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionAccountRead,
			Description: "Read AzerothCore account information linked to a user",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAccountManage,
			Description: "Create and manage AzerothCore accounts",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAccountList,
			Description: "List AzerothCore login accounts",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAccountLink,
			Description: "Create and remove community user / AzerothCore account links",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAccountSelf,
			Description: "Create and link your own AzerothCore account",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminClaimsRead,
			Description: "List pending account claims",
			Owner:       Name,
			Namespace:   "azeroth",
		},
	}
}
