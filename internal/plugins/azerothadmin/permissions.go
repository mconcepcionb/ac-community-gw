package azerothadmin

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionAdminAccountsRead allows reading accounts as an administrator.
	PermissionAdminAccountsRead permissions.Permission = "azeroth.admin.accounts.read"
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
	// PermissionAdminUsersRead allows reading the community user 360 view.
	PermissionAdminUsersRead permissions.Permission = "azeroth.admin.users.read"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionAdminAccountsRead,
			Description: "Read AzerothCore accounts as an administrator",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminAccountsBan,
			Description: "Ban and unban AzerothCore accounts",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminAccountsGMLevel,
			Description: "Change AzerothCore account GM level",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminPlayersRead,
			Description: "List online players",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminPlayersKick,
			Description: "Kick online players",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminPlayersMute,
			Description: "Mute and unmute players",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminCharactersBan,
			Description: "Ban and unban characters",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminAnnounce,
			Description: "Broadcast announcements",
			Owner:       Name,
		},
		{
			Name:        PermissionAdminUsersRead,
			Description: "Read the community user 360 view",
			Owner:       Name,
		},
	}
}
