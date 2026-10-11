package oldgame

import (
	"fmt"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
)

// lootDifferences compares the bag, gold and unopened chests, when the
// recording stored them.
func lootDifferences(g *engine.Game, s Snap) []string {
	var out []string
	if s.Inv == nil {
		return nil
	}
	if a, b := itemNames(g.B.Bag), itemNames(bagOf(s)); fmt.Sprint(a) != fmt.Sprint(b) {
		out = append(out, fmt.Sprintf("bag: Go %v, old game %v", a, b))
	}
	if g.B.Gold != s.Gold {
		out = append(out, fmt.Sprintf("gold: Go %d, old game %d", g.B.Gold, s.Gold))
	}
	if s.ChestsAt != nil {
		var a, b []string
		for _, ch := range g.B.Chests {
			a = append(a, fmt.Sprintf("%d,%d %s", ch.X, ch.Y, ch.Kind))
		}
		for _, ch := range s.ChestsAt {
			b = append(b, fmt.Sprintf("%v,%v %v", ch[0], ch[1], ch[2]))
		}
		slices.Sort(a)
		slices.Sort(b)
		if fmt.Sprint(a) != fmt.Sprint(b) {
			out = append(out, fmt.Sprintf("unopened chests: Go %v, old game %v", a, b))
		}
	}
	return out
}

func itemNames(items []content.Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, fmt.Sprintf("%s(%d)%v", it.Name, it.Rarity, it.Affixes))
	}
	return out
}
