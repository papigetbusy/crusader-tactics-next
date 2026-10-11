package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
)

// prepared is duel() with a Lv 27 Warrior who has learned every skill up to
// Lv 27 and carries the first four.
func prepared() *Game {
	g := duel()
	h := g.B.Unit(1)
	h.Level = 27
	h.Learned, h.Equipped, h.Ultimate = g.C.StartingBuild("Warrior", 27)
	h.Skills = g.C.Bar("Warrior", h.Equipped, h.Ultimate)
	h.Cooldowns = make([]int, len(h.Skills))
	return g
}

func TestBringSkills(t *testing.T) {
	t.Run("Before the battle starts a hero can put a learned skill on the bar in place of another", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		lib := g.library(h)
		if len(lib) == 0 {
			t.Fatal("nothing in the library")
		}
		pick := ""
		for _, s := range lib {
			if !s.Ultimate {
				pick = s.Name
				break
			}
		}
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Choice: pick, Slot: "2"}); err != nil {
			t.Fatal(err)
		}
		if h.Equipped[1] != pick || h.Skills[1].Name != pick || len(h.Cooldowns) != len(h.Skills) {
			t.Fatalf("bar %v", h.Equipped)
		}
	})

	t.Run("A learned ultimate takes the ultimate slot", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		for _, s := range g.library(h) {
			if s.Ultimate {
				if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Choice: s.Name}); err != nil {
					t.Fatal(err)
				}
				if h.Ultimate != s.Name {
					t.Fatalf("ultimate %s", h.Ultimate)
				}
				return
			}
		}
		t.Fatal("no second ultimate learned by Lv 27")
	})

	t.Run("Skills can't be changed once the battle has started, or to a skill not learned, or with the bar full and no slot named", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		pick := ""
		for _, s := range g.library(h) {
			if !s.Ultimate {
				pick = s.Name
			}
		}
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Choice: pick}); code(err) != "bar-full" {
			t.Fatalf("full bar: %v", err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Choice: "Fireball", Slot: "1"}); code(err) != "not-learned" {
			t.Fatalf("not learned: %v", err)
		}
		h.Moved = true
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Choice: pick, Slot: "1"}); code(err) != "battle-started" {
			t.Fatalf("after moving: %v", err)
		}
	})
}
