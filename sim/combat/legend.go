package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// The class legendaries' unique powers and 4-piece class bonuses (approved
// 2026-10-17, docs/legendary-proposal.md). Every power is a bonus kind on
// the weapon (u.Af), so nothing here happens without the gear.

// Roll labels for the powers.
const (
	RollRetaliate = "retaliate roll"
	RollPin       = "pinning shot roll"
	RollEcho      = "arcane echo roll"
)

func pct(n int) float64 { return float64(n) / 100 }

func pn(c *content.Content, key string) int { return int(c.Gear.PowerNumber(key)) }

// legendAttack is how much harder a hit lands because of the attacker's
// powers: Battle Rhythm, Charge, Ambush, the Ranger's long shots and the
// Tank's counterattacks. 1 when none apply.
func legendAttack(c *content.Content, a Attack) float64 {
	u, v := a.By, a.Target
	if u.Team == v.Team {
		return 1
	}
	m := 1.0
	if n := u.Af("rhythm"); n > 0 && u.Rhythm > 0 {
		m *= 1 + pct(n*min(u.Rhythm, pn(c, "rhythmStacks")))
	}
	if n := u.Af("charge"); n > 0 && u.TilesMoved > 0 {
		m *= 1 + pct(min(n*u.TilesMoved, pn(c, "chargeCap")))
	}
	if n := u.Af("ambush"); n > 0 && v.HP >= v.MaxHP {
		m *= 1 + pct(n)
	}
	if n := u.Af("farShot"); n > 0 && Distance(u, v) >= pn(c, "farTiles") {
		m *= 1 + pct(n)
	}
	if n := u.Af("counterBoost"); n > 0 && a.Counter {
		m *= 1 + pct(n)
	}
	return m
}

// legendTaken is how much of a hit the target takes because of powers:
// Unbowed when low, and a Bulwark Aura on an ally next to it. 1 when none.
func legendTaken(c *content.Content, b *state.Battle, a Attack) float64 {
	u, v := a.By, a.Target
	if u.Team == v.Team {
		return 1
	}
	m := 1.0
	if n := v.Af("unbowed"); n > 0 && v.HP*100 < v.MaxHP*pn(c, "unbowedBelow") {
		m *= 1 - pct(n)
	}
	aura := 0
	for _, z := range b.Units {
		if z.Team == v.Team && z != v && Distance(z, v) == 1 {
			aura = max(aura, z.Af("aura"))
		}
	}
	return m * (1 - pct(aura))
}

// firstWard is the Knight's 4-piece bonus: the first hit from an enemy each
// battle does nothing. It reports whether this hit was that one.
func firstWard(a Attack, emit action.Emit) bool {
	u, v := a.By, a.Target
	if u.Team == v.Team || v.Af("firstWard") == 0 || v.Used("firstWard") {
		return false
	}
	v.Use("firstWard")
	emit(action.Event{Kind: action.Immune, Unit: u.ID, Target: v.ID})
	return true
}

// opener is the Assassin's 4-piece bonus: the first attack each battle
// that lands always crits.
func opener(a Attack) bool {
	u := a.By
	if u.Af("opener") == 0 || u.Used("opener") || a.Counter || a.Fixed > 0 || u.Team == a.Target.Team {
		return false
	}
	u.Use("opener")
	return true
}

// critRefund is the Gunslinger's 4-piece bonus: a crit takes a turn off
// every cooldown, the class ability's too.
func critRefund(u *state.Unit) {
	if n := u.Af("critRefund"); n > 0 {
		for i, cd := range u.Cooldowns {
			if cd > 0 && cd < 99 {
				u.Cooldowns[i] = max(0, cd-n)
			}
		}
		u.AbilityCooldown = max(0, u.AbilityCooldown-n)
	}
}

// lifeline is the Healer's Lifeline: once per battle, an ally (or the
// Healer) who would fall stays at 1 HP.
func lifeline(b *state.Battle, v *state.Unit, emit action.Emit) {
	if v.HP > 0 {
		return
	}
	for _, z := range b.Units {
		if z.Team == v.Team && z.Af("lifeline") > 0 && !z.Used("lifeline") {
			z.Use("lifeline")
			v.HP = 1
			emit(action.Event{Kind: action.Unbroken, Unit: v.ID, Skill: "Lifeline"})
			return
		}
	}
}

// dragonblood adds part of a hit as fire (the Dragon Knight's Dragonblood),
// with fire's own multiplier and no weapon triangle.
func dragonblood(c *content.Content, a Attack, v *state.Unit, dm int, emit action.Emit) {
	u := a.By
	if n := u.Af("dragonblood"); n > 0 && a.Fixed == 0 && u.Team != v.Team {
		ex := max(1, int(Round(float64(dm)*pct(n)*Multiplier(c, unarmed(u), v, "fire"))))
		v.HP -= ex
		emit(action.Event{Kind: action.Hit, Unit: u.ID, Target: v.ID, Amount: ex, Skill: "fire"})
	}
}

// legendStatuses are the statuses powers add to a hit: Dragonblood's burn
// and a Pinning Shot's root.
func legendStatuses(c *content.Content, dice rng.Dice, u *state.Unit, fx map[string]int) {
	if u.Af("dragonblood") > 0 {
		fx["burn"] = max(fx["burn"], pn(c, "dragonBurn"))
	}
	if n := u.Af("pin"); n > 0 && dice.Roll(RollPin)*100 < float64(n) {
		fx["snare"] = max(fx["snare"], pn(c, "pinTurns"))
	}
}

// shrugs reports a status that never sticks: Wyrm's Hide stops burns and
// freezes.
func shrugs(v *state.Unit, q string) bool {
	return v.Af("wyrmhide") > 0 && (q == "burn" || q == "frozen")
}

// forTurns is how long a status must be set to last n of the unit's own
// turns: statuses count down at the end of their side's turn, so one put on
// during the unit's own turn needs one more.
func forTurns(b *state.Battle, z *state.Unit, n int) int {
	if b.Side == z.Team {
		return n + 1
	}
	return n
}

// afterLegendHit is what powers do once a hit has landed and its statuses
// are on: Battle Rhythm counts the hit, Retaliate may stun the attacker,
// Ricochet bounces part of the damage, and Arcane Echo may hit again.
func afterLegendHit(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, v *state.Unit, dm int, emit action.Emit) {
	u := a.By
	if u.Team == v.Team {
		return
	}
	if u.Af("rhythm") > 0 {
		u.Rhythm = min(u.Rhythm+1, pn(c, "rhythmStacks"))
	}
	if n := v.Af("retaliate"); n > 0 && v.HP > 0 && b.Unit(u.ID) != nil && dice.Roll(RollRetaliate)*100 < float64(n) {
		putStatus(u, "stun", forTurns(b, u, pn(c, "retaliateStun")), emit)
	}
	if n := u.Af("ricochet"); n > 0 && !a.Bounce && a.Fixed == 0 && dm > 0 {
		if z := bounceTarget(c, b, u, v); z != nil {
			ex := max(1, int(Round(float64(dm)*pct(n))))
			z.HP -= ex
			emit(action.Event{Kind: action.Hit, Unit: u.ID, Target: z.ID, Amount: ex, Skill: "Ricochet"})
			lifeline(b, z, emit)
			if z.HP <= 0 {
				died(c, b, dice, u, z, emit)
			}
		}
	}
	if n := u.Af("echo"); n > 0 && !a.Echo && !a.Counter && a.Fixed == 0 && isMagic(c, u, a.Skill) &&
		b.Unit(v.ID) != nil && b.Unit(u.ID) != nil && dice.Roll(RollEcho)*100 < float64(n) {
		again := a
		again.Echo, again.Target = true, v
		emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: "Arcane Echo"})
		Hurt(c, b, dice, again, emit)
	}
}

// bounceTarget is the enemy a Ricochet jumps to: the nearest other enemy
// within range of the target, the first one on the board on a tie.
func bounceTarget(c *content.Content, b *state.Battle, u, v *state.Unit) *state.Unit {
	var best *state.Unit
	for _, z := range b.Units {
		if z.Team == u.Team || z == v || z.HP <= 0 || Distance(z, v) > pn(c, "ricochetRange") {
			continue
		}
		if best == nil || Distance(z, v) < Distance(best, v) {
			best = z
		}
	}
	return best
}

// onKillPowers are the powers a kill sets off: the Warrior's 4-piece bonus
// resets the class ability, and Shadowstep cloaks the Assassin until its
// next turn.
func onKillPowers(b *state.Battle, u *state.Unit) {
	if u.Af("killReset") > 0 {
		u.AbilityCooldown = 0
	}
	if u.Af("shadowstep") > 0 && u.Status["cloak"] <= 0 {
		u.Status["cloak"] = forTurns(b, u, 1)
		u.Shadowstep = true
	}
}

// statusPowers are the powers that act on statuses a unit causes: Hex
// Mastery lengthens them (n is changed in place by the caller), Soul
// Siphon heals per status, and the Witch's 4-piece bonus spreads them to
// one enemy next to the target.
func statusPowers(c *content.Content, b *state.Battle, a Attack, v *state.Unit, put map[string]int, emit action.Emit) {
	u := a.By
	if u.Team == v.Team || len(put) == 0 {
		return
	}
	if n := u.Af("siphon"); n > 0 {
		heal(u, float64(u.MaxHP)*pct(n)*float64(len(put)))
	}
	if u.Af("hexSpread") == 0 || a.Spread {
		return
	}
	for _, z := range b.Units {
		if z.Team != v.Team || z == v || z.HP <= 0 || Distance(z, v) != 1 {
			continue
		}
		keys := make([]string, 0, len(put))
		for q := range put {
			keys = append(keys, q)
		}
		slices.Sort(keys)
		got := map[string]int{}
		for _, q := range keys {
			if shrugs(z, q) || (z.Ascension == "Juggernaut" && slices.Contains(juggernautImmune, q)) || (z.Ascension == "Frost Wyrm" && q == "snare") {
				continue
			}
			if z.Status == nil {
				z.Status = map[string]int{}
			}
			z.Status[q] = max(z.Status[q], put[q])
			got[q] = z.Status[q]
		}
		if len(got) > 0 {
			emit(action.Event{Kind: action.StatusOn, Target: z.ID, Statuses: got, Skill: "Hex Spread"})
		}
		return
	}
}

// healPowers are what a heal also does: Sanctuary shields the ally, and
// the Healer's 4-piece bonus removes one bad status.
func healPowers(u, z *state.Unit) {
	if n := u.Af("sanctuary"); n > 0 {
		z.Shield = max(z.Shield, int(Round(float64(z.MaxHP)*pct(n))))
	}
	if u.Af("cleanseHeal") > 0 {
		for _, q := range cleansable {
			if z.Status[q] > 0 {
				delete(z.Status, q)
				return
			}
		}
	}
}

// quickHands is the Gunslinger's Quick Hands: once per turn, the first
// attack does not end the turn. It reports whether it kept the turn going.
func quickHands(u *state.Unit, emit action.Emit) bool {
	if u.Af("quick") == 0 || u.QuickHands || !u.Done {
		return false
	}
	u.QuickHands, u.Done, u.Undo = true, false, nil
	emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: "Quick Hands"})
	return true
}

// breathSkills are the Dragon Knight's breath and dive skills.
var breathSkills = []string{"Dragonfire", "Glacial Breath", "Sky Dive", "Hellfire Dive", "Dragon's Fury"}

// Widen is a skill as a unit uses it: the Dragon Knight's and Mage's
// 4-piece bonuses make their breath, dive and area spells reach wider.
func Widen(u *state.Unit, s content.Skill) content.Skill {
	if s.Friendly() {
		return s
	}
	if n := u.Af("breathWide"); n > 0 && slices.Contains(breathSkills, s.Name) {
		s.Area += n
	}
	if n := u.Af("wideSpells"); n > 0 && s.Area > 0 && s.IsMagic() {
		s.Area += n
	}
	return s
}

// StartTurnPowers resets the powers that last until the unit's next turn.
func StartTurnPowers(u *state.Unit) {
	u.Rhythm, u.QuickHands = 0, false
	if u.Shadowstep {
		u.Status["cloak"], u.Shadowstep = 0, false
	}
}
