package azerothcharacter

import (
	"context"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// DirectoryService is the capability to list a community user's characters.
const DirectoryService = "azeroth.character.directory"

// Directory is the cross-plugin capability published by this plugin.
type Directory interface {
	// CharactersByUser resolves the user's linked account and lists its
	// characters.
	CharactersByUser(ctx context.Context, userID string) ([]azerothdb.Character, error)
}

// CharactersByUser implements Directory.
func (p *Plugin) CharactersByUser(ctx context.Context, userID string) ([]azerothdb.Character, error) {
	if p.characters == nil {
		return nil, errCharacterDBNotConfigured
	}
	if p.directory == nil {
		return nil, errDirectoryUnavailable
	}
	_, accountID, err := p.directory.LinkedAccount(ctx, userID)
	if err != nil || accountID == nil {
		return nil, errAccountNotLinked
	}
	return p.characters.ListCharacters(ctx, azerothdb.CharacterQuery{
		AccountID: *accountID,
		Limit:     100,
	})
}
