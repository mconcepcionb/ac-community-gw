package identitydiscord

import (
	"net/http"
	"strings"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var errRolesUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
	"roles_unavailable", "roles are unavailable")

// RoleGrant is one role -> permission grant.
type RoleGrant struct {
	Role       string `json:"role"`
	Permission string `json:"permission"`
} // @name AdminRoleGrant

// DiscordRoleMapping maps a Discord role id to an internal role.
type DiscordRoleMapping struct {
	DiscordRoleID string `json:"discord_role_id"`
	Role          string `json:"role"`
} // @name AdminDiscordRoleMapping

// RolesResponse is the body of GET /api/v1/admin/roles.
type RolesResponse struct {
	Grants   []RoleGrant          `json:"grants"`
	Mappings []DiscordRoleMapping `json:"mappings"`
} // @name AdminRolesResponse

// PermissionRequest is the body of the grant endpoint.
type PermissionRequest struct {
	Permission string `json:"permission"`
} // @name AdminPermissionRequest

// MappingRequest is the body of the Discord mapping endpoint.
type MappingRequest struct {
	Role string `json:"role"`
} // @name AdminMappingRequest

// handleListRoles handles GET /api/v1/admin/roles.
//
//	@Summary		List roles and mappings
//	@Description	Returns every role -> permission grant and Discord role mapping. Requires the identity.roles.manage permission.
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
		Grants:   make([]RoleGrant, 0, len(grants)),
		Mappings: make([]DiscordRoleMapping, 0, len(mappings)),
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

// handleGrantRolePermission handles POST /api/v1/admin/roles/{role}/permissions.
//
//	@Summary		Grant a permission to a role
//	@Description	Grants a permission to a role (creating the role if needed). Requires the identity.roles.manage permission.
//	@Tags			identity
//	@ID				identity.admin.roles.grant
//	@Accept			json
//	@Param			role	path	string				true	"role"
//	@Param			request	body	PermissionRequest	true	"permission"
//	@Success		204	"Granted"
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/roles/{role}/permissions [post]
func (p *Plugin) handleGrantRolePermission(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	role := strings.TrimSpace(r.PathValue("role"))
	var req PermissionRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	permission := strings.TrimSpace(req.Permission)
	if role == "" || permission == "" {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	if err := p.repo.UpsertRole(r.Context(), role); err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	err := p.repo.GrantRolePermission(r.Context(), role, permission)
	p.recordRoles(r, "identity.roles.grant", role+":"+permission, err)
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleRevokeRolePermission handles DELETE /api/v1/admin/roles/{role}/permissions/{permission}.
//
//	@Summary		Revoke a permission from a role
//	@Description	Revokes a permission from a role. Requires the identity.roles.manage permission.
//	@Tags			identity
//	@ID				identity.admin.roles.revoke
//	@Param			role		path	string	true	"role"
//	@Param			permission	path	string	true	"permission"
//	@Success		204	"Revoked"
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/roles/{role}/permissions/{permission} [delete]
func (p *Plugin) handleRevokeRolePermission(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	role := strings.TrimSpace(r.PathValue("role"))
	permission := strings.TrimSpace(r.PathValue("permission"))
	if role == "" || permission == "" {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	err := p.repo.RevokeRolePermission(r.Context(), role, permission)
	p.recordRoles(r, "identity.roles.revoke", role+":"+permission, err)
	if err != nil {
		httpapi.WriteError(w, r, errRolesUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUpsertDiscordMapping handles PUT /api/v1/admin/discord-role-mappings/{discord_role_id}.
//
//	@Summary		Map a Discord role to an internal role
//	@Description	Creates or updates a Discord role mapping. Requires the identity.roles.manage permission.
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
//	@Description	Removes a Discord role mapping. Requires the identity.roles.manage permission.
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
