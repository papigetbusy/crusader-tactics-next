package combat

import "testing"

func TestGroundEffects(t *testing.T) {
	t.Run("A unit in a forest is told it takes 15% less damage there", func(t *testing.T) {
		c, b, _, h := arena('f')
		g := GroundEffects(c, b, h)
		if len(g) != 1 || g[0].Label != "FOREST −15% DMG TAKEN" || !g[0].Good {
			t.Fatalf("%+v", g)
		}
	})
	t.Run("A unit in a fort is told how much it heals each turn", func(t *testing.T) {
		c, b, _, h := arena('t')
		h.MaxHP = 50
		g := GroundEffects(c, b, h)
		if len(g) != 2 || g[1].Label != "FORT +5 HP/t" {
			t.Fatalf("%+v", g)
		}
	})
	t.Run("A unit on a hill is told it has high ground", func(t *testing.T) {
		c, b, _, h := arena('h')
		b.Heights = [][]int{{0, 1, 0}, {0, 0, 0}, {0, 0, 0}}
		g := GroundEffects(c, b, h)
		if len(g) != 2 || g[1].Label != "HIGH GROUND +10% HIT · +10% DMG" {
			t.Fatalf("%+v", g)
		}
	})
	t.Run("Plain ground does nothing, so it says nothing", func(t *testing.T) {
		c, b, _, h := arena('g')
		if g := GroundEffects(c, b, h); len(g) != 0 {
			t.Fatalf("%+v", g)
		}
	})
}
