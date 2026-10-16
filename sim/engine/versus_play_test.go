package engine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// playVersus has a simple bot play both players (tap each enemy in reach,
// else wait, then end the turn), spending a little clock time on every
// action. It returns the actions the referee took and their events.
func playVersus(g *Game, maxTurns int) (took []action.Action, log []byte) {
	var out bytes.Buffer
	ms := 700
	try := func(a action.Action) bool {
		a.Ms = ms
		evs, err := g.Apply(a)
		if err != nil {
			return false
		}
		took = append(took, a)
		b, _ := json.Marshal(evs)
		out.Write(b)
		ms = 300 + (ms*7)%1900
		return true
	}
	for n := 0; n < maxTurns*2 && g.B.Outcome == state.Ongoing; n++ {
		if len(g.B.Decisions) > 0 {
			try(action.Action{Kind: action.Decide, Choice: "ok"})
			continue
		}
		for _, h := range append([]*state.Unit(nil), g.B.Units...) {
			if h.Team != g.B.Side || !g.mine(h) || h.Done || g.B.Unit(h.ID) == nil || g.B.Outcome != state.Ongoing {
				continue
			}
			for _, v := range append([]*state.Unit(nil), g.B.Units...) {
				if v.Team == h.Team || g.B.Unit(v.ID) == nil {
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

func TestVersusMatches(t *testing.T) {
	c := content.MustDefault()
	t.Run("Whole versus matches play out without a hitch and end in a win, a loss or a time out", func(t *testing.T) {
		ends := map[string]int{}
		for seed := uint64(1); seed <= 12; seed++ {
			g, err := NewVersus(c, VersusSetup{Teams: [2][]Pick{picks(c, int(seed)), picks(c, int(seed)+3)}, MapSeed: uint32(seed)}, seed)
			if err != nil {
				t.Fatal(err)
			}
			playVersus(g, 60)
			if g.B.Outcome == state.Ongoing {
				t.Fatalf("seed %d: still going after 60 rounds", seed)
			}
			end := g.B.Outcome
			if g.B.Versus.TimedOut >= 0 {
				end = "time"
			}
			ends[end]++
		}
		t.Logf("results: %v", ends)
	})

	t.Run("A versus match replayed from its seeds and its list of actions comes out identical", func(t *testing.T) {
		for seed := uint64(1); seed <= 4; seed++ {
			s := VersusSetup{Teams: [2][]Pick{picks(c, int(seed)), picks(c, int(seed)+5)}, MapSeed: uint32(seed * 9)}
			g, _ := NewVersus(c, s, seed)
			took, log := playVersus(g, 60)
			r, _ := NewVersus(c, s, seed)
			var again bytes.Buffer
			for i, a := range took {
				evs, err := r.Apply(a)
				if err != nil {
					t.Fatalf("seed %d: action %d (%s) refused on replay: %v", seed, i, a.Kind, err)
				}
				b, _ := json.Marshal(evs)
				again.Write(b)
			}
			if !bytes.Equal(log, again.Bytes()) || r.B.Outcome != g.B.Outcome || r.B.Versus.BankMs != g.B.Versus.BankMs {
				t.Fatalf("seed %d: the replay told a different story", seed)
			}
		}
	})
}
