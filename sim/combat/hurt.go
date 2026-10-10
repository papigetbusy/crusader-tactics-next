package combat

import (
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Result is what one hit did.
type Result struct {
	Missed   bool
	Damage   int            // the damage number (before shields)
	Mult     float64        // above 1: a weak spot; below 1: resisted
	Crit     bool           // critical hit
	Killed   bool           // the target died
	HPAfter  int            // the target's HP afterwards
	Statuses map[string]int // statuses put on the target (name -> turns)
}

// Roll labels, so a failed old-game check can say which roll was out of order.
const (
	RollHit    = "hit roll"
	RollDamage = "damage spread"
	RollCrit   = "crit roll"
)

// Hurt performs one hit: the hit roll, the damage roll, shields, healing
// effects of the skill, statuses and death (old hurt, lines 1687-1710).
// Rewards for a kill (XP, gold, loot) are handled by the caller.
func Hurt(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, sure bool) Result {
	k := c.Combat
	u, v, s := a.By, a.Target, a.Skill
	if !sure {
		if dice.Roll(RollHit)*100 >= float64(HitChance(c, a)) {
			return Result{Missed: true, Mult: 1, HPAfter: v.HP}
		}
	}
	roll := k.DamageRollLow + float64(k.DamageRollSpread*dice.Roll(RollDamage))
	dm, m := Damage(c, b, a, roll)
	res := Result{Damage: dm, Mult: m}
	if u.CritChance > 0 && dice.Roll(RollCrit)*100 < float64(u.CritChance) {
		dm = int(Round(float64(dm) * k.CritDamage))
		res.Damage, res.Crit = dm, true
	}
	hp := dm
	if v.Shield > 0 {
		absorbed := min(v.Shield, hp)
		v.Shield -= absorbed
		hp -= absorbed
	}
	v.HP -= hp

	if s.Drain > 0 {
		share := s.Drain
		if share == 1 {
			share = k.DrainDefault
		}
		heal(u, float64(dm)*share)
	}
	if s.Siphon {
		if w := mostHurtAlly(b, u.Team); w != nil {
			heal(w, float64(dm))
		}
	}

	if v.HP > 0 {
		res.Statuses = applyStatuses(c, a)
	} else {
		res.Killed = true
		b.Remove(v.ID)
		if v.Team == state.Heroes && !v.Summoned {
			b.Fallen = append(b.Fallen, v) // only heroes killed in combat can be revived
		}
	}
	res.HPAfter = v.HP
	return res
}

// applyStatuses puts the skill's statuses on a target that survived. Lightning
// always shocks. A status already there keeps the longer duration.
func applyStatuses(c *content.Content, a Attack) map[string]int {
	fx := map[string]int{}
	for q, n := range a.Skill.Effects {
		fx[q] = n
	}
	el := a.Skill.Element
	if el == content.UserElement || (el != "" && a.By.Team != state.Heroes && a.ClassBase) {
		el = a.By.Element
	}
	if el == "elec" {
		fx["shock"] = max(fx["shock"], c.Combat.ElecShockTurns)
	}
	if a.Target.Status == nil {
		a.Target.Status = map[string]int{}
	}
	for q, n := range fx {
		a.Target.Status[q] = max(a.Target.Status[q], n)
	}
	return fx
}

// heal restores up to n HP (rounded), never above max.
func heal(z *state.Unit, n float64) {
	add := min(z.MaxHP-z.HP, int(Round(n)))
	if add > 0 {
		z.HP += add
	}
}

// mostHurtAlly is the ally with the lowest share of HP who is hurt; ties go
// to the earlier unit, as in the old game's stable sort.
func mostHurtAlly(b *state.Battle, team int) *state.Unit {
	var hurt []*state.Unit
	for _, z := range b.Units {
		if z.Team == team && z.HP < z.MaxHP {
			hurt = append(hurt, z)
		}
	}
	sort.SliceStable(hurt, func(i, j int) bool {
		return float64(hurt[i].HP)/float64(hurt[i].MaxHP) < float64(hurt[j].HP)/float64(hurt[j].MaxHP)
	})
	if len(hurt) == 0 {
		return nil
	}
	return hurt[0]
}

// Forecast is what the screen shows before an attack. It runs the same
// Damage function as a real hit (one math path), at the lowest, middle and
// highest damage roll.
type Forecast struct {
	HitChance  int     `json:"hitChance"`
	Low        int     `json:"low"`
	Mid        int     `json:"mid"`
	High       int     `json:"high"`
	Mult       float64 `json:"mult"`
	CritChance int     `json:"critChance,omitempty"`
	CritLow    int     `json:"critLow,omitempty"`
	CritHi     int     `json:"critHigh,omitempty"`
}

// Preview forecasts an attack without changing anything.
func Preview(c *content.Content, b *state.Battle, a Attack) Forecast {
	k := c.Combat
	lo, m := Damage(c, b, a, k.DamageRollLow)
	mid, _ := Damage(c, b, a, k.DamageRollLow+float64(k.DamageRollSpread*0.5))
	hi, _ := Damage(c, b, a, k.DamageRollLow+k.DamageRollSpread)
	f := Forecast{HitChance: HitChance(c, a), Low: lo, Mid: mid, High: hi, Mult: m, CritChance: a.By.CritChance}
	if f.CritChance > 0 {
		f.CritLow, f.CritHi = int(Round(float64(lo)*k.CritDamage)), int(Round(float64(hi)*k.CritDamage))
	}
	return f
}
