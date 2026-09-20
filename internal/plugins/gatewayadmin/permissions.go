package gatewayadmin

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionAuditRead allows reading the audit log. It is gateway-generic
	// (ADR 0014) and owned by this plugin.
	PermissionAuditRead permissions.Permission = "gw.audit.read"
	// permissionUserRead is the gateway-generic permission that gates the
	// community user 360 view. It is owned by identity-discord; this plugin only
	// enforces it on its aggregate route.
	permissionUserRead permissions.Permission = "gw.identity.user.read"
	// permissionAPIKeysManage is the gateway-generic permission that gates the
	// permission catalog. It is owned by apikeys; this plugin only enforces it.
	permissionAPIKeysManage permissions.Permission = "gw.apikeys.manage"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionAuditRead,
			Description: "Read the audit log",
			Owner:       Name,
			Namespace:   "gw",
		},
	}
}
