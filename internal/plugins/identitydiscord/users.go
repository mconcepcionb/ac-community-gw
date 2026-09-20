package identitydiscord

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
)

// User is the JSON representation of a provisioned community user.
type User struct {
	UserID      string `json:"user_id"`
	DiscordID   string `json:"discord_id"`
	Username    string `json:"username"`
	GlobalName  string `json:"global_name"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
} // @name User

// ListUsersResponse is the body returned by GET /api/v1/identity/users.
type ListUsersResponse struct {
	Users []User `json:"users"`
} // @name ListUsersResponse

// handleListUsers handles GET /api/v1/identity/users.
//
//	@Summary		List community users
//	@Description	Lists provisioned community users. Requires the identity.user.list permission.
//	@Tags			identity
//	@ID				identity.users.list
//	@Produce		json
//	@Param			filter	query	string	false	"filter by username or Discord name"
//	@Param			limit	query	int		false	"page size"	default(50)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	ListUsersResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		500	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/identity/users [get]
func (p *Plugin) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if p.repo == nil {
		httpapi.WriteError(w, r, errIdentityStorageUnavailable)
		return
	}
	query := r.URL.Query()
	limit := 50
	if raw := query.Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if raw := query.Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			offset = parsed
		}
	}

	users, err := p.repo.ListUsers(r.Context(), query.Get("filter"), limit, offset)
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrInternal)
		return
	}
	items := make([]User, 0, len(users))
	for _, user := range users {
		items = append(items, userDTO(user))
	}
	httpapi.WriteJSON(w, http.StatusOK, ListUsersResponse{Users: items})
}

func userDTO(user userdir.User) User {
	return User{
		UserID:      user.ID.String(),
		DiscordID:   user.DiscordID,
		Username:    user.Username,
		GlobalName:  user.GlobalName,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt.UTC().Format(time.RFC3339),
	}
}
