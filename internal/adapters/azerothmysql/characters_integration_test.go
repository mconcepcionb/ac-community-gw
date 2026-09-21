//go:build integration

package azerothmysql

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// TestListCharacters reads the compose MariaDB fixture, which is read-only for
// the gateway user: the assertions target the seeded Thrall/Jaina/Arthas rows.
func TestListCharacters(t *testing.T) {
	dsn := os.Getenv("ACGW_AZEROTH_CHARACTER_DB_DSN")
	if dsn == "" {
		t.Skip("ACGW_AZEROTH_CHARACTER_DB_DSN is not set")
	}
	store, err := NewCharacterStore(Config{DSN: dsn})
	if err != nil {
		t.Fatalf("NewCharacterStore: %v", err)
	}
	defer func() { _ = store.Close() }()
	ctx := context.Background()

	thrall, err := store.FindCharacter(ctx, "Thrall")
	if err != nil {
		t.Fatalf("FindCharacter: %v", err)
	}
	if thrall.Level != 80 || thrall.Class != 7 || !thrall.Online || thrall.GuildName != "Community" {
		t.Fatalf("thrall = %+v", thrall)
	}

	characters, err := store.ListCharacters(ctx, azerothdb.CharacterQuery{
		AccountID: thrall.AccountID,
		Filter:    "Thrall",
	})
	if err != nil || len(characters) != 1 || characters[0].Name != "Thrall" {
		t.Fatalf("ListCharacters = %+v, %v", characters, err)
	}

	count, err := store.CountCharacters(ctx, azerothdb.CharacterQuery{AccountID: thrall.AccountID})
	if err != nil || count != 2 {
		t.Fatalf("CountCharacters = %d, %v (want Thrall and Jaina)", count, err)
	}

	online, err := store.ListCharacters(ctx, azerothdb.CharacterQuery{OnlineOnly: true})
	if err != nil {
		t.Fatalf("ListCharacters online: %v", err)
	}
	foundOnline := false
	for _, character := range online {
		if character.Name == "Thrall" {
			foundOnline = true
		}
	}
	if !foundOnline {
		t.Fatalf("Thrall missing from the online list: %+v", online)
	}

	equipment, err := store.Equipment(ctx, thrall.GUID)
	if err != nil || len(equipment) != 2 {
		t.Fatalf("Equipment = %+v, %v", equipment, err)
	}

	top, err := store.TopCharacters(ctx, azerothdb.BoardProgression, 10, 0)
	if err != nil || len(top) == 0 {
		t.Fatalf("TopCharacters = %+v, %v", top, err)
	}

	if _, err := store.FindCharacter(ctx, "NoSuchCharacter"); !errors.Is(err, azerothdb.ErrCharacterNotFound) {
		t.Fatalf("FindCharacter missing = %v, want ErrCharacterNotFound", err)
	}
}
