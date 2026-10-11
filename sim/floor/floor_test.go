package floor

import (
	"fmt"
	"slices"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func team(n, level int) []Hero {
	var t []Hero
	for range n {
		t = append(t, Hero{Level: level})
	}
	return t
}

func TestFloors(t *testing.T) {
	c := content.MustDefault()

	t.Run("The same run and floor number always give the same floor", func(t *testing.T) {
		a := Build(c, 99, 23, team(5, 20), Memory{}, nil)
		b := Build(c, 99, 23, team(5, 20), Memory{}, nil)
		if fmt.Sprint(a.Ground, a.Biome, a.Chests) != fmt.Sprint(b.Ground, b.Biome, b.Chests) || len(a.Enemies) != len(b.Enemies) {
			t.Fatal("two builds of the same floor differ")
		}
		for i := range a.Enemies {
			x, y := a.Enemies[i], b.Enemies[i]
			if x.X != y.X || x.Y != y.Y || x.Class != y.Class || x.MaxHP != y.MaxHP || x.Element != y.Element {
				t.Fatal("monsters differ")
			}
		}
		if other := Build(c, 100, 23, team(5, 20), Memory{}, nil); fmt.Sprint(other.Ground) == fmt.Sprint(a.Ground) {
			t.Fatal("another run gave the same map")
		}
	})

	t.Run("Maps start at 14x14 and grow one tile every 4 floors, up to 32x32", func(t *testing.T) {
		for floor, want := range map[int]int{1: 14, 4: 14, 5: 15, 40: 23, 73: 32, 200: 32} {
			if got := Size(c, floor); got != want {
				t.Errorf("floor %d: %d", floor, got)
			}
		}
	})

	t.Run("Every 10th floor has a boss at the top middle; every other 5th floor has an elite", func(t *testing.T) {
		f := Build(c, 7, 20, team(5, 20), Memory{}, nil)
		if !f.Enemies[0].Boss || f.Enemies[0].X != f.N>>1 || f.Enemies[0].Y != 1 || f.Enemies[0].Level != f.Level+1 {
			t.Fatalf("floor 20 first monster: %+v", *f.Enemies[0])
		}
		f = Build(c, 7, 15, team(5, 20), Memory{}, nil)
		if !slices.ContainsFunc(f.Enemies, func(u *state.Unit) bool { return u.Elite }) {
			t.Fatal("floor 15 has no elite")
		}
	})

	t.Run("Shops don't use up a floor: floors 7 and 14 are fights", func(t *testing.T) {
		for _, n := range []int{7, 14} {
			if f := Build(c, 7, n, team(5, 10), Memory{}, nil); f.Shop || len(f.Enemies) == 0 {
				t.Errorf("floor %d is a shop", n)
			}
		}
	})

	t.Run("Throne floors (5, 8, 11...) have a throne to seize at the top middle", func(t *testing.T) {
		f := Build(c, 7, 11, team(5, 10), Memory{}, nil)
		if f.Throne == nil || f.Ground[1][f.N>>1] != 'T' {
			t.Fatal("no throne on floor 11")
		}
		if f := Build(c, 7, 12, team(5, 10), Memory{}, nil); f.Throne != nil {
			t.Fatal("a throne on floor 12")
		}
	})

	t.Run("There is always a walking route from the bottom middle to the top middle", func(t *testing.T) {
		for seed := uint32(1); seed < 60; seed++ {
			f := Build(c, seed, int(seed)*3, team(5, 10), Memory{}, nil)
			cost := func(x, y int) float64 {
				if f.Heights[y][x] < 0 {
					return 1e9
				}
				return 1
			}
			if d, _ := paths(f.N, [2]int{f.N >> 1, f.N - 1}, cost); d[f.N>>1] >= 1e9 {
				t.Fatalf("seed %d floor %d has no route", seed, int(seed)*3)
			}
		}
	})

	t.Run("A floor has at most one monster healer", func(t *testing.T) {
		for seed := uint32(1); seed < 80; seed++ {
			f := Build(c, seed, 60, team(8, 60), Memory{}, nil)
			healers := 0
			for _, u := range f.Enemies {
				if u.Class == "Healer" {
					healers++
				}
			}
			if healers > 1 {
				t.Fatalf("seed %d: %d healers", seed, healers)
			}
		}
	})

	t.Run("Bigger teams meet more monsters: one more for each hero above 5", func(t *testing.T) {
		if a, b := Count(c, 9, 5), Count(c, 9, 7); b != a+2 {
			t.Fatalf("5 heroes: %d, 7 heroes: %d", a, b)
		}
	})

	t.Run("Monsters are stronger against a stronger team", func(t *testing.T) {
		weak, strong := Level(c, 20, team(5, 10)), Level(c, 20, team(5, 30))
		if strong <= weak {
			t.Fatalf("level against Lv 10 heroes %d, against Lv 30 heroes %d", weak, strong)
		}
	})

	t.Run("The blue starting tiles are every open tile of the bottom two rows, more than the heroes", func(t *testing.T) {
		f := Build(c, 3, 9, team(5, 10), Memory{}, nil)
		if len(f.Starts) <= 8 {
			t.Fatalf("%d starting tiles", len(f.Starts))
		}
		for _, s := range f.Starts {
			if s.Y < f.N-2 || f.Heights[s.Y][s.X] < 0 {
				t.Fatalf("starting tile %v", s)
			}
		}
	})

	t.Run("Chests opened before stay opened when the team comes back", func(t *testing.T) {
		for seed := uint32(1); seed < 200; seed++ {
			f := Build(c, seed, 30, team(5, 25), Memory{}, nil)
			if len(f.Chests) == 0 {
				continue
			}
			ch := f.Chests[0]
			again := Build(c, seed, 30, team(5, 25), Memory{Opened: []state.Tile{{X: ch.X, Y: ch.Y}}}, nil)
			if len(again.Chests) != len(f.Chests)-1 {
				t.Fatalf("chests %v, after opening one %v", f.Chests, again.Chests)
			}
			return
		}
		t.Fatal("no floor with a chest")
	})

	t.Run("On reinforcement floors up to 2 monsters arrive at the start of the enemy turn on turns 3 and 5", func(t *testing.T) {
		f := Build(c, 5, 15, team(5, 15), Memory{}, nil)
		b := &state.Battle{N: f.N, Ground: f.Ground, Heights: f.Heights, Units: f.Enemies, Reinforce: true, Turn: 2, NextID: 100}
		n := len(b.Units)
		Reinforce(c, b, rng.New(1, 1), func(action.Event) {})
		if len(b.Units) != n {
			t.Fatal("monsters arrived on turn 2")
		}
		b.Turn = 3
		Reinforce(c, b, rng.New(1, 1), func(action.Event) {})
		if len(b.Units) != n+2 || b.Units[n].Y > 1 || b.Units[n].Class == "Healer" {
			t.Fatalf("%d arrived", len(b.Units)-n)
		}
	})
}
