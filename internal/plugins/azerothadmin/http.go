package azerothadmin

import (
	"errors"
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

// handleBanAccount handles POST /api/v1/azeroth/accounts/{username}/ban.
//
//	@Summary		Ban an account
//	@Description	Bans an AzerothCore account. Requires the azeroth.admin.accounts.ban permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.accounts.ban
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string				true	"account username"
//	@Param			request		body	BanAccountRequest	true	"ban details"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts/{username}/ban [post]
func (p *Plugin) handleBanAccount(w http.ResponseWriter, r *http.Request) {
	var req BanAccountRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Username = r.PathValue("username")
	result, err := p.banAccount(r.Context(), req)
	p.record(r, "azeroth.accounts.ban", PermissionAdminAccountsBan, "account", req.Username, err)
	writeCommandResult(w, r, result, err)
}

// handleUnbanAccount handles POST /api/v1/azeroth/accounts/{username}/unban.
//
//	@Summary		Unban an account
//	@Description	Lifts the ban on an AzerothCore account. Requires the azeroth.admin.accounts.ban permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.accounts.unban
//	@Produce		json
//	@Param			username	path	string	true	"account username"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts/{username}/unban [post]
func (p *Plugin) handleUnbanAccount(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	result, err := p.unbanAccount(r.Context(), UnbanAccountRequest{Username: username})
	p.record(r, "azeroth.accounts.unban", PermissionAdminAccountsBan, "account", username, err)
	writeCommandResult(w, r, result, err)
}

// handleSetGMLevel handles PUT /api/v1/azeroth/accounts/{username}/gmlevel.
//
//	@Summary		Set account GM level
//	@Description	Changes the GM level of an AzerothCore account. Requires the azeroth.admin.accounts.gmlevel permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.accounts.set_gmlevel
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string				true	"account username"
//	@Param			request		body	SetGMLevelRequest	true	"GM level"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts/{username}/gmlevel [put]
func (p *Plugin) handleSetGMLevel(w http.ResponseWriter, r *http.Request) {
	var req SetGMLevelRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Username = r.PathValue("username")
	result, err := p.setGMLevel(r.Context(), req)
	p.record(r, "azeroth.accounts.set-gmlevel", PermissionAdminAccountsGMLevel, "account", req.Username, err)
	writeCommandResult(w, r, result, err)
}

// writeCommandResult maps command errors to API errors and writes the result.
func writeCommandResult(w http.ResponseWriter, r *http.Request, result string, err error) {
	if err != nil {
		if errors.Is(err, ErrMissingUsername) ||
			errors.Is(err, azerothcore.ErrInvalidIdentifier) ||
			errors.Is(err, ErrInvalidDuration) ||
			errors.Is(err, ErrInvalidRealm) ||
			errors.Is(err, ErrInvalidGMLevel) {
			httpapi.WriteError(w, r, httpapi.ErrUnprocessable.WithDetails(err.Error()))
			return
		}
		httpapi.WriteError(w, r, httpapi.ErrBadGateway)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.CommandResult{Result: result})
}
