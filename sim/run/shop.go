package run

import (
	"errors"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/items"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Dice streams for choices between battles: each comes from the run seed,
// the floor and a counter, so the same run gives the same shop and gifts.
const (
	streamShop    = 11
	streamGift    = 12
	streamCombine = 13
)

func (r *Run) dice(stream, n int) *rng.PCG {
	return rng.New(uint64(r.Seed)<<32|uint64(n), uint64(stream))
}

// shopBattle is a stand-in battle for making loot between battles: the
// floor and the strongest heroes (loot is made for one of them).
func (r *Run) shopBattle(f int) *state.Battle {
	b := &state.Battle{Floor: f}
	for i, cl := range r.Strongest() {
		h := r.Hero(cl)
		u := &state.Unit{ID: i + 1, Team: state.Heroes, Class: h.Class, Level: h.Level, Gear: h.Gear}
		b.Units = append(b.Units, u)
		b.Party = append(b.Party, u.ID)
	}
	return b
}

// Stock is the shop's stock on a floor, made on the first visit and kept
// (old openShop, line 765): items for the strongest heroes, of the floor's
// rarity, sometimes one better.
func (r *Run) Stock(c *content.Content, f int) []content.Item {
	m := r.Memory(f)
	if m.ShopMade {
		return m.Shop
	}
	s := c.Floors.Shop
	d := r.dice(streamShop, f)
	b := r.shopBattle(f)
	for range s.Stock {
		k := items.Band(c, f, d)
		if d.Roll("shop better") < s.BetterChance {
			k = min(c.Gear.Legendary(), k+1)
		}
		p := b.Units[rng.IntN(d, "shop hero", len(b.Units))]
		m.Shop = append(m.Shop, items.For(c, b, d, p, k))
	}
	m.ShopMade = true
	return m.Shop
}

// Buy buys item i of a floor's shop.
func (r *Run) Buy(c *content.Content, f, i int) error {
	stock := r.Stock(c, f)
	if i < 0 || i >= len(stock) {
		return errors.New("That item is not for sale.")
	}
	price := items.Price(c, stock[i])
	if r.Gold < price {
		return errors.New("Not enough gold.")
	}
	r.Gold -= price
	r.Bag = append(r.Bag, stock[i])
	r.Memory(f).Shop = slices.Delete(stock, i, i+1)
	return nil
}

// Sell sells bag item i for half its price.
func (r *Run) Sell(c *content.Content, i int) error {
	if i < 0 || i >= len(r.Bag) {
		return errors.New("There is no such item in the bag.")
	}
	r.Gold += items.SellPrice(c, r.Bag[i])
	r.Bag = slices.Delete(r.Bag, i, i+1)
	return nil
}

// Combine turns two copies of the same item and rarity in the bag into one
// item a rarity higher. A legendary waits for a Lv 30 hero.
func (r *Run) Combine(c *content.Content, i, j int) (content.Item, error) {
	if i == j || i < 0 || j < 0 || i >= len(r.Bag) || j >= len(r.Bag) {
		return content.Item{}, errors.New("Pick two items from the bag.")
	}
	a, b := r.Bag[i], r.Bag[j]
	if a.Name != b.Name || a.Rarity != b.Rarity {
		return content.Item{}, errors.New("Combining needs two copies of the same item and rarity.")
	}
	if a.Rarity >= c.Gear.Legendary() {
		return content.Item{}, errors.New("Legendaries can't be combined.")
	}
	if items.Allowed(c, r.shopBattle(r.Floor), a.Rarity+1) != a.Rarity+1 {
		return content.Item{}, errors.New("A legendary needs a hero at Lv 30 first.")
	}
	r.Combined++
	it := items.Combine(c, r.Floor, r.dice(streamCombine, r.Combined), a, b)
	hi, lo := max(i, j), min(i, j)
	r.Bag = slices.Delete(r.Bag, hi, hi+1)
	r.Bag = slices.Delete(r.Bag, lo, lo+1)
	r.Bag = append(r.Bag, it)
	return it, nil
}
