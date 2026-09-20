package azerothaccount

import (
	"errors"
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

// handleCreateAccount handles POST /api/v1/azeroth/accounts.
//
//	@Summary		Create an AzerothCore account
//	@Description	Creates a game account. Requires the azeroth.account.manage permission.
//	@Tags			azeroth-account
//	@ID				azeroth.accounts.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateAccountRequest	true	"account to create"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts [post]
func (p *Plugin) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	result, err := p.createAccount(r.Context(), req)
	p.record(r, "azeroth.account.create", req.Username, err)
	writeCommandResult(w, r, result, err)
}

// handleChangePassword handles POST /api/v1/azeroth/accounts/{username}/password.
//
//	@Summary		Change an account password
//	@Description	Sets a new password for a game account. Requires the azeroth.account.manage permission.
//	@Tags			azeroth-account
//	@ID				azeroth.accounts.set_password
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string					true	"account username"
//	@Param			request		body	ChangePasswordRequest	true	"new password"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts/{username}/password [post]
func (p *Plugin) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Username = r.PathValue("username")
	result, err := p.changePassword(r.Context(), req)
	p.record(r, "azeroth.account.change-password", req.Username, err)
	writeCommandResult(w, r, result, err)
}

// handleSetEmail handles PUT /api/v1/azeroth/accounts/{username}/email.
//
//	@Summary		Set an account email
//	@Description	Sets the email of a game account. Requires the azeroth.account.manage permission.
//	@Tags			azeroth-account
//	@ID				azeroth.accounts.set_email
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string			true	"account username"
//	@Param			request		body	SetEmailRequest	true	"new email"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts/{username}/email [put]
func (p *Plugin) handleSetEmail(w http.ResponseWriter, r *http.Request) {
	var req SetEmailRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Username = r.PathValue("username")
	result, err := p.setEmail(r.Context(), req)
	p.record(r, "azeroth.account.set-email", req.Username, err)
	writeCommandResult(w, r, result, err)
}

// writeCommandResult maps command errors to API errors and writes the result.
func writeCommandResult(w http.ResponseWriter, r *http.Request, result string, err error) {
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingUsername),
			errors.Is(err, ErrMissingPassword),
			errors.Is(err, azerothcore.ErrInvalidIdentifier),
			errors.Is(err, azerothcore.ErrInvalidPassword),
			errors.Is(err, azerothcore.ErrInvalidEmail):
			httpapi.WriteError(w, r, httpapi.ErrUnprocessable.WithDetails(err.Error()))
		default:
			httpapi.WriteError(w, r, httpapi.ErrBadGateway)
		}
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.CommandResult{Result: result})
}
