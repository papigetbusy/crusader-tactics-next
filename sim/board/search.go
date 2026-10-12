package board

import "github.com/papigetbusy/crusader-tactics-next/sim/state"

type searchResult struct {
	dist   []int // cost to reach each tile (Blocked if unreachable)
	parent []int // the tile we came from (-1 at a start tile)
}

// neighbours are visited in this exact order, as in the old game, because the
// order decides which of several equally short routes is chosen.
var neighbours = [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

// search is the old game's "dial-bucket" shortest-path search (function dijk,
// index.html line 1222). Tiles are kept in buckets by total cost and handled
// cheapest first; within a bucket, in the order they were added.
func search(b *state.Battle, src []Tile, cost func(x, y int) int, limit int, block func(x, y int) bool, climb ...func(fx, fy, x, y int) int) searchResult {
	n := b.N
	r := searchResult{dist: make([]int, n*n), parent: make([]int, n*n)}
	for i := range r.dist {
		r.dist[i], r.parent[i] = Blocked, -1
	}
	var buckets [][]int
	push := func(d, k int) {
		for len(buckets) <= d {
			buckets = append(buckets, nil)
		}
		buckets[d] = append(buckets[d], k)
	}
	for _, s := range src {
		k := s.Y*n + s.X
		r.dist[k] = 0
		push(0, k)
	}
	for d := 0; d < len(buckets); d++ {
		for i := 0; i < len(buckets[d]); i++ {
			k := buckets[d][i]
			if r.dist[k] < d {
				continue // already reached more cheaply
			}
			x, y := k%n, k/n
			for _, nb := range neighbours {
				a, c := x+nb[0], y+nb[1]
				if !b.Inside(a, c) || (block != nil && block(a, c)) {
					continue
				}
				step := cost(a, c)
				if step >= 100_000_000 {
					continue
				}
				if len(climb) > 0 && climb[0] != nil {
					step += climb[0](x, y, a, c)
				}
				nd, kk := d+step, c*n+a
				if nd > limit || nd >= r.dist[kk] {
					continue
				}
				r.dist[kk], r.parent[kk] = nd, k
				push(nd, kk)
			}
		}
	}
	return r
}
