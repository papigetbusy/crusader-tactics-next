package floor

import (
	"math"

	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
)

// terrain draws a floor's ground (old genTerrain, index.html lines 1213-1228):
// mountain ridges with passes, a river with bridges, forests, hills, a lake,
// forts, clear starting bands at the top and bottom, a guaranteed walking
// route from the bottom middle to the top middle, and a road along it.
// Letters: g plains, m mountain, w water, b bridge, f forest, h hill,
// t fort, r road, T throne.
func terrain(n int, throne bool, r *rng.Stream) [][]byte {
	t := make([][]byte, n)
	for y := range t {
		t[y] = make([]byte, n)
		for x := range t[y] {
			t[y][x] = 'g'
		}
	}
	inb := func(x, y int) bool { return x >= 0 && y >= 0 && x < n && y < n }
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	blob := func(cx, cy, rad int, ch byte, p float64, ok func(byte) bool) {
		for y := cy - rad; y <= cy+rad; y++ {
			for x := cx - rad; x <= cx+rad; x++ {
				if !inb(x, y) {
					continue
				}
				extra := 0
				if r.Next() < .3 {
					extra = 1
				}
				if abs(x-cx)+abs(y-cy) <= rad+extra && r.Next() < p && (ok == nil || ok(t[y][x])) {
					t[y][x] = ch
				}
			}
		}
	}
	top, bot := 3, n-3
	clamp := func(v int) int { return max(top, min(bot-1, v)) }
	step := func(yy *int, p float64) {
		if r.Next() < p {
			if r.Next() < .5 {
				*yy--
			} else {
				*yy++
			}
		}
		*yy = clamp(*yy)
	}

	// mountain ridges across the map, with a few passes
	var ridges int
	switch {
	case n >= 22:
		ridges = 2 + r.N(2)
	case n >= 15:
		ridges = 1 + r.N(2)
	default:
		ridges = r.N(2)
	}
	for k := 0; k < ridges; k++ {
		yy := top + 1 + r.N(max(1, bot-top-2))
		var passes []int
		big := 0
		if n > 22 {
			big = 1
		}
		for i := 0; i < 1+r.N(2)+big; i++ { // the old loop rolls its limit again every time
			passes = append(passes, 1+r.N(n-2))
		}
		x0, x1 := 0, n
		if r.Next() < .3 {
			x0 = r.N(n >> 2)
		}
		if r.Next() < .3 {
			x1 = n - r.N(n>>2)
		}
		for x := x0; x < x1; x++ {
			step(&yy, .35)
			pass := false
			for _, p := range passes {
				if abs(p-x) <= 1 {
					pass = true
				}
			}
			if pass {
				continue
			}
			t[yy][x] = 'm'
			if r.Next() < .45 && yy+1 < bot {
				t[yy+1][x] = 'm'
			}
		}
	}

	// a river with bridges
	if n >= 13 && r.Next() < .65 {
		yy := top + r.N(max(1, bot-top))
		var bridges []int
		for i := 0; i < 2+r.N(2); i++ {
			bridges = append(bridges, r.N(n))
		}
		isBridge := func(x int) bool {
			for _, b := range bridges {
				if b == x {
					return true
				}
			}
			return false
		}
		for x := 0; x < n; x++ {
			step(&yy, .3)
			t[yy][x] = 'w'
			if isBridge(x) {
				t[yy][x] = 'b'
			}
			if r.Next() < .25 && yy+1 < bot && !isBridge(x) {
				t[yy+1][x] = 'w'
			}
		}
	}

	// forests, hills, a lake
	plain := func(c byte) bool { return c == 'g' }
	for i := 0; i < int(math.Round(float64(n*n)/38)); i++ {
		x := r.N(n)
		y := r.N(n)
		blob(x, y, 1+r.N(2), 'f', .75, plain)
	}
	for i := 0; i < int(math.Round(float64(n*n)/110)); i++ {
		x := r.N(n)
		y := r.N(n)
		blob(x, y, 1, 'h', .7, plain)
	}
	if r.Next() < .4 {
		x := r.N(n)
		y := top + r.N(max(1, bot-top))
		blob(x, y, 1, 'w', .8, func(c byte) bool { return c == 'g' || c == 'f' })
	}

	// forts
	forts := 2
	if n >= 20 {
		forts = 3
	}
	for i := 0; i < 1+r.N(forts); i++ {
		x := r.N(n)
		y := top + r.N(max(1, bot-top))
		if t[y][x] != 'w' && t[y][x] != 'm' {
			t[y][x] = 't'
		}
	}

	// clear starting bands (a few trees at the top)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if y < top || y >= bot {
				t[y][x] = 'g'
				if y < top && r.Next() < .12 {
					t[y][x] = 'f'
				}
			}
		}
	}
	if throne {
		t[1][n>>1] = 'T'
	}

	// guarantee a walking route from the bottom middle to the top middle,
	// then lay a road along it
	walk := func(x, y int) float64 {
		switch t[y][x] {
		case 'm', 'w':
			return 1e9
		case 'f':
			return 3
		case 'h':
			return 2
		}
		return 1
	}
	route := func() []int {
		d, prev := paths(n, [2]int{n >> 1, n - 1}, walk)
		if d[n>>1] < 1e9 {
			return prev
		}
		return nil
	}
	prev := route()
	for guard := 0; prev == nil && guard < 6; guard++ {
		x := 1 + r.N(n-2)
		for y := 0; y < n; y++ {
			if t[y][x] == 'm' {
				t[y][x] = 'g'
			}
			if t[y][x] == 'w' {
				t[y][x] = 'b'
			}
		}
		prev = route()
	}
	if prev != nil {
		for k := n >> 1; k >= 0; k = prev[k] {
			x, y := k%n, k/n
			if c := t[y][x]; c == 'g' || c == 'f' || c == 'h' {
				t[y][x] = 'r'
			}
		}
	}
	return t
}

// paths is the old dijk (index.html line 1209) from one tile: the cheapest
// cost to every tile and the tile each was reached from (-1: none). Costs
// are whole numbers; tiles cost 1e8 or more are walls. It keeps the old
// order of visiting, so ties are broken the same way.
func paths(n int, from [2]int, cost func(x, y int) float64) ([]float64, []int) {
	d := make([]float64, n*n)
	prev := make([]int, n*n)
	for i := range d {
		d[i], prev[i] = 1e9, -1
	}
	start := from[1]*n + from[0]
	d[start] = 0
	q := [][]int{{start}}
	for dist := 0; dist < len(q); dist++ {
		for i := 0; i < len(q[dist]); i++ {
			k := q[dist][i]
			if d[k] < float64(dist) {
				continue
			}
			x, y := k%n, k/n
			for _, s := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				a, b := x+s[0], y+s[1]
				if a < 0 || b < 0 || a >= n || b >= n {
					continue
				}
				c := cost(a, b)
				if c >= 1e8 {
					continue
				}
				nd := float64(dist) + c
				kk := b*n + a
				if nd >= d[kk] {
					continue
				}
				d[kk], prev[kk] = nd, k
				for len(q) <= int(nd) {
					q = append(q, nil)
				}
				q[int(nd)] = append(q[int(nd)], kk)
			}
		}
	}
	return d, prev
}

// height is a ground letter's height (old terM): mountains -2 and water -1
// (nobody stands there), hills 1, the rest 0.
func height(c byte) int {
	switch c {
	case 'm':
		return -2
	case 'w':
		return -1
	case 'h':
		return 1
	}
	return 0
}
