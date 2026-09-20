package azerothcharacter

import (
	"net/http"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var (
	errLeaderboardUnknown = httpapi.NewAPIError(http.StatusNotFound,
		"unknown_board", "unknown leaderboard")
	errLeaderboardUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"leaderboard_unavailable", "leaderboards are unavailable")
)

// LeaderboardEntry is one ranked character.
type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	ClassName   string `json:"class_name"`
	RaceName    string `json:"race_name"`
	Guild       string `json:"guild"`
	Money       int64  `json:"money"`
	TotalTime   int    `json:"total_time"`
	ArenaPoints int    `json:"arena_points"`
} // @name AzerothLeaderboardEntry

// LeaderboardResponse is the body of GET /api/v1/azeroth/leaderboards/{board}.
type LeaderboardResponse struct {
	Board   string             `json:"board"`
	Entries []LeaderboardEntry `json:"entries"`
} // @name AzerothLeaderboardResponse

// handleLeaderboard handles GET /api/v1/azeroth/leaderboards/{board}.
//
//	@Summary		Character leaderboard
//	@Description	Ranks opted-in characters by progression, wealth, playtime or PvP. Requires the azeroth.leaderboard.read permission.
//	@Tags			azeroth-character
//	@ID				azeroth.leaderboards.get
//	@Produce		json
//	@Param			board	path	string	true	"board: progression, wealth, playtime or pvp"
//	@Param			limit	query	int		false	"page size"	default(25)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	LeaderboardResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/leaderboards/{board} [get]
func (p *Plugin) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	board := strings.TrimSpace(r.PathValue("board"))
	if !azerothdb.ValidLeaderboard(board) {
		httpapi.WriteError(w, r, errLeaderboardUnknown)
		return
	}
	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	if p.visibility == nil {
		httpapi.WriteError(w, r, errVisibilityUnavailable)
		return
	}
	public, err := p.visibility.PublicNames(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, errLeaderboardUnavailable)
		return
	}
	allowed := make(map[string]bool, len(public))
	for _, name := range public {
		allowed[name] = true
	}

	limit := intParam(r, "limit", 25)
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	offset := intParam(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}
	window := offset + limit + 100

	characters, err := p.characters.TopCharacters(r.Context(), board, window, 0)
	if err != nil {
		httpapi.WriteError(w, r, errCharacterDBUnavailable)
		return
	}
	filtered := make([]azerothdb.Character, 0, len(characters))
	for _, character := range characters {
		if allowed[character.Name] {
			filtered = append(filtered, character)
		}
	}

	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	entries := make([]LeaderboardEntry, 0, end-start)
	for i := start; i < end; i++ {
		character := filtered[i]
		entries = append(entries, LeaderboardEntry{
			Rank:        i + 1,
			Name:        character.Name,
			Level:       character.Level,
			ClassName:   className(character.Class),
			RaceName:    raceName(character.Race),
			Guild:       character.GuildName,
			Money:       character.Money,
			TotalTime:   character.TotalTime,
			ArenaPoints: character.ArenaPoints,
		})
	}
	httpapi.WriteJSON(w, http.StatusOK, LeaderboardResponse{Board: board, Entries: entries})
}
