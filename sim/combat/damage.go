// Package combat is hit chance, damage and what one hit does. It copies the
// old game's hitCh, mult, dcalc and hurt (index.html lines 445-450, 778,
// 1060-1061, 1687-1710).
//
// Ascension passives, class abilities, gear affixes and weapon sets also
// change these numbers in the old game. They arrive with heroes and gear in
// Phase 6; until then a unit simply has none of them.
//
// Exactness: every multiplication happens in the same order as the old game,
// and float64(...) around products stops Go from fusing a multiply and an add
// into one step, which could change the last digit of a result.
package combat

import (
	"math"

	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Attack bundles who hits whom with what.
type Attack struct {
	By    *state.Unit
	Skill content.Skill
	// ClassBase: a monster skill from its type's own list (the old game uses
	// the monster's element for these).
	ClassBase  bool
	Target     *state.Unit
	Extra      float64 // extra damage multiplier (0 or 1 = none): a Guardian's cover, Riposte, splash, Outlaw
	Sure       bool    // never misses (extra hits of Double Strike, Crash Down...)
	Fixed      int     // fixed damage instead of the formula (Clone Burst, Plague Bloom); no crit
	CritBonus  int     // extra crit chance in % (Flurry of Blades)
	Counter    bool    // a counterattack: never countered, never shared by a Sacred Vow
	NoRedirect bool    // a Guardian cannot take this hit
}

// Round is JavaScript's Math.round: halves round up (toward +infinity).
// Go's math.Round rounds halves away from zero, which differs for negatives.
func Round(x float64) float64 { return stats.Round(x) }

// SkillLevel is a hero's level in a skill: it rises with uses (old SKL).
func SkillLevel(c *content.Content, u *state.Unit, skill string) int {
	uses := u.SkillUses[skill]
	need := c.Combat.SkillLevelUses
	lv := 1
	for lv < len(need) && uses >= need[lv] {
		lv++
	}
	return lv
}

// attackElement is the element a hit uses: monsters hit with their own
// element whenever the skill has one (old dcalc, line 445).
func attackElement(a Attack) string {
	if a.Skill.Element != "" && a.By.Team != state.Heroes {
		return a.By.Element
	}
	return a.Skill.Element
}

// Multiplier is the weapon triangle and element wheel (old mult, line 778).
func Multiplier(c *content.Content, by, target *state.Unit, element string) float64 {
	m := 1.0
	if element == "" && by.Weapon != "" && target.Weapon != "" {
		if c.Combat.WeaponBeats[by.Weapon] == target.Weapon {
			m = c.Combat.Strong
		} else if c.Combat.WeaponBeats[target.Weapon] == by.Weapon {
			m = c.Combat.Weak
		}
	}
	ee := element
	if element == "" && by.Team == state.Heroes {
		ee = by.Element
	}
	if ee != "" && ee != content.UserElement && target.Element != "" {
		if c.Beats[ee] == target.Element {
			m *= c.Combat.Strong
		} else if c.Beats[target.Element] == ee {
			m *= c.Combat.Weak
		}
	}
	if by.Status["prism"] > 0 {
		m = math.Max(c.Combat.PrismAtLeast, m)
	}
	return m
}

// Enraged reports a boss below half health, which hits harder (old dcalc,
// line 446).
func Enraged(c *content.Content, u *state.Unit) bool {
	return u.Boss && float64(u.HP) < float64(u.MaxHP)*c.Combat.BossEnrageBelow
}

// Damage is the damage of one hit for a given damage roll (old dcalc, lines
// 445-450). roll is the damage spread, from 0.9 to 1.1. It returns the damage
// and the triangle/element multiplier. The order of the steps is the old
// game's: with fractions, a different order can round differently.
func Damage(c *content.Content, b *state.Battle, a Attack, roll float64) (int, float64) {
	k := c.Combat
	u, v, s := a.By, a.Target, a.Skill
	el := attackElement(a)
	m := Multiplier(c, u, v, el)
	atk := float64(u.Atk) * s.Power
	def := float64(v.Def)
	if Enraged(c, u) {
		atk *= k.BossEnrage
	}
	if u.Team == state.Heroes && !u.Decoy {
		atk = ascensionDamage(c, b, a, el, atk, &def)
	}
	if u.Status["rally"] > 0 {
		atk *= k.RallyAttack
	}
	if s.Execute && float64(v.HP) < float64(v.MaxHP)*k.ExecuteBelow {
		atk *= k.Execute
	}
	if u.Status["weak"] > 0 {
		atk *= k.WeakAttack
	}
	if u.Status["ivy"] > 0 {
		atk *= 1 - u.P("ivyP", 0.3)
	}
	if u.Status["inner"] > 0 {
		atk *= 1 + u.P("inP", 0.3)
	}
	if u.Status["heye"] > 0 && s.Kind != "u" {
		atk *= 1 + u.P("heP", 0.3)
	}
	if u.Status["vow"] > 0 {
		atk *= 1 + u.P("vwP", 0.3)
	}
	if u.Status["imh"] > 0 {
		atk += float64(float64(u.Def) * u.P("imhM", 1))
	}
	if u.Status["frenzy"] > 0 {
		atk *= 1.5
	}
	if u.Team == state.Heroes {
		va := 0.0
		for _, z := range b.Units {
			if z != u && z.Team == state.Heroes && z.Status["vow"] > 0 && z.P("vwA", 0) != 0 && Distance(z, u) <= 1 {
				va = math.Max(va, z.P("vwA", 0))
			}
		}
		atk *= 1 + va
		if InBanner(b, u) {
			atk *= 1.2
		}
	}
	if u.Status["fstorm"] > 0 && (el == "fire" || (el == content.UserElement && u.Element == "fire")) {
		atk *= 1.2
	}
	if u.Status["qd"] > 0 {
		extra := 0.0
		if u.QuickShots == 1 {
			extra = u.P("qd2P", 0)
		}
		atk *= 1 + u.P("qdP", 0) + extra
	}
	if v.Wanted && v.WantedBy == u.ID {
		atk *= 1.5
	}
	if v.Status["tmark"] > 0 && u.Ascension == "Trapper" {
		atk *= c.Ascensions["Trapper"].N("markedDamage")
	}
	if v.Status["frozen"] > 0 && u.Team != v.Team {
		atk *= 1.5
	}
	if v.Status["dmark"] > 0 && u.Team == state.Heroes {
		atk *= 1.5
	}
	if s.Charge {
		atk *= 1 + chargeBonus(c, u)
	}
	if u.Ascension == "Black Knight" && s.Area == 0 && s.ExtraTargets == 0 && s.Line == 0 {
		atk *= c.Ascensions["Black Knight"].N("singleDamage")
	}
	if cursed(v) && anyAscension(b, u.Team, "Hex Queen") {
		atk *= c.Ascensions["Hex Queen"].N("cursedDamage")
	}
	if s.Ambush && v.HP >= v.MaxHP {
		atk *= k.Ambush
	}
	atk *= newAttack(b, a)
	if v.Status["brk"] > 0 {
		def *= k.BrokenDefense
	}
	if s.IgnoreDefense {
		def = 0
	}
	if s.IsMagic() {
		def *= k.MagicDefense
	}
	if v.Status["ivy"] > 0 {
		def *= 1 - v.P("ivyP", 0.3)
	}
	if v.Status["ironm"] > 0 {
		def *= 1 + v.P("ironP", 0.25)
	}
	if u.Status["heye"] > 0 && u.P("heI", 0) != 0 {
		def *= 1 - u.P("heI", 0)
	}
	if u.Status["deadeye"] > 0 {
		def = 0
	}
	if v.Team == state.Heroes && near(b, v, 2, true, func(z *state.Unit) bool { return z.Status["bulwark"] > 0 }) {
		def *= 1.3
	}
	if v.Team == state.Heroes && near(b, v, 2, false, func(z *state.Unit) bool { return z.Status["svow"] > 0 }) {
		def *= 1.2
	}
	defend := 1.0
	if v.Defending {
		defend = k.DefendDamage
	}
	extra := a.Extra
	if extra == 0 {
		extra = 1
	}
	r := (atk - def) * m * defend * extra * damageTaken(c, b, v)
	if v.Status["grd"] > 0 {
		r *= k.GuardDamage
	}
	r *= newTaken(a)
	if v.Status["taunt"] > 0 {
		r *= 0.7
	}
	if v.Status["scales"] > 0 && Distance(u, v) <= 1 {
		r *= 0.6
	}
	if v.Team == state.Heroes && paladinNext(b, v) {
		r *= c.Ascensions["Paladin"].N("aegis")
	}
	dm := int(Round(r * roll))
	if dm < k.MinDamage {
		dm = k.MinDamage
	}
	return dm, m
}

// ascensionDamage is the hero part of dcalc (line 447): skill level, the
// ascension passives and the Warlord's aura. Sniper also lowers def.
func ascensionDamage(c *content.Content, b *state.Battle, a Attack, el string, atk float64, def *float64) float64 {
	k := c.Combat
	u, v, s := a.By, a.Target, a.Skill
	atk *= 1 + float64(k.SkillLevelBonus*float64(SkillLevel(c, u, s.Name)-1))
	asc := c.Ascensions[u.Ascension]
	switch u.Ascension {
	case "Berserker":
		if float64(u.HP) < float64(u.MaxHP)*asc.N("lowHP") {
			atk *= asc.N("lowHPDamage")
		}
	case "Pyromancer", "Inferno Rider":
		if el == "fire" {
			atk *= asc.N("fireDamage")
		}
	case "Blood Witch":
		atk *= asc.N("damage")
	case "Viper":
		if v.Status["bleed"] > 0 {
			atk *= asc.N("bleedingDamage")
		}
	case "Bounty Hunter":
		if v.Elite || v.Boss {
			atk *= asc.N("bigDamage")
		}
	case "Cryomancer":
		if v.Status["snare"] > 0 {
			atk *= asc.N("snaredDamage")
		}
	case "Sniper":
		if !u.Moved {
			atk *= asc.N("stillDamage")
		}
		*def *= asc.N("defense")
	case "Trapper":
		if v.Status["snare"] > 0 || v.Status["stun"] > 0 {
			atk *= asc.N("markedDamage")
		}
	}
	w := c.Ascensions["Warlord"]
	if near(b, u, int(w.N("auraRange")), true, func(z *state.Unit) bool { return z.Ascension == "Warlord" }) {
		atk *= w.N("aura")
	}
	return atk
}

// chargeBonus is the charge skills' extra damage for tiles walked this turn
// (Storm Rider doubles it; a vow raises it).
func chargeBonus(c *content.Content, u *state.Unit) float64 {
	k := c.Combat
	vc := 1.0
	if u.Status["vow"] > 0 {
		vc = 1 + u.P("vwC", 0)
	}
	most, per := k.ChargeMax, k.ChargePerTile
	if u.Ascension == "Storm Rider" {
		sr := c.Ascensions["Storm Rider"]
		most, per = sr.N("chargeMax"), sr.N("chargePerTile")
	}
	return math.Min(float64(most*vc), float64(float64(float64(u.TilesMoved)*per)*vc))
}

// cursed reports a target with one of the Hex Queen's debuffs on it.
func cursed(v *state.Unit) bool {
	for _, q := range []string{"poison", "weak", "brk", "snare", "bleed"} {
		if v.Status[q] > 0 {
			return true
		}
	}
	return false
}

// anyAscension reports a unit of this team with this ascension on the board.
func anyAscension(b *state.Battle, team int, name string) bool {
	for _, z := range b.Units {
		if z.Team == team && z.Ascension == name {
			return true
		}
	}
	return false
}

// near reports a hero within r tiles of v that passes the test; self says
// whether v itself counts.
func near(b *state.Battle, v *state.Unit, r int, self bool, test func(*state.Unit) bool) bool {
	for _, z := range b.Units {
		if z.Team == state.Heroes && (self || z != v) && Distance(z, v) <= r && test(z) {
			return true
		}
	}
	return false
}

// paladinNext reports a Paladin standing right next to v (Aegis).
func paladinNext(b *state.Battle, v *state.Unit) bool {
	for _, z := range b.Units {
		if z.Team == state.Heroes && z != v && z.Ascension == "Paladin" && Distance(z, v) == 1 {
			return true
		}
	}
	return false
}

// InBanner reports a hero standing inside a Warlord's War Banner.
func InBanner(b *state.Battle, u *state.Unit) bool {
	if u.Team != state.Heroes {
		return false
	}
	for _, f := range b.Fields {
		if f.Kind == "banner" && abs(f.X-u.X)+abs(f.Y-u.Y) <= f.Range {
			return true
		}
	}
	return false
}

// damageTaken is the terrain's damage share where the target stands
// (old tRed); unknown ground counts as plains.
func damageTaken(c *content.Content, b *state.Battle, v *state.Unit) float64 {
	t, ok := c.Terrain[board.GroundAt(b, v.X, v.Y)]
	if !ok {
		t = c.Terrain['g']
	}
	return t.DamageTaken
}

// isHero is a player's own hero (not a summon or companion): only heroes
// level their skills.
func isHero(u *state.Unit) bool { return u.Team == state.Heroes && !u.Summoned }
