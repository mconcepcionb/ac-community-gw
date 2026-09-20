package azerothadmin

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
)

var (
	errUser360InvalidID = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_user_id", "id is not a valid UUID")
	errUser360NotFound = httpapi.NewAPIError(http.StatusNotFound,
		"user_not_found", "community user not found")
	errUser360Unavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"user_directory_unavailable", "community user directory is unavailable")
)

// AdminUserCharacter is one character in the 360 view.
type AdminUserCharacter struct {
	GUID   int64  `json:"guid"`
	Name   string `json:"name"`
	Level  int    `json:"level"`
	Race   int    `json:"race"`
	Class  int    `json:"class"`
	Guild  string `json:"guild"`
	Online bool   `json:"online"`
	Money  int64  `json:"money"`
} // @name AdminUserCharacter

// AdminUserOrder is one order in the 360 view.
type AdminUserOrder struct {
	OrderID   string `json:"order_id"`
	SKU       string `json:"sku"`
	Points    int64  `json:"points"`
	Character string `json:"character"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
} // @name AdminUserOrder

// AdminUser is the aggregate returned by the community user 360 view.
type AdminUser struct {
	UserID          string               `json:"user_id"`
	DiscordID       string               `json:"discord_id"`
	Username        string               `json:"username"`
	GlobalName      string               `json:"global_name"`
	DisplayName     string               `json:"display_name"`
	Avatar          string               `json:"avatar"`
	CreatedAt       string               `json:"created_at"`
	Roles           []string             `json:"roles"`
	AccountUsername *string              `json:"account_username"`
	AccountID       *int64               `json:"account_id"`
	Characters      []AdminUserCharacter `json:"characters"`
	Wallet          int64                `json:"wallet"`
	Orders          []AdminUserOrder     `json:"orders"`
} // @name AdminUser

// handleUser360 handles GET /api/v1/admin/users/{id}.
//
//	@Summary		Community user 360 view
//	@Description	Returns the profile, roles, linked account, characters, wallet and recent orders for a community user. Requires the gw.identity.user.read permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.admin.users.get
//	@Produce		json
//	@Param			id	path	string	true	"community user UUID"
//	@Success		200	{object}	AdminUser
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/users/{id} [get]
func (p *Plugin) handleUser360(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, errUser360InvalidID)
		return
	}
	if p.registry == nil {
		httpapi.WriteError(w, r, errUser360Unavailable)
		return
	}
	users, err := services.Consume[userAdmin](p.registry, identityUserAdminService)
	if err != nil {
		httpapi.WriteError(w, r, errUser360Unavailable)
		return
	}
	user, err := users.UserByID(r.Context(), id)
	if err != nil {
		httpapi.WriteError(w, r, errUser360NotFound)
		return
	}

	response := AdminUser{
		UserID:      user.ID.String(),
		DiscordID:   user.DiscordID,
		Username:    user.Username,
		GlobalName:  user.GlobalName,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Roles:       []string{},
		Characters:  []AdminUserCharacter{},
		Orders:      []AdminUserOrder{},
	}
	if !user.CreatedAt.IsZero() {
		response.CreatedAt = user.CreatedAt.UTC().Format(time.RFC3339)
	}
	if roles, rolesErr := users.Roles(r.Context(), id); rolesErr == nil {
		response.Roles = roles
	}

	if accounts, accountErr := services.Consume[accountDirectory](p.registry, accountDirectoryService); accountErr == nil {
		if username, accountID, err := accounts.LinkedAccount(r.Context(), id.String()); err == nil {
			if username != "" {
				response.AccountUsername = &username
			}
			response.AccountID = accountID
		}
	}

	if chars, charErr := services.Consume[characterDirectory](p.registry, characterDirectoryService); charErr == nil {
		if characters, listErr := chars.CharactersByUser(r.Context(), id.String()); listErr == nil {
			response.Characters = make([]AdminUserCharacter, 0, len(characters))
			for _, character := range characters {
				response.Characters = append(response.Characters, adminCharacter(character))
			}
		}
	}

	if store, storeErr := services.Consume[storeAccount](p.registry, storeAccountService); storeErr == nil {
		if balance, walletErr := store.Wallet(r.Context(), id); walletErr == nil {
			response.Wallet = balance
		}
		if orders, orderErr := store.Orders(r.Context(), id, 20, 0); orderErr == nil {
			response.Orders = make([]AdminUserOrder, 0, len(orders))
			for _, order := range orders {
				response.Orders = append(response.Orders, AdminUserOrder{
					OrderID:   order.ID.String(),
					SKU:       order.SKU,
					Points:    order.Points,
					Character: order.Character,
					Status:    order.Status,
					CreatedAt: order.CreatedAt.UTC().Format(time.RFC3339),
				})
			}
		}
	}

	httpapi.WriteJSON(w, http.StatusOK, response)
}

func adminCharacter(character azerothdb.Character) AdminUserCharacter {
	return AdminUserCharacter{
		GUID:   character.GUID,
		Name:   character.Name,
		Level:  character.Level,
		Race:   character.Race,
		Class:  character.Class,
		Guild:  character.GuildName,
		Online: character.Online,
		Money:  character.Money,
	}
}
