package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// HitChance is the chance in % that an attack lands (old hitCh, line 1060):
// Inner Element and a Wanted Poster never miss; a Phantom is harder to hit.
func HitChance(c *content.Content, b *state.Battle, a Attack) int {
	k := c.Combat
	u, v, s := a.By, a.Target, a.Skill
	if s.NoRoll() || u.Status["inner"] > 0 || (v.Wanted && v.WantedBy == u.ID) {
		return 100
	}
	magic := isMagic(c, u, s)
	h := k.HitMelee
	switch {
	case magic:
		h = k.HitMagic
	case s.Strike:
		if stats.HandCategory(c, u) == "r" {
			h = k.HitRanged
		}
	case s.MinRange >= 2:
		h = k.HitRanged
	}
	if u.Team == state.Heroes {
		h += k.HitHeroBonus
	}
	if u.Team == state.Heroes && u.Armed {
		if rk := stats.SetRank(c, u); rk >= 0 {
			h += c.Gear.RankHit[rk]
		}
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
	h += u.Af("hit") - v.Af("evade")
	switch HeightEdge(c, b, u, v) {
	case 1:
		h += k.HeightHit
	case -1:
		h -= k.UphillHit
	}
	if v.Ascension == "Phantom" {
		h -= int(c.Ascensions["Phantom"].N("evasion"))
	}
	return max(k.HitMin, min(k.HitMax, h))
}
