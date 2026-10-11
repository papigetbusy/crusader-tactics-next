package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Threat is where a unit could walk on its next turn, as if it had not
// moved yet (the old zoneOf, line 1847, used by the danger zone and by
// tapping a monster to see its range). A snared, locked or shocked unit
// stays where it is.
func (g *Game) Threat(unit int) []board.Tile {
	u := g.B.Unit(unit)
	if u == nil {
		return nil
	}
	moved, left := u.Moved, u.MoveLeft
	defer func() { u.Moved, u.MoveLeft = moved, left }()
	u.Moved, u.MoveLeft = false, 0
	if !board.CanMove(u) {
		return []board.Tile{{X: u.X, Y: u.Y}}
	}
	return board.Reach(g.C, g.B, u)
}

// inRange is every tile within reach of a unit's ready attacking skills if
// it stood on one of the given tiles (the old atkSet and zoneOf, lines
// 1846-1847): red tiles on the screen.
func (g *Game) inRange(u *state.Unit, from []board.Tile) []board.Tile {
	seen := map[board.Tile]bool{}
	var out []board.Tile
	for i, s := range u.Skills {
		if u.Cooldowns[i] > 0 || s.Friendly() {
			continue
		}
		lo, hi := combat.SkillRange(g.C, u, s)
		if hi <= 0 {
			continue
		}
		hi = min(hi, 2*g.B.N)
		for _, f := range from {
			for dy := -hi; dy <= hi; dy++ {
				for dx := -hi; dx <= hi; dx++ {
					d := abs(dx) + abs(dy)
					t := board.Tile{X: f.X + dx, Y: f.Y + dy}
					if d < max(lo, 1) || d > hi || !g.B.Inside(t.X, t.Y) || seen[t] {
						continue
					}
					seen[t] = true
					out = append(out, t)
				}
			}
		}
	}
	return out
}

// AttackTiles is where a hero could hit from where it stands now (red tiles
// after it moved).
func (g *Game) AttackTiles(unit int) []board.Tile {
	u := g.B.Unit(unit)
	if u == nil || u.Done {
		return nil
	}
	return g.inRange(u, []board.Tile{{X: u.X, Y: u.Y}})
}

// ThreatHit is where a monster could hit next turn, from anywhere it could
// walk (red around its purple move zone).
func (g *Game) ThreatHit(unit int) []board.Tile {
	u := g.B.Unit(unit)
	if u == nil {
		return nil
	}
	return g.inRange(u, g.Threat(unit))
}
