package board

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// strip builds a 1-row battle from a picture like "K.~..": each letter is the
// ground ('.' = plains), and the mover stands on the first tile.
func strip(ground string, class string, mov int) (*content.Content, *state.Battle, *state.Unit) {
	n := len(ground)
	rows := make([]string, n)
	row := []byte(ground)
	for i, ch := range row {
		if ch == '.' {
			row[i] = 'g'
		}
	}
	rows[0] = string(row)
	for y := 1; y < n; y++ {
		rows[y] = string(make([]byte, 0))
		for x := 0; x < n; x++ {
			rows[y] += "w" // everything off the strip is water, so routes stay on it
		}
	}
	u := &state.Unit{ID: 1, Team: state.Heroes, Class: class, Mov: mov}
	b := &state.Battle{N: n, Ground: rows, Units: []*state.Unit{u}}
	return content.MustDefault(), b, u
}

func TestMovement(t *testing.T) {
	t.Run("A unit cannot walk through water", func(t *testing.T) {
		c, b, u := strip("g.w..", "Warrior", 9)
		if CanReach(c, b, u, 3, 0) {
			t.Fatal("walked through water")
		}
	})

	t.Run("A flying unit crosses water and mountains", func(t *testing.T) {
		c, b, u := strip("g.wm.", "Dragon Knight", 9)
		if !CanReach(c, b, u, 4, 0) {
			t.Fatal("flier could not cross")
		}
	})

	t.Run("Move 4 crosses 4 plains tiles but not 5", func(t *testing.T) {
		c, b, u := strip("g.....", "Warrior", 4)
		if !CanReach(c, b, u, 4, 0) || CanReach(c, b, u, 5, 0) {
			t.Fatal("plains distance wrong")
		}
	})

	t.Run("A road tile costs half a move point", func(t *testing.T) {
		c, b, u := strip("grrrrrrrr", "Warrior", 4)
		if !CanReach(c, b, u, 8, 0) {
			t.Fatal("8 road tiles should cost 4 move")
		}
	})

	t.Run("Forest costs 2 move points, and 3 for a mounted Knight", func(t *testing.T) {
		c, b, u := strip("gff", "Warrior", 4)
		if !CanReach(c, b, u, 2, 0) {
			t.Fatal("warrior should cross 2 forest with move 4")
		}
		c, b, u = strip("gff", "Knight", 5)
		if CanReach(c, b, u, 2, 0) {
			t.Fatal("knight should need 6 move for 2 forest")
		}
	})

	t.Run("Allies can be walked through but not stopped on", func(t *testing.T) {
		c, b, u := strip("g...", "Warrior", 3)
		b.Units = append(b.Units, &state.Unit{ID: 2, Team: state.Heroes, X: 1})
		if CanReach(c, b, u, 1, 0) {
			t.Fatal("stopped on an ally")
		}
		if !CanReach(c, b, u, 2, 0) {
			t.Fatal("could not pass the ally")
		}
	})

	t.Run("Enemies block the way", func(t *testing.T) {
		c, b, u := strip("g...", "Warrior", 3)
		b.Units = append(b.Units, &state.Unit{ID: 2, Team: state.Enemies, X: 1})
		if CanReach(c, b, u, 2, 0) {
			t.Fatal("walked through an enemy")
		}
	})

	t.Run("A unit that moved, or is snared, locked or shocked, cannot move", func(t *testing.T) {
		for _, s := range []string{"snare", "lock", "shock"} {
			if CanMove(&state.Unit{Status: map[string]int{s: 1}}) {
				t.Fatalf("%s did not stop movement", s)
			}
		}
		if CanMove(&state.Unit{Moved: true}) || !CanMove(&state.Unit{}) {
			t.Fatal("moved flag wrong")
		}
	})

	t.Run("Where the map has no letter, low ground is water", func(t *testing.T) {
		b := &state.Battle{N: 2, Heights: [][]int{{-1, 0}, {0, 0}}}
		if GroundAt(b, 0, 0) != 'w' || GroundAt(b, 1, 0) != 'g' {
			t.Fatal("height fallback wrong")
		}
	})

	t.Run("A walking route starts where the unit stands and ends at the goal", func(t *testing.T) {
		c, b, u := strip("g...", "Warrior", 3)
		p := Path(c, b, u, 3, 0)
		if len(p) != 4 || p[0] != (Tile{0, 0}) || p[3] != (Tile{3, 0}) {
			t.Fatalf("route %v", p)
		}
	})
}
