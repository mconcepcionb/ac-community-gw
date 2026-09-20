package azerothcharacter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var (
	errCharacterDBNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"character_db_not_configured", "AzerothCore character database is not configured")
	errCharacterDBUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"character_db_unavailable", "AzerothCore character database is unavailable")
	errLoginDBNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"login_db_not_configured", "AzerothCore login database is not configured")
	errLoginDBUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"login_db_unavailable", "AzerothCore login database is unavailable")
	errAccountNotFound = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"account_not_found", "AzerothCore account does not exist")
	errCharacterNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"character_not_found", "character not found")
	errAccountNotLinked = httpapi.NewAPIError(http.StatusNotFound,
		"account_not_linked", "community user has no linked AzerothCore account")
	errNotOwner = httpapi.NewAPIError(http.StatusForbidden,
		"not_owner", "character does not belong to the caller's account")
	errInvalidUserID = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_user_id", "user_id is not a valid UUID")
	errMissingAccount = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"missing_account", "account or account_id is required")
	errDirectoryUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"identity_storage_unavailable", "community user directory is unavailable")
)

// Character is the JSON representation of an AzerothCore character.
type Character struct {
	GUID       int64   `json:"guid"`
	Name       string  `json:"name"`
	Race       int     `json:"race"`
	RaceName   string  `json:"race_name"`
	Class      int     `json:"class"`
	ClassName  string  `json:"class_name"`
	Gender     int     `json:"gender"`
	Level      int     `json:"level"`
	Online     bool    `json:"online"`
	Guild      string  `json:"guild"`
	Money      int64   `json:"money"`
	TotalTime  int     `json:"total_time"`
	LogoutTime *string `json:"logout_time"`
} // @name AzerothCharacter

// CharactersResponse is the body of the character listing endpoints.
type CharactersResponse struct {
	Characters []Character `json:"characters"`
} // @name AzerothCharactersResponse

// handleListCharacters handles GET /api/v1/azeroth/characters.
//
//	@Summary		List characters
//	@Description	Lists characters filtered by account (account_id or account username) with an optional name filter. Requires the azeroth.character.list permission.
//	@Tags			azeroth-character
//	@ID				azeroth.characters.list
//	@Produce		json
//	@Param			account_id	query	int		false	"login account id"
//	@Param			account		query	string	false	"login account username"
//	@Param			filter		query	string	false	"character name filter"
//	@Param			limit		query	int		false	"page size"	default(100)
//	@Param			offset		query	int		false	"page offset"	default(0)
//	@Success		200	{object}	CharactersResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/characters [get]
func (p *Plugin) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	query, err := p.accountQuery(r)
	if err != nil {
		writeCharacterError(w, r, err)
		return
	}
	query.Filter = r.URL.Query().Get("filter")
	query.Limit = intParam(r, "limit", 100)
	query.Offset = intParam(r, "offset", 0)
	p.writeCharacters(w, r, query)
}

// handleListUserCharacters handles GET /api/v1/azeroth/users/{user_id}/characters.
//
//	@Summary		List a community user's characters
//	@Description	Resolves the user's linked AzerothCore account and lists its characters. Requires the azeroth.character.list permission.
//	@Tags			azeroth-character
//	@ID				azeroth.user_characters.list
//	@Produce		json
//	@Param			user_id	path	string	true	"community user UUID"
//	@Param			filter	query	string	false	"character name filter"
//	@Param			limit	query	int		false	"page size"	default(100)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	CharactersResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/users/{user_id}/characters [get]
func (p *Plugin) handleListUserCharacters(w http.ResponseWriter, r *http.Request) {
	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		httpapi.WriteError(w, r, errInvalidUserID)
		return
	}
	if p.directory == nil {
		httpapi.WriteError(w, r, errDirectoryUnavailable)
		return
	}
	_, accountID, err := p.directory.LinkedAccount(r.Context(), userID.String())
	if err != nil || accountID == nil {
		httpapi.WriteError(w, r, errAccountNotLinked)
		return
	}
	query := azerothdb.CharacterQuery{
		AccountID: int64(*accountID),
		Filter:    r.URL.Query().Get("filter"),
		Limit:     intParam(r, "limit", 100),
		Offset:    intParam(r, "offset", 0),
	}
	p.writeCharacters(w, r, query)
}

// handleGetCharacter handles GET /api/v1/azeroth/characters/{name}.
//
//	@Summary		Get a character
//	@Description	Returns one character by name. Requires the azeroth.character.list permission.
//	@Tags			azeroth-character
//	@ID				azeroth.characters.get
//	@Produce		json
//	@Param			name	path	string	true	"character name"
//	@Success		200	{object}	Character
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/characters/{name} [get]
func (p *Plugin) handleGetCharacter(w http.ResponseWriter, r *http.Request) {
	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	character, err := p.characters.FindCharacter(r.Context(), r.PathValue("name"))
	if errors.Is(err, azerothdb.ErrCharacterNotFound) {
		httpapi.WriteError(w, r, errCharacterNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errCharacterDBUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, characterDTO(character))
}

func (p *Plugin) accountQuery(r *http.Request) (azerothdb.CharacterQuery, error) {
	query := r.URL.Query()
	if raw := strings.TrimSpace(query.Get("account_id")); raw != "" {
		accountID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return azerothdb.CharacterQuery{}, errMissingAccount
		}
		return azerothdb.CharacterQuery{AccountID: accountID}, nil
	}
	if username := strings.TrimSpace(query.Get("account")); username != "" {
		if p.accounts == nil {
			return azerothdb.CharacterQuery{}, errLoginDBNotConfigured
		}
		account, err := p.accounts.FindAccountByUsername(r.Context(), username)
		switch {
		case errors.Is(err, azerothdb.ErrAccountNotFound):
			return azerothdb.CharacterQuery{}, errAccountNotFound
		case errors.Is(err, azerothdb.ErrUnavailable):
			return azerothdb.CharacterQuery{}, errLoginDBUnavailable
		case err != nil:
			return azerothdb.CharacterQuery{}, errLoginDBUnavailable
		}
		return azerothdb.CharacterQuery{AccountID: account.ID}, nil
	}
	return azerothdb.CharacterQuery{}, errMissingAccount
}

func (p *Plugin) writeCharacters(w http.ResponseWriter, r *http.Request, query azerothdb.CharacterQuery) {
	characters, err := p.characters.ListCharacters(r.Context(), query)
	if err != nil {
		httpapi.WriteError(w, r, errCharacterDBUnavailable)
		return
	}
	items := make([]Character, 0, len(characters))
	for _, character := range characters {
		items = append(items, characterDTO(character))
	}
	httpapi.WriteJSON(w, http.StatusOK, CharactersResponse{Characters: items})
}

func writeCharacterError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *httpapi.APIError
	if errors.As(err, &apiErr) {
		httpapi.WriteError(w, r, apiErr)
		return
	}
	httpapi.WriteError(w, r, httpapi.ErrInternal)
}

func intParam(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func characterDTO(character azerothdb.Character) Character {
	dto := Character{
		GUID:      character.GUID,
		Name:      character.Name,
		Race:      character.Race,
		RaceName:  raceName(character.Race),
		Class:     character.Class,
		ClassName: className(character.Class),
		Gender:    character.Gender,
		Level:     character.Level,
		Online:    character.Online,
		Guild:     character.GuildName,
		Money:     character.Money,
		TotalTime: character.TotalTime,
	}
	if character.LogoutTime > 0 {
		logout := time.Unix(character.LogoutTime, 0).UTC().Format(time.RFC3339)
		dto.LogoutTime = &logout
	}
	return dto
}

func raceName(race int) string {
	switch race {
	case 1:
		return "Human"
	case 2:
		return "Orc"
	case 3:
		return "Dwarf"
	case 4:
		return "Night Elf"
	case 5:
		return "Undead"
	case 6:
		return "Tauren"
	case 7:
		return "Gnome"
	case 8:
		return "Troll"
	case 10:
		return "Blood Elf"
	case 11:
		return "Draenei"
	default:
		return "Unknown"
	}
}

func className(class int) string {
	switch class {
	case 1:
		return "Warrior"
	case 2:
		return "Paladin"
	case 3:
		return "Hunter"
	case 4:
		return "Rogue"
	case 5:
		return "Priest"
	case 6:
		return "Death Knight"
	case 7:
		return "Shaman"
	case 8:
		return "Mage"
	case 9:
		return "Warlock"
	case 11:
		return "Druid"
	default:
		return "Unknown"
	}
}
