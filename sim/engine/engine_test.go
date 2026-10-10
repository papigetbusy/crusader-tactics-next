package engine

import (
	"errors"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// duel is a hero and a monster on a 5x5 plain.
func duel() *Game {
	c := content.MustDefault()
	rows := []string{"ggggg", "ggggg", "ggggg", "ggggg", "ggggg"}
	h := &state.Unit{ID: 1, Team: state.Heroes, Class: "Warrior", Level: 1, Mov: 2, HP: 30, MaxHP: 30, Atk: 12, Def: 4, Status: map[string]int{}}
	h.Skills = c.HeroSkills("Warrior", 1)
	h.Cooldowns = make([]int, len(h.Skills))
	e := &state.Unit{ID: 2, Team: state.Enemies, Class: "Warrior", Level: 1, X: 4, Y: 4, Mov: 4, HP: 26, MaxHP: 26, Atk: 9, Def: 4, Status: map[string]int{}}
	e.Skills = c.MonsterSkills("Warrior", 1, false)
	e.Cooldowns = make([]int, len(e.Skills))
	b := &state.Battle{N: 5, Ground: rows, Units: []*state.Unit{h, e}, Turn: 1}
	return &Game{C: c, B: b, Dice: rng.New(7, 1)}
}

func code(err error) string {
	var r *Refused
	if errors.As(err, &r) {
		return r.Code
	}
	return ""
}

func TestFrontDoor(t *testing.T) {
	t.Run("A move out of reach is refused and changes nothing", func(t *testing.T) {
		g := duel()
		_, err := g.Apply(action.Action{Kind: action.Move, Unit: 1, X: 4, Y: 0})
		if code(err) != "too-far" || g.B.Unit(1).X != 0 {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("A hero moves once per turn", func(t *testing.T) {
		g := duel()
		if _, err := g.Apply(action.Action{Kind: action.Move, Unit: 1, X: 1, Y: 1}); err != nil {
			t.Fatal(err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Move, Unit: 1, X: 1, Y: 2}); code(err) != "cannot-move" {
			t.Fatalf("second move: %v", err)
		}
	})

	t.Run("An attack on a target out of range is refused", func(t *testing.T) {
		g := duel()
		if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: 1, Skill: 0, Target: 2}); code(err) != "bad-target" {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("After attacking, a hero is finished for the turn", func(t *testing.T) {
		g := duel()
		g.B.Unit(2).X, g.B.Unit(2).Y = 1, 0
		if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: 1, Skill: 0, Target: 2}); err != nil {
			t.Fatal(err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Wait, Unit: 1}); code(err) != "finished" {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("Ending the turn lets every monster act, then a new turn starts", func(t *testing.T) {
		g := duel()
		evs, err := g.Apply(action.Action{Kind: action.EndTurn})
		if err != nil {
			t.Fatal(err)
		}
		moved := false
		for _, e := range evs {
			moved = moved || (e.Kind == action.Moved && e.Unit == 2)
		}
		if !moved || g.B.Turn != 2 || g.B.Side != state.Heroes {
			t.Fatalf("monster moved %v, turn %d, side %d", moved, g.B.Turn, g.B.Side)
		}
	})

	t.Run("Nothing can be done after the battle is over", func(t *testing.T) {
		g := duel()
		g.B.Outcome = state.Won
		if _, err := g.Apply(action.Action{Kind: action.EndTurn}); code(err) != "battle-over" {
			t.Fatalf("err %v", err)
		}
	})
}
