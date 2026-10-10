package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func has(evs []action.Event, k string) bool {
	for _, e := range evs {
		if e.Kind == k {
			return true
		}
	}
	return false
}

func TestMoreActions(t *testing.T) {
	t.Run("Defending finishes the hero for the turn", func(t *testing.T) {
		g := duel()
		if _, err := g.Apply(action.Action{Kind: action.Defend, Unit: 1}); err != nil {
			t.Fatal(err)
		}
		if u := g.B.Unit(1); !u.Defending || !u.Done {
			t.Fatalf("defending %v done %v", u.Defending, u.Done)
		}
	})

	t.Run("A potion heals 10 plus 3 per level, never above full health", func(t *testing.T) {
		g := duel()
		u := g.B.Unit(1)
		u.Potions, u.Level, u.HP = 2, 2, 10
		if _, err := g.Apply(action.Action{Kind: action.Potion, Unit: 1}); err != nil {
			t.Fatal(err)
		}
		if u.HP != 26 || u.Potions != 1 || !u.Done {
			t.Fatalf("hp %d potions %d done %v", u.HP, u.Potions, u.Done)
		}
		g = duel()
		u = g.B.Unit(1)
		u.Potions, u.HP = 1, 28
		_, _ = g.Apply(action.Action{Kind: action.Potion, Unit: 1})
		if u.HP != u.MaxHP {
			t.Fatalf("hp %d of %d", u.HP, u.MaxHP)
		}
	})

	t.Run("A potion is refused with none left or at full health", func(t *testing.T) {
		g := duel()
		u := g.B.Unit(1)
		u.HP = 5
		if _, err := g.Apply(action.Action{Kind: action.Potion, Unit: 1}); code(err) != "no-potions" {
			t.Fatalf("err %v", err)
		}
		u.Potions, u.HP = 3, u.MaxHP
		if _, err := g.Apply(action.Action{Kind: action.Potion, Unit: 1}); code(err) != "full-hp" {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("Undo puts a hero back where it stood and lets it move again", func(t *testing.T) {
		g := duel()
		if _, err := g.Apply(action.Action{Kind: action.Move, Unit: 1, X: 1, Y: 1}); err != nil {
			t.Fatal(err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Undo, Unit: 1}); err != nil {
			t.Fatal(err)
		}
		u := g.B.Unit(1)
		if u.X != 0 || u.Y != 0 || u.Moved {
			t.Fatalf("at %d,%d moved %v", u.X, u.Y, u.Moved)
		}
		if _, err := g.Apply(action.Action{Kind: action.Undo, Unit: 1}); code(err) != "nothing-to-undo" {
			t.Fatalf("second undo: %v", err)
		}
	})

	t.Run("Undo is refused when someone now stands on the old tile", func(t *testing.T) {
		g := duel()
		_, _ = g.Apply(action.Action{Kind: action.Move, Unit: 1, X: 1, Y: 1})
		g.B.Unit(2).X, g.B.Unit(2).Y = 0, 0
		if _, err := g.Apply(action.Action{Kind: action.Undo, Unit: 1}); code(err) != "tile-taken" {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("A hero who reaches the throne wins: every enemy flees", func(t *testing.T) {
		g := duel()
		g.B.Throne = &state.Tile{X: 1, Y: 1}
		evs, err := g.Apply(action.Action{Kind: action.Move, Unit: 1, X: 1, Y: 1})
		if err != nil {
			t.Fatal(err)
		}
		if !has(evs, action.Seized) || g.B.Unit(2) != nil || g.B.Outcome != state.Won {
			t.Fatalf("outcome %q", g.B.Outcome)
		}
	})

	t.Run("Resurrection brings back the hero who fell last, beside the healer, with 40% health", func(t *testing.T) {
		g := duel()
		h := g.B.Unit(1)
		h.Class, h.Level = "Healer", 20
		h.Skills = g.C.HeroSkills("Healer", 20)
		h.Cooldowns = make([]int, len(h.Skills))
		slot := -1
		for i, s := range h.Skills {
			if s.Revive {
				slot = i
			}
		}
		if slot < 0 {
			t.Fatal("the Healer has no revive skill")
		}
		first := &state.Unit{ID: 8, Team: state.Heroes, Class: "Tank", MaxHP: 50, Status: map[string]int{"poison": 2}}
		last := &state.Unit{ID: 9, Team: state.Heroes, Class: "Mage", MaxHP: 41, Status: map[string]int{"burn": 1}}
		g.B.Fallen = []*state.Unit{first, last}
		if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: 1, Skill: slot, Target: 1}); err != nil {
			t.Fatal(err)
		}
		v := g.B.Unit(9)
		if v == nil || v.HP != 16 || v.X != 1 || v.Y != 0 || len(v.Status) != 0 || !v.Done {
			t.Fatalf("revived %+v", v)
		}
		if len(g.B.Fallen) != 1 || g.B.Fallen[0] != first {
			t.Fatal("the other fallen hero should still be waiting")
		}
	})

	t.Run("Tapping a monster in reach attacks it from where the hero stands", func(t *testing.T) {
		g := duel()
		g.B.Unit(2).X, g.B.Unit(2).Y = 1, 0
		p, ok := g.AttackPlan(1, 2)
		if !ok || p.Walk != nil {
			t.Fatalf("plan %+v ok %v", p, ok)
		}
	})

	t.Run("Tapping a monster out of reach walks to the nearest tile that can hit it", func(t *testing.T) {
		g := duel()
		g.B.Unit(2).X, g.B.Unit(2).Y = 2, 1
		p, ok := g.AttackPlan(1, 2)
		if !ok || p.Walk == nil || abs(p.Walk.X-2)+abs(p.Walk.Y-1) != 1 || p.Walk.X+p.Walk.Y != 2 {
			t.Fatalf("plan %+v ok %v", p, ok)
		}
	})

	t.Run("Tapping a monster too far to reach this turn gives no plan", func(t *testing.T) {
		g := duel()
		if _, ok := g.AttackPlan(1, 2); ok {
			t.Fatal("expected no plan")
		}
	})
}

func TestScreenInfo(t *testing.T) {
	t.Run("A boss below half health is shown as enraged", func(t *testing.T) {
		g := duel()
		e := g.B.Unit(2)
		e.Boss, e.HP = true, 12
		for _, u := range g.View().Units {
			if u.ID == 2 && !u.Enraged {
				t.Fatal("not shown as enraged")
			}
		}
	})

	t.Run("The forecast of an area skill names everyone else caught in the blast", func(t *testing.T) {
		g := duel()
		h := g.B.Unit(1)
		h.Skills[0].Area, h.Skills[0].MaxRange = 1, 4
		e := g.B.Unit(2)
		e.X, e.Y = 2, 0
		g.B.Units = append(g.B.Units, &state.Unit{ID: 3, Team: state.Enemies, Class: "Mage", X: 3, Y: 0, HP: 5, MaxHP: 5, Status: map[string]int{}})
		f, ok := g.Preview(1, 0, 2)
		if !ok || len(f.Also) != 1 || f.Also[0] != 3 {
			t.Fatalf("forecast %+v", f)
		}
	})

	t.Run("A new practice hero carries 3 potions", func(t *testing.T) {
		g := duel()
		for _, s := range practiceSetups(t) {
			for _, u := range NewFight(g.C, s, 1).B.Units {
				if u.Team == state.Heroes && u.Potions != 3 {
					t.Fatalf("%s has %d potions", u.Class, u.Potions)
				}
			}
		}
	})
}
