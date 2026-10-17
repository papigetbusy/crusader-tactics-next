package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// innerStatus is the status Inner Element adds for each element (old hurt,
// line 1702); anything else (no element) bleeds.
var innerStatus = map[string][2]any{
	"fire": {"burn", 3}, "ice": {"snare", 2}, "elec": {"shock", 1}, "light": {"weak", 2}, "dark": {"poison", 3},
}

// lengthened are the statuses an upgraded Inner Element makes last a turn longer.
var lengthened = []string{"burn", "snare", "shock", "weak", "poison", "bleed"}

// hexed are the debuffs a Hex Queen makes last a turn longer.
var hexed = []string{"poison", "weak", "brk", "snare"}

// juggernautImmune are the statuses that never stick to a Juggernaut.
var juggernautImmune = []string{"stun", "snare", "brk", "shock", "lock", "ivy"}

// applyStatuses puts the hit's statuses on a target that survived (old
// hurt, lines 1702-1703): the skill's own, lightning's shock, Beast Mode's
// burn chance, Inner Element, Poison Ivy, and the ascension passives. A
// status already there keeps the longer duration.
func applyStatuses(c *content.Content, b *state.Battle, dice rng.Dice, a Attack, v *state.Unit) map[string]int {
	u, s := a.By, a.Skill
	fixed := a.Fixed > 0
	fx := map[string]int{}
	for q, n := range s.Effects {
		fx[q] = n
	}
	el := s.Element
	if el == content.UserElement || (el != "" && !u.Player() && a.ClassBase) {
		el = u.Element
	}
	if el == "elec" {
		fx["shock"] = max(fx["shock"], c.Combat.ElecShockTurns)
	}
	if u.Status["beast"] > 0 && !fixed && dice.Roll(RollBeastBurn) < u.P("bmB", 0.35) {
		fx["burn"] = 3
	}
	if u.Status["inner"] > 0 {
		if is, ok := innerStatus[el]; ok {
			q, n := is[0].(string), is[1].(int)
			if q == "burn" {
				fx[q] = n
			} else {
				fx[q] = max(fx[q], n)
			}
		} else {
			fx["bleed"] = max(fx["bleed"], 3)
		}
		if u.P("inExt", 0) != 0 {
			for _, q := range lengthened {
				if fx[q] != 0 {
					fx[q]++
				}
			}
		}
	}
	if u.Status["ivyA"] > 0 && !fixed {
		fx["lock"] = max(fx["lock"], 2)
		fx["ivy"] = max(fx["ivy"], 2)
		if u.P("ivPo", 0) != 0 {
			fx["poison"] = max(fx["poison"], 2)
		}
		setIvy(v, u.P("ivP", 0.3))
	}
	gearStatuses(c, dice, u, fx)
	asc := c.Ascensions[u.Ascension]
	switch u.Ascension {
	case "Pyromancer", "Inferno Rider":
		if s.Element == "fire" {
			fx["bleed"] = max(fx["bleed"], int(asc.N("fireBleed")))
		}
	case "Viper", "Plague Doctor":
		fx["bleed"] = max(fx["bleed"], int(asc.N("bleed")))
	case "Frost Wyrm":
		fx["snare"] = max(fx["snare"], int(asc.N("snare")))
	case "Cryomancer":
		if s.Element == "ice" {
			fx["snare"] = max(fx["snare"], int(asc.N("iceSnare")))
		}
	case "Trapper":
		if fx["snare"] != 0 {
			fx["snare"] = int(asc.N("snare"))
		}
	}
	if v.Status == nil {
		v.Status = map[string]int{}
	}
	put := map[string]int{}
	order := make([]string, 0, len(fx))
	for q := range fx {
		order = append(order, q)
	}
	slices.Sort(order) // the same order every time, so Warding's rolls are repeatable
	for _, q := range order {
		n := fx[q]
		if u.Team != v.Team && warded(dice, v) {
			continue
		}
		if u.Ascension == "Hex Queen" && slices.Contains(hexed, q) {
			n++
		}
		if h := u.Af("hexMastery"); h > 0 && u.Team != v.Team {
			n += h // Hex Mastery (approved 2026-10-17)
		}
		if shrugs(v, q) {
			continue
		}
		if (v.Ascension == "Juggernaut" && slices.Contains(juggernautImmune, q)) || (v.Ascension == "Frost Wyrm" && q == "snare") {
			continue
		}
		v.Status[q] = max(v.Status[q], n)
		put[q] = n
	}
	if u.Status["ivyA"] > 0 && u.P("ivS", 0) != 0 && !fixed {
		spread := 0
		for _, z := range b.Units {
			if spread >= int(u.P("ivS", 0)) {
				break
			}
			if z.Team == v.Team && z != v && Distance(z, v) <= 1 {
				z.Status["ivy"] = max(z.Status["ivy"], 2)
				z.Status["lock"] = max(z.Status["lock"], 2)
				setIvy(z, u.P("ivP", 0.3))
				spread++
			}
		}
	}
	if u.Status["vamp"] > 0 && !fixed {
		v.Status["bleed"] = max(v.Status["bleed"], 5)
		v.HeavyBleed, v.HeavyBleedBy = true, u.ID
		put["bleed"] = v.Status["bleed"]
	}
	if u.Status["imh"] > 0 && u.P("imhStun", 0) != 0 {
		v.Status["stun"] = max(v.Status["stun"], 1)
		put["stun"] = v.Status["stun"]
	}
	return put
}

// setIvy keeps the strongest Ivy weakening on a unit.
func setIvy(z *state.Unit, share float64) {
	if z.Params == nil {
		z.Params = map[string]float64{}
	}
	z.Params["ivyP"] = max(z.Params["ivyP"], share)
}
