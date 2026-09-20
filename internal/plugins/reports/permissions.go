package reports

import "github.com/mconcepcionb/ac-community-gw/internal/core/permissions"

const (
	// PermissionCreate allows submitting a report and reading your own reports.
	PermissionCreate permissions.Permission = "gw.report.create"
	// PermissionRead allows reading and closing every report.
	PermissionRead permissions.Permission = "gw.report.read"
)

func permissionDefs() []permissions.Definition {
	return []permissions.Definition{
		{
			Name:        PermissionCreate,
			Description: "Submit a player report and read your own reports",
			Owner:       Name,
			Namespace:   "gw",
		},
		{
			Name:        PermissionRead,
			Description: "Read and close player reports",
			Owner:       Name,
			Namespace:   "gw",
		},
	}
}
