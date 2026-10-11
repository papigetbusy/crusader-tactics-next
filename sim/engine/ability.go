package engine

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// checkAbility is "may this hero use its class ability now?" (old
// useAbility, lines 1193-1195, and its target picking, line 1759).
func (g *Game) checkAbility(u *state.Unit, a action.Action) error {
	b := g.B
	ab, ok := combat.AbilityOf(g.C, u)
	if !ok {
		return no("no-ability", "%s has no class ability.", u.Class)
	}
	if u.Done && !ab.Free {
		return no("finished", "%s has finished for this turn.", u.Class)
	}
	if u.AbilityCooldown > 0 && !ab.Free {
		return no("on-cooldown", "%s is not ready yet (%d turns).", ab.Name, u.AbilityCooldown)
	}
	if ab.Still && u.Moved {
		return no("moved", "%s only works before you move this turn.", ab.Name)
	}
	t := ab.Target
	switch {
	case t == nil:
	case t.Kind == "unit":
		if !slices.Contains(combat.AbilityUnits(b, u, t), b.Unit(a.Target)) {
			return no("bad-target", "%s cannot be aimed at that.", ab.Name)
		}
	default:
		n := max(1, t.N)
		if len(a.Tiles) != n {
			return no("bad-target", "%s needs %d tile(s).", ab.Name, n)
		}
		for i, p := range a.Tiles {
			if !combat.AbilityTileOK(b, u, t, p[0], p[1]) || slices.Contains(a.Tiles[:i], p) {
				return no("bad-target", "%s cannot be aimed there.", ab.Name)
			}
		}
	}
	return nil
}

// skillReady is the old skAvail rule (line 2050): an active Cloak cannot be
// used again, and Quick Draw's shots must be different skills unless the
// upgraded version allows repeats.
func skillReady(u *state.Unit, slot int) error {
	s := u.Skills[slot]
	if s.Cloak && (u.Status["cloak"] > 0 || u.Cooldowns[slot] >= 99) {
		return no("cloaked", "Cloak is already active.")
	}
	if u.Status["qd"] > 0 && u.P("qdSame", 0) == 0 && slices.Contains(u.QuickUsed, s.Name) {
		return no("used", "Quick Draw: pick a different skill for this shot.")
	}
	return nil
}
