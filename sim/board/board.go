// Package board answers "where can this unit walk?". It copies the old game's
// movement rules exactly (index.html lines 1216-1230).
package board

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Blocked is the cost of a tile nobody can enter.
const Blocked = 1_000_000_000

// Tile is a board position.
type Tile = state.Tile

// GroundAt returns the terrain letter at (x, y). Where the map has no letter,
// the old game used the height: below 0 is water, otherwise plains.
func GroundAt(b *state.Battle, x, y int) byte {
	if y >= 0 && y < len(b.Ground) && x >= 0 && x < len(b.Ground[y]) {
		return b.Ground[y][x]
	}
	if y >= 0 && y < len(b.Heights) && x >= 0 && x < len(b.Heights[y]) && b.Heights[y][x] < 0 {
		return 'w'
	}
	return 'g'
}

// StepCost is what entering (x, y) costs this unit, in half-steps.
func StepCost(c *content.Content, b *state.Battle, u *state.Unit, x, y int) int {
	g := GroundAt(b, x, y)
	fly := c.Flying[u.Class]
	if c.Blocked[g] {
		if fly {
			return c.FlyingCost
		}
		return Blocked
	}
	if fly {
		return c.FlyingCost
	}
	if g == 'f' && c.Mounted[u.Class] {
		return c.MountedForestCost
	}
	t, ok := c.Terrain[g]
	if !ok {
		t = c.Terrain['g']
	}
	return t.MoveCost
}

// CanMove reports whether a unit may still move this turn: not yet moved,
// and not snared, locked or shocked.
func CanMove(u *state.Unit) bool {
	return !u.Moved && u.Status["snare"] <= 0 && u.Status["lock"] <= 0 && u.Status["shock"] <= 0
}

// Budget is how far a unit may walk this turn, in half-steps (old
// moveBudget, line 1225): its move, plus haste, Bloodlust, Beast Mode, Vow of
// Honor, and standing in a War Banner.
func Budget(b *state.Battle, u *state.Unit) int {
	mov := u.Mov
	if u.Status["haste"] > 0 {
		mov += 2
	}
	if u.Status["blood"] > 0 {
		mov += int(u.P("blM", 3))
	}
	if u.Status["beast"] > 0 {
		mov += int(u.P("bmM", 7))
	}
	if u.Status["vow"] > 0 {
		mov += int(u.P("vwM", 5))
	}
	if inBanner(b, u) {
		mov++
	}
	return 2 * mov
}

// inBanner reports a hero standing inside a War Banner.
func inBanner(b *state.Battle, u *state.Unit) bool {
	if u.Team != state.Heroes {
		return false
	}
	for _, f := range b.Fields {
		if f.Kind == "banner" && abs(f.X-u.X)+abs(f.Y-u.Y) <= f.Range {
			return true
		}
	}
	return false
}

// MoveUnit walks a unit to (x, y) and returns the route. It counts the tiles
// walked this turn (charge skills use it), and a hero in an upgraded Beast
// Mode sets the tile it lands on ablaze (old startMove, line 788). It does
// not check the move is allowed; callers do that first. Walking zero tiles
// does nothing.
func MoveUnit(c *content.Content, b *state.Battle, u *state.Unit, x, y int, emit action.Emit) []Tile {
	n := abs(u.X-x) + abs(u.Y-y)
	if n == 0 {
		return nil
	}
	if u.Team == state.Heroes && u.Status["beast"] > 0 && u.P("bmT", 0) != 0 {
		b.Fields = append(b.Fields, state.Field{X: x, Y: y, Turns: 2, Kind: "fire", Owner: u.ID})
	}
	path := Path(c, b, u, x, y)
	if len(path) < 2 {
		path = []Tile{{X: u.X, Y: u.Y}, {X: x, Y: y}}
	}
	u.TilesMoved += max(n, len(path)-1)
	u.X, u.Y = x, y
	if emit != nil {
		ev := action.Event{Kind: action.Moved, Unit: u.ID}
		for _, t := range path {
			ev.Path = append(ev.Path, [2]int{t.X, t.Y})
		}
		emit(ev)
	}
	return path
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// enemyBlocks is the old rule: you can pass through allies but not enemies.
func enemyBlocks(b *state.Battle, u *state.Unit) func(x, y int) bool {
	return func(x, y int) bool {
		o := b.UnitAt(x, y)
		return o != nil && o != u && o.Team != u.Team
	}
}

// Reach lists every tile this unit can walk to and stop on, in the old game's
// order (row by row). It includes the tile the unit stands on.
func Reach(c *content.Content, b *state.Battle, u *state.Unit) []Tile {
	if u.Status["vamp"] > 0 { // Blood Pact: any open tile (old bfs, line 1226)
		var out []Tile
		for y := 0; y < b.N; y++ {
			for x := 0; x < b.N; x++ {
				if o := b.UnitAt(x, y); o == nil || o == u {
					out = append(out, Tile{X: x, Y: y})
				}
			}
		}
		return out
	}
	cost := func(x, y int) int { return StepCost(c, b, u, x, y) }
	r := search(b, []Tile{{X: u.X, Y: u.Y}}, cost, Budget(b, u), enemyBlocks(b, u))
	var out []Tile
	for k, d := range r.dist {
		if d >= Blocked {
			continue
		}
		x, y := k%b.N, k/b.N
		if o := b.UnitAt(x, y); o == nil || o == u {
			out = append(out, Tile{X: x, Y: y})
		}
	}
	return out
}

// CanReach reports whether the unit can walk to (x, y) this turn.
func CanReach(c *content.Content, b *state.Battle, u *state.Unit, x, y int) bool {
	for _, t := range Reach(c, b, u) {
		if t.X == x && t.Y == y {
			return true
		}
	}
	return false
}

// Path is the route the unit walks to (x, y) with no distance limit, start
// tile first, or nil if there is no route. Ties are broken exactly as the old
// game broke them, because traps and auto-walk depend on the route taken.
func Path(c *content.Content, b *state.Battle, u *state.Unit, x, y int) []Tile {
	cost := func(a, bb int) int { return StepCost(c, b, u, a, bb) }
	if u.Status["vamp"] > 0 { // Blood Pact walks over anything (old bfsPath, line 1228)
		cost = func(int, int) int { return 2 }
	}
	r := search(b, []Tile{{X: u.X, Y: u.Y}}, cost, 10_000_000, enemyBlocks(b, u))
	k := y*b.N + x
	if r.dist[k] >= Blocked {
		return nil
	}
	var out []Tile
	for k >= 0 {
		out = append([]Tile{{X: k % b.N, Y: k / b.N}}, out...)
		k = r.parent[k]
	}
	return out
}

// Distances gives, for every tile, how many tiles of walking it is from the
// nearest goal for this unit (999 if unreachable). The enemy AI uses this.
func Distances(c *content.Content, b *state.Battle, u *state.Unit, goals []Tile) func(x, y int) float64 {
	cost := func(x, y int) int { return StepCost(c, b, u, x, y) }
	r := search(b, goals, cost, 10_000_000, nil)
	return func(x, y int) float64 {
		d := r.dist[y*b.N+x]
		if d >= Blocked {
			return 999
		}
		return float64(d) / 2
	}
}
