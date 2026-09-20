package azerothaccount

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
)

var (
	errLoginDBNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"login_db_not_configured", "AzerothCore login database is not configured")
	errLoginDBUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"login_db_unavailable", "AzerothCore login database is unavailable")
	errAccountMissing = httpapi.NewAPIError(http.StatusNotFound,
		"account_not_found", "account not found")
)

// AccountOwner is the community user that owns a linked account.
type AccountOwner struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
	DiscordID   string `json:"discord_id,omitempty"`
	LinkedAt    string `json:"linked_at,omitempty"`
} // @name AzerothAccountOwner

// Account is the JSON representation of an AzerothCore login account.
type Account struct {
	ID        int64         `json:"id"`
	Username  string        `json:"username"`
	Email     string        `json:"email"`
	GMLevel   int           `json:"gm_level"`
	Expansion int           `json:"expansion"`
	Online    bool          `json:"online"`
	Banned    bool          `json:"banned"`
	BanReason string        `json:"ban_reason"`
	LastIP    string        `json:"last_ip"`
	LastLogin *string       `json:"last_login"`
	Claimed   bool          `json:"claimed"`
	ClaimedBy *AccountOwner `json:"claimed_by,omitempty"`
} // @name AzerothAccount

// AccountDetail is the body of GET /api/v1/azeroth/accounts/{username}.
type AccountDetail struct {
	Account Account     `json:"account"`
	Claim   *AdminClaim `json:"claim,omitempty"`
} // @name AzerothAccountDetail

// AccountsResponse is the body of GET /api/v1/azeroth/accounts.
type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
} // @name AzerothAccountsResponse

// handleListAccounts handles GET /api/v1/azeroth/accounts by reading the
// AzerothCore login database.
//
//	@Summary		List accounts
//	@Description	Lists AzerothCore login accounts. Requires the azeroth.account.list permission.
//	@Tags			azeroth-account
//	@ID				azeroth.accounts.list
//	@Produce		json
//	@Param			filter	query	string	false	"username filter"
//	@Param			limit	query	int		false	"page size"	default(50)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	AccountsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts [get]
func (p *Plugin) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	if p.accounts == nil {
		httpapi.WriteError(w, r, errLoginDBNotConfigured)
		return
	}

	query := azerothdb.Query{Filter: r.URL.Query().Get("filter")}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			query.Limit = parsed
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			query.Offset = parsed
		}
	}

	accounts, err := p.accounts.ListAccounts(r.Context(), query)
	if err != nil {
		httpapi.WriteError(w, r, errLoginDBUnavailable)
		return
	}

	owners := p.accountOwners(r.Context())
	items := make([]Account, 0, len(accounts))
	for _, account := range accounts {
		dto := accountDTO(account)
		if owner, ok := owners[strings.ToLower(account.Username)]; ok {
			dto.Claimed = true
			dto.ClaimedBy = &owner
		}
		items = append(items, dto)
	}
	httpapi.WriteJSON(w, http.StatusOK, AccountsResponse{Accounts: items})
}

// handleGetAccount handles GET /api/v1/azeroth/accounts/{username}.
//
//	@Summary		Account detail
//	@Description	Returns one AzerothCore account with its community owner and any pending claim. Requires the azeroth.account.list permission.
//	@Tags			azeroth-account
//	@ID				azeroth.accounts.get
//	@Produce		json
//	@Param			username	path	string	true	"account username"
//	@Success		200	{object}	AccountDetail
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/accounts/{username} [get]
func (p *Plugin) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	if p.accounts == nil {
		httpapi.WriteError(w, r, errLoginDBNotConfigured)
		return
	}
	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	account, err := p.accounts.FindAccountByUsername(r.Context(), username)
	if errors.Is(err, azerothdb.ErrAccountNotFound) {
		httpapi.WriteError(w, r, errAccountMissing)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errLoginDBUnavailable)
		return
	}

	dto := accountDTO(account)
	detail := AccountDetail{Account: dto}
	if owners := p.accountOwners(r.Context()); owners != nil {
		if owner, ok := owners[strings.ToLower(account.Username)]; ok {
			detail.Account.Claimed = true
			detail.Account.ClaimedBy = &owner
		}
	}
	if claim := p.claimForUsername(r.Context(), account.Username); claim != nil {
		detail.Claim = claim
	}
	httpapi.WriteJSON(w, http.StatusOK, detail)
}

func accountDTO(account azerothdb.Account) Account {
	dto := Account{
		ID:        account.ID,
		Username:  account.Username,
		Email:     account.Email,
		GMLevel:   account.GMLevel,
		Expansion: account.Expansion,
		Online:    account.Online,
		Banned:    account.Banned,
		BanReason: account.BanReason,
		LastIP:    account.LastIP,
	}
	if account.LastLogin != nil {
		lastLogin := account.LastLogin.UTC().Format(time.RFC3339)
		dto.LastLogin = &lastLogin
	}
	return dto
}

// accountOwners maps lowercased account usernames to their owning user.
func (p *Plugin) accountOwners(ctx context.Context) map[string]AccountOwner {
	owners := make(map[string]AccountOwner)
	if p.links == nil {
		return owners
	}
	links, err := p.links.List(ctx)
	if err != nil {
		return owners
	}
	for _, link := range links {
		owner := AccountOwner{
			UserID:   link.UserID.String(),
			LinkedAt: link.LinkedAt.UTC().Format(time.RFC3339),
		}
		if p.users != nil {
			if users, lookupErr := p.users.ListUsers(ctx, link.UserID.String(), 1, 0); lookupErr == nil {
				for _, user := range users {
					if user.ID == link.UserID {
						owner.DisplayName = displayName(user)
						owner.DiscordID = user.DiscordID
						break
					}
				}
			}
		}
		owners[strings.ToLower(link.AccountUsername)] = owner
	}
	return owners
}

// claimForUsername returns the pending claim for an account, if any.
func (p *Plugin) claimForUsername(ctx context.Context, username string) *AdminClaim {
	if p.claims == nil {
		return nil
	}
	claims, err := p.claims.ListClaims(ctx, 200, 0)
	if err != nil {
		return nil
	}
	for _, claim := range claims {
		if strings.EqualFold(claim.AccountUsername, username) {
			return &AdminClaim{
				UserID:          claim.UserID.String(),
				AccountUsername: claim.AccountUsername,
				ExpiresAt:       claim.ExpiresAt.UTC().Format(time.RFC3339),
				Attempts:        claim.Attempts,
			}
		}
	}
	return nil
}

func displayName(user userdir.User) string {
	if user.DisplayName != "" {
		return user.DisplayName
	}
	if user.GlobalName != "" {
		return user.GlobalName
	}
	return user.Username
}
