package engine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// playTurns plays heroes' turns with random skills, then ends the turn.
func playTurns(t *testing.T, g *Game, dice rng.Dice, n int) {
	for range n {
		if g.B.Outcome != state.Ongoing {
			return
		}
		for _, id := range g.B.Party {
			u := g.B.Unit(id)
			if u == nil || u.Done || g.B.Outcome != state.Ongoing || len(g.B.Decisions) > 0 {
				continue
			}
			aimAndUse(t, g, dice, u, rng.IntN(dice, "soak", len(u.Skills)))
		}
		for len(g.B.Decisions) > 0 {
			if _, err := g.Apply(action.Action{Kind: action.Decide, Choice: g.Options()[0]}); err != nil {
				t.Fatal(err)
			}
		}
		if g.B.Outcome == state.Ongoing {
			if _, err := g.Apply(action.Action{Kind: action.EndTurn}); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSaves(t *testing.T) {
	c := content.MustDefault()
	strongClimb := func() *Game {
		g := NewClimb(c, 77, 5)
		for _, h := range g.climb.run.Heroes {
			h.Level = 24
			h.Learned, h.Equipped, h.Ultimate = c.StartingBuild(h.Class, 24)
		}
		g.climb.run.Floor, g.climb.run.Best = 12, 12
		g.prepare()
		return g
	}

	t.Run("A saved battle resumes exactly where it was, mid-fight; only the dice are fresh", func(t *testing.T) {
		g := strongClimb()
		if _, err := g.Apply(action.Action{Kind: action.Start}); err != nil {
			t.Fatal(err)
		}
		playTurns(t, g, rng.New(1, 99), 2)
		if g.B.Turn != 3 || g.B.Outcome != state.Ongoing {
			t.Fatalf("turn %d, outcome %v: not mid-fight", g.B.Turn, g.B.Outcome)
		}
		data, err := g.Save()
		if err != nil {
			t.Fatal(err)
		}
		h, err := Load(c, data, 1234)
		if err != nil {
			t.Fatal(err)
		}
		again, _ := h.Save()
		if !bytes.Equal(data, again) {
			t.Fatal("saving the loaded game gives a different save")
		}
		g.Dice = rng.New(1234, streamCombat) // the same fresh dice as the loaded game
		playTurns(t, g, rng.New(2, 99), 3)
		playTurns(t, h, rng.New(2, 99), 3)
		if g.B.Turn < 5 && g.B.Outcome == state.Ongoing {
			t.Fatalf("only reached turn %d", g.B.Turn)
		}
		a, _ := g.Save()
		b, _ := h.Save()
		if !bytes.Equal(a, b) {
			t.Fatal("the loaded game played differently from the original")
		}
	})

	t.Run("A climb saved between battles keeps the run: floors, roster, gold and bag", func(t *testing.T) {
		g := strongClimb()
		g.climb.run.Gold, g.climb.run.Bag = 77, []content.Item{{Name: "Steel Sword", Slot: "w", Rarity: 1, WeaponType: "sword"}}
		g.climb.run.Memory(3).Cleared = true
		data, _ := g.Save()
		h, err := Load(c, data, 9)
		if err != nil {
			t.Fatal(err)
		}
		r := h.climb.run
		if r.Floor != 12 || r.Gold != 77 || len(r.Bag) != 1 || !r.Memory(3).Cleared || r.Hero("Witch").Level != 24 || h.climb.phase != PhasePrepare {
			t.Fatalf("%+v", r)
		}
		if _, err := h.Apply(action.Action{Kind: action.Start}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Saves carry a version stamp; something else, or a save from a newer game, is refused in plain words", func(t *testing.T) {
		data, _ := NewClimb(c, 1, 1).Save()
		if !strings.Contains(string(data), `"version":1`) {
			t.Fatal("no version stamp")
		}
		if _, err := Load(c, []byte(`{"hello":1}`), 1); err == nil || !strings.Contains(err.Error(), "not a save") {
			t.Fatalf("junk: %v", err)
		}
		newer := strings.Replace(string(data), `"version":1`, `"version":99`, 1)
		if _, err := Load(c, []byte(newer), 1); err == nil || !strings.Contains(err.Error(), "newer version") {
			t.Fatalf("newer: %v", err)
		}
	})
}
