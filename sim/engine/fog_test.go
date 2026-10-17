package engine

import (
	"strings"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// near puts unit u d steps above unit v (same column), on open ground.
func near(g *Game, u, v *state.Unit, d int) {
	u.X, u.Y = v.X, v.Y-d
	row := []byte(g.B.Ground[u.Y])
	row[u.X] = 'g'
	g.B.Ground[u.Y] = string(row)
	g.B.Heights[u.Y][u.X] = 0
}

func TestFogOfWar(t *testing.T) {
	t.Run("A versus map is the climb's biggest size", func(t *testing.T) {
		g := versus(t)
		if g.B.N != g.C.Floors.MapMax || g.B.N != 32 {
			t.Fatalf("map %d", g.B.N)
		}
	})
	t.Run("At the start neither player can see the other's heroes across the map", func(t *testing.T) {
		g := versus(t)
		for s := 0; s < 2; s++ {
			for _, u := range g.ViewFor(s).Units {
				if u.Team != s {
					t.Fatalf("player %d sees %s", s+1, u.Class)
				}
			}
		}
	})
	t.Run("An enemy within 6 steps of one of your heroes is seen; one 7 steps away is not", func(t *testing.T) {
		g := versus(t)
		me, foe := side(g, 0)[0], side(g, 1)[0]
		near(g, foe, me, 6)
		if !g.Sees(0, foe) {
			t.Fatal("6 steps away was not seen")
		}
		near(g, foe, me, 7)
		for _, u := range side(g, 0)[1:] { // keep the rest of Blue far away
			u.X, u.Y = 0, 31
		}
		if g.Sees(0, foe) {
			t.Fatal("7 steps away was seen")
		}
	})
	t.Run("The screen is told which tiles a player sees, to shade the rest", func(t *testing.T) {
		g := versus(t)
		v := g.ViewFor(0)
		me := side(g, 0)[0]
		if len(v.Seen) != g.B.N || v.Seen[me.Y][me.X] != '1' || v.Seen[0][0] != '0' {
			t.Fatalf("seen rows %d", len(v.Seen))
		}
	})
	t.Run("You can't attack a unit you can't see", func(t *testing.T) {
		g := versus(t)
		me, foe := side(g, 0)[0], side(g, 1)[0]
		if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: me.ID, Skill: 0, Target: foe.ID}); code(err) != "unseen" {
			t.Fatalf("attack on an unseen unit: %v", err)
		}
		for _, id := range g.Targets(me.ID, 0) {
			if id == foe.ID {
				t.Fatal("an unseen unit was offered as a target")
			}
		}
	})
	t.Run("Events about enemies a player can't see are left out of what that player is shown", func(t *testing.T) {
		g := versus(t)
		before := g.Visible(0)
		if _, err := g.Apply(action.Action{Kind: action.EndTurn}); err != nil {
			t.Fatal(err)
		}
		foe := side(g, 1)[0]
		evs := []action.Event{{Kind: action.Moved, Unit: foe.ID}, {Kind: action.TurnStart, Side: 1}}
		got := g.EventsFor(0, evs, before)
		if len(got) != 1 || got[0].Kind != action.TurnStart {
			t.Fatalf("Blue was shown %v", got)
		}
		if len(g.EventsFor(1, evs, g.Visible(1))) != 2 {
			t.Fatal("Red was not shown its own move")
		}
	})
	t.Run("A player's save has no enemies they can't see, so it can't be used to peek", func(t *testing.T) {
		g := versus(t)
		b, err := g.SaveFor(0)
		if err != nil {
			t.Fatal(err)
		}
		l, err := Load(g.C, b, 1)
		if err != nil {
			t.Fatal(err)
		}
		for _, u := range l.B.Units {
			if u.Team == 1 {
				t.Fatalf("Blue's save has Red's %s", u.Class)
			}
		}
		if !strings.Contains(string(b), side(g, 0)[0].Class) {
			t.Fatal("Blue's own heroes are missing")
		}
	})
	t.Run("When the match is over the fog lifts and both players see the whole board", func(t *testing.T) {
		g := versus(t)
		g.B.Outcome = state.Won
		if len(g.ViewFor(0).Units) != len(g.B.Units) {
			t.Fatal("still fogged after the match")
		}
	})
	t.Run("Watchers see everything", func(t *testing.T) {
		g := versus(t)
		if len(g.ViewFor(-1).Units) != len(g.B.Units) || g.ViewFor(-1).Seen != nil {
			t.Fatal("a watcher missed units")
		}
	})
	t.Run("Single player has no fog of war", func(t *testing.T) {
		g := climbGame()
		if g.Fogged() {
			t.Fatal("fog in single player")
		}
	})
}
