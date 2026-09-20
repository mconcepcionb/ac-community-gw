package azerothcharacter

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionCharacterList allows reading AzerothCore characters.
	PermissionCharacterList permissions.Permission = "azeroth.character.list"
	// PermissionMailSend allows sending in-game mail, items and money.
	PermissionMailSend permissions.Permission = "azeroth.mail.send"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionCharacterList,
			Description: "Read AzerothCore characters",
			Owner:       Name,
		},
		{
			Name:        PermissionMailSend,
			Description: "Send in-game mail, items and money",
			Owner:       Name,
		},
	}
}
