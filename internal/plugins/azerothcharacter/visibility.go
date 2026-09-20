package azerothcharacter

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

// VisibilityStore persists the per-character public visibility flag.
type VisibilityStore interface {
	SetVisibility(ctx context.Context, characterName string, userID uuid.UUID, public bool) (bool, error)
	GetVisibility(ctx context.Context, characterName string) (bool, error)
	ListByUser(ctx context.Context, userID uuid.UUID) (map[string]bool, error)
	PublicNames(ctx context.Context) ([]string, error)
}

var errVisibilityUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
	"visibility_unavailable", "character visibility is unavailable")

// VisibilityEntry is one character's public flag.
type VisibilityEntry struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
} // @name AzerothCharacterVisibility

// VisibilityResponse is the body of the visibility listing.
type VisibilityResponse struct {
	Items []VisibilityEntry `json:"items"`
} // @name AzerothCharacterVisibilityResponse

// SetVisibilityRequest is the body of the visibility update.
type SetVisibilityRequest struct {
	Public bool `json:"public"`
} // @name AzerothSetVisibilityRequest

// handleListVisibility handles GET /api/v1/azeroth/me/characters/visibility.
//
//	@Summary		List my character visibility
//	@Description	Returns the public-board visibility flag for each of the user's characters. Requires the azeroth.character.self permission.
//	@Tags			azeroth-character
//	@ID				azeroth.me.characters.visibility.list
//	@Produce		json
//	@Success		200	{object}	VisibilityResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/characters/visibility [get]
func (p *Plugin) handleListVisibility(w http.ResponseWriter, r *http.Request) {
	if p.visibility == nil {
		httpapi.WriteError(w, r, errVisibilityUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	flags, err := p.visibility.ListByUser(r.Context(), principal.UserID)
	if err != nil {
		httpapi.WriteError(w, r, errVisibilityUnavailable)
		return
	}
	items := make([]VisibilityEntry, 0, len(flags))
	for name, public := range flags {
		items = append(items, VisibilityEntry{Name: name, Public: public})
	}
	httpapi.WriteJSON(w, http.StatusOK, VisibilityResponse{Items: items})
}

// handleSetVisibility handles PUT /api/v1/azeroth/me/characters/{name}/visibility.
//
//	@Summary		Set my character visibility
//	@Description	Toggles whether a character owned by the user appears on public boards. Requires the azeroth.character.self permission.
//	@Tags			azeroth-character
//	@ID				azeroth.me.characters.visibility.set
//	@Accept			json
//	@Produce		json
//	@Param			name	path	string					true	"character name"
//	@Param			request	body	SetVisibilityRequest	true	"visibility"
//	@Success		200	{object}	VisibilityEntry
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/characters/{name}/visibility [put]
func (p *Plugin) handleSetVisibility(w http.ResponseWriter, r *http.Request) {
	if p.visibility == nil {
		httpapi.WriteError(w, r, errVisibilityUnavailable)
		return
	}
	name := strings.TrimSpace(r.PathValue("name"))
	if !characterNamePattern.MatchString(name) {
		httpapi.WriteError(w, r, errInvalidRecipient)
		return
	}
	var req SetVisibilityRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	if p.directory == nil {
		httpapi.WriteError(w, r, errDirectoryUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	_, accountID, err := p.directory.LinkedAccount(r.Context(), principal.UserID.String())
	if err != nil || accountID == nil {
		httpapi.WriteError(w, r, errAccountNotLinked)
		return
	}
	character, err := p.characters.FindCharacter(r.Context(), name)
	if errors.Is(err, azerothdb.ErrCharacterNotFound) {
		httpapi.WriteError(w, r, errCharacterNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errCharacterDBUnavailable)
		return
	}
	if character.AccountID != *accountID {
		httpapi.WriteError(w, r, errNotOwner)
		return
	}
	public, err := p.visibility.SetVisibility(r.Context(), name, principal.UserID, req.Public)
	if err != nil {
		httpapi.WriteError(w, r, errVisibilityUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, VisibilityEntry{Name: name, Public: public})
}
