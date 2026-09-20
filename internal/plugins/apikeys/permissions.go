package apikeys

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionManage allows managing API keys.
	PermissionManage permissions.Permission = "apikeys.manage"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionManage,
			Description: "Create, rotate and revoke API keys",
			Owner:       Name,
		},
	}
}
