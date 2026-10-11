package engine

import (
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

// Guide is the fixed rules the screen describes in words: what each kind of
// ground does, which element beats which, and the weapon triangle. It never
// changes during a game, so the screen asks for it once.
type Guide struct {
	Terrain     []content.Terrain `json:"terrain"`
	Beats       map[string]string `json:"beats"`       // element wheel: fire beats ice...
	WeaponBeats map[string]string `json:"weaponBeats"` // sword beats lance...
}

// RulesGuide is the Guide for these rules.
func RulesGuide(c *content.Content) Guide {
	g := Guide{Beats: c.Beats, WeaponBeats: c.Combat.WeaponBeats}
	for _, t := range c.Terrain {
		g.Terrain = append(g.Terrain, t)
	}
	sort.Slice(g.Terrain, func(i, j int) bool { return g.Terrain[i].Letter < g.Terrain[j].Letter })
	return g
}
