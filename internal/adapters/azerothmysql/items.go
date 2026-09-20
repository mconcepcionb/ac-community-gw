package azerothmysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

const itemColumns = `
       entry,
       name,
       class,
       subclass,
       Quality,
       ItemLevel,
       RequiredLevel,
       InventoryType,
       BuyPrice,
       SellPrice,
       displayid,
       description,
       stackable,
       maxcount,
       Flags,
       armor,
       bonding,
       ContainerSlots,
       itemset,
       Delay,
       holy_res,
       fire_res,
       nature_res,
       frost_res,
       shadow_res,
       arcane_res,
       stat_type1, stat_value1,
       stat_type2, stat_value2,
       stat_type3, stat_value3,
       stat_type4, stat_value4,
       stat_type5, stat_value5,
       stat_type6, stat_value6,
       stat_type7, stat_value7,
       stat_type8, stat_value8,
       stat_type9, stat_value9,
       stat_type10, stat_value10,
       dmg_min1, dmg_max1, dmg_type1,
       dmg_min2, dmg_max2, dmg_type2,
       spellid_1, spelltrigger_1, spellcooldown_1,
       spellid_2, spelltrigger_2, spellcooldown_2,
       spellid_3, spelltrigger_3, spellcooldown_3,
       spellid_4, spelltrigger_4, spellcooldown_4,
       spellid_5, spelltrigger_5, spellcooldown_5`

const listItemsQuery = `SELECT ` + itemColumns + `
FROM item_template
WHERE (? = '' OR name LIKE ?)
  AND (? = 0 OR class = ?)
ORDER BY ItemLevel DESC, name ASC
LIMIT ? OFFSET ?`

const findItemQuery = `SELECT ` + itemColumns + `
FROM item_template
WHERE entry = ?
LIMIT 1`

// ItemStore is a read-only AzerothCore world database item client.
type ItemStore struct {
	db *sql.DB
}

var _ azerothdb.ItemReader = (*ItemStore)(nil)

// NewItemStore opens and verifies a connection pool to the world DB.
func NewItemStore(cfg Config) (*ItemStore, error) {
	db, err := openPool(cfg.DSN, cfg, "world")
	if err != nil {
		return nil, err
	}
	return &ItemStore{db: db}, nil
}

// ListItems implements azerothdb.ItemReader.
func (s *ItemStore) ListItems(ctx context.Context, query azerothdb.ItemQuery) ([]azerothdb.Item, error) {
	limit := clampLimit(query.Limit)
	offset := clampOffset(query.Offset)
	filter := strings.TrimSpace(query.Filter)
	pattern := "%" + escapeLike(filter) + "%"

	rows, err := s.db.QueryContext(ctx, listItemsQuery, filter, pattern, query.Class, query.Class, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("azerothmysql: list items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]azerothdb.Item, 0, limit)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("azerothmysql: scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azerothmysql: iterate items: %w", err)
	}
	return items, nil
}

// FindItem implements azerothdb.ItemReader.
func (s *ItemStore) FindItem(ctx context.Context, entry int64) (azerothdb.Item, error) {
	row := s.db.QueryRowContext(ctx, findItemQuery, entry)
	item, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return azerothdb.Item{}, azerothdb.ErrItemNotFound
	}
	if err != nil {
		return azerothdb.Item{}, fmt.Errorf("azerothmysql: find item: %w", err)
	}
	return item, nil
}

// Close closes the pool.
func (s *ItemStore) Close() error { return s.db.Close() }

func scanItem(row rowScanner) (azerothdb.Item, error) {
	var (
		item          azerothdb.Item
		statType      [10]int
		statValue     [10]int
		dmgMin        [2]float64
		dmgMax        [2]float64
		dmgType       [2]int
		spellID       [5]int
		spellTrigger  [5]int
		spellCooldown [5]int
	)
	if err := row.Scan(
		&item.Entry,
		&item.Name,
		&item.Class,
		&item.Subclass,
		&item.Quality,
		&item.ItemLevel,
		&item.RequiredLevel,
		&item.InventoryType,
		&item.BuyPrice,
		&item.SellPrice,
		&item.DisplayID,
		&item.Description,
		&item.Stackable,
		&item.MaxCount,
		&item.Flags,
		&item.Armor,
		&item.Bonding,
		&item.ContainerSlots,
		&item.ItemSet,
		&item.Delay,
		&item.HolyRes,
		&item.FireRes,
		&item.NatureRes,
		&item.FrostRes,
		&item.ShadowRes,
		&item.ArcaneRes,
		&statType[0], &statValue[0],
		&statType[1], &statValue[1],
		&statType[2], &statValue[2],
		&statType[3], &statValue[3],
		&statType[4], &statValue[4],
		&statType[5], &statValue[5],
		&statType[6], &statValue[6],
		&statType[7], &statValue[7],
		&statType[8], &statValue[8],
		&statType[9], &statValue[9],
		&dmgMin[0], &dmgMax[0], &dmgType[0],
		&dmgMin[1], &dmgMax[1], &dmgType[1],
		&spellID[0], &spellTrigger[0], &spellCooldown[0],
		&spellID[1], &spellTrigger[1], &spellCooldown[1],
		&spellID[2], &spellTrigger[2], &spellCooldown[2],
		&spellID[3], &spellTrigger[3], &spellCooldown[3],
		&spellID[4], &spellTrigger[4], &spellCooldown[4],
	); err != nil {
		return azerothdb.Item{}, err
	}
	for i := range statType {
		if statType[i] != 0 || statValue[i] != 0 {
			item.Stats = append(item.Stats, azerothdb.ItemStat{Type: statType[i], Value: statValue[i]})
		}
	}
	for i := range dmgMin {
		if dmgMin[i] != 0 || dmgMax[i] != 0 {
			item.Damage = append(item.Damage, azerothdb.ItemDamage{Min: dmgMin[i], Max: dmgMax[i], Type: dmgType[i]})
		}
	}
	for i := range spellID {
		if spellID[i] != 0 {
			item.Spells = append(item.Spells, azerothdb.ItemSpell{ID: spellID[i], Trigger: spellTrigger[i], Cooldown: spellCooldown[i]})
		}
	}
	return item, nil
}
