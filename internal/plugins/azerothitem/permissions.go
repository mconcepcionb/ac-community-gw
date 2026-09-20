package azerothitem

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionItemList allows searching the item catalog.
	PermissionItemList permissions.Permission = "azeroth.item.list"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionItemList,
			Description: "Search the AzerothCore item catalog",
			Owner:       Name,
		},
	}
}
