package gatewayadmin

import (
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var errCatalogUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
	"permission_catalog_unavailable", "the permission catalog is unavailable")

// Permission describes a registered permission for the scope picker.
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
} // @name ApiPermission

// PermissionsResponse is the body of GET /api/v1/admin/permissions.
type PermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
} // @name ApiPermissionsResponse

// handlePermissions handles GET /api/v1/admin/permissions.
//
//	@Summary		List registered permissions
//	@Description	Returns every permission registered by the plugins, for the API key scope picker. Requires the gw.apikeys.manage permission.
//	@Tags			gateway-admin
//	@ID				gateway.admin.permissions.list
//	@Produce		json
//	@Success		200	{object}	PermissionsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/permissions [get]
func (p *Plugin) handlePermissions(w http.ResponseWriter, r *http.Request) {
	if p.permissionRegistry == nil {
		httpapi.WriteError(w, r, errCatalogUnavailable)
		return
	}
	defs := p.permissionRegistry.Definitions()
	items := make([]Permission, 0, len(defs))
	for _, def := range defs {
		items = append(items, Permission{
			Name:        string(def.Name),
			Description: def.Description,
			Owner:       def.Owner,
		})
	}
	httpapi.WriteJSON(w, http.StatusOK, PermissionsResponse{Permissions: items})
}
