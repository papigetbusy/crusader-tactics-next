package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// BagFit is how one bag item suits a hero: can it be worn, and how its
// numbers compare with what the hero wears in that slot now (the old gear
// sheet's "vs worn" line, index.html line 1654).
type BagFit struct {
	OK   bool           `json:"ok"`
	Why  string         `json:"why,omitempty"`  // why it can't be worn
	Worn bool           `json:"worn,omitempty"` // the hero wears something there now
	Diff map[string]int `json:"diff,omitempty"` // atk, def, hp, mov: this item minus the worn one
}

// bagFits is BagFit for every bag item, for one hero.
func (g *Game) bagFits(u *state.Unit) []BagFit {
	var out []BagFit
	for _, it := range g.B.Bag {
		f := BagFit{}
		f.OK, f.Why = stats.CanUse(g.C, u.Class, u.Level, it)
		var worn content.Item
		slot := it.Slot
		if slot == content.SlotWeapon {
			slot = stats.WeaponSlotsOf(g.C, u.Class)[0]
		}
		worn, f.Worn = u.Gear[slot]
		if f.OK {
			f.Diff = map[string]int{}
			for k, d := range map[string]int{"atk": it.Atk - worn.Atk, "def": it.Def - worn.Def, "hp": it.HP - worn.HP, "mov": it.Mov - worn.Mov} {
				if d != 0 {
					f.Diff[k] = d
				}
			}
		}
		out = append(out, f)
	}
	return out
}
