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
		return content.Item{Name: w.Name, Slot: "w", Rarity: k, WeaponType: w.Type, Atk: pw, Affixes: bonuses(c, dice, k, "w", class, w.Type)}
	case "c":
		name := g.CharmNames[rng.IntN(dice, RollItem, len(g.CharmNames))]
		return content.Item{Name: name, Slot: "c", Rarity: k, HP: pw * 3, Affixes: bonuses(c, dice, k, "c", class, "")}
	}
	if at == "" {
		at = string("LMH"[rng.IntN(dice, RollItem, 3)])
	}
	name := g.ArmorNames[at][sl][0]
	if k >= 2 {
		name = g.ArmorNames[at][sl][1]
	}
	it := sized(content.Item{Name: name, Slot: sl, Rarity: k, Weight: at}, at)
	if sl == "b" && k >= 2 {
		it.Mov = 1
	}
	it.Affixes = bonuses(c, dice, k, sl, class, "")
	return it
}

// weaponTypeOrder is the old WTYPES order (Object.keys(WTN)).
var weaponTypeOrder = []string{"sword", "axe", "lance", "bow", "gun", "staff", "wand", "dagger"}

func round(x float64) int { return int(stats.Round(x)) }

// bonuses rolls a new item's bonuses (old rollAffixes, line 1203): 0-1 for
// common, 1-2 uncommon, 2-3 rare; only kinds that fit the slot, the class
// and the weapon; each rolled within its rarity's range.
func bonuses(c *content.Content, dice rng.Dice, k int, sl, class, wt string) []content.Affix {
	counts := []int{rng.IntN(dice, RollBonus, 2), 1 + rng.IntN(dice, RollBonus, 2), 2 + rng.IntN(dice, RollBonus, 2), 3}
	n := counts[min(k, 3)]
	allowedKeys := c.Gear.Loot.ClassAffixes[class]
	var cand []content.AffixKind
	for _, a := range c.Gear.Affixes {
		if strings.Contains(a.Slots, sl) && a.Ranges[k][1] > 0 && !a.LegendaryOnly &&
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
