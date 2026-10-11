package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// bodyguard sends a hit aimed at a guarded unit to its Bodyguard, for 70% of
// the damage; a Guardian next to the bodyguard does not take it further.
func bodyguard(b *state.Battle, a Attack, emit action.Emit) Attack {
	v := a.Target
	if v.Status["guarded"] <= 0 || a.NoRedirect {
		return a
	}
	g := b.Unit(v.GuardedBy)
	if g == nil || g == v || g.Team != v.Team {
		return a
	}
	emit(action.Event{Kind: action.Guarded, Unit: g.ID, Target: v.ID})
	a.Target, a.Extra, a.NoRedirect = g, extraOf(a)*0.7, true
	return a
}

// strikeBack is what a hit sets off on the even-schedule skills' side: a
// Riposte Stance striking back at a melee attacker (power 0.8; a Duelist
// already does its own), Dragon Scales throwing back 20% of a melee hit,
// and a Voodoo Link passing half the damage to the linked unit.
func strikeBack(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, v *state.Unit, dm int, emit action.Emit) {
	u := a.By
	if a.Counter || b.Unit(u.ID) == nil {
		return
	}
	melee := Distance(u, v) == 1
	if v.HP > 0 && v.Status["riposte"] > 0 && v.Ascension != "Duelist" && melee {
		emit(action.Event{Kind: action.Riposte, Unit: v.ID, Target: u.ID})
		Hurt(c, b, dice, Attack{By: v, Skill: c.Skills["Slash"], Target: u, Counter: true, NoRedirect: true, Extra: 0.8}, emit)
	}
	if v.Status["dscale"] > 0 && melee && dm > 0 && b.Unit(u.ID) != nil {
		Hurt(c, b, dice, Attack{By: v, Skill: content.Skill{Name: "Dragon Scales"}, Target: u, Fixed: max(1, int(Round(float64(dm)*0.2))), Counter: true, NoRedirect: true}, emit)
	}
	if p := b.Unit(v.LinkedTo); v.Status["link"] > 0 && p != nil && p != v && p.Status["link"] > 0 && dm > 1 {
		Hurt(c, b, dice, Attack{By: u, Skill: content.Skill{Name: "Voodoo Link"}, Target: p, Fixed: dm / 2, Counter: true, NoRedirect: true}, emit)
	}
}
