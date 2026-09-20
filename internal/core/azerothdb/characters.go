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
	GUID       int64
	AccountID  int64
	Name       string
	Race       int
	Class      int
	Gender     int
	Level      int
	Online     bool
	LogoutTime int64
	TotalTime  int
	Money      int64
	GuildName  string
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
	// FindCharacter returns one character by name or ErrCharacterNotFound.
	FindCharacter(ctx context.Context, name string) (Character, error)
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

// FindCharacter implements CharacterReader.
func (u UnavailableCharacters) FindCharacter(context.Context, string) (Character, error) {
	if u.Err != nil {
		return Character{}, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return Character{}, ErrUnavailable
}
