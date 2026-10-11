package engine

import "github.com/papigetbusy/crusader-tactics-next/sim/board"

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
