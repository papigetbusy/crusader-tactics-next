package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// newAttack is how much harder a hit lands because of the even-schedule
// skills: Battle Trance +25%, a Hunter's Mark on the target +20%, and High
// Noon +10% per enemy in range (at most +60%). 1 when none apply.
func newAttack(b *state.Battle, a Attack) float64 {
	u, v, s := a.By, a.Target, a.Skill
	m := 1.0
	if u.Status["trance"] > 0 {
		m *= 1.25
	}
	if v.Status["hmark"] > 0 && u.Team != v.Team {
		m *= 1.2
	}
	if s.Trick == "highNoon" {
		n := 0
		for _, z := range b.Units {
			if d := Distance(u, z); z.Team != u.Team && d >= s.MinRange && d <= s.MaxRange+u.Reach(s.MaxRange) {
				n++
			}
		}
		m *= 1 + min(0.6, float64(n)*0.1)
	}
	return m
}

// newTaken is how much of a hit the target takes because of the
// even-schedule skills: +15% in a Battle Trance, -20% after a Rescue, and
// -30% from spells behind Dragon Scales. 1 when none apply.
func newTaken(a Attack) float64 {
	v := a.Target
	m := 1.0
	if v.Status["trance"] > 0 {
		m *= 1.15
	}
	if v.Status["cover"] > 0 {
		m *= 0.8
	}
	if v.Status["dscale"] > 0 && a.Skill.IsMagic() {
		m *= 0.7
	}
	return m
}
