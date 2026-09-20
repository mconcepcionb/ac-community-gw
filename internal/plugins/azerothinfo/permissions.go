package azerothinfo

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionInfoPublicRead allows reading public AzerothCore information.
	PermissionInfoPublicRead permissions.Permission = "azeroth.info.public.read"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionInfoPublicRead,
			Description: "Read public AzerothCore server information",
			Owner:       Name,
			Namespace:   "azeroth",
		},
	}
}
