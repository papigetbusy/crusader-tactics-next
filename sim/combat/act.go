package combat

import (
	"fmt"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// RollShuffle labels the rolls of the fair shuffle that picks the extra
// targets of multi-target skills (approved difference from the old game).
const RollShuffle = "shuffle"

// UseSkill is one unit using the skill in a slot of its bar on a target:
// buffs, heals and attacks, including area, multi-target, multi-hit and
// piercing skills (old act, index.html lines 1711-1734). It assumes the
// caller already checked the use is allowed.
//
// Not yet ported (they arrive with their phases): summons and revives,
// ascension passives, class-ability statuses, XP, gold and loot.
func UseSkill(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, slot int, v *state.Unit, emit action.Emit) error {
	s := u.Skills[slot]
	if s.Summon != nil || s.Revive {
		return fmt.Errorf("%s: summons and revives arrive in a later phase", s.Name)
	}
	_, classBase, _ := c.SkillFor(u.Team, u.Class, s.Name)
	emit(action.Event{Kind: action.SkillUsed, Unit: u.ID, Target: v.ID, Skill: s.Name})
	switch {
	case s.Support:
		support(b, u, s, emit)
	case s.Heal:
		healSkill(c, b, u, s, v, emit)
	default:
		attack(c, b, dice, Attack{By: u, Skill: s, ClassBase: classBase, Target: v}, emit)
	}
	finishSkill(c, u, slot, s)
	return nil
}

func support(b *state.Battle, u *state.Unit, s content.Skill, emit action.Emit) {
	for _, z := range b.Units {
		if z.Team != u.Team || Distance(z, u) > s.Area {
			continue
		}
		for q, n := range s.Buffs {
			z.Status[q] = max(z.Status[q], n)
		}
		if s.ResetCooldown && z != u {
			for i := range z.Cooldowns {
				z.Cooldowns[i] = 0
			}
		}
		emit(action.Event{Kind: action.StatusOn, Target: z.ID, Statuses: s.Buffs})
	}
}

// HealAmount is how much a heal skill restores before the target's room
// is considered (old healAmt0, line 451, for units without gear).
func HealAmount(c *content.Content, u *state.Unit, s content.Skill) float64 {
	skill := 1.0
	if isHero(u) {
		skill = 1 + float64(c.Combat.SkillLevelBonus*float64(SkillLevel(c, u, s.Name)-1))
	}
	return float64(1*1) * s.Power * float64(6+2*u.Level) * 1 * skill
}

func healSkill(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, v *state.Unit, emit action.Emit) {
	list := []*state.Unit{v}
	if s.AllAllies || s.Area > 0 {
		list = nil
		for _, z := range b.Units {
			if z.Team == u.Team && (s.AllAllies || Distance(z, v) <= s.Area) {
				list = append(list, z)
			}
		}
	}
	for _, z := range list {
		if z.Status["frozen"] > 0 {
			continue
		}
		if s.Cleanse {
			for _, q := range cleansable {
				delete(z.Status, q)
			}
		}
		before := z.HP
		heal(z, HealAmount(c, u, s))
		emit(action.Event{Kind: action.Healed, Unit: u.ID, Target: z.ID, Amount: z.HP - before})
	}
}

func attack(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, emit action.Emit) {
	u, s, v := a.By, a.Skill, a.Target
	targets := []*state.Unit{v}
	if s.ExtraTargets > 0 {
		var pool []*state.Unit
		for _, z := range b.Units {
			if z.Team != u.Team && z != v && Distance(u, z) <= s.MaxRange+u.RangeBonus {
				pool = append(pool, z)
			}
		}
		rng.Shuffle(dice, RollShuffle, pool)
		targets = append(targets, pool[:min(len(pool), s.ExtraTargets-1)]...)
	}
	hits := max(1, s.Hits)
	for _, z := range targets {
		for h := 0; h < hits; h++ {
			if b.Unit(z.ID) == nil {
				break
			}
			strike(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z}, emit)
		}
	}
	if s.Area > 0 {
		var near []*state.Unit
		for _, z := range b.Units {
			if z.Team != u.Team && !slices.Contains(targets, z) && Distance(z, v) <= s.Area {
				near = append(near, z)
			}
		}
		for _, z := range near {
			strike(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z}, emit)
		}
	}
	if s.Line > 0 {
		dx, dy := 0, sign(v.Y-u.Y)
		if abs(v.X-u.X) >= abs(v.Y-u.Y) {
			dx, dy = sign(v.X-u.X), 0
		}
		for i := 1; i <= s.Line; i++ {
			if z := b.UnitAt(v.X+dx*i, v.Y+dy*i); z != nil && z.Team != u.Team {
				strike(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z}, emit)
			}
		}
	}
	if s.Splash > 0 {
		for _, z := range b.Units {
			if z.Team == u.Team && Distance(z, u) <= 1 {
				heal(z, float64(z.MaxHP)*s.Splash)
			}
		}
	}
}

// strike is one Hurt plus the events that describe it.
func strike(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, emit action.Emit) {
	r := Hurt(c, b, dice, a, false)
	switch {
	case r.Missed:
		emit(action.Event{Kind: action.Missed, Unit: a.By.ID, Target: a.Target.ID})
		return
	default:
		emit(action.Event{Kind: action.Hit, Unit: a.By.ID, Target: a.Target.ID, Amount: r.Damage, Mult: r.Mult, Crit: r.Crit})
	}
	if len(r.Statuses) > 0 {
		emit(action.Event{Kind: action.StatusOn, Target: a.Target.ID, Statuses: r.Statuses})
	}
	if r.Killed {
		emit(action.Event{Kind: action.Died, Unit: a.Target.ID})
	}
}

// finishSkill is what every skill use does afterwards: heroes' skills gain
// uses (2 when switching skills), cooldowns start, the unit is done.
func finishSkill(c *content.Content, u *state.Unit, slot int, s content.Skill) {
	if isHero(u) {
		if u.SkillUses == nil {
			u.SkillUses = map[string]int{}
		}
		add := 1
		if u.LastSkill != "" && u.LastSkill != s.Name {
			add = 2
		}
		u.SkillUses[s.Name] += add
		u.LastSkill = s.Name
	}
	if s.Cloak {
		u.Status["cloak"] = 3
	}
	if s.LockSelf {
		u.Status["lock"] = 3
	}
	if s.Cooldown > 0 {
		u.Cooldowns[slot] = max(1, s.Cooldown)
	}
	u.Done = true
	if s.Cloak {
		u.Cooldowns[slot] = 99
	}
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}
