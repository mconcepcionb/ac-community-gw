package azerothaccount

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
)

var (
	errLinkStoreUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"link_store_unavailable", "account link store is unavailable")
	errMissingTarget = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"missing_user", "user_id or discord_id is required")
	errBadUserID = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_user_id", "user_id is not a valid UUID")
	errUserNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"user_not_found", "community user not found")
	errAmbiguousUser = httpapi.NewAPIError(http.StatusConflict,
		"ambiguous_user", "multiple users match the given name; use user_id or discord_id")
	errIdentityUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"identity_storage_unavailable", "community user directory is unavailable")
	errMissingAccount = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"missing_account_username", "account_username is required")
	errAccountNotFound = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"account_not_found", "AzerothCore account does not exist")
	errAccountAlreadyLinked = httpapi.NewAPIError(http.StatusConflict,
		"account_already_linked", "account is already linked to another user")
	errLinkNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"link_not_found", "account link not found")
)

// CreateLinkRequest is the body of POST /api/v1/azeroth/account-links.
type CreateLinkRequest struct {
	UserID          string `json:"user_id"`
	DiscordID       string `json:"discord_id"`
	DiscordUsername string `json:"discord_username"`
	AccountUsername string `json:"account_username"`
} // @name CreateLinkRequest

// Link is the JSON representation of a community user <-> account link.
type Link struct {
	UserID          string `json:"user_id"`
	AccountUsername string `json:"account_username"`
	AccountID       *int64 `json:"account_id"`
	LinkedAt        string `json:"linked_at"`
} // @name AzerothAccountLink

// LinksResponse is the body of GET /api/v1/azeroth/account-links.
type LinksResponse struct {
	Links []Link `json:"links"`
} // @name AzerothAccountLinksResponse

// handleCreateLink handles POST /api/v1/azeroth/account-links.
//
//	@Summary		Link a community user to an account
//	@Description	Creates or replaces the link between a community user and an AzerothCore account. Requires the azeroth.account.link permission.
//	@Tags			azeroth-account
//	@ID				azeroth.account_links.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateLinkRequest	true	"link to create"
//	@Success		201	{object}	Link
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/account-links [post]
func (p *Plugin) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	if p.links == nil {
		httpapi.WriteError(w, r, errLinkStoreUnavailable)
		return
	}
	var req CreateLinkRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	if strings.TrimSpace(req.AccountUsername) == "" {
		httpapi.WriteError(w, r, errMissingAccount)
		return
	}

	userID, err := p.resolveTargetUser(r.Context(), req)
	if err != nil {
		writeLinkError(w, r, err)
		return
	}

	accountID, err := p.verifyAccount(r.Context(), req.AccountUsername)
	if err != nil {
		writeLinkError(w, r, err)
		return
	}

	link, err := p.links.Upsert(r.Context(), userID, req.AccountUsername, accountID)
	if err != nil {
		writeLinkError(w, r, err)
		return
	}
	p.recordLink(r.Context(), "azeroth.account.link", userID, audit.ResultSuccess)
	httpapi.WriteJSON(w, http.StatusCreated, linkDTO(link))
}

// handleGetLink handles GET /api/v1/azeroth/account-links/{user_id}.
//
//	@Summary		Get an account link
//	@Description	Returns the link for a community user. Requires the azeroth.account.read permission.
//	@Tags			azeroth-account
//	@ID				azeroth.account_links.get
//	@Produce		json
//	@Param			user_id	path	string	true	"community user UUID"
//	@Success		200	{object}	Link
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/account-links/{user_id} [get]
func (p *Plugin) handleGetLink(w http.ResponseWriter, r *http.Request) {
	if p.links == nil {
		httpapi.WriteError(w, r, errLinkStoreUnavailable)
		return
	}
	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		httpapi.WriteError(w, r, errBadUserID)
		return
	}
	link, err := p.links.Get(r.Context(), userID)
	if err != nil {
		writeLinkError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, linkDTO(link))
}

// handleDeleteLink handles DELETE /api/v1/azeroth/account-links/{user_id}.
//
//	@Summary		Delete an account link
//	@Description	Removes the link for a community user. Requires the azeroth.account.link permission.
//	@Tags			azeroth-account
//	@ID				azeroth.account_links.delete
//	@Param			user_id	path	string	true	"community user UUID"
//	@Success		204	"Link deleted"
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/account-links/{user_id} [delete]
func (p *Plugin) handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	if p.links == nil {
		httpapi.WriteError(w, r, errLinkStoreUnavailable)
		return
	}
	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		httpapi.WriteError(w, r, errBadUserID)
		return
	}
	if err := p.links.Delete(r.Context(), userID); err != nil {
		httpapi.WriteError(w, r, httpapi.ErrInternal)
		return
	}
	p.recordLink(r.Context(), "azeroth.account.unlink", userID, audit.ResultSuccess)
	w.WriteHeader(http.StatusNoContent)
}

// handleListLinks handles GET /api/v1/azeroth/account-links.
//
//	@Summary		List account links
//	@Description	Lists every community user <-> account link. Requires the azeroth.account.read permission.
//	@Tags			azeroth-account
//	@ID				azeroth.account_links.list
//	@Produce		json
//	@Success		200	{object}	LinksResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/account-links [get]
func (p *Plugin) handleListLinks(w http.ResponseWriter, r *http.Request) {
	if p.links == nil {
		httpapi.WriteError(w, r, errLinkStoreUnavailable)
		return
	}
	links, err := p.links.List(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrInternal)
		return
	}
	items := make([]Link, 0, len(links))
	for _, link := range links {
		items = append(items, linkDTO(link))
	}
	httpapi.WriteJSON(w, http.StatusOK, LinksResponse{Links: items})
}

func (p *Plugin) resolveTargetUser(ctx context.Context, req CreateLinkRequest) (uuid.UUID, error) {
	if value := strings.TrimSpace(req.UserID); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return uuid.Nil, errBadUserID
		}
		return id, nil
	}
	if value := strings.TrimSpace(req.DiscordID); value != "" {
		if p.users == nil {
			return uuid.Nil, errIdentityUnavailable
		}
		id, found, err := p.users.ResolveUser(ctx, value)
		if err != nil {
			return uuid.Nil, errIdentityUnavailable
		}
		if !found {
			return uuid.Nil, errUserNotFound
		}
		return id, nil
	}
	if value := strings.TrimSpace(req.DiscordUsername); value != "" {
		if p.users == nil {
			return uuid.Nil, errIdentityUnavailable
		}
		users, err := p.users.ResolveUserByName(ctx, value)
		if err != nil {
			return uuid.Nil, errIdentityUnavailable
		}
		switch len(users) {
		case 0:
			return uuid.Nil, errUserNotFound
		case 1:
			return users[0].ID, nil
		default:
			return uuid.Nil, errAmbiguousUser.WithDetails(map[string]any{"candidates": usersJSON(users)})
		}
	}
	return uuid.Nil, errMissingTarget
}

func usersJSON(users []userdir.User) []map[string]any {
	items := make([]map[string]any, 0, len(users))
	for _, user := range users {
		items = append(items, map[string]any{
			"user_id":      user.ID.String(),
			"discord_id":   user.DiscordID,
			"username":     user.Username,
			"global_name":  user.GlobalName,
			"display_name": user.DisplayName,
		})
	}
	return items
}

// verifyAccount checks the account exists in the login DB and returns its id.
func (p *Plugin) verifyAccount(ctx context.Context, username string) (*int64, error) {
	if p.accounts == nil {
		return nil, errLoginDBNotConfigured
	}
	account, err := p.accounts.FindAccountByUsername(ctx, username)
	switch {
	case errors.Is(err, azerothdb.ErrAccountNotFound):
		return nil, errAccountNotFound
	case errors.Is(err, azerothdb.ErrUnavailable):
		return nil, errLoginDBUnavailable
	case err != nil:
		return nil, errLoginDBUnavailable
	}
	id := int64(account.ID)
	return &id, nil
}

func (p *Plugin) recordLink(ctx context.Context, action string, target uuid.UUID, result audit.Result) {
	principal, _ := auth.PrincipalFromContext(ctx)
	_ = p.audit.Record(ctx, audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         action,
		Permission:     string(PermissionAccountLink),
		TargetType:     "community_user",
		TargetID:       target.String(),
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(ctx),
	})
}

func writeLinkError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *httpapi.APIError
	switch {
	case errors.As(err, &apiErr):
		httpapi.WriteError(w, r, apiErr)
	case errors.Is(err, domain.ErrAccountAlreadyLinked):
		httpapi.WriteError(w, r, errAccountAlreadyLinked)
	case errors.Is(err, domain.ErrLinkNotFound):
		httpapi.WriteError(w, r, errLinkNotFound)
	default:
		httpapi.WriteError(w, r, httpapi.ErrInternal)
	}
}

func linkDTO(link domain.Link) Link {
	dto := Link{
		UserID:          link.UserID.String(),
		AccountUsername: link.AccountUsername,
		LinkedAt:        link.LinkedAt.UTC().Format(time.RFC3339),
	}
	if link.AccountID != nil {
		dto.AccountID = link.AccountID
	}
	return dto
}
