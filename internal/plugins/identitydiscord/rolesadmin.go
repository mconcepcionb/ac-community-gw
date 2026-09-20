package identitydiscord

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
)

var (
	errRolesUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"roles_unavailable", "roles are unavailable")
	errUnknownPermission = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"unknown_permission", "one or more permissions are not registered")
)

// RoleGrant is one role -> permission grant.
type RoleGrant struct {
	Role       string `json:"role"`
	Permission string `json:"permission"`
} // @name AdminRoleGrant

// PermissionDefinition is one registered permission.
type PermissionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner,omitempty"`
} // @name AdminPermissionDefinition

// DiscordRoleMapping maps a Discord role id to an internal role.
type DiscordRoleMapping struct {
	DiscordRoleID string `json:"discord_role_id"`
	Role          string `json:"role"`
} // @name AdminDiscordRoleMapping

// RolesResponse is the body of GET /api/v1/admin/roles.
type RolesResponse struct {
	Roles       []string               `json:"roles"`
	Permissions []PermissionDefinition `json:"permissions"`
	Grants      []RoleGrant            `json:"grants"`
	Mappings    []DiscordRoleMapping   `json:"mappings"`
} // @name AdminRolesResponse

// ReplaceRolePermissionsRequest is the body of the batch permission endpoint.
type ReplaceRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
} // @name AdminReplaceRolePermissionsRequest

// MappingRequest is the body of the Discord mapping endpoint.
type MappingRequest struct {
	Role string `json:"role"`
} // @name AdminMappingRequest

// handleListRoles handles GET /api/v1/admin/roles.
//
//	@Summary		List roles, permissions and mappings
//	@Description	Returns every internal role, the registered permission catalog, every role -> permission grant and every Discord role mapping. Requires the gw.identity.roles.manage permission.
//	@Tags			identity
//	@ID				identity.admin.roles.list
//	@Produce		json
//	@Success		200	{object}	RolesResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/roles [get]
func (p *Plugin) handleListRoles(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	roles, err := p.repo.ListRoles(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	grants, err := p.repo.ListRolePermissions(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	mappings, err := p.repo.ListDiscordRoleMappings(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	response := RolesResponse{
		Roles:       roles,
		Permissions: p.permissionCatalog(),
		Grants:      make([]RoleGrant, 0, len(grants)),
		Mappings:    make([]DiscordRoleMapping, 0, len(mappings)),
	}
	for _, grant := range grants {
		response.Grants = append(response.Grants, RoleGrant{Role: grant.Role, Permission: grant.Permission})
	}
	for _, mapping := range mappings {
		response.Mappings = append(response.Mappings, DiscordRoleMapping{
			DiscordRoleID: mapping.DiscordRoleID,
			Role:          mapping.Role,
		})
	}
	httpapi.WriteJSON(w, http.StatusOK, response)
}

// handleReplaceRolePermissions handles PUT /api/v1/admin/roles/{role}/permissions.
//
//	@Summary		Replace a role's permissions
//	@Description	Replaces the full set of permission grants of a role (creating the role if needed) in one transaction. Requires the gw.identity.roles.manage permission.
//	@Tags			identity
//	@ID				identity.admin.roles.replace_permissions
//	@Accept			json
//	@Param			role	path	string							true	"role"
//	@Param			request	body	ReplaceRolePermissionsRequest	true	"permissions"
//	@Success		204	"Replaced"
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/roles/{role}/permissions [put]
func (p *Plugin) handleReplaceRolePermissions(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	role := strings.TrimSpace(r.PathValue("role"))
	if role == "" {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	var req ReplaceRolePermissionsRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	normalized := normalizePermissions(req.Permissions)
	if p.permissions != nil {
		for _, permission := range normalized {
			if !p.permissions.Has(permissions.Permission(permission)) {
				httpapi.WriteError(w, r, errUnknownPermission)
				return
			}
		}
	}
	err := p.repo.ReplaceRolePermissions(r.Context(), role, normalized)
	p.recordRoles(r, "identity.roles.replace", role, err)
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUpsertDiscordMapping handles PUT /api/v1/admin/discord-role-mappings/{discord_role_id}.
//
//	@Summary		Map a Discord role to an internal role
//	@Description	Creates or updates a Discord role mapping. Requires the gw.identity.roles.manage permission.
//	@Tags			identity
//	@ID				identity.admin.discord_mappings.upsert
//	@Accept			json
//	@Param			discord_role_id	path	string			true	"Discord role id"
//	@Param			request			body	MappingRequest	true	"internal role"
//	@Success		204	"Mapped"
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/discord-role-mappings/{discord_role_id} [put]
func (p *Plugin) handleUpsertDiscordMapping(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	discordRoleID := strings.TrimSpace(r.PathValue("discord_role_id"))
	var req MappingRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	role := strings.TrimSpace(req.Role)
	if discordRoleID == "" || role == "" {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	if err := p.repo.UpsertRole(r.Context(), role); err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	err := p.repo.UpsertDiscordRoleMapping(r.Context(), discordRoleID, role)
	p.recordRoles(r, "identity.discord_mappings.upsert", discordRoleID, err)
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteDiscordMapping handles DELETE /api/v1/admin/discord-role-mappings/{discord_role_id}.
//
//	@Summary		Delete a Discord role mapping
//	@Description	Removes a Discord role mapping. Requires the gw.identity.roles.manage permission.
//	@Tags			identity
//	@ID				identity.admin.discord_mappings.delete
//	@Param			discord_role_id	path	string	true	"Discord role id"
//	@Success		204	"Deleted"
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/discord-role-mappings/{discord_role_id} [delete]
func (p *Plugin) handleDeleteDiscordMapping(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	discordRoleID := strings.TrimSpace(r.PathValue("discord_role_id"))
	if discordRoleID == "" {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	err := p.repo.DeleteDiscordRoleMapping(r.Context(), discordRoleID)
	p.recordRoles(r, "identity.discord_mappings.delete", discordRoleID, err)
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// permissionCatalog returns the registered permissions sorted by name.
func (p *Plugin) permissionCatalog() []PermissionDefinition {
	if p.permissions == nil {
		return []PermissionDefinition{}
	}
	defs := p.permissions.Definitions()
	catalog := make([]PermissionDefinition, 0, len(defs))
	for _, def := range defs {
		catalog = append(catalog, PermissionDefinition{
			Name:        string(def.Name),
			Description: def.Description,
			Owner:       def.Owner,
		})
	}
	return catalog
}

// normalizePermissions trims, drops empties and deduplicates permission names.
func normalizePermissions(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func (p *Plugin) recordRoles(r *http.Request, action, target string, err error) {
	principal, _ := auth.PrincipalFromContext(r.Context())
	result := audit.ResultSuccess
	if err != nil {
		result = audit.ResultFailure
	}
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         action,
		Permission:     string(PermissionRolesManage),
		TargetType:     "role",
		TargetID:       target,
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}
