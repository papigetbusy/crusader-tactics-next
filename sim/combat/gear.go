package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Roll labels for gear.
const (
	RollStatusChance = "gear status chance"
	RollDouble       = "gear double strike"
)

// isMagic is a spell: a magic skill, or Strike with a staff or wand in hand
// (old isMag, line 914).
func isMagic(c *content.Content, u *state.Unit, s content.Skill) bool {
	return s.IsMagic() || (s.Strike && stats.HandCategory(c, u) == "g")
}

// strikeRange is how far Strike reaches with the weapon in hand: melee 1
// tile, bows, guns, staffs and wands 2-6 (old tgt, line 741).
func strikeRange(c *content.Content, u *state.Unit) (int, int) {
	if stats.HandCategory(c, u) == "m" {
		return 1, 1
	}
	return 2, 6
}

// rankDamage is the weapon rank's damage multiplier for the weapon in hand
// (old dcalc, line 445).
func rankDamage(c *content.Content, u *state.Unit) float64 {
	if rk := stats.SetRank(c, u); rk >= 0 {
		return c.Gear.RankDamage[rk]
	}
	return 1
}

// gearDamage is the gear part of the hero damage formula (old dcalc, line
// 447): more damage, more damage of the hit's element (or the hero's own),
// more physical damage, more damage to elites and bosses.
func gearDamage(c *content.Content, a Attack, el string, atk float64) float64 {
	u, v := a.By, a.Target
	ee := el
	if el == content.UserElement || el == "" {
		ee = u.Element
	}
	atk *= 1 + float64(u.Af("dmg"))/100
	if ee != "" && u.Af(ee) != 0 {
		atk *= 1 + float64(u.Af(ee))/100
	}
	if !isMagic(c, u, a.Skill) && u.Af("phys") != 0 {
		atk *= 1 + float64(u.Af("phys"))/100
	}
	if u.Af("bossDmg") != 0 && (v.Boss || v.Elite) {
		atk *= 1 + float64(u.Af("bossDmg"))/100
	}
	return atk
}

// addedElements is gear that adds part of a hit's damage as fire or thunder,
// with its own element multiplier and no weapon triangle (old hurt, line
// 1695).
func addedElements(c *content.Content, a Attack, v *state.Unit, dm int, emit action.Emit) {
	u := a.By
	for _, key := range []string{"addFire", "addElec"} {
		if u.Af(key) == 0 {
			continue
		}
		el := c.Gear.AddedElements[key]
		ex := max(1, int(Round(float64(float64(dm)*float64(u.Af(key))/100)*Multiplier(c, unarmed(u), v, el))))
		v.HP -= ex
		emit(action.Event{Kind: action.Hit, Unit: u.ID, Target: v.ID, Amount: ex, Skill: el})
	}
}

// thorns hurts an attacker for part of the damage it dealt (old hurt, line
// 1697); thorns never kill.
func thorns(b *state.Battle, u, v *state.Unit, dm int, emit action.Emit) {
	if v.Af("thorns") == 0 || u == v || b.Unit(u.ID) == nil {
		return
	}
	if th := int(Round(float64(dm) * float64(v.Af("thorns")) / 100)); th > 0 {
		u.HP = max(1, u.HP-th)
		emit(action.Event{Kind: action.Hit, Unit: v.ID, Target: u.ID, Amount: th, Skill: "thorns"})
	}
}

// gearStatuses rolls the gear's chances to bleed, poison or stun (old hurt,
// line 1702), in that order.
func gearStatuses(c *content.Content, dice rng.Dice, u *state.Unit, fx map[string]int) {
	for _, key := range []string{"bleedCh", "poisonCh", "stunCh"} {
		if u.Af(key) != 0 && dice.Roll(RollStatusChance)*100 < float64(u.Af(key)) {
			st := c.Gear.StatusChances[key]
			fx[st.Key] = max(fx[st.Key], st.Value)
		}
	}
}

// doubleStrike is gear's chance to hit the main target once more at 60%
// (old act, line 1722).
func doubleStrike(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, v *state.Unit, emit action.Emit) {
	u := a.By
	if u.Af("double") == 0 || v == nil || b.Unit(v.ID) == nil {
		return
	}
	if dice.Roll(RollDouble)*100 < float64(u.Af("double")) {
		Hurt(c, b, dice, Attack{By: u, Skill: a.Skill, ClassBase: a.ClassBase, Target: v, Extra: c.Gear.DoubleStrike}, emit)
	}
}
