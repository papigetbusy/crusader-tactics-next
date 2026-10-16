package combat

import (
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
// revives, buffs, heals and attacks (old act, index.html lines 1711-1734).
// It assumes the caller already checked the use is allowed.
func UseSkill(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, slot int, v *state.Unit, emit action.Emit) error {
	s := u.Skills[slot]
	if err := useSkill(c, b, dice, u, slot, v, emit); err != nil {
		return err
	}
	counter(c, b, dice, u, s, v, emit)
	return nil
}

// useSkill is UseSkill before any counterattack.
func useSkill(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, slot int, v *state.Unit, emit action.Emit) error {
	s := u.Skills[slot]
	if IsNew(s) {
		return useNew(c, b, dice, u, slot, s, Pick{Target: v}, emit)
	}
	_, classBase, _ := c.SkillFor(u.Team, u.Class, s.Name)
	emit(action.Event{Kind: action.SkillUsed, Unit: u.ID, Target: v.ID, Skill: s.Name})
	switch {
	case s.Summon != nil:
		raise(c, b, u, *s.Summon, emit)
	case s.Revive:
		revive(c, b, u, emit)
	case s.Support:
		support(c, b, u, s, emit)
	case s.Heal:
		healSkill(c, b, dice, u, s, v, emit)
	default:
		attack(c, b, dice, Attack{By: u, Skill: s, ClassBase: classBase, Target: v}, emit)
	}
	finishSkill(c, b, u, slot, s, emit)
	return nil
}

// raise is a Necromancer's raise skill: the monster appears on the first open
// tile around the caster, one level lower, undead, and shares the caster's
// action (old act, line 1714).
func raise(c *content.Content, b *state.Battle, u *state.Unit, kind int, emit action.Emit) {
	for _, d := range reviveSpots {
		t := state.Tile{X: u.X + d[0], Y: u.Y + d[1]}
		if !b.Inside(t.X, t.Y) || heightAt(b, t.X, t.Y) < 0 || b.UnitAt(t.X, t.Y) != nil {
			continue
		}
		z := monsterUnit(c, b, c.Monsters[kind], max(1, u.Level-1), "dark", u.Team, t)
		z.Summoned, z.Undead, z.Owner, z.Done = true, true, u.ID, true
		b.Units = append(b.Units, z)
		emit(action.Event{Kind: action.Summoned, Unit: u.ID, Target: z.ID})
		return
	}
}

// FlushRaises raises the skeletons Grave Bloom owes for this action's kills
// (the old game raised them a moment after each kill).
func FlushRaises(c *content.Content, b *state.Battle, emit action.Emit) {
	for len(b.Raises) > 0 {
		r := b.Raises[0]
		b.Raises = b.Raises[1:]
		if nb := b.Unit(r.By); nb != nil {
			RaiseSkeleton(c, b, nb, r.X, r.Y, true, emit)
		}
	}
}

// attack is a damaging skill (old act, lines 1721-1727): breaking a cloak,
// Blood Magic's HP cost, the targets (multi-target picks, Hunter's Eye,
// Unbroken Maiden's cleave), the hits (Outlaw's two shots, Double Strike),
// area, Inner Element and Beast Mode splash, piercing, healing splashes,
// and the Stormcaller's and Trickshot's arcs.
func attack(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, emit action.Emit) {
	u, s, v := a.By, a.Skill, a.Target
	if u.Status["cloak"] > 0 { // attacking breaks the cloak
		u.Status["cloak"] = 0
		CloakCooldown(u)
	}
	if u.Ascension == "Blood Witch" || s.HPCost > 0 {
		cost := s.HPCost
		if cost == 0 {
			cost = c.Ascensions["Blood Witch"].N("hpCost")
		}
		u.HP = max(1, u.HP-int(Round(float64(u.MaxHP)*cost)))
	}
	targets := []*state.Unit{v}
	if s.ExtraTargets > 0 {
		targets = append(targets, extraTargets(b, dice, u, s, v)...)
	}
	if u.Status["heye"] > 0 && s.Kind != "u" && s.Area == 0 && s.Line == 0 && len(targets) < int(u.P("heN", 3)) {
		reach := s.MaxRange + u.RangeBonus + int(u.P("heR", 3))
		for _, z := range b.Units {
			if len(targets) >= int(u.P("heN", 3)) {
				break
			}
			if z.Team != u.Team && !slices.Contains(targets, z) && Distance(u, z) >= s.MinRange && Distance(u, z) <= reach {
				targets = append(targets, z)
			}
		}
	}
	if u.Status["imh"] > 0 && u.P("imhCl", 0) != 0 {
		for _, z := range b.Units {
			if z.Team != u.Team && !slices.Contains(targets, z) && Distance(z, u) <= 1 {
				targets = append(targets, z)
			}
		}
	}
	outlaw := u.Ascension == "Outlaw" && s.Hits == 0 && s.ExtraTargets == 0
	double := u.Status["dbl"] > 0 && s.Kind != "u"
	hits := max(1, s.Hits)
	if outlaw {
		hits = int(c.Ascensions["Outlaw"].N("shots"))
	}
	if double {
		hits *= int(u.P("dblN", 2))
	}
	for _, z := range targets {
		for h := 0; h < hits; h++ {
			if b.Unit(z.ID) == nil {
				break
			}
			hit := Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z, Extra: 1}
			if z == v {
				hit.Extra = extraOf(a) // a Shoulder Check's slam
			}
			if outlaw {
				hit.Extra = c.Ascensions["Outlaw"].N("shotPower")
			}
			if double {
				if h > 0 {
					hit.Extra *= 1 + u.P("dblB", 0)
					hit.Sure = u.P("dblSure", 0) != 0
				}
				hit.CritBonus = int(u.P("dblC", 0))
			}
			Hurt(c, b, dice, hit, emit)
		}
	}
	doubleStrike(c, b, dice, a, v, emit)
	if s.Area > 0 {
		for _, z := range areaOf(b, u, s, v, targets) {
			Hurt(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z}, emit)
		}
	}
	if (u.Status["inner"] > 0 && u.P("inArea", 0) != 0) || (u.Status["beast"] > 0 && u.P("bmS", 0) != 0) {
		rr := s.Area + 1
		mul := u.P("bmS", 0.5)
		if u.Status["inner"] > 0 && u.P("inArea", 0) != 0 {
			mul = 1
		}
		var ring []*state.Unit
		for _, z := range b.Units {
			if z.Team != u.Team && !slices.Contains(targets, z) && Distance(z, v) <= rr && (s.Area == 0 || Distance(z, v) > s.Area) {
				ring = append(ring, z)
			}
		}
		for _, z := range ring {
			Hurt(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z, Extra: mul}, emit)
		}
	}
	if s.Line > 0 {
		for _, t := range lineOf(u, s, v) {
			if z := b.UnitAt(t[0], t[1]); z != nil && z.Team != u.Team {
				Hurt(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z}, emit)
			}
		}
	}
	if u.Ascension == "Templar" && float64(u.TilesMoved) >= c.Ascensions["Templar"].N("chargeTiles") {
		for _, z := range members(b, u.Team) {
			if Distance(z, v) <= 1 && z.Status["frozen"] <= 0 {
				heal(z, float64(z.MaxHP)*c.Ascensions["Templar"].N("chargeHeal"))
			}
		}
	}
	if s.Splash > 0 {
		for _, z := range members(b, u.Team) {
			if Distance(z, u) <= 1 && z.Status["frozen"] <= 0 {
				heal(z, float64(z.MaxHP)*s.Splash)
			}
		}
	}
	if u.Ascension == "Stormcaller" || u.Ascension == "Trickshot" {
		asc := c.Ascensions[u.Ascension]
		for _, z := range b.Units {
			if z.Team != u.Team && !slices.Contains(targets, z) && Distance(z, v) <= int(asc.N("arcRange")) {
				Hurt(c, b, dice, Attack{By: u, Skill: s, ClassBase: a.ClassBase, Target: z, Extra: asc.N("arc")}, emit)
				break
			}
		}
	}
}

// extraTargets picks a multi-target skill's other targets in range: the
// old game's own picks when replaying it, otherwise the fair shuffle.
func extraTargets(b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, v *state.Unit) []*state.Unit {
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
		var out []*state.Unit
		for _, id := range p.PickExtra(ids, s.ExtraTargets-1) {
			out = append(out, b.Unit(id))
		}
		return out
	}
	rng.Shuffle(dice, RollShuffle, pool)
	return pool[:min(len(pool), s.ExtraTargets-1)]
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

// finishSkill is what every skill use does afterwards (old act, lines
// 1728-1733): skill uses (2 when switching skills), Cloak, the one-shot
// ability statuses wearing off, self-locking skills, the cooldown, then who
// may act again: Hit and Run, Quick Draw, Oath of Ruin and Blade Dance.
func finishSkill(c *content.Content, b *state.Battle, u *state.Unit, slot int, s content.Skill, emit action.Emit) {
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
	offensive := !s.Friendly()
	if offensive {
		if u.Status["aim"] > 0 {
			u.Status["aim"] = 0 // Steady Aim is spent on the next shot
		}
		if s.IsMagic() && s.Aim != content.AimEcho {
			u.LastSpell = s.Name
		}
		if s.Kind != "u" {
			u.Status["dbl"] = 0
			if u.P("heK", 0) == 0 {
				u.Status["heye"] = 0
			}
		}
		u.Status["imh"] = 0
		if u.Status["inner"] > 0 {
			setParam(u, "inN", u.P("inN", 1)-1)
			if u.Params["inN"] <= 0 {
				u.Status["inner"] = 0
			}
		}
		u.Status["deadeye"] = 0
	}
	if s.LockSelf && u.Ascension != "Oracle" {
		u.Status["lock"] = 3
	}
	if s.Cooldown > 0 && !(u.Ascension == "Storm Dragoon" && s.Name == "Sky Dive") {
		u.Cooldowns[slot] = max(1, s.Cooldown-u.Af("cdr"))
	}
	if u.Player() && u.Ascension == "Tempest Archer" && !u.Moved && offensive {
		u.Acted, u.Done = true, false
		emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: "Hit and Run"})
	} else {
		u.Done = true
	}
	if s.Cloak {
		u.Cooldowns[slot] = 99
	}
	if u.Player() && u.Status["qd"] > 0 && offensive {
		u.QuickShots++
		u.QuickUsed = append(u.QuickUsed, s.Name)
		if float64(u.QuickShots) < u.P("qdN", 2) {
			u.Done, u.Moved = false, true
			emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: "Quick Draw"})
		} else {
			u.Status["qd"], u.QuickShots, u.QuickUsed = 0, 0, nil
		}
	}
	if u.ExtraTurn {
		u.ExtraTurn = false
		u.Done, u.Moved, u.Acted, u.Undo = false, false, false, nil
		emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: "Oath of Ruin"})
	}
	if u.Player() && u.Ascension == "Blade Dancer" && u.KilledThisAction && !u.Danced {
		u.Danced = true
		u.Done, u.Moved, u.Undo = false, false, nil
		emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: "Blade Dance"})
	}
	u.KilledThisAction = false
	if u.Player() {
		PairDone(b, u)
	}
}

// setParam sets one of a unit's ability numbers.
func setParam(u *state.Unit, key string, v float64) {
	if u.Params == nil {
		u.Params = map[string]float64{}
	}
	u.Params[key] = v
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
