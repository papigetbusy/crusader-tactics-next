package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"

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
	t.Run("A monster's red zone is every tile it could hit from anywhere it could walk", func(t *testing.T) {
		g := climbGame()
		for _, u := range g.B.Units {
			if u.Team == state.Enemies {
				walk, hit := g.Threat(u.ID), g.ThreatHit(u.ID)
				if len(hit) < len(walk) {
					t.Fatalf("walk %d tiles, hit %d", len(walk), len(hit))
				}
				return
			}
		}
	})
	t.Run("A hero's red tiles are where its ready attacks reach from where it stands", func(t *testing.T) {
		g := climbGame()
		must(t, g, action.Action{Kind: action.Start})
		h := g.B.Unit(g.B.Party[0])
		ts := g.AttackTiles(h.ID)
		if len(ts) == 0 {
			t.Fatal("no red tiles")
		}
		for _, x := range ts {
			if d := abs(x.X-h.X) + abs(x.Y-h.Y); d == 0 {
				t.Fatal("the hero's own tile is red")
			}
		}
	})
}
