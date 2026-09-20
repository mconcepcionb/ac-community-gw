package azerothcharacter

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionCharacterList allows reading AzerothCore characters.
	PermissionCharacterList permissions.Permission = "azeroth.character.list"
	// PermissionMailSend allows sending in-game mail, items and money.
	PermissionMailSend permissions.Permission = "azeroth.mail.send"
	// PermissionCharacterSelf allows a user to read their own characters.
	PermissionCharacterSelf permissions.Permission = "azeroth.character.self"
	// PermissionMailSelf allows a user to mail their own characters.
	PermissionMailSelf permissions.Permission = "azeroth.mail.self"
	// PermissionLeaderboardRead allows reading character leaderboards.
	PermissionLeaderboardRead permissions.Permission = "azeroth.leaderboard.read"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionCharacterList,
			Description: "Read AzerothCore characters",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionMailSend,
			Description: "Send in-game mail, items and money",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionCharacterSelf,
			Description: "Read your own characters",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionMailSelf,
			Description: "Mail your own characters",
			Owner:       Name,
			Namespace:   "azeroth",
		},
		{
			Name:        PermissionLeaderboardRead,
			Description: "Read character leaderboards",
			Owner:       Name,
			Namespace:   "azeroth",
		},
	}
}
