package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// ItemView is one item as the screen shows it.
type ItemView struct {
	Slot   string `json:"slot"` // where it is worn (w, w2, w3, h, a, b, c), or its kind in the bag
	Name   string `json:"name"`
	Rarity string `json:"rarity"` // common, uncommon, rare, legendary
	Color  string `json:"color"`
	Kind   string `json:"kind"` // "Sword · needs rank C", "Heavy armor", "Charm"
	Text   string `json:"text"` // its numbers and bonuses in a line
	Lore   string `json:"lore,omitempty"`
	// For the item's picture: its rarity as a number (0 common, up to the
	// legendary), weapon type and armor weight.
	Tier       int    `json:"tier"`
	WeaponType string `json:"weaponType,omitempty"`
	Weight     string `json:"weight,omitempty"`
}

// SetView is one weapon set.
type SetView struct {
	Weapons string `json:"weapons"`        // "Steel Sword", "Unarmed"
	Rank    string `json:"rank,omitempty"` // the hero's rank with it
	InHand  bool   `json:"inHand,omitempty"`
}

// GearView is what a hero wears.
type GearView struct {
	Worn       []ItemView `json:"worn"`
	Sets       []SetView  `json:"sets"`
	Bonuses    string     `json:"bonuses,omitempty"`     // every bonus active now
	Weapons    string     `json:"weapons"`               // what the class can use: "Sword C · Axe D · Medium and Heavy armor"
	FreeSlots  []string   `json:"freeSlots,omitempty"`   // empty weapon slots
	WeaponSlot []string   `json:"weaponSlots,omitempty"` // every weapon slot
	Bag        []BagFit   `json:"bag,omitempty"`         // how each bag item suits this hero
	Capped     bool       `json:"capped,omitempty"`      // at the level cap: earns no XP until the highest floor is cleared
}

// itemText describes an item's numbers and bonuses in one line.
func (g *Game) itemText(it content.Item) string {
	var parts []string
	for _, p := range []struct {
		n int
		w string
	}{{it.Atk, "ATK"}, {it.Def, "DEF"}, {it.HP, "HP"}, {it.Mov, "move"}} {
		if p.n != 0 {
			parts = append(parts, fmt.Sprintf("+%d %s", p.n, p.w))
		}
	}
	for _, a := range it.Affixes {
		parts = append(parts, g.affixText(a.Key, a.Value))
	}
	if it.Set != "" {
		parts = append(parts, "part of "+g.C.Gear.Sets[it.Set].Name)
	}
	return strings.Join(parts, " · ")
}

func (g *Game) affixText(key string, v int) string {
	for _, k := range g.C.Gear.Affixes {
		if k.Key == key {
			t := strings.TrimSuffix(strings.TrimSuffix(k.Text, " (legendary)"), " (epic)")
			if strings.HasPrefix(t, "%") {
				return fmt.Sprintf("%d%s", v, t)
			}
			return fmt.Sprintf("+%d %s", v, t)
		}
	}
	return fmt.Sprintf("%s %d", key, v)
}

func (g *Game) itemView(slot string, it content.Item) ItemView {
	gr := g.C.Gear
	iv := ItemView{Slot: slot, Name: it.Name, Text: g.itemText(it), Tier: it.Rarity, WeaponType: it.WeaponType, Weight: it.Weight}
	if it.Rarity >= 0 && it.Rarity < len(gr.Rarities) {
		iv.Rarity, iv.Color = gr.Rarities[it.Rarity].Name, gr.Rarities[it.Rarity].Color
	}
	switch it.Slot {
	case content.SlotWeapon:
		iv.Kind = fmt.Sprintf("%s · needs rank %s", gr.WeaponNames[it.WeaponType], gr.Ranks[gr.RarityRank[it.Rarity]])
	case content.SlotCharm:
		iv.Kind = "Charm"
	default:
		iv.Kind = map[string]string{"L": "Light", "M": "Medium", "H": "Heavy"}[it.Weight] + " armor"
	}
	for _, u := range gr.Uniques {
		if u.Name == it.Name {
			iv.Lore = u.Lore
		}
	}
	return iv
}

// gearView is a hero's gear sheet.
func (g *Game) gearView(u *state.Unit) *GearView {
	c := g.C
	_, ok := c.Gear.Classes[u.Class]
	if !ok || !u.Armed {
		return nil
	}
	gv := &GearView{WeaponSlot: stats.WeaponSlotsOf(c, u.Class)}
	for _, sl := range append(stats.WeaponSlotsOf(c, u.Class), "h", "a", "b", "c") {
		if it, ok := u.Gear[sl]; ok {
			gv.Worn = append(gv.Worn, g.itemView(sl, it))
		} else if strings.HasPrefix(sl, "w") {
			gv.FreeSlots = append(gv.FreeSlots, sl)
		}
	}
	for i := range stats.SetsOf(c, u.Class) {
		sv := SetView{Weapons: "Unarmed", InHand: i == u.WeaponSet}
		var names []string
		for _, it := range stats.SetItems(c, u.Class, u.Gear, i) {
			names = append(names, it.Name)
		}
		if len(names) > 0 {
			sv.Weapons = strings.Join(names, " + ")
			sv.Rank = c.Gear.Ranks[stats.RankFor(c, u.Class, u.Level, stats.SetWeaponType(c, u.Class, u.Gear, i))]
		}
		gv.Sets = append(gv.Sets, sv)
	}
	var keys []string
	for k, v := range u.Affix {
		if v != 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var bon []string
	for _, k := range keys {
		bon = append(bon, g.affixText(k, u.Affix[k]))
	}
	gv.Bonuses = strings.Join(bon, " · ")
	gv.Weapons = g.weaponsText(u.Class, u.Level)
	gv.Bag = g.bagFits(u)
	gv.Capped = g.B.LevelCap > 0 && u.Level >= g.B.LevelCap
	return gv
}

// bagView is the bag as the screen shows it.
func (g *Game) bagView() []ItemView {
	var out []ItemView
	for _, it := range g.B.Bag {
		out = append(out, g.itemView(it.Slot, it))
	}
	return out
}

// weaponsText is what a class can use at a level: "Sword C · Axe D · Medium
// and Heavy armor".
func (g *Game) weaponsText(class string, level int) string {
	c := g.C
	cg := c.Gear.Classes[class]
	var prof []string
	seen := map[string]bool{}
	for _, w := range cg.Weapons {
		if !seen[w.Type] {
			seen[w.Type] = true
			prof = append(prof, c.Gear.WeaponNames[w.Type]+" "+c.Gear.Ranks[stats.RankFor(c, class, level, w.Type)])
		}
	}
	var arm []string
	for _, a := range cg.Armor {
		arm = append(arm, map[string]string{"L": "Light", "M": "Medium", "H": "Heavy"}[a])
	}
	return strings.Join(prof, " · ") + " · " + strings.Join(arm, " and ") + " armor"
}
