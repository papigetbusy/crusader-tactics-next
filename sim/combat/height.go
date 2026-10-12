package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Height (owner, 2026-10-11; design-lock 20): one step counts, hill over
// everything else.

// high reports whether (x, y) is high ground (a hill).
func high(b *state.Battle, x, y int) bool { return heightAt(b, x, y) > 0 }

// HeightEdge is +1 when the attacker stands higher than its target, -1 when
// it attacks uphill, 0 otherwise (or with height rules off).
func HeightEdge(c *content.Content, b *state.Battle, u, v *state.Unit) int {
	if !c.Combat.Height || b == nil {
		return 0
	}
	hu, hv := high(b, u.X, u.Y), high(b, v.X, v.Y)
	switch {
	case hu && !hv:
		return 1
	case hv && !hu:
		return -1
	}
	return 0
}
