// Package itemview renders an AzerothCore item template into a presentation
// shape shared by the item catalog and the store.
package itemview

import (
	"fmt"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// Stat is a rendered stat bonus.
type Stat struct {
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Value int    `json:"value"`
} // @name AzerothItemStat

// Damage is a rendered damage range.
type Damage struct {
	Type     int     `json:"type"`
	TypeName string  `json:"type_name"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	DPS      float64 `json:"dps"`
} // @name AzerothItemDamage

// Spell is a rendered item spell (names require client DBC data).
type Spell struct {
	ID          int    `json:"id"`
	Trigger     int    `json:"trigger"`
	TriggerName string `json:"trigger_name"`
	CooldownMS  int    `json:"cooldown_ms"`
} // @name AzerothItemSpell

// Resistances is the school resistance block.
type Resistances struct {
	Holy   int `json:"holy"`
	Fire   int `json:"fire"`
	Nature int `json:"nature"`
	Frost  int `json:"frost"`
	Shadow int `json:"shadow"`
	Arcane int `json:"arcane"`
} // @name AzerothItemResistances

// View is the renderable presentation of an item.
type View struct {
	Entry             int64       `json:"entry"`
	Name              string      `json:"name"`
	Quality           int         `json:"quality"`
	QualityName       string      `json:"quality_name"`
	QualityColor      string      `json:"quality_color"`
	Class             int         `json:"class"`
	ClassName         string      `json:"class_name"`
	Subclass          int         `json:"subclass"`
	SubclassName      string      `json:"subclass_name"`
	InventoryType     int         `json:"inventory_type"`
	InventoryTypeName string      `json:"inventory_type_name"`
	ItemLevel         int         `json:"item_level"`
	RequiredLevel     int         `json:"required_level"`
	Bonding           int         `json:"bonding"`
	BondingName       string      `json:"bonding_name"`
	Armor             int         `json:"armor"`
	Delay             int         `json:"delay"`
	Stackable         int         `json:"stackable"`
	MaxCount          int         `json:"max_count"`
	ContainerSlots    int         `json:"container_slots"`
	ItemSet           int         `json:"item_set"`
	Description       string      `json:"description"`
	BuyPrice          int64       `json:"buy_price"`
	SellPrice         int64       `json:"sell_price"`
	DisplayID         int         `json:"display_id"`
	Stats             []Stat      `json:"stats"`
	Damage            []Damage    `json:"damage"`
	Spells            []Spell     `json:"spells"`
	Resistances       Resistances `json:"resistances"`
} // @name AzerothItem

// Build renders an item template.
func Build(item azerothdb.Item) View {
	view := View{
		Entry:             item.Entry,
		Name:              item.Name,
		Quality:           item.Quality,
		QualityName:       qualityName(item.Quality),
		QualityColor:      qualityColor(item.Quality),
		Class:             item.Class,
		ClassName:         className(item.Class),
		Subclass:          item.Subclass,
		SubclassName:      subclassName(item.Class, item.Subclass),
		InventoryType:     item.InventoryType,
		InventoryTypeName: inventoryTypeName(item.InventoryType),
		ItemLevel:         item.ItemLevel,
		RequiredLevel:     item.RequiredLevel,
		Bonding:           item.Bonding,
		BondingName:       bondingName(item.Bonding),
		Armor:             item.Armor,
		Delay:             item.Delay,
		Stackable:         item.Stackable,
		MaxCount:          item.MaxCount,
		ContainerSlots:    item.ContainerSlots,
		ItemSet:           item.ItemSet,
		Description:       item.Description,
		BuyPrice:          item.BuyPrice,
		SellPrice:         item.SellPrice,
		DisplayID:         item.DisplayID,
		Stats:             make([]Stat, 0, len(item.Stats)),
		Damage:            make([]Damage, 0, len(item.Damage)),
		Spells:            make([]Spell, 0, len(item.Spells)),
		Resistances: Resistances{
			Holy:   item.HolyRes,
			Fire:   item.FireRes,
			Nature: item.NatureRes,
			Frost:  item.FrostRes,
			Shadow: item.ShadowRes,
			Arcane: item.ArcaneRes,
		},
	}
	for _, stat := range item.Stats {
		view.Stats = append(view.Stats, Stat{Type: stat.Type, Name: statName(stat.Type), Value: stat.Value})
	}
	for _, damage := range item.Damage {
		dps := 0.0
		if item.Delay > 0 {
			dps = (damage.Min + damage.Max) / 2 / (float64(item.Delay) / 1000)
		}
		view.Damage = append(view.Damage, Damage{
			Type:     damage.Type,
			TypeName: damageTypeName(damage.Type),
			Min:      damage.Min,
			Max:      damage.Max,
			DPS:      dps,
		})
	}
	for _, spell := range item.Spells {
		view.Spells = append(view.Spells, Spell{
			ID:          spell.ID,
			Trigger:     spell.Trigger,
			TriggerName: spellTriggerName(spell.Trigger),
			CooldownMS:  spell.Cooldown,
		})
	}
	return view
}

func qualityName(quality int) string {
	switch quality {
	case 0:
		return "Poor"
	case 1:
		return "Common"
	case 2:
		return "Uncommon"
	case 3:
		return "Rare"
	case 4:
		return "Epic"
	case 5:
		return "Legendary"
	case 6:
		return "Artifact"
	case 7:
		return "Heirloom"
	default:
		return "Unknown"
	}
}

func qualityColor(quality int) string {
	switch quality {
	case 0:
		return "#9d9d9d"
	case 1:
		return "#ffffff"
	case 2:
		return "#1eff00"
	case 3:
		return "#0070dd"
	case 4:
		return "#a335ee"
	case 5:
		return "#ff8000"
	case 6, 7:
		return "#e6cc80"
	default:
		return "#ffffff"
	}
}

func className(class int) string {
	switch class {
	case 0:
		return "Consumable"
	case 1:
		return "Container"
	case 2:
		return "Weapon"
	case 3:
		return "Gem"
	case 4:
		return "Armor"
	case 5:
		return "Reagent"
	case 6:
		return "Projectile"
	case 7:
		return "Trade Goods"
	case 9:
		return "Recipe"
	case 11:
		return "Quiver"
	case 12:
		return "Quest"
	case 13:
		return "Key"
	case 15:
		return "Miscellaneous"
	case 16:
		return "Glyph"
	default:
		return "Unknown"
	}
}

func subclassName(class, subclass int) string {
	switch class {
	case 0:
		return mapOr(consumableSubclasses, subclass)
	case 1:
		return mapOr(containerSubclasses, subclass)
	case 2:
		return mapOr(weaponSubclasses, subclass)
	case 3:
		return mapOr(gemSubclasses, subclass)
	case 4:
		return mapOr(armorSubclasses, subclass)
	case 7:
		return mapOr(tradeGoodsSubclasses, subclass)
	default:
		return fmt.Sprintf("Subclass %d", subclass)
	}
}

var consumableSubclasses = map[int]string{
	0: "Consumable", 1: "Potion", 2: "Elixir", 3: "Flask", 4: "Scroll",
	5: "Food & Drink", 6: "Item Enhancement", 7: "Bandage", 8: "Other",
}

var containerSubclasses = map[int]string{
	0: "Bag", 1: "Soul Bag", 2: "Herb Bag", 3: "Enchanting Bag", 4: "Engineering Bag",
	5: "Gem Bag", 6: "Mining Bag", 7: "Leatherworking Bag", 8: "Inscription Bag",
}

var weaponSubclasses = map[int]string{
	0: "One-Handed Axe", 1: "Two-Handed Axe", 2: "Bow", 3: "Gun", 4: "One-Handed Mace",
	5: "Two-Handed Mace", 6: "Polearm", 7: "One-Handed Sword", 8: "Two-Handed Sword",
	10: "Staff", 13: "Fist Weapon", 14: "Miscellaneous", 15: "Dagger", 16: "Thrown",
	17: "Spear", 18: "Crossbow", 19: "Wand", 20: "Fishing Pole",
}

var gemSubclasses = map[int]string{
	0: "Red", 1: "Blue", 2: "Yellow", 3: "Purple", 4: "Green", 5: "Orange",
	6: "Meta", 7: "Simple", 8: "Prismatic",
}

var armorSubclasses = map[int]string{
	0: "Miscellaneous", 1: "Cloth", 2: "Leather", 3: "Mail", 4: "Plate",
	5: "Buckler", 6: "Shield", 7: "Libram", 8: "Idol", 9: "Totem", 10: "Sigil",
}

var tradeGoodsSubclasses = map[int]string{
	0: "Trade Goods", 1: "Parts", 2: "Explosives", 3: "Devices", 4: "Jewelcrafting",
	5: "Cloth", 6: "Leather", 7: "Metal & Stone", 8: "Meat", 9: "Herb",
	10: "Elemental", 11: "Other", 12: "Enchanting", 13: "Inscription",
}

func inventoryTypeName(inventoryType int) string {
	names := map[int]string{
		0: "Non-equip", 1: "Head", 2: "Neck", 3: "Shoulder", 4: "Shirt", 5: "Chest",
		6: "Waist", 7: "Legs", 8: "Feet", 9: "Wrist", 10: "Hands", 11: "Finger",
		12: "Trinket", 13: "One-Hand", 14: "Shield", 15: "Ranged", 16: "Back",
		17: "Two-Hand", 18: "Bag", 19: "Tabard", 20: "Robe", 21: "Main Hand",
		22: "Off Hand", 23: "Holdable", 24: "Ammo", 25: "Thrown", 26: "Ranged",
		27: "Quiver", 28: "Relic",
	}
	return mapOr(names, inventoryType)
}

func bondingName(bonding int) string {
	switch bonding {
	case 0:
		return "No binding"
	case 1:
		return "Binds when picked up"
	case 2:
		return "Binds when equipped"
	case 3:
		return "Binds when used"
	case 4:
		return "Quest item"
	case 5:
		return "Binds to account"
	default:
		return "Unknown"
	}
}

func damageTypeName(damageType int) string {
	switch damageType {
	case 0:
		return "Physical"
	case 1:
		return "Holy"
	case 2:
		return "Fire"
	case 3:
		return "Nature"
	case 4:
		return "Frost"
	case 5:
		return "Shadow"
	case 6:
		return "Arcane"
	default:
		return "Unknown"
	}
}

func spellTriggerName(trigger int) string {
	switch trigger {
	case 0:
		return "On use"
	case 1:
		return "On equip"
	case 2:
		return "Chance on hit"
	case 3:
		return "Soulstone"
	case 4:
		return "On use (no delay)"
	case 5:
		return "Learn spell"
	case 6:
		return "On learn"
	default:
		return "Unknown"
	}
}

func statName(stat int) string {
	if name, ok := statNames[stat]; ok {
		return name
	}
	return fmt.Sprintf("Stat %d", stat)
}

var statNames = map[int]string{
	0:  "Mana",
	1:  "Health",
	3:  "Agility",
	4:  "Strength",
	5:  "Intellect",
	6:  "Spirit",
	7:  "Stamina",
	12: "Defense Rating",
	13: "Dodge Rating",
	14: "Parry Rating",
	15: "Block Rating",
	16: "Hit Rating",
	31: "Hit Rating",
	32: "Crit Rating",
	35: "Resilience Rating",
	36: "Haste Rating",
	37: "Expertise Rating",
	38: "Attack Power",
	39: "Ranged Attack Power",
	41: "Spell Healing",
	42: "Spell Damage",
	43: "Mana per 5 sec",
	44: "Armor Penetration Rating",
	45: "Spell Power",
	46: "Health per 5 sec",
	47: "Spell Penetration",
	48: "Block Value",
}

func mapOr(names map[int]string, key int) string {
	if name, ok := names[key]; ok {
		return name
	}
	return "Unknown"
}
