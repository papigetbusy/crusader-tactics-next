package items

import (
	"slices"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
)

// RollCombine labels the rolls of combining two items.
const RollCombine = "combine roll"

// Combine makes one item a rarity higher from two copies of the same item
// (old shop Combine and combineAf, index.html lines 768 and 1043): a new
// item of the same kind whose bonuses mix both: a bonus on both keeps the
// higher value plus half the lower, capped at 1.25 times the new rarity's
// best roll; the mix is picked fairly, and topped up with fresh bonuses.
// A purple item keeps the purple bonus it is made with.
func Combine(c *content.Content, floor int, dice rng.Dice, a, b content.Item) content.Item {
	k := a.Rarity + 1
	at := ""
	if strings.Contains("hab", a.Slot) {
		at = a.Weight
	}
	it := Make(c, floor, dice, k, a.Slot, a.WeaponType, at, "")
	if k >= c.Gear.Legendary() {
		return it
	}
	keep := 0
	if k == c.Gear.Epic() {
		for _, p := range c.Gear.Prefixes {
			if strings.HasPrefix(it.Name, p.Name+" ") {
				keep = len(p.Bonuses)
			}
		}
	}
	pre := slices.Clone(it.Affixes[:keep])
	n := len(it.Affixes) - keep
	merged := map[string]int{}
	var order []string
	for _, af := range append(slices.Clone(a.Affixes), b.Affixes...) {
		if v, ok := merged[af.Key]; ok {
			merged[af.Key] = max(v, af.Value) + int(round(float64(min(v, af.Value))*.5))
			continue
		}
		merged[af.Key] = af.Value
		order = append(order, af.Key)
	}
	var pool []content.Affix
	for _, key := range order {
		kind, ok := affixKind(c, key)
		if !ok || !strings.Contains(kind.Slots, a.Slot) || kind.LegendaryOnly || (a.Slot == content.SlotWeapon && !weaponAllows(c, key, a.WeaponType)) ||
			slices.ContainsFunc(pre, func(p content.Affix) bool { return p.Key == key }) {
			continue
		}
		v := merged[key]
		if hi := kind.Ranges[k][1]; hi > 0 {
			v = min(v, round(float64(hi)*1.25))
		}
		pool = append(pool, content.Affix{Key: key, Value: v})
	}
	rng.Shuffle(dice, RollCombine, pool)
	pool = pool[:min(n, len(pool))]
	for _, f := range it.Affixes[keep:] {
		if len(pool) >= n {
			break
		}
		if !slices.ContainsFunc(pool, func(p content.Affix) bool { return p.Key == f.Key }) {
			pool = append(pool, f)
		}
	}
	it.Affixes = append(pre, pool...)
	return it
}

func affixKind(c *content.Content, key string) (content.AffixKind, bool) {
	for _, k := range c.Gear.Affixes {
		if k.Key == key {
			return k, true
		}
	}
	return content.AffixKind{}, false
}

// Price is what the shop asks for an item (old itemVal, line 762): a price
// for its rarity plus 3 per point of attack, defense, a third of its HP, and
// 3 per move (as the old game counted).
func Price(c *content.Content, it content.Item) int {
	s := c.Floors.Shop
	base := 0
	if it.Rarity < len(s.Prices) {
		base = s.Prices[it.Rarity]
	}
	return round(float64(base) + s.PerPoint*(float64(it.Atk)+float64(it.Def)+float64(it.HP)/3+float64(it.Mov*3)))
}

// SellPrice is what the shop pays: half the price, rounded.
func SellPrice(c *content.Content, it content.Item) int {
	return round(float64(Price(c, it)) * c.Floors.Shop.SellShare)
}
