// Package stats is how heroes grow: their numbers at each level, experience
// and level-ups, and evolving into new forms.
package stats

import (
	"math"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Round is JavaScript's Math.round: halves round up (toward +infinity).
// Go's math.Round rounds halves away from zero, which differs for negatives.
func Round(x float64) float64 {
	f := math.Floor(x)
	if x-f >= 0.5 {
		return f + 1
	}
	return f
}

// Numbers are a hero's main numbers.
type Numbers struct {
	MaxHP, Atk, Def, Mov int
}

// Form is which form a hero of this level is in: 0 base, 1 and 2 the
// evolutions (old initP, index.html line 437). Ascensions arrive in step 6b.
func Form(c *content.Content, level int) int {
	f := 0
	for i, at := range c.Progress.FormLevels {
		if level >= at {
			f = i
		}
	}
	return f
}

// FormName is what a hero of this class and level is called ("Vanguard").
func FormName(c *content.Content, class string, level int) string {
	names := c.Progress.FormNames[class]
	if f := Form(c, level); f < len(names) {
		return names[f]
	}
	return class
}

// Hero is a hero's numbers at a level, without gear (old calc, index.html
// lines 736-737). Gear and ascensions are added in later steps.
func Hero(c *content.Content, class string, level int) Numbers {
	p, b := c.Progress, c.Classes[class]
	lv := float64(level - 1)
	n := Numbers{
		MaxHP: int(Round(float64(b.HP) + float64(p.GrowthHP*lv))),
		Atk:   int(Round(float64(b.Atk) + float64(lv*p.GrowthAtk) + p.AtkBonus)),
		Def:   int(Round(float64(b.Def) + float64(lv*p.GrowthDef))),
		Mov:   b.Mov,
	}
	f := p.Forms[Form(c, level)]
	n.MaxHP = int(Round(float64(n.MaxHP) * f.HP))
	n.Atk = int(Round(float64(n.Atk) * f.Atk))
	n.Def += f.Def
	n.Mov += f.Mov
	return n
}

// XPToNext is the experience a hero needs to leave this level.
func XPToNext(c *content.Content, level int) int { return level * c.Progress.XPPerLevel }

// gains is whether this unit earns experience: heroes, not summons.
func gains(u *state.Unit) bool { return u.Team == state.Heroes && !u.Summoned }

// Gain gives a hero experience and levels it up as often as it earns (old
// gain, index.html line 780). A hero at the level cap gains nothing; reaching
// the cap empties the bar.
func Gain(c *content.Content, b *state.Battle, u *state.Unit, n int, emit action.Emit) {
	if !gains(u) || (b.LevelCap > 0 && u.Level >= b.LevelCap) {
		return
	}
	u.XP += max(1, int(Round(float64(n)*c.Progress.XPRate)))
	for u.XP >= XPToNext(c, u.Level) {
		u.XP -= XPToNext(c, u.Level)
		levelUp(c, u, emit)
		if b.LevelCap > 0 && u.Level >= b.LevelCap {
			u.XP = 0
			break
		}
	}
}

// levelUp raises a hero one level: new numbers, the extra max HP is also
// healed, and at levels 5 and 20 the hero evolves.
func levelUp(c *content.Content, u *state.Unit, emit action.Emit) {
	before := Form(c, u.Level)
	u.Level++
	emit(action.Event{Kind: action.LevelUp, Unit: u.ID, Amount: u.Level})
	if f := Form(c, u.Level); f != before {
		emit(action.Event{Kind: action.Evolved, Unit: u.ID, Skill: FormName(c, u.Class, u.Level)})
	}
	n := Hero(c, u.Class, u.Level)
	u.HP += n.MaxHP - u.MaxHP
	u.MaxHP, u.Atk, u.Def, u.Mov = n.MaxHP, n.Atk, n.Def, n.Mov
}

// Share gives the experience for a kill: the killer keeps most of it and the
// other heroes on the board split the rest (old xpShare, index.html line 779).
func Share(c *content.Content, b *state.Battle, u *state.Unit, n int, emit action.Emit) {
	var others []*state.Unit
	for _, z := range b.Units {
		if z != u && gains(z) {
			others = append(others, z)
		}
	}
	k := c.Progress.KillShare
	if k == 0 || len(others) == 0 {
		Gain(c, b, u, n, emit)
		return
	}
	Gain(c, b, u, int(Round(float64(n)*(1-k))), emit)
	each := max(1, int(Round(float64(n)*k/float64(len(others)))))
	for _, z := range others {
		Gain(c, b, z, each, emit)
	}
}

// KillXP is the experience for defeating an enemy.
func KillXP(c *content.Content, v *state.Unit) int {
	p := c.Progress
	n := p.KillBase + v.Level
	switch {
	case v.Boss:
		n *= p.KillBoss
	case v.Elite:
		n *= p.KillElite
	}
	return n
}

// HitXP is the experience for a hit that does not kill.
func HitXP(c *content.Content, damage int, putsStatuses bool) int {
	p := c.Progress
	n := p.HitBase + int(Round(float64(damage)/p.HitPerDamage))
	if putsStatuses {
		n += p.HitStatusBonus
	}
	return n
}

// HealXP is the experience for healing: more for bigger heals, up to a limit.
func HealXP(c *content.Content, healed int) int {
	p := c.Progress
	return p.HealBase + min(p.HealMax, int(Round(float64(healed)/p.HealPer)))
}
