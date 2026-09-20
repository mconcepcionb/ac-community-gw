package identitydiscord

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionUserRead allows reading and searching community users.
	PermissionUserRead permissions.Permission = "gw.identity.user.read"
	// PermissionRolesManage allows managing roles, permission grants and Discord
	// role mappings.
	PermissionRolesManage permissions.Permission = "gw.identity.roles.manage"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionUserRead,
			Description: "Read and search community users",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionRolesManage,
			Description: "Manage roles, permission grants and Discord role mappings",
			Owner:       Name,
			Namespace:   "gw",
		},
	}
}
