package floor

import "testing"

func TestArena(t *testing.T) {
	t.Run("A versus map is mirrored, so neither player gets better ground", func(t *testing.T) {
		for seed := uint32(1); seed <= 50; seed++ {
			a := NewArena(12, 8, 7, seed)
			for y := 0; y < a.N; y++ {
				for x := 0; x < a.N; x++ {
					if a.Ground[y][x] != a.Ground[a.N-1-y][a.N-1-x] {
						t.Fatalf("seed %d: (%d,%d) %c but its mirror %c", seed, x, y, a.Ground[y][x], a.Ground[a.N-1-y][a.N-1-x])
					}
				}
			}
		}
	})
	t.Run("Each player gets 8 starting tiles on plains at their own end", func(t *testing.T) {
		a := NewArena(12, 8, 7, 3)
		for side, ts := range a.Starts {
			if len(ts) != 8 {
				t.Fatalf("side %d has %d tiles", side, len(ts))
			}
			for _, s := range ts {
				if a.Ground[s.Y][s.X] != 'g' || (side == 0 && s.Y < 10) || (side == 1 && s.Y > 1) {
					t.Fatalf("side %d start %+v on %c", side, s, a.Ground[s.Y][s.X])
				}
			}
		}
	})
	t.Run("The two ends of a versus map are always joined by a walkable route", func(t *testing.T) {
		for seed := uint32(1); seed <= 200; seed++ {
			a := NewArena(12, 8, 7, seed)
			var rows [][]byte
			for _, r := range a.Ground {
				rows = append(rows, []byte(r))
			}
			if !joined(rows) {
				t.Fatalf("seed %d is cut in two", seed)
			}
		}
	})
}
