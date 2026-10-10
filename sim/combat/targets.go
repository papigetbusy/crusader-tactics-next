package combat

import (
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
	switch {
	case s.Summon != nil:
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
		if len(b.Fallen) > 0 {
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
	lo, hi := s.MinRange, s.MaxRange
	if s.Strike {
		// The basic Strike without a ranged weapon set is a melee hit
		// (weapon sets arrive with gear in Phase 6).
		lo, hi = 1, 1
	}
	if !s.Heal {
		hi += u.RangeBonus
	}
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
