package azerothcharacter

import (
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

// handleMyCharacters handles GET /api/v1/azeroth/me/characters.
//
//	@Summary		List my characters
//	@Description	Lists the characters of the authenticated user's linked account. Requires the azeroth.character.self permission.
//	@Tags			azeroth-character
//	@ID				azeroth.me.characters.list
//	@Produce		json
//	@Param			filter	query	string	false	"character name filter"
//	@Param			limit	query	int		false	"page size"	default(100)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	CharactersResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/characters [get]
func (p *Plugin) handleMyCharacters(w http.ResponseWriter, r *http.Request) {
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
	query := azerothdb.CharacterQuery{
		AccountID: *accountID,
		Filter:    r.URL.Query().Get("filter"),
		Limit:     intParam(r, "limit", 100),
		Offset:    intParam(r, "offset", 0),
	}
	p.writeCharacters(w, r, query)
}

// handleMyMail handles POST /api/v1/azeroth/me/mail.
//
//	@Summary		Mail my character
//	@Description	Delivers items and/or money to a character owned by the authenticated user's linked account. Requires the azeroth.mail.self permission.
//	@Tags			azeroth-character
//	@ID				azeroth.me.mail.send
//	@Accept			json
//	@Produce		json
//	@Param			request	body	SendMailRequest	true	"delivery request"
//	@Success		200	{object}	SendMailResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/mail [post]
func (p *Plugin) handleMyMail(w http.ResponseWriter, r *http.Request) {
	p.handleSendMail(w, r)
}
