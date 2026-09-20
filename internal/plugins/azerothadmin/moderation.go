package azerothadmin

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var (
	errInvalidName = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_name", "player/character name is invalid")
	errInvalidDuration = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_duration", "duration is invalid")
	errEmptyMessage = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"empty_message", "message is required")
	errCommandFailed = httpapi.NewAPIError(http.StatusBadGateway,
		"command_failed", "AzerothCore rejected the command")
	errOnlineUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"online_unavailable", "the online list is unavailable")
)

// playerNamePattern keeps names safe to embed in the command string.
var playerNamePattern = regexp.MustCompile(`^[A-Za-z]{2,32}$`)

// kickPlayer builds and executes `.kick <name> [reason]`.
func (p *Plugin) kickPlayer(ctx context.Context, name, reason string) (string, error) {
	return p.executor.Execute(ctx, fmt.Sprintf(".kick %s %s", name, reason))
}

// mutePlayer builds and executes `.mute <name> <duration> <reason]`.
func (p *Plugin) mutePlayer(ctx context.Context, name, duration, reason string) (string, error) {
	return p.executor.Execute(ctx, fmt.Sprintf(".mute %s %s %s", name, duration, reason))
}

// unmutePlayer builds and executes `.unmute <name>`.
func (p *Plugin) unmutePlayer(ctx context.Context, name string) (string, error) {
	return p.executor.Execute(ctx, fmt.Sprintf(".unmute %s", name))
}

// banCharacter builds and executes `.ban character <name> <duration> <reason>`.
func (p *Plugin) banCharacter(ctx context.Context, name, duration, reason string) (string, error) {
	return p.executor.Execute(ctx, fmt.Sprintf(".ban character %s %s %s", name, duration, reason))
}

// unbanCharacter builds and executes `.unban character <name>`.
func (p *Plugin) unbanCharacter(ctx context.Context, name string) (string, error) {
	return p.executor.Execute(ctx, fmt.Sprintf(".unban character %s", name))
}

// announce builds and executes `.announce <message>`.
func (p *Plugin) announce(ctx context.Context, message string) (string, error) {
	return p.executor.Execute(ctx, fmt.Sprintf(".announce %s", message))
}

// KickPlayerRequest is the optional body of POST /api/v1/azeroth/players/{name}/kick.
type KickPlayerRequest struct {
	Reason string `json:"reason"`
} // @name KickPlayerRequest

// MutePlayerRequest is the body of POST /api/v1/azeroth/players/{name}/mute.
type MutePlayerRequest struct {
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
} // @name MutePlayerRequest

// BanCharacterRequest is the body of POST /api/v1/azeroth/characters/{name}/ban.
type BanCharacterRequest struct {
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
} // @name BanCharacterRequest

// AnnounceRequest is the body of POST /api/v1/azeroth/announce.
type AnnounceRequest struct {
	Message string `json:"message"`
} // @name AnnounceRequest

// OnlinePlayer is one character currently online.
type OnlinePlayer struct {
	GUID      int64  `json:"guid"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	Class     int    `json:"class"`
	ClassName string `json:"class_name"`
	Race      int    `json:"race"`
	RaceName  string `json:"race_name"`
	Guild     string `json:"guild"`
	AccountID int64  `json:"account_id"`
} // @name AzerothOnlinePlayer

// OnlineListResponse is the body of GET /api/v1/azeroth/online.
type OnlineListResponse struct {
	Players []OnlinePlayer `json:"players"`
} // @name AzerothOnlineListResponse

// handleOnlineList handles GET /api/v1/azeroth/online.
//
//	@Summary		List online players
//	@Description	Lists characters currently online, read from the character database. Requires the azeroth.admin.players.read permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.online.list
//	@Produce		json
//	@Success		200	{object}	OnlineListResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/online [get]
func (p *Plugin) handleOnlineList(w http.ResponseWriter, r *http.Request) {
	if p.registry == nil {
		httpapi.WriteError(w, r, errOnlineUnavailable)
		return
	}
	value, err := p.registry.Resolve(characterDirectoryService)
	if err != nil {
		httpapi.WriteError(w, r, errOnlineUnavailable)
		return
	}
	directory, ok := value.(characterDirectory)
	if !ok {
		httpapi.WriteError(w, r, errOnlineUnavailable)
		return
	}
	characters, err := directory.OnlineCharacters(r.Context(), 200, 0)
	if err != nil {
		httpapi.WriteError(w, r, errOnlineUnavailable)
		return
	}
	players := make([]OnlinePlayer, 0, len(characters))
	for _, character := range characters {
		players = append(players, OnlinePlayer{
			GUID:      character.GUID,
			Name:      character.Name,
			Level:     character.Level,
			Class:     character.Class,
			ClassName: adminClassName(character.Class),
			Race:      character.Race,
			RaceName:  adminRaceName(character.Race),
			Guild:     character.GuildName,
			AccountID: character.AccountID,
		})
	}
	httpapi.WriteJSON(w, http.StatusOK, OnlineListResponse{Players: players})
}

func adminClassName(class int) string {
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

func adminRaceName(race int) string {
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

// handleKick handles POST /api/v1/azeroth/players/{name}/kick.
//
//	@Summary		Kick a player
//	@Description	Kicks an online player. Requires the azeroth.admin.players.kick permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.players.kick
//	@Accept			json
//	@Produce		json
//	@Param			name	path	string				true	"player name"
//	@Param			request	body	KickPlayerRequest	false	"optional reason"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/players/{name}/kick [post]
func (p *Plugin) handleKick(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if !playerNamePattern.MatchString(name) {
		httpapi.WriteError(w, r, errInvalidName)
		return
	}
	var req KickPlayerRequest
	if r.ContentLength > 0 {
		if err := httpapi.DecodeJSON(r, &req); err != nil {
			httpapi.WriteError(w, r, httpapi.ErrBadRequest)
			return
		}
	}
	output, err := p.kickPlayer(r.Context(), name, sanitizeText(req.Reason))
	p.record(r, "azeroth.players.kick", PermissionAdminPlayersKick, "player", name, err)
	writeCommandOutput(w, r, output, err)
}

// handleMute handles POST /api/v1/azeroth/players/{name}/mute.
//
//	@Summary		Mute a player
//	@Description	Mutes an online player. Requires the azeroth.admin.players.mute permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.players.mute
//	@Accept			json
//	@Produce		json
//	@Param			name	path	string				true	"player name"
//	@Param			request	body	MutePlayerRequest	true	"mute details"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/players/{name}/mute [post]
func (p *Plugin) handleMute(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if !playerNamePattern.MatchString(name) {
		httpapi.WriteError(w, r, errInvalidName)
		return
	}
	var req MutePlayerRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	duration, err := sanitizeDuration(req.Duration)
	if err != nil {
		httpapi.WriteError(w, r, errInvalidDuration)
		return
	}
	output, err := p.mutePlayer(r.Context(), name, duration, sanitizeReason(req.Reason))
	p.record(r, "azeroth.players.mute", PermissionAdminPlayersMute, "player", name, err)
	writeCommandOutput(w, r, output, err)
}

// handleUnmute handles POST /api/v1/azeroth/players/{name}/unmute.
//
//	@Summary		Unmute a player
//	@Description	Unmutes a player. Requires the azeroth.admin.players.mute permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.players.unmute
//	@Produce		json
//	@Param			name	path	string	true	"player name"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/players/{name}/unmute [post]
func (p *Plugin) handleUnmute(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if !playerNamePattern.MatchString(name) {
		httpapi.WriteError(w, r, errInvalidName)
		return
	}
	output, err := p.unmutePlayer(r.Context(), name)
	p.record(r, "azeroth.players.unmute", PermissionAdminPlayersMute, "player", name, err)
	writeCommandOutput(w, r, output, err)
}

// handleBanCharacter handles POST /api/v1/azeroth/characters/{name}/ban.
//
//	@Summary		Ban a character
//	@Description	Bans a character. Requires the azeroth.admin.characters.ban permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.characters.ban
//	@Accept			json
//	@Produce		json
//	@Param			name	path	string				true	"character name"
//	@Param			request	body	BanCharacterRequest	true	"ban details"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/characters/{name}/ban [post]
func (p *Plugin) handleBanCharacter(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if !playerNamePattern.MatchString(name) {
		httpapi.WriteError(w, r, errInvalidName)
		return
	}
	var req BanCharacterRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	duration, err := sanitizeDuration(req.Duration)
	if err != nil {
		httpapi.WriteError(w, r, errInvalidDuration)
		return
	}
	output, err := p.banCharacter(r.Context(), name, duration, sanitizeReason(req.Reason))
	p.record(r, "azeroth.characters.ban", PermissionAdminCharactersBan, "character", name, err)
	writeCommandOutput(w, r, output, err)
}

// handleUnbanCharacter handles POST /api/v1/azeroth/characters/{name}/unban.
//
//	@Summary		Unban a character
//	@Description	Lifts the ban on a character. Requires the azeroth.admin.characters.ban permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.characters.unban
//	@Produce		json
//	@Param			name	path	string	true	"character name"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/characters/{name}/unban [post]
func (p *Plugin) handleUnbanCharacter(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if !playerNamePattern.MatchString(name) {
		httpapi.WriteError(w, r, errInvalidName)
		return
	}
	output, err := p.unbanCharacter(r.Context(), name)
	p.record(r, "azeroth.characters.unban", PermissionAdminCharactersBan, "character", name, err)
	writeCommandOutput(w, r, output, err)
}

// handleAnnounce handles POST /api/v1/azeroth/announce.
//
//	@Summary		Broadcast an announcement
//	@Description	Sends a server-wide announcement. Requires the azeroth.admin.announce permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.announce.send
//	@Accept			json
//	@Produce		json
//	@Param			request	body	AnnounceRequest	true	"announcement"
//	@Success		200	{object}	httpapi.CommandResult
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/announce [post]
func (p *Plugin) handleAnnounce(w http.ResponseWriter, r *http.Request) {
	var req AnnounceRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	message := sanitizeText(req.Message)
	if message == "" {
		httpapi.WriteError(w, r, errEmptyMessage)
		return
	}
	output, err := p.announce(r.Context(), message)
	p.record(r, "azeroth.announce", PermissionAdminAnnounce, "server", "", err)
	writeCommandOutput(w, r, output, err)
}

func writeCommandOutput(w http.ResponseWriter, r *http.Request, output string, err error) {
	if err != nil {
		httpapi.WriteError(w, r, errCommandFailed)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.CommandResult{Result: output})
}

func sanitizeDuration(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "1d", nil
	}
	return azerothcore.SafeBanDuration(value)
}

func sanitizeReason(value string) string {
	cleaned := sanitizeText(value)
	if cleaned == "" {
		return "No reason"
	}
	return cleaned
}

// sanitizeText removes characters that could break the command line and caps
// the length.
func sanitizeText(value string) string {
	return azerothcore.SafeQuotedText(value)
}
