package identitydiscord

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionSelfRead allows reading the authenticated user's profile.
	PermissionSelfRead permissions.Permission = "identity.self.read"
	// PermissionSessionRevoke allows revoking user sessions.
	PermissionSessionRevoke permissions.Permission = "identity.session.revoke"
	// PermissionUserList allows searching community users.
	PermissionUserList permissions.Permission = "identity.user.list"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionSelfRead,
			Description: "Read the authenticated user's own profile",
			Owner:       Name,
		},
		{
			Name:        PermissionSessionRevoke,
			Description: "Revoke user sessions",
			Owner:       Name,
		},
		{
			Name:        PermissionUserList,
			Description: "Search community users",
			Owner:       Name,
		},
	}
}
