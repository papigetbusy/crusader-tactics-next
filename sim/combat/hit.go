package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// HitChance is the chance in % that an attack lands (old hitCh, line 1060).
func HitChance(c *content.Content, a Attack) int {
	k := c.Combat
	u, v, s := a.By, a.Target, a.Skill
	if s.NoRoll() {
		return 100
	}
	magic := s.IsMagic()
	h := k.HitMelee
	switch {
	case magic:
		h = k.HitMagic
	case s.Strike:
		// The basic Strike is melee without a ranged weapon set equipped
		// (weapon sets arrive with gear in Phase 6).
	case s.MinRange >= 2:
		h = k.HitRanged
	}
	if u.Team == state.Heroes {
		h += k.HitHeroBonus
	} else if u.Boss {
		h += k.HitBossBonus
	} else if u.Elite {
		h += k.HitEliteBonus
	}
	tri := k.TriangleWeapons
	if !magic && slices.Contains(tri, u.Weapon) && slices.Contains(tri, v.Weapon) {
		if k.WeaponBeats[u.Weapon] == v.Weapon {
			h += k.HitTriangle
		} else if k.WeaponBeats[v.Weapon] == u.Weapon {
			h -= k.HitTriangle
		}
	}
	return max(k.HitMin, min(k.HitMax, h))
}
