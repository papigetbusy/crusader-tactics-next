package run

import (
	"errors"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/items"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Won records a won battle (old clearFloor and milestones, lines 760-761):
// the floor is cleared; winning the highest floor opens the next one, and
// every 100th floor reached brings a free legendary. It returns the gift,
// if any. Replaying a lower floor stays on that floor (to farm it).
func (r *Run) Won(c *content.Content, b *state.Battle, startChests []state.Chest) *content.Item {
	r.Keep(b, startChests)
	r.Memory(b.Floor).Cleared = true
	var gift *content.Item
	if r.Floor == r.Best {
		r.Best++
		r.Floor = r.Best
		if r.Best%100 == 0 && r.Gifts < r.Best/100 {
			r.Gifts++
			d := r.dice(streamGift, r.Gifts)
			slot := []string{"w", "h", "a", "b", "c"}[rng.IntN(d, "gift slot", 5)]
			it := items.Make(c, r.Best, d, c.Gear.Legendary(), slot, "", "", "")
			r.Bag = append(r.Bag, it)
			gift = &it
		}
	}
	r.Visit++
	return gift
}

// Lost records a lost battle: progress is kept and the floor is fought again
// (old: a defeat re-runs the floor; owner, 2026-10-11: no penalty).
func (r *Run) Lost(b *state.Battle, startChests []state.Chest) { r.Keep(b, startChests) }

// Travel goes to a floor already reached (old floor travel list, line 1590).
func (r *Run) Travel(f int) error {
	if f < 1 || f > r.Best {
		return errors.New("That floor hasn't been reached yet.")
	}
	if f != r.Floor {
		r.Floor = f
		r.Visit++
	}
	return nil
}

// ShopFirst reports a floor whose shop comes up before it is fought: every
// 7th floor, once per arrival.
func (r *Run) ShopFirst(c *content.Content) bool {
	n := c.Floors.ShopBefore
	return n > 0 && r.Floor%n == 0 && r.ShopVisit != r.Floor
}

// ChangeAscension switches a Lv 30 hero to another of its class's
// ascensions, once per floor (old re-spec, line 1771): learned skills stay;
// the ultimate switches to the new ascension's.
func (r *Run) ChangeAscension(c *content.Content, class, choice string) error {
	h := r.Hero(class)
	if h == nil {
		return errors.New("That is not one of your heroes.")
	}
	if h.Ascension == "" {
		return errors.New("Choose an ascension at Lv 30 first.")
	}
	if h.AscensionVisit == r.Visit {
		return errors.New("Ascension can only change once per floor.")
	}
	if !slices.Contains(c.AscensionChoices[class], choice) || choice == h.Ascension {
		return errors.New("Pick another of this class's ascensions.")
	}
	a := c.Ascensions[choice]
	h.Ascension, h.Ultimate, h.AscensionVisit = choice, a.Ultimate, r.Visit
	if !slices.Contains(h.Learned, a.Ultimate) {
		h.Learned = append(h.Learned, a.Ultimate)
	}
	return nil
}
