package combat

import (
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Result is what one hit did.
type Result struct {
	Missed   bool
	Immune   bool           // Glacial Scales: ranged hits do nothing
	Target   *state.Unit    // who was really hit (a Guardian can take the hit)
	Damage   int            // the damage number (before shields)
	Mult     float64        // above 1: a weak spot; below 1: resisted
	Crit     bool           // critical hit
	Killed   bool           // the target died
	HPAfter  int            // the target's HP afterwards
	Statuses map[string]int // statuses put on the target (name -> turns)
}

// Roll labels, so a failed old-game check can say which roll was out of order.
const (
	RollHit       = "hit roll"
	RollDamage    = "damage spread"
	RollCrit      = "crit roll"
	RollBeastBurn = "beast burn roll"
)

// Hurt performs one hit and everything that comes with it (old hurt, lines
// 1687-1710): a Guardian taking the hit, the hit roll, damage, crits,
// shields, a Sacred Vow sharing the damage, life steal and healing effects,
// statuses, death and its triggers, experience, and a Duelist's Riposte.
func Hurt(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, emit action.Emit) Result {
	k := c.Combat
	u, s := a.By, a.Skill
	a = bodyguard(b, a, emit)
	if u.Status["aim"] > 0 && !s.Friendly() {
		a.Sure = true // Steady Aim
	}
	if a.Target.Team == state.Heroes && a.Target.Ascension != "Guardian" && !a.NoRedirect {
		for _, z := range b.Units {
			if z.Team == state.Heroes && z.Ascension == "Guardian" && z != a.Target && Distance(z, a.Target) == 1 {
				emit(action.Event{Kind: action.Guarded, Unit: z.ID, Target: a.Target.ID})
				a.Target, a.Extra = z, extraOf(a)*c.Ascensions["Guardian"].N("bodyguard")
				break
			}
		}
	}
	v := a.Target
	fixed := a.Fixed > 0
	hc := 100 // the old game rolls the hit dice even for hits that cannot miss
	if !a.Sure && !fixed {
		hc = HitChance(c, b, a)
	}
	if dice.Roll(RollHit)*100 >= float64(hc) {
		emit(action.Event{Kind: action.Missed, Unit: u.ID, Target: v.ID})
		stats.Gain(c, b, u, c.Progress.MissXP, emit)
		return Result{Missed: true, Target: v, Mult: 1, HPAfter: v.HP}
	}
	res := Result{Target: v, Mult: 1, Damage: a.Fixed}
	if !fixed {
		roll := k.DamageRollLow + float64(k.DamageRollSpread*dice.Roll(RollDamage))
		res.Damage, res.Mult = Damage(c, b, a, roll)
		if v.Status["scales"] > 0 && Distance(u, v) > 1 {
			emit(action.Event{Kind: action.Immune, Unit: u.ID, Target: v.ID})
			res.Immune, res.Damage, res.HPAfter = true, 0, v.HP
			return res
		}
	}
	cc := critChance(u, a)
	if u.Status["coin"] > 0 && !fixed && !a.Counter {
		cc = 100 // Lucky Coin
		u.Status["coin"]--
	}
	if cc > 0 && dice.Roll(RollCrit)*100 < float64(cc) {
		res.Damage = int(Round(float64(res.Damage) * critDamage(c, u)))
		res.Crit = true
	}
	dm := res.Damage
	hp := dm
	if v.Shield > 0 {
		absorbed := min(v.Shield, hp)
		v.Shield -= absorbed
		hp -= absorbed
	}
	if v.Team == state.Heroes && hp > 1 && !a.Counter {
		for _, z := range b.Units {
			if z.Team == state.Heroes && z != v && z.Status["svow"] > 0 && Distance(z, v) <= 2 {
				half := hp / 2
				hp -= half
				z.HP = max(1, z.HP-half)
				emit(action.Event{Kind: action.Shared, Unit: z.ID, Target: v.ID, Amount: half})
				break
			}
		}
	}
	v.HP -= hp
	if v.HP <= 0 && v.Status["unb"] > 0 && !v.UnbrokenUsed {
		v.HP, v.UnbrokenUsed = 1, true
		emit(action.Event{Kind: action.Unbroken, Unit: v.ID})
	}
	if v.HP <= 0 && v.Status["lastlight"] > 0 {
		v.HP = 1 // Last Light
		emit(action.Event{Kind: action.Unbroken, Unit: v.ID})
	}
	emit(action.Event{Kind: action.Hit, Unit: u.ID, Target: v.ID, Amount: dm, Mult: res.Mult, Crit: res.Crit})
	addedElements(c, a, v, dm, emit)
	if u.Status["beast"] > 0 && !fixed {
		ex := max(1, int(Round(float64(float64(dm)*u.P("bmF", 0.25))*Multiplier(c, unarmed(u), v, "fire"))))
		v.HP -= ex
		emit(action.Event{Kind: action.Hit, Unit: u.ID, Target: v.ID, Amount: ex, Skill: "fire"})
	}
	afterHit(c, b, a, v, dm, emit)

	if v.HP > 0 {
		res.Statuses = applyStatuses(c, b, dice, a, v)
		if len(res.Statuses) > 0 {
			emit(action.Event{Kind: action.StatusOn, Target: v.ID, Statuses: res.Statuses})
		}
	} else {
		res.Killed = true
		died(c, b, dice, u, v, emit)
	}
	if !res.Killed && u.Team == state.Heroes {
		stats.Gain(c, b, u, stats.HitXP(c, dm, len(s.Effects) > 0), emit)
	}
	res.HPAfter = v.HP
	strikeBack(c, b, dice, a, v, dm, emit)
	if v.HP > 0 && v.Ascension == "Duelist" && !a.Counter && Distance(u, v) == 1 && b.Unit(u.ID) != nil {
		riposte := Attack{By: v, Skill: c.Skills["Slash"], Target: u, Counter: true, NoRedirect: true,
			Extra: c.Ascensions["Duelist"].N("riposte")}
		emit(action.Event{Kind: action.Riposte, Unit: v.ID, Target: u.ID})
		Hurt(c, b, dice, riposte, emit)
		if v.Status["flurry"] > 0 && b.Unit(u.ID) != nil {
			Hurt(c, b, dice, riposte, emit)
		}
	}
	return res
}

func extraOf(a Attack) float64 {
	if a.Extra == 0 {
		return 1
	}
	return a.Extra
}

// unarmed is the unit without its weapon, for added elemental damage (the
// old game called mult with w:null).
func unarmed(u *state.Unit) *state.Unit {
	cp := *u
	cp.Weapon = ""
	return &cp
}

// critChance is the chance in % of a critical hit: Deadeye always crits;
// gear, Bloodlust and Flurry of Blades add chance.
func critChance(u *state.Unit, a Attack) int {
	if a.Fixed > 0 {
		return 0
	}
	if u.Status["deadeye"] > 0 {
		return 100
	}
	cc := u.Af("crit") + a.CritBonus
	if u.Status["blood"] > 0 {
		cc = u.Af("crit") + int(u.P("blC", 50)) + a.CritBonus
	}
	return cc
}

// critDamage is how much a critical hit multiplies the damage.
func critDamage(c *content.Content, u *state.Unit) float64 {
	if u.Status["deadeye"] > 0 {
		return 3
	}
	m := c.Combat.CritDamage + float64(u.Af("critDmg"))/100
	if u.Status["blood"] > 0 {
		m += u.P("blD", 0) / 100
	}
	return m
}

// afterHit is the healing and snaring that follows a hit (old hurt, lines
// 1697-1700): life steal, a taunting Paladin's heal, Frost Bulwark and
// Glacial Scales rooting the attacker, drain and siphon skills, and the
// Paladin, Blood Witch and Plague Doctor passives.
func afterHit(c *content.Content, b *state.Battle, a Attack, v *state.Unit, dm int, emit action.Emit) {
	u, s := a.By, a.Skill
	if u.Af("leech") != 0 {
		heal(u, float64(dm)*float64(u.Af("leech"))/100)
	}
	if u.Status["frenzy"] > 0 {
		heal(u, float64(dm)*0.25)
	}
	if u.Status["vamp"] > 0 {
		heal(u, float64(dm)*0.2)
	}
	if v.Status["taunt"] > 0 && v.HP > 0 {
		heal(v, float64(v.MaxHP)*0.1)
	}
	if u.Team == state.Enemies && v.Team == state.Heroes && near(b, v, 2, true, func(z *state.Unit) bool { return z.Status["bulwark"] > 0 }) {
		u.Status["snare"] = max(u.Status["snare"], 1)
	}
	if v.Status["scales"] > 0 && Distance(u, v) <= 1 {
		u.Status["snare"] = max(u.Status["snare"], 2)
	}
	thorns(b, u, v, dm, emit)
	if s.Drain > 0 {
		share := s.Drain
		if share == 1 {
			share = c.Combat.DrainDefault
		}
		heal(u, float64(dm)*share)
	}
	switch u.Ascension {
	case "Paladin", "Blood Witch":
		heal(u, float64(dm)*c.Ascensions[u.Ascension].N("lifeSteal"))
	}
	if s.Siphon {
		if w := mostHurtAlly(b, u.Team); w != nil {
			heal(w, float64(dm))
		}
	}
	if u.Ascension == "Plague Doctor" {
		if w := mostHurtAlly(b, state.Heroes); w != nil {
			heal(w, float64(dm)*c.Ascensions["Plague Doctor"].N("healShare"))
		}
	}
}

// heal restores up to n HP (rounded), never above max.
func heal(z *state.Unit, n float64) int {
	if z.Status["wither"] > 0 {
		n /= 2 // Wither
	}
	add := min(z.MaxHP-z.HP, int(Round(n)))
	if add > 0 {
		z.HP += add
		return add
	}
	return 0
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
