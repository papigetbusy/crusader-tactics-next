package engine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// playBot plays a climb the simple way a player could (tap each monster in
// reach, else wait, then end the turn) and returns every action the referee
// took, in order, with the events of each.
func playBot(g *Game, steps int) (took []action.Action, log []byte) {
	var out bytes.Buffer
	try := func(a action.Action) bool {
		evs, err := g.Apply(a)
		if err != nil {
			return false
		}
		took = append(took, a)
		b, _ := json.Marshal(evs)
		out.Write(b)
		return true
	}
	for n := 0; n < steps; n++ {
		if try(action.Action{Kind: action.Decide, Choice: "ok"}) || try(action.Action{Kind: action.Start}) ||
			try(action.Action{Kind: action.Continue}) {
			continue
		}
		if g.B == nil || g.B.Outcome != state.Ongoing {
			break
		}
		for _, id := range g.B.Party {
			h := g.B.Unit(id)
			if h == nil || h.HP <= 0 {
				continue
			}
			for _, v := range g.B.Units {
				if v.Team != state.Enemies {
					continue
				}
				p, ok := g.AttackPlan(h.ID, v.ID)
				if !ok {
					continue
				}
				if p.Walk != nil {
					try(action.Action{Kind: action.Move, Unit: h.ID, X: p.Walk.X, Y: p.Walk.Y})
				}
				if try(action.Action{Kind: action.Attack, Unit: h.ID, Skill: p.Slot, Target: v.ID}) {
					break
				}
			}
			try(action.Action{Kind: action.Wait, Unit: h.ID})
		}
		try(action.Action{Kind: action.EndTurn})
	}
	return took, out.Bytes()
}

func TestReplay(t *testing.T) {
	c := content.MustDefault()
	t.Run("A climb replayed from its seeds and its list of actions comes out identical", func(t *testing.T) {
		for _, seed := range []uint32{4242, 7, 90210} {
			g := NewClimb(c, seed, uint64(seed)*3)
			took, log := playBot(g, 120)
			if len(took) < 20 {
				t.Fatalf("seed %d: only %d actions were played", seed, len(took))
			}
			t.Logf("seed %d: %d actions, %d bytes of events, floor %d", seed, len(took), len(log), g.climb.run.Floor)
			want, err := g.Save()
			if err != nil {
				t.Fatal(err)
			}
			r := NewClimb(c, seed, uint64(seed)*3)
			var again bytes.Buffer
			for i, a := range took {
				evs, err := r.Apply(a)
				if err != nil {
					t.Fatalf("seed %d: action %d (%s) refused on replay: %v", seed, i, a.Kind, err)
				}
				b, _ := json.Marshal(evs)
				again.Write(b)
			}
			if !bytes.Equal(log, again.Bytes()) {
				t.Fatalf("seed %d: the replay told a different story", seed)
			}
			got, _ := r.Save()
			if !bytes.Equal(want, got) {
				t.Fatalf("seed %d: the replay ended in a different game", seed)
			}
		}
	})
}
