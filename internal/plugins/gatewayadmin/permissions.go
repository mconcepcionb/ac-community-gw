package gatewayadmin

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionAuditRead allows reading the audit log. It is gateway-generic
	// (ADR 0014) and owned by this plugin.
	PermissionAuditRead permissions.Permission = "gw.audit.read"
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
