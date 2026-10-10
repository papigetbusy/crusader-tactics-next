package turn

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func field(ground string) (*content.Content, *state.Battle, *state.Unit, *state.Unit) {
	c := content.MustDefault()
	b := &state.Battle{N: 3, Ground: []string{ground, "ggg", "ggg"}, Side: state.Heroes}
	hero := &state.Unit{ID: 1, Team: state.Heroes, Class: "Warrior", HP: 50, MaxHP: 100, Status: map[string]int{}, Cooldowns: []int{0, 2, 99}}
	foe := &state.Unit{ID: 2, Team: state.Enemies, Class: "Warrior", X: 2, Y: 2, HP: 30, MaxHP: 30, Status: map[string]int{}}
	b.Units = []*state.Unit{hero, foe}
	return c, b, hero, foe
}

func quiet(action.Event) {}

func TestTurns(t *testing.T) {
	t.Run("Poison takes 7% of max HP at turn start, never less than 3", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		h.Status["poison"] = 2
		Start(c, b, state.Heroes, quiet)
		if h.HP != 43 || h.Status["poison"] != 1 {
			t.Fatalf("hp %d, poison %d", h.HP, h.Status["poison"])
		}
		h.MaxHP, h.HP = 20, 20
		Start(c, b, state.Heroes, quiet)
		if h.HP != 17 {
			t.Fatalf("small unit lost %d, want 3", 20-h.HP)
		}
	})

	t.Run("Burn and bleed each take 6% of max HP at turn start", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		h.Status["burn"], h.Status["bleed"] = 1, 1
		Start(c, b, state.Heroes, quiet)
		if h.HP != 38 {
			t.Fatalf("hp %d, want 50-6-6", h.HP)
		}
	})

	t.Run("A unit killed by poison leaves the board", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		h.HP, h.Status["poison"] = 2, 1
		Start(c, b, state.Heroes, quiet)
		if b.Unit(h.ID) != nil || b.Outcome != state.Lost {
			t.Fatal("the poisoned hero should be gone and the battle lost")
		}
	})

	t.Run("Standing in a fort heals 10% of max HP at turn start", func(t *testing.T) {
		c, b, h, _ := field("tgg")
		Start(c, b, state.Heroes, quiet)
		if h.HP != 60 {
			t.Fatalf("hp %d", h.HP)
		}
	})

	t.Run("A stunned unit skips its turn", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		h.Status["stun"] = 1
		Start(c, b, state.Heroes, quiet)
		if !h.Done {
			t.Fatal("stunned hero can act")
		}
	})

	t.Run("Cooldowns count down by one at turn start (a cooldown of 99 never does)", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		Start(c, b, state.Heroes, quiet)
		if h.Cooldowns[1] != 1 || h.Cooldowns[2] != 99 {
			t.Fatalf("cooldowns %v", h.Cooldowns)
		}
	})

	t.Run("Timed statuses count down by one at the end of their side's turn", func(t *testing.T) {
		c, b, h, f := field("ggg")
		h.Status["snare"], f.Status["snare"] = 2, 2
		End(c, b, state.Heroes, quiet)
		if h.Status["snare"] != 1 || f.Status["snare"] != 2 {
			t.Fatalf("hero %v, foe %v", h.Status, f.Status)
		}
	})

	t.Run("The battle is won when every enemy is gone", func(t *testing.T) {
		_, b, _, f := field("ggg")
		b.Remove(f.ID)
		Check(b, quiet)
		if b.Outcome != state.Won {
			t.Fatal("not won")
		}
	})
}
