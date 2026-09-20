package azerothdb

import (
	"context"
	"errors"
	"fmt"
)

// ErrCharacterNotFound is returned when a character name does not exist.
var ErrCharacterNotFound = errors.New("azerothdb: character not found")

// Character is a read-only view of an AzerothCore character.
type Character struct {
	GUID        int64
	AccountID   int64
	Name        string
	Race        int
	Class       int
	Gender      int
	Level       int
	Online      bool
	LogoutTime  int64
	TotalTime   int
	Money       int64
	GuildName   string
	ArenaPoints int
	Banned      bool
	BanReason   string
}

// Equipment is one equipped item in a character slot.
type Equipment struct {
	Slot  int
	Entry int64
	Count int
}

// Leaderboard board names.
const (
	// BoardProgression ranks by character level.
	BoardProgression = "progression"
	// BoardWealth ranks by money.
	BoardWealth = "wealth"
	// BoardPlaytime ranks by total played time.
	BoardPlaytime = "playtime"
	// BoardPvP ranks by arena points.
	BoardPvP = "pvp"
)

// ValidLeaderboard reports whether a board name is supported.
func ValidLeaderboard(board string) bool {
	switch board {
	case BoardProgression, BoardWealth, BoardPlaytime, BoardPvP:
		return true
	default:
		return false
	}
}

// CharacterQuery filters and paginates a character listing.
type CharacterQuery struct {
	// AccountID restricts the listing to one login account.
	AccountID int64
	// Filter matches character names with a case-insensitive substring.
	Filter string
	Limit  int
	Offset int
}

// CharacterReader reads AzerothCore characters.
type CharacterReader interface {
	ListCharacters(ctx context.Context, query CharacterQuery) ([]Character, error)
	// CountCharacters returns the number of characters matching the query,
	// ignoring limit and offset.
	CountCharacters(ctx context.Context, query CharacterQuery) (int, error)
	// Equipment returns the character's equipped items, ordered by slot.
	Equipment(ctx context.Context, guid int64) ([]Equipment, error)
	// FindCharacter returns one character by name or ErrCharacterNotFound.
	FindCharacter(ctx context.Context, name string) (Character, error)
	// TopCharacters returns one leaderboard page ordered by the board metric.
	TopCharacters(ctx context.Context, board string, limit, offset int) ([]Character, error)
}

// UnavailableCharacters is a CharacterReader used when the character database
// was configured but could not be reached.
type UnavailableCharacters struct {
	Err error
}

// ListCharacters implements CharacterReader.
func (u UnavailableCharacters) ListCharacters(context.Context, CharacterQuery) ([]Character, error) {
	if u.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return nil, ErrUnavailable
}

// CountCharacters implements CharacterReader.
func (u UnavailableCharacters) CountCharacters(context.Context, CharacterQuery) (int, error) {
	if u.Err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return 0, ErrUnavailable
}

// Equipment implements CharacterReader.
func (u UnavailableCharacters) Equipment(context.Context, int64) ([]Equipment, error) {
	if u.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return nil, ErrUnavailable
}

// FindCharacter implements CharacterReader.
func (u UnavailableCharacters) FindCharacter(context.Context, string) (Character, error) {
	if u.Err != nil {
		return Character{}, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return Character{}, ErrUnavailable
}

// TopCharacters implements CharacterReader.
func (u UnavailableCharacters) TopCharacters(context.Context, string, int, int) ([]Character, error) {
	if u.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return nil, ErrUnavailable
}
