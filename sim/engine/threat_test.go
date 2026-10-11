package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestThreat(t *testing.T) {
	t.Run("A monster's danger zone is where it could walk next turn, even after it has moved", func(t *testing.T) {
		g := climbGame()
		var m *state.Unit
		for _, u := range g.B.Units {
			if u.Team == state.Enemies {
				m = u
				break
			}
		}
		fresh := len(g.Threat(m.ID))
		m.Moved = true
		if got := len(g.Threat(m.ID)); got != fresh || fresh < 2 || !m.Moved {
			t.Fatalf("fresh %d, after moving %d", fresh, got)
		}
	})
	t.Run("A snared monster's danger zone is only its own tile", func(t *testing.T) {
		g := climbGame()
		for _, u := range g.B.Units {
			if u.Team == state.Enemies {
				u.Status["snare"] = 2
				if z := g.Threat(u.ID); len(z) != 1 || z[0] != (state.Tile{X: u.X, Y: u.Y}) {
					t.Fatalf("zone %v", z)
				}
				return
			}
		}
	})
}
