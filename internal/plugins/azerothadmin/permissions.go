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
	// permissionAuditRead is the gateway-generic permission that gates the audit
	// viewer. Its definition is owned by gateway-admin (ADR 0014); this plugin
	// only enforces it on its aggregate route until the route moves.
	permissionAuditRead permissions.Permission = "gw.audit.read"
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
	}
}
