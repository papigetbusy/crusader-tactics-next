package engine

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// setFor is the weapon set a hero uses for a skill: the one asked for (1-3),
// or the one in hand, or else the first set the skill works with (the old
// screen switched sets the same way: setOk and tgtFor, index.html lines 1114,
// 1826 and 2051).
func (g *Game) setFor(u *state.Unit, s content.Skill, want int) (int, error) {
	if !u.Armed {
		return u.WeaponSet, nil
	}
	sets := stats.SetsOf(g.C, u.Class)
	if want > 0 {
		if want > len(sets) {
			return 0, no("no-such-set", "%s has no weapon set %d.", u.Class, want)
		}
		if !stats.SetFits(g.C, u, s, want-1) {
			return 0, no("wrong-weapon", "%s needs a %s.", s.Name, categoryName(stats.ClassCategory(g.C, u.Class)))
		}
		return want - 1, nil
	}
	if stats.SetFits(g.C, u, s, u.WeaponSet) {
		return u.WeaponSet, nil
	}
	for i := range sets {
		if stats.SetFits(g.C, u, s, i) {
			return i, nil
		}
	}
	return 0, no("wrong-weapon", "%s needs a %s.", s.Name, categoryName(stats.ClassCategory(g.C, u.Class)))
}

func categoryName(cat string) string {
	return map[string]string{"m": "melee weapon", "r": "bow or gun", "g": "staff or wand"}[cat]
}

// inSet runs fn with set i in the hero's hand, then puts the old set back.
func (g *Game) inSet(u *state.Unit, i int, fn func()) {
	keep := u.WeaponSet
	stats.UseSet(g.C, u, i)
	fn()
	stats.UseSet(g.C, u, keep)
}

// checkGear is "may this hero put on that item?" and "may that item be
// dropped?" (old equip and the drop button, index.html lines 1786 and 1657).
func (g *Game) checkGear(a action.Action) error {
	b := g.B
	if a.Item < 0 || a.Item >= len(b.Bag) {
		return no("no-such-item", "There is no such item in the bag.")
	}
	if a.Kind == action.Drop {
		return nil
	}
	u := b.Unit(a.Unit)
	if u == nil || u.Team != state.Heroes || u.Companion || u.Decoy || u.Summoned {
		return no("no-such-hero", "That is not one of your heroes.")
	}
	if u.Done && g.C.Combat.EquipEndsTurn {
		return no("finished", "%s has finished for this turn.", u.Class)
	}
	it := b.Bag[a.Item]
	if ok, why := stats.CanUse(g.C, u.Class, u.Level, it); !ok {
		return no("cannot-use", "%s", why)
	}
	if it.Slot == content.SlotWeapon {
		slots := stats.WeaponSlotsOf(g.C, u.Class)
		if a.Slot != "" && !slices.Contains(slots, a.Slot) {
			return no("no-such-slot", "%s has no weapon slot %q.", u.Class, a.Slot)
		}
		if a.Slot == "" && freeWeaponSlot(u, slots) == "" {
			return no("choose-slot", "Every weapon slot is full: choose which weapon to replace.")
		}
	}
	return nil
}

func freeWeaponSlot(u *state.Unit, slots []string) string {
	for _, sl := range slots {
		if _, ok := u.Gear[sl]; !ok {
			return sl
		}
	}
	return ""
}

// equip puts a bag item on a hero; what it replaces goes back in the bag,
// and the hero's numbers are worked out again (old equip, line 1786).
func (g *Game) equip(a action.Action, emit action.Emit) {
	b := g.B
	u := b.Unit(a.Unit)
	it := b.Bag[a.Item]
	sl := it.Slot
	if it.Slot == content.SlotWeapon {
		sl = a.Slot
		if sl == "" {
			sl = freeWeaponSlot(u, stats.WeaponSlotsOf(g.C, u.Class))
		}
	}
	b.Bag = append(b.Bag[:a.Item:a.Item], b.Bag[a.Item+1:]...)
	if u.Gear == nil {
		u.Gear = map[string]content.Item{}
	}
	if old, ok := u.Gear[sl]; ok {
		b.Bag = append(b.Bag, old)
	}
	u.Gear[sl] = it
	stats.Refresh(g.C, b, u)
	emit(action.Event{Kind: action.GearOn, Unit: u.ID, Skill: it.Name})
	if g.C.Combat.EquipEndsTurn {
		u.Done = true
		combat.PairDone(b, u)
		emit(action.Event{Kind: action.Finished, Unit: u.ID})
	}
}

// checkSet is "may this hero take that weapon set in hand?" (the old ⇄
// button, index.html line 1548): a hero with more than one set, not
// finished for the turn.
func (g *Game) checkSet(a action.Action) error {
	u := g.B.Unit(a.Unit)
	if u == nil || u.Team != state.Heroes || !u.Armed {
		return no("no-such-hero", "That is not one of your heroes.")
	}
	if u.Done {
		return no("finished", "%s has finished for this turn.", u.Class)
	}
	if a.Set < 1 || a.Set > len(stats.SetsOf(g.C, u.Class)) {
		return no("no-such-set", "%s has no weapon set %d.", u.Class, a.Set)
	}
	return nil
}
