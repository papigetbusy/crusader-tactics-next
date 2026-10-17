package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/run"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestOldSavesGetEpicSets(t *testing.T) {
	t.Run("Epic armor from a save made before epic sets joins the epic set of its weight when loaded", func(t *testing.T) {
		c := content.MustDefault()
		e := c.Gear.Epic()
		helm := content.Item{Name: "Vigil Greathelm", Slot: "h", Rarity: e, Weight: "H"}
		sword := content.Item{Name: "Gravewarden Blade", Slot: "w", Rarity: e, WeaponType: "sword"}
		rare := content.Item{Name: "Steel Helm", Slot: "h", Rarity: 2, Weight: "H"}
		s := saved{Run: &run.Run{Bag: []content.Item{helm, sword, rare},
			Heroes: []*run.Hero{{Class: "Knight", Gear: map[string]content.Item{"h": helm}}}},
			Battle: &state.Battle{Bag: []content.Item{helm}}}
		upgradeItems(c, &s)
		if s.Run.Bag[0].Set != "eH" || s.Run.Bag[1].Set != "" || s.Run.Bag[2].Set != "" {
			t.Fatalf("bag sets: %q %q %q", s.Run.Bag[0].Set, s.Run.Bag[1].Set, s.Run.Bag[2].Set)
		}
		if s.Run.Heroes[0].Gear["h"].Set != "eH" || s.Battle.Bag[0].Set != "eH" {
			t.Fatal("worn or battle gear was not upgraded")
		}
	})
}
