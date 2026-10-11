package content

import "slices"

// The purple (epic) tier, approved by the owner on 2026-10-11 (items P1-P42
// in docs/purple-tier-proposal.md). It sits between rare and legendary.

// epicRanges are the ordinary bonuses' epic ranges (P4), about halfway
// between rare and legendary.
var epicRanges = map[string][2]int{
	"dmg": {13, 20}, "fire": {18, 28}, "ice": {18, 28}, "elec": {18, 28}, "light": {18, 28}, "dark": {18, 28},
	"phys": {15, 23}, "addFire": {11, 18}, "addElec": {11, 18}, "crit": {9, 14}, "critDmg": {28, 48},
	"hpPct": {12, 20}, "defPct": {18, 27}, "dr": {6, 10}, "leech": {6, 10}, "evade": {6, 9}, "healPct": {18, 28},
	"bossDmg": {19, 32}, "bleedCh": {13, 20}, "poisonCh": {13, 20}, "thorns": {14, 24}, "regen": {3, 5}, "hit": {10, 15},
}

// epicPrefixes are the epic-only bonuses (P8-P21).
var epicPrefixes = []Prefix{
	{"Vampiric", "wc", []PrefixBonus{{"leech", 9, 11}}},
	{"Searing", "wc", []PrefixBonus{{"addFire", 16, 20}}},
	{"Thundering", "wc", []PrefixBonus{{"addElec", 16, 20}}},
	{"Keen", "wbc", []PrefixBonus{{"crit", 13, 15}}},
	{"Brutal", "wc", []PrefixBonus{{"critDmg", 45, 55}}},
	{"Huntsman's", "w", []PrefixBonus{{"bossDmg", 30, 36}}},
	{"Venomous", "wc", []PrefixBonus{{"poisonCh", 20, 24}}},
	{"Serrated", "w", []PrefixBonus{{"bleedCh", 20, 24}}},
	{"Echoing", "w", []PrefixBonus{{"double", 6, 8}}},
	{"Bulwark", "ahc", []PrefixBonus{{"dr", 9, 11}}},
	{"Thorned", "ah", []PrefixBonus{{"thorns", 22, 26}}},
	{"Mending", "wc", []PrefixBonus{{"healPct", 28, 32}, {"regen", 3, 3}}},
	{"Executioner's", "w", []PrefixBonus{{"execute", 25, 25}}},
	{"Warding", "ahbc", []PrefixBonus{{"ward", 25, 25}}},
}

// epicWeapons are the purple weapons, one per type (P22-P29).
var epicWeapons = []WeaponBase{
	{"Gravewarden Blade", "sword", 3}, {"Bloodmoon Cleaver", "axe", 3}, {"Wyrmspine Pike", "lance", 3},
	{"Duskwind Recurve", "bow", 3}, {"Ashen Repeater", "gun", 3}, {"Candlewood Crozier", "staff", 3},
	{"Starfall Rod", "wand", 3}, {"Widow's Kiss", "dagger", 3},
}

// epicArmor are the purple armor pieces (P30-P38), by weight and slot.
var epicArmor = map[string]map[string]string{
	"H": {"h": "Vigil Greathelm", "a": "Cathedral Plate", "b": "Rampart Sabatons"},
	"M": {"h": "Wolfsteel Coif", "a": "Brigandine", "b": "Pathfinder Greaves"},
	"L": {"h": "Silk Cowl", "a": "Shadowweave Jerkin", "b": "Moonstep Boots"},
}

// addPurple turns the old four tiers into five.
func addPurple(c *Content) {
	g := &c.Gear
	const epic = 3
	g.Rarities = slices.Insert(g.Rarities, epic, Rarity{"epic", 6, "#b46cf0"}) // P1
	g.RarityRank = slices.Insert(g.RarityRank, epic, 4)                        // P2: rank A
	for i := range g.Weapons {
		if g.Weapons[i].Rarity >= epic {
			g.Weapons[i].Rarity++
		}
	}
	g.Weapons = append(g.Weapons, epicWeapons...)
	for w, slots := range epicArmor {
		for sl, name := range slots {
			g.ArmorNames[w][sl] = append(g.ArmorNames[w][sl], name)
		}
	}
	c.Floors.Shop.Prices = slices.Insert(c.Floors.Shop.Prices, epic, 200)          // between rare (100) and legendary (400)
	g.EpicCharms = []string{"Saint's Reliquary", "Blood Ruby", "Hourglass of Ash"} // P39-P41
	for i, a := range g.Affixes {
		g.Affixes[i].Ranges = slices.Insert(a.Ranges, epic, epicRanges[a.Key])
	}
	g.Affixes = append(g.Affixes,
		AffixKind{Key: "execute", Text: "% more damage to a target below 30% HP (epic)", Ranges: make([][2]int, 5), Slots: "w", LegendaryOnly: true},
		AffixKind{Key: "ward", Text: "% chance to shrug off a harmful status (epic)", Ranges: make([][2]int, 5), Slots: "ahbc", LegendaryOnly: true})
	g.Prefixes = epicPrefixes
	for class, keys := range g.Loot.ClassAffixes {
		if class != "Healer" {
			keys = append(keys, "execute")
		}
		g.Loot.ClassAffixes[class] = append(keys, "ward")
	}
	l := &g.Loot
	l.Bands = []LootBand{ // P5
		{10, []Odds{{.7, 0}, {1, 1}}},
		{20, []Odds{{.7, 1}, {1, 2}}},
		{30, []Odds{{.5, 1}, {.9, 2}, {1, 3}}},
		{60, []Odds{{.55, 2}, {.95, 3}, {1, 4}}},
		{0, []Odds{{.3, 2}, {.8, 3}, {1, 4}}},
	}
	l.Chests["gold"] = []Odds{{.65, 2}, {.95, 3}, {1, 4}} // P7, with a 5% legendary from Lv 30 (owner's answer)
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-11",
		What: "A purple (epic) tier between rare and legendary: power 6, weapons need rank A, each item has its own " +
			"epic bonus (its name prefix) and 2-3 ordinary bonuses; 8 weapons, 9 armor pieces, 3 charms; drops from " +
			"floor 21; gold chests give rare, purple or (from Lv 30) legendary. The old game had four tiers.",
		Apply: addPurple,
	})
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-11",
		What: "Shops no longer take a floor number: every 7 floors the shop comes up between floors, then that floor is " +
			"fought. Before, floors 7, 14, 21... were shop floors with no fight.",
		Apply: func(c *Content) { c.Floors.ShopEvery, c.Floors.ShopBefore = 0, 7 },
	}, Change{
		Date: "2026-10-10",
		What: "Heroes start on blue starting tiles the player picks: every open tile of the bottom 2 rows. " +
			"Before, the heroes stood in fixed spots spread along the bottom row.",
		Apply: func(c *Content) { c.Floors.StartRows = 2 },
	})
}
