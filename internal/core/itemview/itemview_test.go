package itemview

import (
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

func TestBuildRendersWeapon(t *testing.T) {
	view := Build(azerothdb.Item{
		Entry: 200000, Name: "Community Blade", Class: 2, Subclass: 7, Quality: 4,
		ItemLevel: 70, RequiredLevel: 60, InventoryType: 13, Delay: 2000,
		Bonding: 1,
		Stats:   []azerothdb.ItemStat{{Type: 4, Value: 20}, {Type: 7, Value: 15}},
		Damage:  []azerothdb.ItemDamage{{Min: 50, Max: 90, Type: 0}},
	})

	if view.QualityName != "Epic" || view.QualityColor != "#a335ee" {
		t.Fatalf("quality = %s %s", view.QualityName, view.QualityColor)
	}
	if view.ClassName != "Weapon" || view.SubclassName != "One-Handed Sword" {
		t.Fatalf("class = %s / %s", view.ClassName, view.SubclassName)
	}
	if view.InventoryTypeName != "One-Hand" {
		t.Fatalf("inventory = %s", view.InventoryTypeName)
	}
	if view.BondingName != "Binds when picked up" {
		t.Fatalf("bonding = %s", view.BondingName)
	}
	if len(view.Stats) != 2 || view.Stats[0].Name != "Strength" || view.Stats[0].Value != 20 {
		t.Fatalf("stats = %+v", view.Stats)
	}
	if len(view.Damage) != 1 || view.Damage[0].TypeName != "Physical" || view.Damage[0].DPS != 35 {
		t.Fatalf("damage = %+v", view.Damage)
	}
}

func TestBuildRendersContainerAndSpells(t *testing.T) {
	view := Build(azerothdb.Item{
		Entry: 4496, Name: "Traveler's Backpack", Class: 1, Subclass: 0, Quality: 1,
		ContainerSlots: 16, InventoryType: 18,
		Spells: []azerothdb.ItemSpell{{ID: 21992, Trigger: 2}},
	})
	if view.SubclassName != "Bag" || view.ContainerSlots != 16 {
		t.Fatalf("container = %s slots=%d", view.SubclassName, view.ContainerSlots)
	}
	if len(view.Spells) != 1 || view.Spells[0].TriggerName != "Chance on hit" {
		t.Fatalf("spells = %+v", view.Spells)
	}
}
