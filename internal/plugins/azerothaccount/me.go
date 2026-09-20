package azerothaccount

import (
	"errors"
	"net/http"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
)

var (
	errAlreadyLinked = httpapi.NewAPIError(http.StatusConflict,
		"already_linked", "your Discord account is already linked to a game account")
	errLinksUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"links_unavailable", "account links are unavailable")
)

// SelfAccountRequest is the body of POST /api/v1/azeroth/me/account.
type SelfAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
} // @name AzerothSelfAccountRequest

// SelfAccountResponse is the link status of the authenticated user.
type SelfAccountResponse struct {
	Linked          bool   `json:"linked"`
	AccountUsername string `json:"account_username,omitempty"`
	AccountID       *int64 `json:"account_id,omitempty"`
} // @name AzerothSelfAccountResponse

// handleMyAccount handles GET /api/v1/azeroth/me/account.
//
//	@Summary		My linked account
//	@Description	Returns whether the authenticated user has linked a game account. Requires the azeroth.account.self permission.
//	@Tags			azeroth-account
//	@ID				azeroth.me.account.get
//	@Produce		json
//	@Success		200	{object}	SelfAccountResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/account [get]
func (p *Plugin) handleMyAccount(w http.ResponseWriter, r *http.Request) {
	if p.links == nil {
		httpapi.WriteError(w, r, errLinksUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	link, err := p.links.Get(r.Context(), principal.UserID)
	if errors.Is(err, domain.ErrLinkNotFound) {
		httpapi.WriteJSON(w, http.StatusOK, SelfAccountResponse{Linked: false})
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errLinksUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SelfAccountResponse{
		Linked:          true,
		AccountUsername: link.AccountUsername,
		AccountID:       link.AccountID,
	})
}

// handleCreateMyAccount handles POST /api/v1/azeroth/me/account.
//
//	@Summary		Create and link my account
//	@Description	Creates a game account with the provided credentials and links it to the authenticated user. One account per user. Requires the azeroth.account.self permission.
//	@Tags			azeroth-account
//	@ID				azeroth.me.account.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	SelfAccountRequest	true	"self-chosen credentials"
//	@Success		201	{object}	SelfAccountResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/account [post]
func (p *Plugin) handleCreateMyAccount(w http.ResponseWriter, r *http.Request) {
	if p.links == nil {
		httpapi.WriteError(w, r, errLinksUnavailable)
		return
	}
	var req SelfAccountRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())

	if _, err := p.links.Get(r.Context(), principal.UserID); err == nil {
		httpapi.WriteError(w, r, errAlreadyLinked)
		return
	} else if !errors.Is(err, domain.ErrLinkNotFound) {
		httpapi.WriteError(w, r, errLinksUnavailable)
		return
	}

	_, err := p.createAccount(r.Context(), CreateAccountRequest{
		Username: req.Username,
		Password: req.Password,
	})
	p.recordSelf(r, "azeroth.account.self.create", req.Username, err)
	if err != nil {
		writeSelfCommandError(w, r, err)
		return
	}

	var accountID *int64
	if p.accounts != nil {
		if account, lookupErr := p.accounts.FindAccountByUsername(r.Context(), req.Username); lookupErr == nil {
			id := account.ID
			accountID = &id
		}
	}
	link, err := p.links.Upsert(r.Context(), principal.UserID, req.Username, accountID)
	if err != nil {
		httpapi.WriteError(w, r, errLinksUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, SelfAccountResponse{
		Linked:          true,
		AccountUsername: link.AccountUsername,
		AccountID:       link.AccountID,
	})
}

func writeSelfCommandError(w http.ResponseWriter, r *http.Request, err error) {
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
}

func (p *Plugin) recordSelf(r *http.Request, action, targetID string, err error) {
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
		Permission:     string(PermissionAccountSelf),
		TargetType:     "account",
		TargetID:       targetID,
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}

func rateLimit(reg *plugins.Registry, next http.Handler) http.Handler {
	if reg.RateLimit == nil {
		return next
	}
	return reg.RateLimit(next)
}
