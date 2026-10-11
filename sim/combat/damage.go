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
	ClassBase bool
	Target    *state.Unit
	Extra     float64 // extra damage multiplier (1 = none); Riposte, splash, Outlaw use it later
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

// Damage is the damage of one hit for a given damage roll (old dcalc).
// roll is the damage spread, from 0.9 to 1.1. It returns the damage and the
// triangle/element multiplier.
func Damage(c *content.Content, b *state.Battle, a Attack, roll float64) (int, float64) {
	k := c.Combat
	u, v, s := a.By, a.Target, a.Skill
	m := Multiplier(c, u, v, attackElement(a))
	atk := float64(u.Atk) * s.Power
	def := float64(v.Def)
	if Enraged(c, u) {
		atk *= k.BossEnrage
	}
	if isHero(u) {
		atk *= 1 + float64(k.SkillLevelBonus*float64(SkillLevel(c, u, s.Name)-1))
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
	if s.Charge {
		atk *= 1 + math.Min(k.ChargeMax, float64(u.TilesMoved)*k.ChargePerTile)
	}
	if s.Ambush && v.HP >= v.MaxHP {
		atk *= k.Ambush
	}
	if v.Status["brk"] > 0 {
		def *= k.BrokenDefense
	}
	if s.IgnoreDefense {
		def = 0
	}
	if s.IsMagic() {
		def *= k.MagicDefense
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
	dm := int(Round(r * roll))
	if dm < k.MinDamage {
		dm = k.MinDamage
	}
	return dm, m
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
