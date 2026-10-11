package items

import (
	"slices"
	"strings"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
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

	t.Run("Legendaries only drop once a hero has reached Lv 30; before that a legendary roll gives a purple item", func(t *testing.T) {
		c, b, w, _ := party(25)
		Drop(c, b, dice(0.1), monster(true, false), w, nop)
		if len(b.Bag) != 1 || b.Bag[0].Rarity != c.Gear.Epic() || b.BossLegendary {
			t.Fatalf("bag %v", b.Bag)
		}
		c, b, w, _ = party(30)
		Drop(c, b, dice(0.1), monster(true, false), w, nop)
		if len(b.Bag) != 1 || b.Bag[0].Rarity != c.Gear.Legendary() || !b.Bag[0].Unique {
			t.Fatalf("bag %v", b.Bag)
		}
	})

	t.Run("A boss drops a legendary 25% of the time, once per floor", func(t *testing.T) {
		c, b, w, _ := party(30)
		Drop(c, b, dice(0.1), monster(true, false), w, nop)
		Drop(c, b, dice(0.1, 0.1), monster(true, false), w, nop)
		if !b.BossLegendary || len(b.Bag) != 2 || b.Bag[1].Rarity == c.Gear.Legendary() {
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
		it := lootPick(c, b, dice(0.1, 0.1, 0.1, 0.1, 0.1), w, c.Gear.Legendary())
		if it.Slot == "w" || it.Rarity != c.Gear.Legendary() {
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

	t.Run("A gold chest gives a rare item 65% of the time, a purple one 30% and a legendary 5% (purple until a hero is Lv 30)", func(t *testing.T) {
		for _, tc := range []struct {
			roll  float64
			level int
			want  int
		}{{0.5, 20, 2}, {0.9, 20, 3}, {0.97, 20, 3}, {0.97, 30, 4}} {
			c, b, w, _ := party(tc.level)
			b.Chests = []state.Chest{{X: 0, Y: 0, Kind: "gold"}}
			OpenChest(c, b, dice(tc.roll), w, nop)
			if len(b.Chests) != 0 || len(b.Bag) != 1 || b.Bag[0].Rarity != tc.want {
				t.Fatalf("roll %v at Lv %d: chests %v bag %v", tc.roll, tc.level, b.Chests, b.Bag)
			}
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

	t.Run("Purple first drops on floors 21-30, 10% of the time", func(t *testing.T) {
		c := content.MustDefault()
		if Band(c, 20, dice(0.95)) != 2 || Band(c, 25, dice(0.95)) != c.Gear.Epic() || Band(c, 25, dice(0.6)) != 2 {
			t.Fatal("wrong floor bands")
		}
	})

	t.Run("A purple item has one purple bonus, named in front of the item, plus 2 or 3 ordinary bonuses", func(t *testing.T) {
		c, _, _, _ := party(25)
		for i := 0; i < 200; i++ {
			it := Make(c, 25, rng.New(uint64(i), 3), c.Gear.Epic(), "w", "sword", "", "Warrior")
			pre := ""
			for _, p := range c.Gear.Prefixes {
				if strings.HasPrefix(it.Name, p.Name+" ") && p.Bonuses[0].Key == it.Affixes[0].Key {
					pre = p.Name
				}
			}
			if pre == "" || !strings.HasSuffix(it.Name, "Gravewarden Blade") || len(it.Affixes) < 3 || len(it.Affixes) > 4 {
				t.Fatalf("%+v", it)
			}
			seen := map[string]bool{}
			for _, a := range it.Affixes {
				if seen[a.Key] {
					t.Fatalf("%s twice on %s", a.Key, it.Name)
				}
				seen[a.Key] = true
			}
		}
	})

	t.Run("A purple weapon needs weapon rank A: a Lv 17 hero gets a rare one instead", func(t *testing.T) {
		c, b, w, _ := party(17)
		w.Gear = map[string]content.Item{"h": {}, "a": {}, "b": {}, "c": {}, "w2": {}}
		for i := 0; i < 30; i++ {
			if it := lootPick(c, b, rng.New(uint64(i), 4), w, c.Gear.Epic()); it.Slot == "w" && it.Rarity > 2 {
				t.Fatalf("a Lv 17 Warrior got %+v", it)
			}
		}
		w.Level = 18
		if ok, _ := stats.CanUse(c, "Warrior", 18, Make(c, 25, dice(), c.Gear.Epic(), "w", "sword", "", "Warrior")); !ok {
			t.Fatal("a Lv 18 Warrior can't hold a purple sword")
		}
	})

	t.Run("Purple charms and armor have their own names", func(t *testing.T) {
		c, _, _, _ := party(25)
		if n := Make(c, 25, dice(0.1), c.Gear.Epic(), "a", "", "H", "Warrior").Name; !strings.HasSuffix(n, "Cathedral Plate") {
			t.Fatal(n)
		}
		if n := Make(c, 25, dice(0.1), 2, "c", "", "", "Warrior").Name; slices.Contains(c.Gear.EpicCharms, n) {
			t.Fatalf("a rare charm called %s", n)
		}
	})
}
