package azerothdb

import (
	"context"
	"errors"
	"fmt"
)

// ErrItemNotFound is returned when an item entry does not exist.
var ErrItemNotFound = errors.New("azerothdb: item not found")

// ItemStat is one stat bonus of an item.
type ItemStat struct {
	Type  int
	Value int
}

// ItemDamage is one damage range of an item.
type ItemDamage struct {
	Min  float64
	Max  float64
	Type int
}

// ItemSpell is one spell bound to an item.
type ItemSpell struct {
	ID       int
	Trigger  int
	Cooldown int
}

// Item is a read-only view of an AzerothCore item template.
type Item struct {
	Entry          int64
	Name           string
	Class          int
	Subclass       int
	Quality        int
	ItemLevel      int
	RequiredLevel  int
	InventoryType  int
	BuyPrice       int64
	SellPrice      int64
	DisplayID      int
	Description    string
	Stackable      int
	MaxCount       int
	Flags          int64
	Armor          int
	Bonding        int
	ContainerSlots int
	ItemSet        int
	Delay          int
	HolyRes        int
	FireRes        int
	NatureRes      int
	FrostRes       int
	ShadowRes      int
	ArcaneRes      int
	Stats          []ItemStat
	Damage         []ItemDamage
	Spells         []ItemSpell
}

// ItemQuery filters and paginates an item catalog search.
type ItemQuery struct {
	// Filter matches item names with a case-insensitive substring.
	Filter string
	// Class restricts the search to one item class (0 = any).
	Class  int
	Limit  int
	Offset int
}

// ItemReader reads AzerothCore item templates.
type ItemReader interface {
	ListItems(ctx context.Context, query ItemQuery) ([]Item, error)
	// FindItem returns one item by entry or ErrItemNotFound.
	FindItem(ctx context.Context, entry int64) (Item, error)
}

// Catalog is the cross-plugin item lookup capability published by azeroth-item.
type Catalog interface {
	LookupItem(ctx context.Context, entry int64) (Item, bool, error)
}

// UnavailableItems is an ItemReader used when the world database was configured
// but could not be reached.
type UnavailableItems struct {
	Err error
}

// ListItems implements ItemReader.
func (u UnavailableItems) ListItems(context.Context, ItemQuery) ([]Item, error) {
	if u.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return nil, ErrUnavailable
}

// FindItem implements ItemReader.
func (u UnavailableItems) FindItem(context.Context, int64) (Item, error) {
	if u.Err != nil {
		return Item{}, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return Item{}, ErrUnavailable
}
