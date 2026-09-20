package adminnotes

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionRead allows reading staff annotations.
	PermissionRead permissions.Permission = "gw.notes.read"
	// PermissionWrite allows creating annotations and editing your own.
	PermissionWrite permissions.Permission = "gw.notes.write"
	// PermissionManage allows editing and deleting any annotation.
	PermissionManage permissions.Permission = "gw.notes.manage"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionRead,
			Description: "Read staff annotations",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionWrite,
			Description: "Create staff annotations and edit your own",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionManage,
			Description: "Edit and delete any staff annotation",
			Owner:       Name,
			Namespace:   "gw",
		},
	}
}
