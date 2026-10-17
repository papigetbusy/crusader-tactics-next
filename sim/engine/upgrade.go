package engine

import (
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// upgradeItems brings gear from saves made before the legendary rework
// (2026-10-17) up to date: epic armor made before epic sets existed joins
// the epic set of its weight. Heroes in a saved battle get their bonuses
// worked out again.
func upgradeItems(c *content.Content, s *saved) {
	fix := func(it *content.Item) {
		if it.Rarity == c.Gear.Epic() && it.Set == "" && strings.Contains("hab", it.Slot) && it.Weight != "" {
			if _, ok := c.Gear.Sets["e"+it.Weight]; ok {
				it.Set = "e" + it.Weight
			}
		}
	}
	list := func(items []content.Item) {
		for i := range items {
			fix(&items[i])
		}
	}
	gear := func(g map[string]content.Item) {
		for k, it := range g {
			fix(&it)
			g[k] = it
		}
	}
	if s.Run != nil {
		list(s.Run.Bag)
		for _, h := range s.Run.Heroes {
			gear(h.Gear)
		}
	}
	if b := s.Battle; b != nil {
		list(b.Bag)
		list(b.RedBag)
		for _, u := range b.Units {
			if u.Armed {
				gear(u.Gear)
				stats.UseSet(c, u, u.WeaponSet)
			}
		}
	}
}
