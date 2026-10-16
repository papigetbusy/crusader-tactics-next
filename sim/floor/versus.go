package floor

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Arena is a versus map (owner, 2026-10-16): the climb's terrain drawn for
// an n x n map, then made mirrored, so the bottom half is the top half turned
// around (point symmetry) and neither player gets better ground. The two rows
// at each end are plains for placing heroes, and both ends are joined by a
// walkable route.
type Arena struct {
	N       int
	Ground  []string
	Heights [][]int
	Biome   int             // index into the biome list
	Starts  [2][]state.Tile // each side's starting tiles: side 0 at the bottom, side 1 at the top
}

// NewArena draws the versus map from a seed.
func NewArena(n, team, biomes int, seed uint32) Arena {
	t := terrain(n, false, rng.NewStream(seed+1))
	for y := 0; y < n/2; y++ {
		for x := 0; x < n; x++ {
			t[n-1-y][n-1-x] = t[y][x]
		}
	}
	if n%2 == 1 { // the middle row mirrors onto itself
		m := n / 2
		for x := 0; x < n/2; x++ {
			t[m][n-1-x] = t[m][x]
		}
	}
	for _, y := range []int{0, 1, n - 2, n - 1} {
		for x := range t[y] {
			t[y][x] = 'g'
		}
	}
	if !joined(t) { // open the two middle columns; they mirror onto each other
		for y := range t {
			t[y][n/2-1], t[y][n-1-(n/2-1)] = 'r', 'r'
		}
	}
	a := Arena{N: n, Biome: rng.NewStream(seed + 2).N(max(1, biomes))}
	for _, row := range t {
		a.Ground = append(a.Ground, string(row))
		h := make([]int, n)
		for x, ch := range row {
			h[x] = height(ch)
		}
		a.Heights = append(a.Heights, h)
	}
	x0 := (n - team) / 2
	for i := 0; i < team; i++ {
		x, y := x0+i%n, n-1-i/n
		a.Starts[0] = append(a.Starts[0], state.Tile{X: x, Y: y})
		a.Starts[1] = append(a.Starts[1], state.Tile{X: n - 1 - x, Y: n - 1 - y})
	}
	return a
}

// joined reports whether a walker can get from the bottom row to the top row
// over ground that is not mountain or water.
func joined(t [][]byte) bool {
	n := len(t)
	open := func(x, y int) bool { return t[y][x] != 'm' && t[y][x] != 'w' }
	seen := make([][]bool, n)
	for y := range seen {
		seen[y] = make([]bool, n)
	}
	var q [][2]int
	for x := 0; x < n; x++ {
		if open(x, n-1) {
			seen[n-1][x] = true
			q = append(q, [2]int{x, n - 1})
		}
	}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		if p[1] == 0 {
			return true
		}
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			x, y := p[0]+d[0], p[1]+d[1]
			if x >= 0 && y >= 0 && x < n && y < n && !seen[y][x] && open(x, y) {
				seen[y][x] = true
				q = append(q, [2]int{x, y})
			}
		}
	}
	return false
}
