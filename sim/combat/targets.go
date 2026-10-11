package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Distance is the number of tiles between two units, counting only straight
// steps (old dist).
func Distance(a, b *state.Unit) int { return abs(a.X-b.X) + abs(a.Y-b.Y) }

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// cleansable are the statuses a cleanse removes (old hasSt, line 776).
var cleansable = []string{"bleed", "snare", "stun", "brk", "poison", "weak", "shock", "burn", "ivy"}

func hasBadStatus(v *state.Unit) bool {
	for _, q := range cleansable {
		if v.Status[q] > 0 {
			return true
		}
	}
	return false
}

// hidden reports a hero that monsters cannot target (cloaked, or in the sky).
func hidden(v *state.Unit) bool {
	return v.Status["cloak"] > 0 || v.Status["blood"] > 0 || v.Status["sky"] > 0
}

// Targets lists who a unit can use a skill on right now, in board order
// (old tgt, line 777). Heals pick hurt allies; attacks pick enemies in range.
func Targets(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill) []*state.Unit {
	return restrain(b, u, s, targets(c, b, u, s))
}

// restrain applies Garrote and Taunting Roar: a silenced hero can only use
// its basic Strike, and a taunted unit's attacks can only go at the unit
// that taunted it.
func restrain(b *state.Battle, u *state.Unit, s content.Skill, out []*state.Unit) []*state.Unit {
	if u.Status["silence"] > 0 && !s.Strike {
		return nil
	}
	if t := b.Unit(u.TauntedBy); u.Status["taunted"] > 0 && t != nil && !s.Friendly() {
		if slices.Contains(out, t) {
			return []*state.Unit{t}
		}
		return nil
	}
	return out
}

func targets(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill) []*state.Unit {
	if IsNew(s) {
		return newTargets(c, b, u, s)
	}
	switch {
	case s.Summon != nil: // one raised monster at a time (old tgt, line 777)
		if PetOf(b, u) != nil {
			return nil
		}
		return []*state.Unit{u}
	case s.Support:
		var out []*state.Unit
		for _, z := range b.Units {
			if z.Team == u.Team && Distance(z, u) <= s.Area {
				out = append(out, z)
			}
		}
		return out
	case s.Revive:
		if (u.Team == state.Heroes && len(b.Fallen) > 0) || (u.Team != state.Heroes && len(b.FallenFoes) > 0) {
			return []*state.Unit{u}
		}
		return nil
	case s.Heal && s.AllAllies:
		for _, v := range b.Units {
			if v.Team == u.Team && v.HP < v.MaxHP {
				return []*state.Unit{u}
			}
		}
		return nil
	}
	lo, hi := SkillRange(c, u, s)
	var out []*state.Unit
	for _, v := range b.Units {
		var want bool
		if s.Heal {
			want = v.Team == u.Team && ((s.Cleanse && hasBadStatus(v)) || (!s.Cleanse && v.HP < v.MaxHP))
		} else {
			want = v.Team != u.Team
		}
		d := Distance(u, v)
		if want && d >= lo && d <= hi && !(u.Team == state.Enemies && v.Team == state.Heroes && hidden(v)) {
			out = append(out, v)
		}
	}
	return out
}

// PetOf is the monster a hero raised and still has on the board, or nil.
func PetOf(b *state.Battle, u *state.Unit) *state.Unit {
	for _, z := range b.Units {
		if z.Summoned && z.Owner == u.ID {
			return z
		}
	}
	return nil
}

// SkillRange is how near and how far a unit can use a skill now: the skill's
// range (a Strike's comes from the weapon in hand), plus range bonuses for
// attacks (gear, Hunter's Eye, Deadeye, Steady Aim).
func SkillRange(c *content.Content, u *state.Unit, s content.Skill) (lo, hi int) {
	lo, hi = s.MinRange, s.MaxRange
	if s.Strike {
		lo, hi = strikeRange(c, u)
	}
	if !s.Heal {
		hi += u.RangeBonus
		if u.Status["heye"] > 0 && s.Kind != "u" {
			hi += int(u.P("heR", 3))
		}
		if u.Status["deadeye"] > 0 {
			hi += 99
		}
		if u.Status["aim"] > 0 {
			hi += 2 // Steady Aim
		}
	}
	return lo, hi
}
