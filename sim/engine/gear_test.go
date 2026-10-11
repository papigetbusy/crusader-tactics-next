package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

func blade(name string, atk int) content.Item {
	return content.Item{Name: name, Slot: "w", Rarity: 0, WeaponType: "sword", Atk: atk}
}

// armedDuel is duel() with the hero holding gear.
func armedDuel(class string, gear map[string]content.Item) *Game {
	g := duel()
	h := g.B.Unit(1)
	h.Class, h.Level = class, 10
	h.Skills = g.C.HeroSkills(class, 10)
	h.Cooldowns = make([]int, len(h.Skills))
	h.Gear, h.BaseAtk = gear, h.Atk
	stats.InitSets(g.C, h)
	return g
}

func TestEquip(t *testing.T) {
	t.Run("Putting on an item from the bag works out the hero's numbers again and puts what it replaced in the bag", func(t *testing.T) {
		g := armedDuel("Warrior", map[string]content.Item{"w": blade("Rusty Blade", 3)})
		g.B.Bag = []content.Item{{Name: "Plate Armor", Slot: "a", Weight: "H", Def: 7, HP: 10}, blade("Steel Sword", 8)}
		h := g.B.Unit(1)
		def := h.Def
		if _, err := g.Apply(action.Action{Kind: action.Equip, Unit: 1, Item: 0}); err != nil {
			t.Fatal(err)
		}
		if h.Gear["a"].Name != "Plate Armor" || h.Def <= def || len(g.B.Bag) != 1 {
			t.Fatalf("armor %q, DEF %d (was %d), bag %d", h.Gear["a"].Name, h.Def, def, len(g.B.Bag))
		}
		h.Done = false // a new turn
		if _, err := g.Apply(action.Action{Kind: action.Equip, Unit: 1, Item: 0, Slot: "w"}); err != nil {
			t.Fatal(err)
		}
		if h.Gear["w"].Name != "Steel Sword" || h.Atk != h.BaseAtk+8 || g.B.Bag[0].Name != "Rusty Blade" {
			t.Fatalf("weapon %q, attack %d (base %d), bag %v", h.Gear["w"].Name, h.Atk, h.BaseAtk, g.B.Bag)
		}
	})

	t.Run("Putting on gear in battle uses up the hero's turn", func(t *testing.T) {
		g := armedDuel("Warrior", nil)
		g.B.Bag = []content.Item{blade("Steel Sword", 8), blade("Rusty Blade", 3)}
		if _, err := g.Apply(action.Action{Kind: action.Equip, Unit: 1, Item: 0}); err != nil {
			t.Fatal(err)
		}
		if !g.B.Unit(1).Done {
			t.Fatal("the hero can still act")
		}
		if _, err := g.Apply(action.Action{Kind: action.Equip, Unit: 1, Item: 0}); code(err) != "finished" {
			t.Fatalf("second swap: %v", err)
		}
	})

	t.Run("A hero can't put on an item its class can't use", func(t *testing.T) {
		g := armedDuel("Warrior", nil)
		g.B.Bag = []content.Item{{Name: "Yew Bow", Slot: "w", Rarity: 1, WeaponType: "bow"}}
		if _, err := g.Apply(action.Action{Kind: action.Equip, Unit: 1, Item: 0}); code(err) != "cannot-use" {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("When every weapon slot is full, the player must choose which weapon to replace", func(t *testing.T) {
		g := armedDuel("Warrior", map[string]content.Item{"w": blade("Rusty Blade", 3), "w2": {Name: "Iron Axe", Slot: "w", WeaponType: "axe", Atk: 3}})
		g.B.Bag = []content.Item{blade("Steel Sword", 8)}
		if _, err := g.Apply(action.Action{Kind: action.Equip, Unit: 1, Item: 0}); code(err) != "choose-slot" {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("Dropping an item throws it away for good", func(t *testing.T) {
		g := duel()
		g.B.Bag = []content.Item{blade("Rusty Blade", 3)}
		if _, err := g.Apply(action.Action{Kind: action.Drop, Item: 0}); err != nil || len(g.B.Bag) != 0 {
			t.Fatalf("err %v, bag %d", err, len(g.B.Bag))
		}
	})
}

func TestWeaponSets(t *testing.T) {
	bowAndSword := func() *Game {
		g := armedDuel("Ranger", map[string]content.Item{"w": {Name: "Yew Bow", Slot: "w", Rarity: 1, WeaponType: "bow", Atk: 5}, "w2": blade("Steel Sword", 5)})
		g.B.Unit(2).X, g.B.Unit(2).Y = 2, 0
		return g
	}
	shot := func(g *Game) int {
		for i, s := range g.B.Unit(1).Skills {
			if s.Name == "Shot" {
				return i
			}
		}
		return -1
	}

	t.Run("A skill used with the wrong weapon in hand switches to the first weapon set it works with", func(t *testing.T) {
		g := bowAndSword()
		h := g.B.Unit(1)
		stats.UseSet(g.C, h, 1)
		if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: 1, Skill: shot(g), Target: 2}); err != nil {
			t.Fatal(err)
		}
		if h.WeaponSet != 0 {
			t.Fatalf("weapon set %d", h.WeaponSet+1)
		}
	})

	t.Run("Choosing a weapon set the skill can't be used with is refused", func(t *testing.T) {
		g := bowAndSword()
		if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: 1, Skill: shot(g), Target: 2, Set: 2}); code(err) != "wrong-weapon" {
			t.Fatalf("got %v", err)
		}
	})
}
