package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// reviveSpots are checked in this order around the healer (old act, line 1715).
var reviveSpots = [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}}

// revive brings back the hero who fell most recently, next to the healer,
// with 40% of max HP, no statuses, finished for this turn. With no free
// spot, nobody comes back.
func revive(c *content.Content, b *state.Battle, u *state.Unit, emit action.Emit) {
	fallen := &b.Fallen
	if u.Team != state.Heroes {
		fallen = &b.FallenFoes // a monster healer brings back a monster
	}
	if len(*fallen) == 0 {
		return
	}
	f := (*fallen)[len(*fallen)-1]
	for _, d := range reviveSpots {
		x, y := u.X+d[0], u.Y+d[1]
		if !b.Inside(x, y) || heightAt(b, x, y) < 0 || b.UnitAt(x, y) != nil {
			continue
		}
		*fallen = (*fallen)[:len(*fallen)-1]
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
			if z.Status["wither"] <= 0 {
				z.Status[q] = max(z.Status[q], n)
			}
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
// is considered (old healAmt and healAmt0, lines 451-452; gear arrives in
// step 6c): skill power, the healer's level, the Saint's bonus, the skill's
// level, and Enlightenment.
func HealAmount(c *content.Content, u *state.Unit, s content.Skill) float64 {
	saint := 1.0
	if a, ok := c.Ascensions[u.Ascension]; ok && a.Heal != 0 {
		saint = a.Heal
	}
	skill := 1.0
	if isHero(u) {
		skill = 1 + float64(c.Combat.SkillLevelBonus*float64(SkillLevel(c, u, s.Name)-1))
	}
	amt := 0 + float64(1*1)*s.Power*float64(6+2*u.Level)*saint*skill
	if u.Status["enl"] > 0 {
		amt *= 1 + u.P("enlP", 0.3)
	}
	return amt
}

// enlCleansed are the debuffs Enlightenment can remove.
var enlCleansed = []string{"bleed", "snare", "stun", "brk", "poison", "weak", "shock", "burn", "ivy", "lock"}

// healSkill is a heal or cleanse (old act, lines 1718-1720). Enlightenment
// spreads it to the whole party and cleanses; a Saint turns overhealing into
// a shield; an Oracle hastes; Plague Bloom hurts and poisons enemies near
// each healed ally. Frozen allies cannot be healed.
func healSkill(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, v *state.Unit, emit action.Emit) {
	list := []*state.Unit{v}
	if s.AllAllies || u.Status["enl"] > 0 {
		list = members(b, u.Team)
	} else if s.Area > 0 {
		list = nil
		for _, z := range b.Units {
			if z.Team == u.Team && Distance(z, v) <= s.Area {
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
		if u.Status["enl"] > 0 && u.P("enlC", 0) != 0 {
			var dq []string
			for _, q := range enlCleansed {
				if z.Status[q] > 0 {
					dq = append(dq, q)
				}
			}
			if u.P("enlC", 0) == 1 && len(dq) > 1 {
				dq = dq[:1]
			}
			for _, q := range dq {
				delete(z.Status, q)
			}
		}
		amt := HealAmount(c, u, s)
		room := z.MaxHP - z.HP
		n := heal(z, amt)
		healed += n
		emit(action.Event{Kind: action.Healed, Unit: u.ID, Target: z.ID, Amount: n})
		if u.Ascension == "Saint" && amt > float64(room) {
			cap, mul := c.Ascensions["Saint"].N("shieldCap"), 1.0
			if u.Status["choir"] > 0 {
				cap, mul = c.Ascensions["Saint"].N("choirShieldCap"), c.Ascensions["Saint"].N("choirShield")
			}
			z.Shield = min(int(Round(float64(z.MaxHP)*cap)), z.Shield+int(Round((amt-float64(room))*mul)))
		}
		if u.Ascension == "Oracle" {
			z.Status["haste"] = max(z.Status["haste"], int(c.Ascensions["Oracle"].N("haste")))
		}
		if u.Status["pbloom"] > 0 {
			pd := max(1, int(Round(amt*0.5)))
			bloom := content.Skill{Name: "Plague Bloom", MaxRange: 9, Power: 1, Element: "dark"}
			for _, e := range snapshot(b) {
				if e.Team != u.Team && Distance(e, z) <= 2 {
					e.Status["poison"] = max(e.Status["poison"], 3)
					Hurt(c, b, dice, Attack{By: u, Skill: bloom, Target: e, Fixed: pd, Sure: true}, emit)
				}
			}
		}
	}
	stats.Gain(c, b, u, stats.HealXP(c, healed), emit) // old act, line 1720
}

// members lists a team's units in board order.
func members(b *state.Battle, team int) []*state.Unit {
	var out []*state.Unit
	for _, z := range b.Units {
		if z.Team == team {
			out = append(out, z)
		}
	}
	return out
}

// snapshot is a copy of the board's unit list, for loops that may kill.
func snapshot(b *state.Battle) []*state.Unit { return append([]*state.Unit(nil), b.Units...) }
