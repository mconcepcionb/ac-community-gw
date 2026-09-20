package azerothaccount

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var (
	errLoginDBNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"login_db_not_configured", "AzerothCore login database is not configured")
	errLoginDBUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"login_db_unavailable", "AzerothCore login database is unavailable")
)

// Account is the JSON representation of an AzerothCore login account.
type Account struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	GMLevel   int     `json:"gm_level"`
	Expansion int     `json:"expansion"`
	Online    bool    `json:"online"`
	Banned    bool    `json:"banned"`
	BanReason string  `json:"ban_reason"`
	LastIP    string  `json:"last_ip"`
	LastLogin *string `json:"last_login"`
} // @name AzerothAccount

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

	items := make([]Account, 0, len(accounts))
	for _, account := range accounts {
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
		items = append(items, dto)
	}
	httpapi.WriteJSON(w, http.StatusOK, AccountsResponse{Accounts: items})
}
