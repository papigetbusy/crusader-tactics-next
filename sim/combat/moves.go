package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Moves of the even-schedule skills: shoves, pulls, leaps, teleports and
// swaps. A unit is only ever put on an open tile it could stand on.

// Open reports a tile z could stand on: on the map, empty (or z itself) and
// ground z can enter.
func Open(c *content.Content, b *state.Battle, z *state.Unit, x, y int) bool {
	if !b.Inside(x, y) {
		return false
	}
	if o := b.UnitAt(x, y); o != nil && o != z {
		return false
	}
	return board.StepCost(c, b, z, x, y) < board.Blocked
}

// direction is the straight step from a toward b along the longer axis
// (ties go along x), as piercing skills use.
func direction(ax, ay, bx, by int) (int, int) {
	if abs(bx-ax) >= abs(by-ay) {
		return sign(bx - ax), 0
	}
	return 0, sign(by - ay)
}

// place moves z to (x, y) at once (a jump, a teleport or a shove) and tells
// the screen the tiles it went through.
func place(z *state.Unit, path [][2]int, emit action.Emit) {
	if len(path) < 2 {
		return
	}
	z.X, z.Y = path[len(path)-1][0], path[len(path)-1][1]
	emit(action.Event{Kind: action.Moved, Unit: z.ID, Path: path})
}

// Blocked reports whether shoving v n tiles away from (fx, fy) would stop
// early against a wall, the map's edge or another unit.
func shoveBlocked(c *content.Content, b *state.Battle, fx, fy int, v *state.Unit, n int) bool {
	dx, dy := direction(fx, fy, v.X, v.Y)
	for i := 1; i <= n; i++ {
		if !Open(c, b, v, v.X+dx*i, v.Y+dy*i) {
			return true
		}
	}
	return false
}

// shove pushes v up to n tiles straight away from (fx, fy), stopping at the
// first tile it cannot enter. It reports whether it was stopped.
func shove(c *content.Content, b *state.Battle, fx, fy int, v *state.Unit, n int, emit action.Emit) bool {
	dx, dy := direction(fx, fy, v.X, v.Y)
	path := [][2]int{{v.X, v.Y}}
	stopped := false
	for i := 0; i < n; i++ {
		x, y := path[len(path)-1][0]+dx, path[len(path)-1][1]+dy
		if !Open(c, b, v, x, y) {
			stopped = true
			break
		}
		path = append(path, [2]int{x, y})
	}
	place(v, path, emit)
	return stopped
}

// drag moves v step by step toward (tx, ty) until it is next to it, it has
// taken max steps (0 = no limit), or it is stuck. Each step goes along the
// longer axis, or the other one if that is blocked.
func drag(c *content.Content, b *state.Battle, v *state.Unit, tx, ty, steps int, emit action.Emit) {
	path := [][2]int{{v.X, v.Y}}
	for i := 0; steps == 0 || i < steps; i++ {
		x, y := path[len(path)-1][0], path[len(path)-1][1]
		if abs(tx-x)+abs(ty-y) <= 1 {
			break
		}
		dx, dy := direction(x, y, tx, ty)
		if !Open(c, b, v, x+dx, y+dy) {
			if dx != 0 {
				dx, dy = 0, sign(ty-y)
			} else {
				dx, dy = sign(tx-x), 0
			}
			if dx == 0 && dy == 0 || !Open(c, b, v, x+dx, y+dy) {
				break
			}
		}
		path = append(path, [2]int{x + dx, y + dy})
	}
	place(v, path, emit)
}

// landing is the open tile next to v that u can jump to within n tiles,
// nearest to u first (ties in board order); ok is false if there is none.
func landing(c *content.Content, b *state.Battle, u, v *state.Unit, n int) (state.Tile, bool) {
	best, bd, ok := state.Tile{}, 1<<30, false
	for _, d := range [4][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}} {
		x, y := v.X+d[0], v.Y+d[1]
		dist := abs(x-u.X) + abs(y-u.Y)
		if dist <= n && dist < bd && (dist == 0 || Open(c, b, u, x, y)) {
			best, bd, ok = state.Tile{X: x, Y: y}, dist, true
		}
	}
	return best, ok
}

// behindOf is the tile on the far side of v from u.
func behindOf(u, v *state.Unit) state.Tile {
	dx, dy := direction(u.X, u.Y, v.X, v.Y)
	return state.Tile{X: v.X + dx, Y: v.Y + dy}
}

// jump moves u straight to a tile (no walking: leaps and teleports).
func jump(u *state.Unit, t state.Tile, emit action.Emit) {
	place(u, [][2]int{{u.X, u.Y}, {t.X, t.Y}}, emit)
}

// swap trades the places of two units.
func swap(u, z *state.Unit, emit action.Emit) {
	ux, uy := u.X, u.Y
	jump(u, state.Tile{X: z.X, Y: z.Y}, func(action.Event) {})
	z.X, z.Y = ux, uy
	emit(action.Event{Kind: action.Moved, Unit: u.ID, Path: [][2]int{{z.X, z.Y}, {u.X, u.Y}}})
	emit(action.Event{Kind: action.Moved, Unit: z.ID, Path: [][2]int{{u.X, u.Y}, {z.X, z.Y}}})
}

// lineTiles are the tiles from u to (x, y), not counting u's own, if (x, y)
// is in a straight line no more than n tiles away; nil otherwise.
func lineTiles(u *state.Unit, x, y, n int) []state.Tile {
	d := abs(x-u.X) + abs(y-u.Y)
	if d == 0 || d > n || (x != u.X && y != u.Y) {
		return nil
	}
	dx, dy := sign(x-u.X), sign(y-u.Y)
	var out []state.Tile
	for i := 1; i <= d; i++ {
		out = append(out, state.Tile{X: u.X + dx*i, Y: u.Y + dy*i})
	}
	return out
}

// sideStep shoves z one tile to the side of a line going (dx, dy): left
// first, then right. ok is false when both sides are taken.
func sideStep(c *content.Content, b *state.Battle, z *state.Unit, dx, dy int, emit action.Emit) bool {
	for _, s := range [2][2]int{{-dy, dx}, {dy, -dx}} {
		if Open(c, b, z, z.X+s[0], z.Y+s[1]) {
			jump(z, state.Tile{X: z.X + s[0], Y: z.Y + s[1]}, emit)
			return true
		}
	}
	return false
}
