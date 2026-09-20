//go:build integration

package azerothmysql

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

func TestListItems(t *testing.T) {
	dsn := os.Getenv("ACGW_AZEROTH_WORLD_DB_DSN")
	if dsn == "" {
		t.Skip("ACGW_AZEROTH_WORLD_DB_DSN is not set")
	}
	store, err := NewItemStore(Config{DSN: dsn})
	if err != nil {
		t.Fatalf("NewItemStore: %v", err)
	}
	defer func() { _ = store.Close() }()

	items, err := store.ListItems(context.Background(), azerothdb.ItemQuery{Filter: "Hearth"})
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if len(items) != 1 || items[0].Entry != 6948 {
		t.Fatalf("items = %+v", items)
	}

	item, err := store.FindItem(context.Background(), 4496)
	if err != nil {
		t.Fatalf("FindItem: %v", err)
	}
	if item.Name != "Traveler's Backpack" || item.Class != 1 || item.InventoryType != 18 {
		t.Fatalf("item = %+v", item)
	}

	if _, err := store.FindItem(context.Background(), 1); !errors.Is(err, azerothdb.ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
}
