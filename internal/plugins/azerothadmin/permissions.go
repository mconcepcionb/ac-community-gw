package azerothadmin

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionAdminAccountsBan allows banning and unbanning accounts.
	PermissionAdminAccountsBan permissions.Permission = "azeroth.admin.accounts.ban"
	// PermissionAdminAccountsGMLevel allows changing account GM level.
	PermissionAdminAccountsGMLevel permissions.Permission = "azeroth.admin.accounts.gmlevel"
	// PermissionAdminPlayersRead allows listing online players.
	PermissionAdminPlayersRead permissions.Permission = "azeroth.admin.players.read"
	// PermissionAdminPlayersKick allows kicking online players.
	PermissionAdminPlayersKick permissions.Permission = "azeroth.admin.players.kick"
	// PermissionAdminPlayersMute allows muting and unmuting players.
	PermissionAdminPlayersMute permissions.Permission = "azeroth.admin.players.mute"
	// PermissionAdminCharactersBan allows banning and unbanning characters.
	PermissionAdminCharactersBan permissions.Permission = "azeroth.admin.characters.ban"
	// PermissionAdminAnnounce allows broadcasting announcements.
	PermissionAdminAnnounce permissions.Permission = "azeroth.admin.announce"
	// PermissionAdminAuditRead allows reading the audit log. It is
	// gateway-generic (ADR 0014); the definition is registered here until the
	// generic admin routes move to a gateway plugin.
	PermissionAdminAuditRead permissions.Permission = "gw.audit.read"
	// permissionUserRead is the gateway-generic permission that gates the
	// community user 360 view. It is owned by identity-discord; this plugin only
	// enforces it on its aggregate route.
	permissionUserRead permissions.Permission = "gw.identity.user.read"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionAdminAccountsBan,
			Description: "Ban and unban AzerothCore accounts",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminAccountsGMLevel,
			Description: "Change AzerothCore account GM level",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminPlayersRead,
			Description: "List online players",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminPlayersKick,
			Description: "Kick online players",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminPlayersMute,
			Description: "Mute and unmute players",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminCharactersBan,
			Description: "Ban and unban characters",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminAnnounce,
			Description: "Broadcast announcements",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionAdminAuditRead,
			Description: "Read the audit log",
			Owner:       Name,
			Namespace:   "gw",
		},
	}
}
