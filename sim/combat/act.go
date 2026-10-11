package combat

import (
	"fmt"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
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
	if s.Summon != nil {
		return fmt.Errorf("%s: summons arrive with the Necromancer in Phase 6", s.Name)
	}
	_, classBase, _ := c.SkillFor(u.Team, u.Class, s.Name)
	emit(action.Event{Kind: action.SkillUsed, Unit: u.ID, Target: v.ID, Skill: s.Name})
	switch {
	case s.Revive:
		revive(c, b, u, emit)
	case s.Support:
		support(c, b, u, s, emit)
	case s.Heal:
		healSkill(c, b, u, s, v, emit)
	default:
		attack(c, b, dice, Attack{By: u, Skill: s, ClassBase: classBase, Target: v}, emit)
	}
	finishSkill(c, u, slot, s)
	return nil
}

// reviveSpots are checked in this order around the healer (old act, line 1715).
var reviveSpots = [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}}

// revive brings back the hero who fell most recently, next to the healer,
// with 40% of max HP, no statuses, finished for this turn. With no free
// spot, nobody comes back.
func revive(c *content.Content, b *state.Battle, u *state.Unit, emit action.Emit) {
	if len(b.Fallen) == 0 {
		return
	}
	f := b.Fallen[len(b.Fallen)-1]
	for _, d := range reviveSpots {
		x, y := u.X+d[0], u.Y+d[1]
		if !b.Inside(x, y) || heightAt(b, x, y) < 0 || b.UnitAt(x, y) != nil {
			continue
		}
		b.Fallen = b.Fallen[:len(b.Fallen)-1]
		f.X, f.Y = x, y
		f.HP = int(Round(float64(f.MaxHP) * c.Combat.ReviveShare))
		f.Status = map[string]int{}
		f.Done = true
		b.Units = append(b.Units, f)
		emit(action.Event{Kind: action.Revived, Unit: u.ID, Target: f.ID, Amount: f.HP})
		return
	}
}

// heightAt is the map height (below 0 is water or mountain).
func heightAt(b *state.Battle, x, y int) int {
	if y < len(b.Heights) && x < len(b.Heights[y]) {
		return b.Heights[y][x]
	}
	return 0
}

func support(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, emit action.Emit) {
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
	stats.Gain(c, b, u, c.Progress.SupportXP, emit) // old act, line 1717
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
	healed := 0
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
		healed += z.HP - before
		emit(action.Event{Kind: action.Healed, Unit: u.ID, Target: z.ID, Amount: z.HP - before})
	}
	stats.Gain(c, b, u, stats.HealXP(c, healed), emit) // old act, line 1720
}

func attack(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, emit action.Emit) {
	u, s, v := a.By, a.Skill, a.Target
	if u.Status["cloak"] > 0 { // attacking breaks the cloak
		u.Status["cloak"] = 0
		CloakCooldown(u)
	}
	targets := []*state.Unit{v}
	if s.ExtraTargets > 0 {
		var pool []*state.Unit
		for _, z := range b.Units {
			if z.Team != u.Team && z != v && Distance(u, z) <= s.MaxRange+u.RangeBonus {
				pool = append(pool, z)
			}
		}
		if p, ok := dice.(rng.Picker); ok {
			ids := make([]int, len(pool))
			for i, z := range pool {
				ids[i] = z.ID
			}
			for _, id := range p.PickExtra(ids, s.ExtraTargets-1) {
				targets = append(targets, b.Unit(id))
			}
		} else {
			rng.Shuffle(dice, RollShuffle, pool)
			targets = append(targets, pool[:min(len(pool), s.ExtraTargets-1)]...)
		}
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
		for _, z := range areaOf(b, u, s, v, targets) {
			strike(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z}, emit)
		}
	}
	if s.Line > 0 {
		for _, t := range lineOf(u, s, v) {
			if z := b.UnitAt(t[0], t[1]); z != nil && z.Team != u.Team {
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

// areaOf are the other enemies caught in an area skill's blast around v.
func areaOf(b *state.Battle, u *state.Unit, s content.Skill, v *state.Unit, targets []*state.Unit) []*state.Unit {
	var near []*state.Unit
	for _, z := range b.Units {
		if z.Team != u.Team && !slices.Contains(targets, z) && Distance(z, v) <= s.Area {
			near = append(near, z)
		}
	}
	return near
}

// lineOf are the tiles a piercing skill continues through behind v.
func lineOf(u *state.Unit, s content.Skill, v *state.Unit) [][2]int {
	dx, dy := 0, sign(v.Y-u.Y)
	if abs(v.X-u.X) >= abs(v.Y-u.Y) {
		dx, dy = sign(v.X-u.X), 0
	}
	var out [][2]int
	for i := 1; i <= s.Line; i++ {
		out = append(out, [2]int{v.X + dx*i, v.Y + dy*i})
	}
	return out
}

// AlsoHit lists the other units an area or piercing skill aimed at v would
// hit as well, for the forecast (random extra targets are not known ahead).
func AlsoHit(b *state.Battle, u *state.Unit, s content.Skill, v *state.Unit) []int {
	var out []int
	if s.Area > 0 {
		for _, z := range areaOf(b, u, s, v, []*state.Unit{v}) {
			out = append(out, z.ID)
		}
	}
	for _, t := range lineOf(u, s, v) {
		if z := b.UnitAt(t[0], t[1]); z != nil && z.Team != u.Team {
			out = append(out, z.ID)
		}
	}
	return out
}

// strike is one Hurt plus the events that describe it.
func strike(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, emit action.Emit) {
	r := Hurt(c, b, dice, a, false)
	switch {
	case r.Missed:
		emit(action.Event{Kind: action.Missed, Unit: a.By.ID, Target: a.Target.ID})
		stats.Gain(c, b, a.By, c.Progress.MissXP, emit) // old hurt, line 1689
		return
	default:
		emit(action.Event{Kind: action.Hit, Unit: a.By.ID, Target: a.Target.ID, Amount: r.Damage, Mult: r.Mult, Crit: r.Crit})
	}
	if len(r.Statuses) > 0 {
		emit(action.Event{Kind: action.StatusOn, Target: a.Target.ID, Statuses: r.Statuses})
	}
	if r.Killed {
		emit(action.Event{Kind: action.Died, Unit: a.Target.ID})
		if a.By.Team == state.Heroes { // old hurt, line 1707
			stats.Share(c, b, a.By, stats.KillXP(c, a.Target), emit)
		}
	} else {
		stats.Gain(c, b, a.By, stats.HitXP(c, r.Damage, len(a.Skill.Effects) > 0), emit)
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
		u.CloakUsed = true
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

// CloakCooldown starts the Cloak skill's 2-turn cooldown once the cloak is
// over (old cloakCd, index.html line 1094).
func CloakCooldown(u *state.Unit) {
	for i, s := range u.Skills {
		if s.Cloak {
			u.Cooldowns[i] = 2
			break
		}
	}
	u.CloakUsed = false
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
