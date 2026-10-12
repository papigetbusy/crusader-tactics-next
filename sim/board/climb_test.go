package board

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// hillRow is a 5x1 strip: plains, a hill in the middle, plains.
func hillRow(class string) (*content.Content, *state.Battle, *state.Unit) {
	c := content.MustDefault()
	u := &state.Unit{ID: 1, Team: state.Heroes, Class: class, Mov: 2, Status: map[string]int{}}
	b := &state.Battle{N: 5, Ground: []string{"gghgg", "wwwww", "wwwww", "wwwww", "wwwww"},
		Heights: [][]int{{0, 0, 1, 0, 0}, {-1, -1, -1, -1, -1}, {-1, -1, -1, -1, -1}, {-1, -1, -1, -1, -1}, {-1, -1, -1, -1, -1}}, Units: []*state.Unit{u}}
	return c, b, u
}

func TestClimbing(t *testing.T) {
	t.Run("Climbing onto a hill costs 1 more move", func(t *testing.T) {
		c, b, u := hillRow("Warrior")
		u.Mov = 3 // plains 1 + hill 1.5 + climb 1 = 3.5: out of reach
		if CanReach(c, b, u, 2, 0) {
			t.Fatal("a Warrior with 3 move climbed the hill two tiles away")
		}
		u.Mov = 4
		if !CanReach(c, b, u, 2, 0) {
			t.Fatal("a Warrior with 4 move could not climb the hill")
		}
	})
	t.Run("Rangers and Knights climb hills at no extra cost", func(t *testing.T) {
		for _, cl := range []string{"Ranger", "Knight"} {
			c, b, u := hillRow(cl)
			u.Mov = 3
			if !CanReach(c, b, u, 2, 0) {
				t.Fatalf("a %s with 3 move could not climb the hill", cl)
			}
		}
	})
}
