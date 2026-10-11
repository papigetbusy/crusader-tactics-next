package items

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// dice return these fractions, then 0.5 forever.
func dice(fs ...float64) *rng.Feed {
	var v []uint32
	for _, f := range fs {
		v = append(v, uint32(f*4294967296.0))
	}
	half := 0.5
	for range 60 {
		v = append(v, uint32(half*4294967296.0))
	}
	return rng.NewFeed(v)
}

func nop(action.Event) {}

// party is a fight on floor 20 with a Warrior and a Mage.
func party(level int) (*content.Content, *state.Battle, *state.Unit, *state.Unit) {
	c := content.MustDefault()
	w := &state.Unit{ID: 1, Team: state.Heroes, Class: "Warrior", Level: level, Gear: map[string]content.Item{}}
	m := &state.Unit{ID: 2, Team: state.Heroes, Class: "Mage", Level: level, Gear: map[string]content.Item{
		"h": {Name: "Cloth Hood", Slot: "h", Rarity: 2}, "a": {Name: "Cloth Tunic", Slot: "a", Rarity: 2}}}
	b := &state.Battle{Floor: 20, Units: []*state.Unit{w, m}, Party: []int{1, 2}}
	return c, b, w, m
}

func monster(boss, elite bool) *state.Unit {
	return &state.Unit{ID: 9, Team: state.Enemies, Class: "Warrior", Level: 20, Boss: boss, Elite: elite}
}

func TestLoot(t *testing.T) {
	t.Run("A normal monster drops an item 60% of the time, of the floor's rarity band", func(t *testing.T) {
		c, b, w, _ := party(20)
		Drop(c, b, dice(0.7), monster(false, false), w, nop)
		if len(b.Bag) != 0 {
			t.Fatal("dropped on a 0.7 roll")
		}
		Drop(c, b, dice(0.1, 0.1), monster(false, false), w, nop)
		if len(b.Bag) != 1 || b.Bag[0].Rarity != 1 {
			t.Fatalf("bag %v (floor 20 band: uncommon at 70%%)", b.Bag)
		}
	})

	t.Run("Legendaries only drop once a hero has reached Lv 30; before that a legendary roll gives a rare item", func(t *testing.T) {
		c, b, w, _ := party(25)
		Drop(c, b, dice(0.1), monster(true, false), w, nop)
		if len(b.Bag) != 1 || b.Bag[0].Rarity != 2 || b.BossLegendary {
			t.Fatalf("bag %v", b.Bag)
		}
		c, b, w, _ = party(30)
		Drop(c, b, dice(0.1), monster(true, false), w, nop)
		if len(b.Bag) != 1 || b.Bag[0].Rarity != 3 || !b.Bag[0].Unique {
			t.Fatalf("bag %v", b.Bag)
		}
	})

	t.Run("A boss drops a legendary 25% of the time, once per floor", func(t *testing.T) {
		c, b, w, _ := party(30)
		Drop(c, b, dice(0.1), monster(true, false), w, nop)
		Drop(c, b, dice(0.1, 0.1), monster(true, false), w, nop)
		if !b.BossLegendary || len(b.Bag) != 2 || b.Bag[1].Rarity == 3 {
			t.Fatalf("bag %v", b.Bag)
		}
	})

	t.Run("Loot goes to the hero who made the kill half the time, else to the least geared hero", func(t *testing.T) {
		c, b, w, m := party(20)
		// the Mage made the kill but the roll says "least geared": the Warrior wears nothing
		it, _ := lootFor(c, b, dice(0.9, 0.1, 0.1, 0.4), m, 1)
		if it.Slot == "w" && it.WeaponType != "sword" && it.WeaponType != "axe" {
			t.Fatalf("a %s for the Warrior?", it.WeaponType)
		}
		if it.Slot == "a" && it.Weight == "L" {
			t.Fatalf("light armor for the Warrior: %+v", it)
		}
		_ = w
	})

	t.Run("A party hero who has died still counts as the least geared hero for loot", func(t *testing.T) {
		c, b, w, m := party(20)
		b.Units = []*state.Unit{m}
		b.Lost = []*state.Unit{w} // died where no revive reaches him; still in the party
		it, ok := lootFor(c, b, dice(0.9), nil, 1)
		if !ok || it.Slot == "w" && it.WeaponType != "sword" && it.WeaponType != "axe" {
			t.Fatalf("%v %+v (the Warrior wears nothing, so the item is his)", ok, it)
		}
	})

	t.Run("Weapons only come in types the hero uses, and no rarer than its weapon rank allows", func(t *testing.T) {
		c, b, w, _ := party(1)
		for i := 0; i < 50; i++ {
			it := lootPick(c, b, rng.New(uint64(i), 1), w, 2)
			if it.Slot == "w" && (it.WeaponType != "sword" && it.WeaponType != "axe" || it.Rarity > 1) {
				t.Fatalf("a Lv 1 Warrior got %+v", it)
			}
		}
	})

	t.Run("A legendary weapon a hero can't use yet becomes a legendary armor piece or charm", func(t *testing.T) {
		c, b, w, _ := party(29)
		w.Gear = map[string]content.Item{"h": {}, "a": {}, "b": {}, "c": {}, "w2": {}} // only the sword slot is empty
		it := lootPick(c, b, dice(0.1, 0.1, 0.1, 0.1, 0.1), w, 3)
		if it.Slot == "w" || it.Rarity != 3 {
			t.Fatalf("got %+v", it)
		}
	})

	t.Run("An item's bonuses fit its slot, the class and the weapon", func(t *testing.T) {
		c, _, _, _ := party(20)
		for i := 0; i < 200; i++ {
			it := Make(c, 20, rng.New(uint64(i), 2), 2, "w", "sword", "", "Warrior")
			for _, a := range it.Affixes {
				if a.Key == "healPct" || a.Key == "evade" {
					t.Fatalf("a Warrior's sword rolled %s", a.Key)
				}
			}
			if len(it.Affixes) < 2 || len(it.Affixes) > 3 {
				t.Fatalf("a rare item has %d bonuses", len(it.Affixes))
			}
		}
	})

	t.Run("Opening a chest gives an item of the chest's rarity", func(t *testing.T) {
		c, b, w, _ := party(20)
		b.Chests = []state.Chest{{X: 0, Y: 0, Kind: "gold"}}
		OpenChest(c, b, dice(0.9), w, nop)
		if len(b.Chests) != 0 || len(b.Bag) != 1 || b.Bag[0].Rarity != 2 {
			t.Fatalf("chests %v bag %v", b.Chests, b.Bag)
		}
	})

	t.Run("A kill pays gold: 2 plus 0.6 per monster level, 3x for elites and 8x for bosses", func(t *testing.T) {
		c, b, w, _ := party(20)
		if g := Gold(c, b, w, monster(false, false)); g != 14 {
			t.Fatalf("normal: %d", g)
		}
		if g := Gold(c, b, w, monster(false, true)); g != 42 {
			t.Fatalf("elite: %d", g)
		}
		if g := Gold(c, b, w, monster(true, false)); g != 112 {
			t.Fatalf("boss: %d", g)
		}
		b.Replayed = true
		if g := Gold(c, b, w, monster(false, false)); g != 7 {
			t.Fatalf("replayed floor: %d", g)
		}
	})
}
