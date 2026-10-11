package engine

import (
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Plan answers the tap-to-attack question: to use a skill on that unit, which
// skill and, if needed, which tile to walk to first. It copies the old game's
// shortcut (findSk and the auto-walk, index.html lines 2053-2067): the first
// ready skill on the bar that can reach the target; if none can, the nearest
// tile (by straight steps) from which one can.
type Plan struct {
	Slot int         `json:"slot"`
	Walk *board.Tile `json:"walk,omitempty"` // nil: attack from where you stand
}

// offensive skills are aimed at enemies; heals, buffs and revives at allies.
func offensive(s content.Skill) bool {
	return !s.Friendly()
}

// oneTap reports skills used by tapping a single unit (not tiles, lines,
// two units or a roll).
func oneTap(s content.Skill) bool {
	switch s.Aim {
	case content.AimUnit, content.AimAlly, content.AimAny, content.AimEcho:
		return true
	}
	return false
}

func (g *Game) firstSkillFor(u, v *state.Unit) int {
	foe := v.Team != u.Team
	for i, s := range u.Skills {
		if u.Cooldowns[i] > 0 || (s.Cloak && u.Status["cloak"] > 0) || offensive(s) != foe || !oneTap(s) || s.Free {
			continue
		}
		for _, t := range combat.Targets(g.C, g.B, u, s) {
			if t == v {
				return i
			}
		}
	}
	return -1
}

// AttackPlan finds the skill and walk for a hero to use on the target unit.
func (g *Game) AttackPlan(unit, target int) (Plan, bool) {
	u, v := g.B.Unit(unit), g.B.Unit(target)
	if u == nil || v == nil || u.Team != state.Heroes || u.Done || u.Acted || g.B.Outcome != state.Ongoing {
		return Plan{}, false
	}
	if s := g.firstSkillFor(u, v); s >= 0 {
		return Plan{Slot: s}, true
	}
	if u.Moved || !board.CanMove(u) {
		return Plan{}, false
	}
	ox, oy := u.X, u.Y
	tiles := board.Reach(g.C, g.B, u)
	sort.SliceStable(tiles, func(i, j int) bool {
		return abs(tiles[i].X-ox)+abs(tiles[i].Y-oy) < abs(tiles[j].X-ox)+abs(tiles[j].Y-oy)
	})
	defer func() { u.X, u.Y = ox, oy }()
	for _, t := range tiles {
		if (t.X == ox && t.Y == oy) || g.B.UnitAt(t.X, t.Y) != nil {
			continue
		}
		u.X, u.Y = t.X, t.Y
		if s := g.firstSkillFor(u, v); s >= 0 {
			walk := t
			return Plan{Slot: s, Walk: &walk}, true
		}
	}
	return Plan{}, false
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
