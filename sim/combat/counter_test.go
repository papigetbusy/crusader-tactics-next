package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// strikes counts the counterattacks in one use of skill s by h on f.
func strikes(c *content.Content, b *state.Battle, h, f *state.Unit, s content.Skill) (n int, back int) {
	h.Skills, h.Cooldowns = []content.Skill{s}, []int{0}
	hp := h.HP
	_ = UseSkill(c, b, hits(), h, 0, f, func(e action.Event) {
		if e.Kind == action.Countered && e.Unit == f.ID && e.Target == h.ID {
			n++
		}
	})
	return n, hp - h.HP
}

func TestCounterattacks(t *testing.T) {
	t.Run("A unit that survives a hit strikes back once with a full Strike", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 400, 400
		n, lost := strikes(c, b, h, f, plain(1))
		want := Preview(c, b, CounterAttack(c, h, f)).Mid
		if n != 1 || lost != want {
			t.Fatalf("%d counters, hero lost %d (a full Strike is %d)", n, lost, want)
		}
	})
	t.Run("Heroes strike back at monsters too", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.HP, h.MaxHP = 400, 400
		if n, _ := strikes(c, b, f, h, plain(1)); n != 1 {
			t.Fatalf("%d counters", n)
		}
	})
	t.Run("A unit defeated by the hit cannot strike back", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP = 1
		if n, _ := strikes(c, b, h, f, plain(1)); n != 0 {
			t.Fatalf("%d counters", n)
		}
	})
	t.Run("A melee unit cannot strike back at an attacker out of its reach", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 400, 400
		h.X, h.Y = 1, 2
		s := plain(1)
		s.MinRange, s.MaxRange = 2, 6
		if n, _ := strikes(c, b, h, f, s); n != 0 {
			t.Fatalf("%d counters", n)
		}
		if ok, why := CanCounter(c, h, f, s, h.X, h.Y); ok || why != "out of range" {
			t.Fatalf("CanCounter %v %q", ok, why)
		}
	})
	t.Run("A ranged unit strikes back at range", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Class, f.HP, f.MaxHP = "Ranger", 400, 400
		h.X, h.Y = 1, 2
		s := plain(1)
		s.MinRange, s.MaxRange = 2, 6
		if n, _ := strikes(c, b, h, f, s); n != 1 {
			t.Fatalf("%d counters", n)
		}
	})
	t.Run("Stunned, frozen, rooted and cloaked units cannot strike back", func(t *testing.T) {
		for _, st := range []string{"stun", "frozen", "snare", "cloak"} {
			c, b, h, f := arena('g')
			f.HP, f.MaxHP = 400, 400
			f.Status[st] = 2
			if n, _ := strikes(c, b, h, f, plain(1)); n != 0 {
				t.Fatalf("%s: %d counters", st, n)
			}
		}
	})
	t.Run("Area skills, ultimates, heals and buffs are never struck back at", func(t *testing.T) {
		area, ult, heal := plain(1), plain(1), plain(1)
		area.Area, ult.Kind, heal.Heal = 1, "u", true
		for _, s := range []content.Skill{area, ult, heal} {
			if ok, _ := Countered(s); ok {
				t.Fatalf("%+v can be countered", s)
			}
		}
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 400, 400
		if n, _ := strikes(c, b, h, f, ult); n != 0 {
			t.Fatalf("an ultimate was struck back at %d times", n)
		}
	})
	t.Run("A counterattack is never struck back at", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.HP, h.MaxHP, f.HP, f.MaxHP = 400, 400, 400, 400
		n := 0
		h.Skills, h.Cooldowns = []content.Skill{plain(1)}, []int{0}
		_ = UseSkill(c, b, hits(), h, 0, f, func(e action.Event) {
			if e.Kind == action.Countered {
				n++
			}
		})
		if n != 1 {
			t.Fatalf("%d counters in one exchange", n)
		}
	})
}
