package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Fire Emblem counterattacks (owner, 2026-10-11; design-lock 20): a unit
// hit by a single-target attack that survives strikes back once with its
// Strike, if the attacker is within its Strike range.

// counterStatuses stop a unit from striking back.
var counterStatuses = []string{"stun", "frozen", "snare", "cloak", "blood", "sky"}

// Countered reports whether skill s can be struck back at, and if not, why.
func Countered(s content.Skill) (bool, string) {
	switch {
	case s.Heal || s.Support || s.Revive || s.Summon != nil || s.NoDamage:
		return false, "not an attack"
	case s.Kind == "u":
		return false, "ultimate"
	case s.Area > 0 || s.Line > 0 || s.ExtraTargets > 0:
		return false, "area attack"
	case s.Aim != content.AimUnit && s.Aim != content.AimRoll && s.Aim != content.AimShove:
		return false, "not aimed at one unit"
	}
	return true, ""
}

// strikeOf is the Strike a unit counters with, and whether it is from its
// type's own list.
func strikeOf(c *content.Content, v *state.Unit) (content.Skill, bool) {
	for _, s := range v.Skills {
		if s.Strike {
			_, cb, _ := c.SkillFor(v.Team, v.Class, s.Name)
			return s, cb
		}
	}
	return c.Skills["Strike"], false
}

// CounterRange is how near and far a unit strikes back: the weapon in hand
// for an armed hero; otherwise its class's usual weapon (melee 1, bows,
// guns, staffs and wands 2-6), plus range bonuses.
func CounterRange(c *content.Content, v *state.Unit) (lo, hi int) {
	s, _ := strikeOf(c, v)
	if v.Armed {
		return SkillRange(c, v, s)
	}
	lo, hi = 1, 1
	if stats.ClassCategory(c, v.Class) != "m" {
		lo, hi = 2, 6
	}
	return lo, hi + v.RangeBonus
}

// CanCounter reports whether v would strike back at u standing at (x, y)
// after a hit with skill s, and if not, why. It does not know yet whether v
// survives the hit.
func CanCounter(c *content.Content, u, v *state.Unit, s content.Skill, x, y int) (bool, string) {
	if !c.Combat.Counterattacks || v == nil || u == nil || v == u || v.Team == u.Team || v.Decoy {
		return false, ""
	}
	if ok, why := Countered(s); !ok {
		return false, why
	}
	if v.Ascension == "Duelist" { // design-lock 20: Riposte replaces the counter (owner, 2026-10-16)
		return false, "ripostes instead"
	}
	for _, st := range counterStatuses {
		if v.Status[st] > 0 {
			return false, stName(st)
		}
	}
	lo, hi := CounterRange(c, v)
	if d := abs(v.X-x) + abs(v.Y-y); d < lo || d > hi {
		return false, "out of range"
	}
	return true, ""
}

func stName(k string) string {
	return map[string]string{"stun": "stunned", "frozen": "frozen", "snare": "rooted", "cloak": "cloaked", "blood": "cloaked", "sky": "in the sky"}[k]
}

// CounterAttack is the counter's hit, for the forecast and the real blow.
func CounterAttack(c *content.Content, u, v *state.Unit) Attack {
	s, cb := strikeOf(c, v)
	return Attack{By: v, Skill: s, ClassBase: cb, Target: u, Counter: true, NoRedirect: true, Extra: c.Combat.CounterPower}
}

// counter strikes back after unit u used skill s on v, if v survived.
func counter(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, v *state.Unit, emit action.Emit) {
	if v == nil || b.Unit(u.ID) == nil || b.Unit(v.ID) == nil || v.HP <= 0 || u.HP <= 0 {
		return
	}
	if ok, _ := CanCounter(c, u, v, s, u.X, u.Y); !ok {
		return
	}
	emit(action.Event{Kind: action.Countered, Unit: v.ID, Target: u.ID, Skill: "Strike"})
	Hurt(c, b, dice, CounterAttack(c, u, v), emit)
}
