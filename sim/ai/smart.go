package ai

import (
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Smarter enemies as you climb (owner, 2026-10-11; design-lock 20). Before
// SmartFrom, monsters think exactly like the old game.

// smart reports whether monster e weighs who to hit on this floor.
func smart(c *content.Content, b *state.Battle, e *state.Unit) bool {
	if e.Team != state.Enemies {
		return false
	}
	from := c.AI.SmartFrom
	if e.Boss || e.Elite {
		from = c.AI.SmartBossFrom
	}
	return from > 0 && b.Floor >= from
}

// cautious reports whether monster e avoids ending next to several heroes.
func cautious(c *content.Content, b *state.Battle, e *state.Unit) bool {
	return e.Team == state.Enemies && c.AI.CautiousFrom > 0 && b.Floor >= c.AI.CautiousFrom
}

// healer is a hero who heals: a Healer, or anyone carrying a heal.
func healer(v *state.Unit) bool {
	if v.Class == "Healer" {
		return true
	}
	for _, s := range v.Skills {
		if s.Heal {
			return true
		}
	}
	return false
}

// pull is how much monster e wants to hit v with skill s: a hero it can
// defeat this turn most, then healers, then heroes weak to the attack.
func pull(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill, v *state.Unit) float64 {
	if s.Heal || s.Support || v.Team == e.Team {
		return 0
	}
	p := 0.0
	_, classBase, _ := c.SkillFor(e.Team, e.Class, s.Name)
	f := combat.Preview(c, b, combat.Attack{By: e, Skill: s, ClassBase: classBase, Target: v})
	if f.HitChance > 0 && f.Mid >= v.HP {
		p += c.AI.KillWeight
	}
	if healer(v) {
		p += c.AI.HealerWeight
	}
	if f.Mult > 1 {
		p += c.AI.WeakWeight
	}
	return p
}

// smartOrder puts the most wanted target first; ties go to the lowest HP,
// as in the old game.
func smartOrder(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill, cands []*state.Unit) {
	w := make(map[int]float64, len(cands))
	for _, v := range cands {
		w[v.ID] = pull(c, b, e, s, v)
	}
	sort.SliceStable(cands, func(i, j int) bool {
		if w[cands[i].ID] != w[cands[j].ID] {
			return w[cands[i].ID] > w[cands[j].ID]
		}
		return cands[i].HP < cands[j].HP
	})
}

// tileBonus is how much better standing on (x, y) is for a smart monster:
// the pull of the most wanted foe it could hit from there, and, for a
// cautious one, a penalty for ending within 2 tiles of two or more heroes when it
// could hit nobody.
func tileBonus(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill, fs []*state.Unit, pulls map[int]float64, x, y int) float64 {
	best, inRange, near := 0.0, false, 0
	for _, f := range fs {
		d := abs(f.X-x) + abs(f.Y-y)
		if d >= s.MinRange && d <= s.MaxRange {
			inRange = true
			best = max(best, pulls[f.ID])
		}
		if d <= 2 && f.Team == state.Heroes {
			near++
		}
	}
	bonus := 0.0
	if smart(c, b, e) {
		bonus -= best * c.AI.TileWeight
	}
	if cautious(c, b, e) && !inRange && near >= 2 {
		bonus += c.AI.CautiousPenalty
	}
	return bonus
}
