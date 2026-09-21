package itemview

import (
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

func TestQualityMapping(t *testing.T) {
	cases := []struct {
		quality int
		name    string
		color   string
	}{
		{0, "Poor", "#9d9d9d"},
		{1, "Common", "#ffffff"},
		{2, "Uncommon", "#1eff00"},
		{3, "Rare", "#0070dd"},
		{4, "Epic", "#a335ee"},
		{5, "Legendary", "#ff8000"},
		{6, "Artifact", "#e6cc80"},
		{7, "Heirloom", "#e6cc80"},
		{99, "Unknown", "#ffffff"},
	}
	for _, tc := range cases {
		view := Build(azerothdb.Item{Quality: tc.quality})
		if view.QualityName != tc.name || view.QualityColor != tc.color {
			t.Errorf("quality %d = %s/%s, want %s/%s", tc.quality, view.QualityName, view.QualityColor, tc.name, tc.color)
		}
	}
}

func TestMappingHelpers(t *testing.T) {
	if className(3) != "Gem" || className(99) != "Unknown" {
		t.Fatalf("className = %s/%s", className(3), className(99))
	}
	if subclassName(7, 11) != "Other" || subclassName(99, 3) != "Subclass 3" {
		t.Fatalf("subclassName = %s/%s", subclassName(7, 11), subclassName(99, 3))
	}
	if inventoryTypeName(18) != "Bag" || inventoryTypeName(99) != "Unknown" {
		t.Fatalf("inventoryTypeName = %s/%s", inventoryTypeName(18), inventoryTypeName(99))
	}
	if bondingName(5) != "Binds to account" || bondingName(99) != "Unknown" {
		t.Fatalf("bondingName = %s/%s", bondingName(5), bondingName(99))
	}
	if damageTypeName(6) != "Arcane" || damageTypeName(99) != "Unknown" {
		t.Fatalf("damageTypeName = %s/%s", damageTypeName(6), damageTypeName(99))
	}
	if spellTriggerName(4) != "On use (no delay)" || spellTriggerName(99) != "Unknown" {
		t.Fatalf("spellTriggerName = %s/%s", spellTriggerName(4), spellTriggerName(99))
	}
	if statName(45) != "Spell Power" || statName(99) != "Stat 99" {
		t.Fatalf("statName = %s/%s", statName(45), statName(99))
	}
}

func TestDamageStatsSpellsAndResistances(t *testing.T) {
	view := Build(azerothdb.Item{
		Delay:   2000,
		Damage:  []azerothdb.ItemDamage{{Min: 10, Max: 20, Type: 4}},
		Stats:   []azerothdb.ItemStat{{Type: 7, Value: 5}},
		Spells:  []azerothdb.ItemSpell{{ID: 1, Trigger: 5}},
		HolyRes: 1, FireRes: 2, NatureRes: 3, FrostRes: 4, ShadowRes: 5, ArcaneRes: 6,
	})

	if view.Damage[0].TypeName != "Frost" || view.Damage[0].DPS != 7.5 {
		t.Fatalf("damage = %+v", view.Damage)
	}
	if view.Stats[0].Name != "Stamina" || view.Stats[0].Value != 5 {
		t.Fatalf("stats = %+v", view.Stats)
	}
	if view.Spells[0].TriggerName != "Learn spell" {
		t.Fatalf("spells = %+v", view.Spells)
	}
	if view.Resistances.Holy != 1 || view.Resistances.Arcane != 6 {
		t.Fatalf("resistances = %+v", view.Resistances)
	}
}

func TestNoDelayMeansNoDPS(t *testing.T) {
	view := Build(azerothdb.Item{Damage: []azerothdb.ItemDamage{{Min: 10, Max: 20, Type: 0}}})
	if view.Damage[0].DPS != 0 {
		t.Fatalf("dps without delay = %v", view.Damage[0].DPS)
	}
}
