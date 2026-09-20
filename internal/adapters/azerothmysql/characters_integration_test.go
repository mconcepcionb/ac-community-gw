//go:build integration

package azerothmysql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

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

	// The compose fixture seeds account id 1 (ADMIN).
	const accountID = 1
	guid := int64(400000 + time.Now().UnixNano()%500000)
	name := fmt.Sprintf("It%d", time.Now().UnixNano()%1_000_000)
	_, err = store.db.Exec(`
		INSERT INTO characters (guid, account, name, race, class, gender, level, money, online, taximask, innTriggerId)
		VALUES (?, ?, ?, 2, 7, 0, 42, 123, 1, '', 0)`, guid, accountID, name)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.db.Exec(`DELETE FROM characters WHERE guid = ?`, guid)
	})

	characters, err := store.ListCharacters(context.Background(), azerothdb.CharacterQuery{
		AccountID: accountID,
		Filter:    name,
	})
	if err != nil {
		t.Fatalf("ListCharacters: %v", err)
	}
	if len(characters) != 1 || characters[0].Name != name || characters[0].Level != 42 {
		t.Fatalf("characters = %+v", characters)
	}

	found, err := store.FindCharacter(context.Background(), name)
	if err != nil {
		t.Fatalf("FindCharacter: %v", err)
	}
	if found.Class != 7 || !found.Online {
		t.Fatalf("found = %+v", found)
	}

	if _, err := store.FindCharacter(context.Background(), "NoSuchCharacter"); !errors.Is(err, azerothdb.ErrCharacterNotFound) {
		t.Fatalf("expected ErrCharacterNotFound, got %v", err)
	}
}
