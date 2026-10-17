package items

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
)

func TestLegendaryRework(t *testing.T) {
	c := content.MustDefault()
	g := c.Gear

	t.Run("Items roll exactly 1, 2 or 3 bonuses by rarity, and an epic 3 plus its epic bonus", func(t *testing.T) {
		for seed := uint64(1); seed <= 40; seed++ {
			for k, want := range []int{1, 2, 3, 4} {
				it := Make(c, 30, rng.New(seed, 0), k, "w", "sword", "", "Warrior")
				if len(it.Affixes) != want {
					t.Fatalf("rarity %d: %d bonuses (%v)", k, len(it.Affixes), it.Affixes)
				}
			}
		}
	})

	t.Run("Every legendary has 5 bonuses, one more than an epic", func(t *testing.T) {
		for _, u := range g.Uniques {
			if len(u.Affixes) != 5 {
				t.Errorf("%s has %d", u.Name, len(u.Affixes))
			}
		}
	})

	t.Run("Every class has its own legendary weapon with 2 unique powers and a 4-piece bonus", func(t *testing.T) {
		powers := map[string]bool{}
		for _, a := range g.Affixes {
			powers[a.Key] = a.Power
		}
		have := map[string]bool{}
		for _, u := range g.Uniques {
			if u.Slot != content.SlotWeapon {
				continue
			}
			n := 0
			for _, a := range u.Affixes {
				if powers[a.Key] {
					n++
				}
			}
			if u.Class == "" || n != 2 || len(u.Full) != 1 || !powers[u.Full[0].Key] {
				t.Errorf("%s: class %q, %d powers, set of 4 %v", u.Name, u.Class, n, u.Full)
			}
			if have[u.Class] {
				t.Errorf("%s has two legendaries", u.Class)
			}
			have[u.Class] = true
		}
		if len(have) != len(g.Classes) {
			t.Fatalf("%d of %d classes have one", len(have), len(g.Classes))
		}
	})

	t.Run("A legendary weapon drops only for its own class, even when another class shares the weapon type", func(t *testing.T) {
		for class, cg := range g.Classes {
			for seed := uint64(1); seed <= 10; seed++ {
				for _, w := range cg.Weapons {
					it := Make(c, 30, rng.New(seed, 0), g.Legendary(), "w", w.Type, "", class)
					if it.Class != class || len(it.Full) == 0 {
						t.Fatalf("%s got %s (%s)", class, it.Name, it.Class)
					}
				}
			}
		}
	})

	t.Run("Epic armor belongs to the epic set of its weight", func(t *testing.T) {
		for _, w := range []string{"L", "M", "H"} {
			it := Make(c, 30, rng.New(3, 0), g.Epic(), "a", "", w, "Knight")
			if it.Set != "e"+w {
				t.Fatalf("%s armor: set %q", w, it.Set)
			}
		}
		if it := Make(c, 30, rng.New(3, 0), 2, "a", "", "H", "Knight"); it.Set != "" {
			t.Fatalf("rare armor is in set %q", it.Set)
		}
	})
}
