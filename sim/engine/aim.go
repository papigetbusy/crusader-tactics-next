package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// pick turns an action's target, second target and tiles into a combat.Pick.
func (g *Game) pick(a action.Action) combat.Pick {
	p := combat.Pick{Target: g.B.Unit(a.Target), Second: g.B.Unit(a.Second)}
	for _, t := range a.Tiles {
		p.Tiles = append(p.Tiles, state.Tile{X: t[0], Y: t[1]})
	}
	return p
}

// checkNew is "may this hero use this even-schedule skill like this?":
// Garrote's silence, Taunting Roar, Steady Aim's standing still, and what
// was picked.
func (g *Game) checkNew(u *state.Unit, s content.Skill, a action.Action) error {
	p := g.pick(a)
	if u.Status["silence"] > 0 && !s.Strike {
		return no("silenced", "%s is silenced: only Strike can be used.", u.Class)
	}
	if t := g.B.Unit(u.TauntedBy); u.Status["taunted"] > 0 && t != nil && !s.Friendly() && p.Target != t {
		return no("taunted", "%s was taunted and must attack the %s.", u.Class, t.Class)
	}
	if s.Still && u.Moved {
		return no("moved", "%s only works before you move this turn.", s.Name)
	}
	if !combat.PickOK(g.C, g.B, u, s, p) {
		return no("bad-target", "%s cannot be aimed like that.", s.Name)
	}
	return nil
}

// SkillAim is the next thing to pick for a skill, given the picks so far
// (target, second and tiles as in the action): unit IDs, or tiles. Skills
// of the old game always ask for one unit.
func (g *Game) SkillAim(unit, slot int, so action.Action) (units []int, tiles [][2]int) {
	u := g.B.Unit(unit)
	if u == nil || slot < 0 || slot >= len(u.Skills) || u.Cooldowns[slot] > 0 {
		return nil, nil
	}
	set, err := g.setFor(u, u.Skills[slot], so.Set)
	if err != nil {
		return nil, nil
	}
	g.inSet(u, set, func() { units, tiles = g.skillAim(u, slot, so) })
	return units, tiles
}

func (g *Game) skillAim(u *state.Unit, slot int, so action.Action) (units []int, tiles [][2]int) {
	s, p := u.Skills[slot], g.pick(so)
	ids := func(zs []*state.Unit) []int {
		var out []int
		for _, z := range zs {
			if z != p.Target {
				out = append(out, z.ID)
			}
		}
		return out
	}
	tl := func(ts []state.Tile) [][2]int {
		var out [][2]int
		for _, t := range ts {
			out = append(out, [2]int{t.X, t.Y})
		}
		return out
	}
	switch s.Aim {
	case content.AimTile, content.AimLine:
		return nil, tl(combat.AimTiles(g.C, g.B, u, s, p))
	case content.AimRoll:
		if len(p.Tiles) == 0 {
			return nil, tl(combat.AimTiles(g.C, g.B, u, s, p))
		}
		return ids(combat.RollTargets(g.C, g.B, u, s, p.Tiles[0])), nil
	case content.AimShove:
		if p.Target != nil {
			return nil, tl(combat.AimTiles(g.C, g.B, u, s, p))
		}
	}
	return ids(combat.Targets(g.C, g.B, u, s)), nil
}
