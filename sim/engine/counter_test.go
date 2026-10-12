package engine

import "testing"

func TestCounterForecast(t *testing.T) {
	t.Run("The forecast shows the target's counterattack", func(t *testing.T) {
		g := duel()
		e := g.B.Unit(2)
		e.X, e.Y = 1, 0
		slot := -1
		for i, s := range g.B.Unit(1).Skills {
			if !s.Heal && !s.Support && s.Kind != "u" && s.MaxRange == 1 {
				slot = i
				break
			}
		}
		f, ok := g.Preview(1, slot, 2)
		if !ok || f.Counter == nil || f.Counter.Mid <= 0 {
			t.Fatalf("forecast %+v", f)
		}
	})
	t.Run("The forecast says why there is no counterattack", func(t *testing.T) {
		g := duel()
		e := g.B.Unit(2)
		e.X, e.Y = 1, 0
		e.Status["stun"] = 1
		f, _ := g.Preview(1, 0, 2)
		if f.Counter != nil || f.NoCounter != "stunned" {
			t.Fatalf("forecast %+v", f)
		}
	})
}

func TestHeightForecast(t *testing.T) {
	t.Run("The forecast says when a hero attacks from high ground", func(t *testing.T) {
		g := duel()
		e := g.B.Unit(2)
		e.X, e.Y = 1, 0
		g.B.Heights = [][]int{{1, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}}
		if f, _ := g.Preview(1, 0, 2); f.Height != 1 {
			t.Fatalf("forecast %+v", f)
		}
	})
}
