package items

import (
	"math"
	"slices"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Make creates an item of rarity k for a slot (old genItem, lines 1062-1068).
// Its numbers grow with the floor; legendaries are the named uniques.
func Make(c *content.Content, floor int, dice rng.Dice, k int, sl, wt, at, class string) content.Item {
	g := c.Gear
	pw := (1 + int(math.Floor(float64(floor)*g.Loot.Power/4))) * g.Rarities[k].Power
	hf := int(math.Ceil(float64(pw) / 2))
	sized := func(it content.Item, weight string) content.Item {
		w, ok := g.ArmorWeights[weight]
		if !ok {
			w = content.ArmorWeight{Def: 1, HP: 1}
		}
		switch sl {
		case "a":
			it.Def, it.HP = round(float64(pw)*w.Def), round(float64(pw*2)*w.HP)
		case "h":
			it.Def, it.HP = round(float64(hf)*w.Def), round(float64(pw)*w.HP)
		case "b":
			it.Def = round(float64(hf) * w.Def)
		}
		return it
	}
	legendary := g.Legendary()
	finish := func(it content.Item, wt string) content.Item {
		var pre []content.Affix
		var name string
		if k == g.Epic() {
			name, pre = prefix(c, dice, sl, class, wt)
		}
		var skip []string
		for _, a := range pre {
			skip = append(skip, a.Key)
		}
		it.Affixes = append(pre, bonuses(c, dice, k, sl, class, wt, skip)...)
		if name != "" {
			it.Name = name + " " + it.Name
		}
		return it
	}
	if k >= legendary {
		var pool []content.Unique
		for _, u := range g.Uniques {
			if u.Slot == sl && (wt == "" || u.WeaponType == wt) && (at == "" || u.Weight == at) {
				pool = append(pool, u)
			}
		}
		if len(pool) == 0 {
			for _, u := range g.Uniques {
				if u.Slot == sl {
					pool = append(pool, u)
				}
			}
		}
		u := pool[rng.IntN(dice, RollItem, len(pool))]
		it := content.Item{Name: u.Name, Slot: sl, Rarity: legendary, Unique: true, Set: u.Set, WeaponType: u.WeaponType,
			Weight: u.Weight, Affixes: slices.Clone(u.Affixes)}
		switch sl {
		case "w":
			it.Atk = pw
		case "c":
			it.HP = pw * 3
		}
		return sized(it, u.Weight)
	}
	switch sl {
	case "w":
		if wt == "" {
			wt = weaponTypeOrder[rng.IntN(dice, RollItem, len(weaponTypeOrder))]
		}
		var pool []content.WeaponBase
		for _, w := range g.Weapons {
			if w.Rarity == k && w.Type == wt {
				pool = append(pool, w)
			}
		}
		if len(pool) == 0 {
			for _, w := range g.Weapons {
				if w.Rarity == k {
					pool = append(pool, w)
				}
			}
		}
		w := pool[rng.IntN(dice, RollItem, len(pool))]
		return finish(content.Item{Name: w.Name, Slot: "w", Rarity: k, WeaponType: w.Type, Atk: pw}, w.Type)
	case "c":
		names := g.CharmNames
		if k == g.Epic() && len(g.EpicCharms) > 0 {
			names = g.EpicCharms
		}
		name := names[rng.IntN(dice, RollItem, len(names))]
		return finish(content.Item{Name: name, Slot: "c", Rarity: k, HP: pw * 3}, "")
	}
	if at == "" {
		at = string("LMH"[rng.IntN(dice, RollItem, 3)])
	}
	names := g.ArmorNames[at][sl]
	name := names[0]
	if k >= 2 {
		name = names[1]
	}
	if k == g.Epic() && len(names) > 2 {
		name = names[2]
	}
	it := sized(content.Item{Name: name, Slot: sl, Rarity: k, Weight: at}, at)
	if sl == "b" && k >= 2 {
		it.Mov = 1
	}
	return finish(it, "")
}

// prefix rolls an epic item's own bonus (approved 2026-10-11): one of the
// prefixes that fit the slot, the class and the weapon. It returns the
// prefix's name and its bonuses.
func prefix(c *content.Content, dice rng.Dice, sl, class, wt string) (string, []content.Affix) {
	keys := c.Gear.Loot.ClassAffixes[class]
	var cand []content.Prefix
	for _, p := range c.Gear.Prefixes {
		k := p.Bonuses[0].Key
		if strings.Contains(p.Slots, sl) && (len(keys) == 0 || slices.Contains(keys, k)) && (sl != "w" || weaponAllows(c, k, wt)) {
			cand = append(cand, p)
		}
	}
	if len(cand) == 0 {
		return "", nil
	}
	p := cand[rng.IntN(dice, RollPrefix, len(cand))]
	var out []content.Affix
	for _, b := range p.Bonuses {
		out = append(out, content.Affix{Key: b.Key, Value: b.Low + rng.IntN(dice, RollPrefix, b.High-b.Low+1)})
	}
	return p.Name, out
}

// weaponTypeOrder is the old WTYPES order (Object.keys(WTN)).
var weaponTypeOrder = []string{"sword", "axe", "lance", "bow", "gun", "staff", "wand", "dagger"}

func round(x float64) int { return int(stats.Round(x)) }

// bonuses rolls a new item's bonuses (old rollAffixes, line 1203): 0-1 for
// common, 1-2 uncommon, 2-3 rare or epic; only kinds that fit the slot, the
// class and the weapon, and not one the epic bonus already gives (skip);
// each rolled within its rarity's range.
func bonuses(c *content.Content, dice rng.Dice, k int, sl, class, wt string, skip []string) []content.Affix {
	counts := []int{rng.IntN(dice, RollBonus, 2), 1 + rng.IntN(dice, RollBonus, 2), 2 + rng.IntN(dice, RollBonus, 2), 3}
	n := counts[min(k, 3)]
	if k == c.Gear.Epic() {
		n = counts[2]
	}
	allowedKeys := c.Gear.Loot.ClassAffixes[class]
	var cand []content.AffixKind
	for _, a := range c.Gear.Affixes {
		if strings.Contains(a.Slots, sl) && a.Ranges[k][1] > 0 && !a.LegendaryOnly && !slices.Contains(skip, a.Key) &&
			(len(allowedKeys) == 0 || slices.Contains(allowedKeys, a.Key)) && (sl != "w" || weaponAllows(c, a.Key, wt)) {
			cand = append(cand, a)
		}
	}
	order := pickOrder(dice, cand, min(n, len(cand)))
	var out []content.Affix
	for _, a := range order {
		lo, hi := a.Ranges[k][0], a.Ranges[k][1]
		out = append(out, content.Affix{Key: a.Key, Value: lo + rng.IntN(dice, RollBonus, hi-lo+1)})
	}
	return out
}

// pickOrder chooses which n bonuses an item gets, in order: a fair shuffle
// (approved 2026-10-09), or the old game's own picks when replaying it.
func pickOrder(dice rng.Dice, cand []content.AffixKind, n int) []content.AffixKind {
	if p, ok := dice.(rng.AffixPicker); ok {
		keys := make([]string, len(cand))
		for i, a := range cand {
			keys[i] = a.Key
		}
		if picked, ok := p.PickAffixes(keys, n); ok && len(picked) == n {
			var out []content.AffixKind
			for _, k := range picked {
				if i := slices.Index(keys, k); i >= 0 {
					out = append(out, cand[i])
				}
			}
			if len(out) == n {
				return out
			}
		}
	}
	cand = slices.Clone(cand)
	rng.Shuffle(dice, RollOrder, cand)
	return cand[:n]
}

// weaponAllows is the old WGATE (line 1202): healing bonuses only on staffs
// and wands; no physical or bleed bonuses on staffs and wands.
func weaponAllows(c *content.Content, key, wt string) bool {
	if wt == "" {
		return true
	}
	if key == "healPct" {
		return wt == "staff" || wt == "wand"
	}
	if (key == "phys" || key == "bleedCh") && c.Gear.Category(wt) == "g" {
		return false
	}
	return true
}
