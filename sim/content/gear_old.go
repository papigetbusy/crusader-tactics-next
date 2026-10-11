package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// oldGear reads the gear tables (WPN, AFX, SETS, UNQ, RANKS, RREQ, WTN) and
// adds the gear numbers that were constants in the old code.
func oldGear(tables fs.FS) (Gear, error) {
	var wpn [][]any
	var afx [][]any
	var sets map[string]struct {
		N  string  `json:"n"`
		B2 [][]any `json:"b2"`
		B3 [][]any `json:"b3"`
	}
	var unq map[string]struct {
		Sl  string  `json:"sl"`
		Wt  string  `json:"wt"`
		At  string  `json:"at"`
		Set string  `json:"set"`
		Af  [][]any `json:"af"`
		Fl  string  `json:"fl"`
	}
	g := Gear{
		About: "Gear: weapon ranks, what each class can carry, rarities, item names, the bonuses items can roll, " +
			"legendary items and armor sets. From the old game's WPN, AFX, SETS, UNQ, RANKS and RREQ tables and its gear code.",
		RankDamage: []float64{.9, .95, 1, 1.05, 1.1, 1.2}, // RKD, line 1020
		RankHit:    []int{-10, -5, 0, 5, 10, 15},          // RKH, line 1020
		RankSteps: map[string]RankStep{ // rankIdx, line 1043
			"P":  {Start: 2, Levels: []int{8, 18, 30}},
			"P2": {Start: 2, Levels: []int{10, 20, 32}},
			"S":  {Start: 1, Levels: []int{10, 20, 32}},
		},
		Classes:        oldClassGear,
		WeaponCategory: map[string]string{"sword": "m", "axe": "m", "lance": "m", "dagger": "m", "bow": "r", "gun": "r", "staff": "g", "wand": "g"}, // WCAT, line 1109
		StaffHeal:      .5,                                                                                                                          // healAmt0, line 452
		MaxReduction:   50,                                                                                                                          // dcalc, line 449
		DoubleStrike:   .6,                                                                                                                          // act, line 1722
		AddedElements:  map[string]string{"addFire": "fire", "addElec": "elec"},                                                                     // hurt, line 1695
		StatusChances:  map[string]Affix{"bleedCh": {"bleed", 3}, "poisonCh": {"poison", 3}, "stunCh": {"stun", 1}},                                 // hurt, line 1702
		Rarities: []Rarity{{"common", 1, "#cfd3e6"}, {"uncommon", 2, "#7be0a8"}, {"rare", 4, "#5da8ff"}, // RAR, line 841
			{"legendary", 10, "#f2b84b"}},
		ArmorNames: map[string]map[string][]string{ // ARN, line 1027
			"L": {"h": {"Cloth Hood", "Leather Cap"}, "a": {"Cloth Tunic", "Leather Vest"}, "b": {"Worn Boots", "Leather Boots"}},
			"M": {"h": {"Chain Coif", "Scale Helm"}, "a": {"Chain Mail", "Scale Mail"}, "b": {"Chain Boots", "Scale Greaves"}},
			"H": {"h": {"Iron Helm", "Steel Helm"}, "a": {"Plate Armor", "Full Plate"}, "b": {"Iron Greaves", "Plate Sabatons"}},
		},
		CharmNames:   []string{"Bone Charm", "Lucky Coin", "Wolf Fang"},                    // NM.c, line 843
		ArmorWeights: map[string]ArmorWeight{"H": {1.35, .9}, "L": {.7, 1.2}, "M": {1, 1}}, // genItem, lines 1064, 1067
		Sets:         map[string]GearSet{},
		Loot: Loot{
			Power:      .8,                                                                                                  // TUNE.loot, line 845
			Bands:      []LootBand{{10, odds(.7, 0, 1)}, {30, odds(.7, 1, 2)}, {60, odds(.93, 2, 3)}, {0, odds(.75, 2, 3)}}, // bandRoll, line 1005
			DropChance: .6, BossLegendary: .25, EliteBetter: .3,                                                             // drop, line 784
			ForKiller: .5,                  // lootFor, line 1074
			EmptySlot: .75, MainWeapon: .6, // lootPick, lines 1071-1072
			Chests:       map[string][]Odds{"wood": odds(.7, 0, 1), "silver": odds(.7, 1, 2), "gold": odds(.7, 2, 3)},    // CHT, line 999; openChest, line 1003
			ClassAffixes: oldClassAffixes,                                                                                // CAF, lines 1197-1201
			GoldBase:     2, GoldPerLevel: .6, GoldBoss: 8, GoldElite: 3, GoldBounty: 2, GoldWanted: 3, GoldReplayed: .5, // hurt, line 1707; gmul, line 1460
		},
	}
	for name, v := range map[string]any{"WPN.json": &wpn, "AFX.json": &afx, "SETS.json": &sets, "UNQ.json": &unq,
		"RANKS.json": &g.Ranks, "RREQ.json": &g.RarityRank, "WTN.json": &g.WeaponNames} {
		b, err := fs.ReadFile(tables, name)
		if err != nil {
			return g, fmt.Errorf("old table %s is missing: %w", name, err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			return g, fmt.Errorf("old table %s: %w", name, err)
		}
	}
	for _, w := range wpn {
		g.Weapons = append(g.Weapons, WeaponBase{Name: w[0].(string), Type: w[1].(string), Rarity: int(w[2].(float64))})
	}
	for _, a := range afx {
		k := AffixKind{Key: a[0].(string), Text: a[1].(string), Slots: a[3].(string), LegendaryOnly: len(a) > 4}
		for _, r := range a[2].([]any) {
			p := r.([]any)
			k.Ranges = append(k.Ranges, [2]int{int(p[0].(float64)), int(p[1].(float64))})
		}
		g.Affixes = append(g.Affixes, k)
	}
	for key, s := range sets {
		g.Sets[key] = GearSet{Name: s.N, Two: affixList(s.B2), Three: affixList(s.B3)}
	}
	b, _ := fs.ReadFile(tables, "UNQ.json")
	order, err := objectKeys(b)
	if err != nil {
		return g, err
	}
	for _, name := range order {
		u := unq[name]
		g.Uniques = append(g.Uniques, Unique{Name: name, Slot: u.Sl, WeaponType: u.Wt, Weight: u.At, Set: u.Set, Affixes: affixList(u.Af), Lore: u.Fl})
	}
	return g, nil
}

func affixList(pairs [][]any) []Affix {
	out := make([]Affix, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, Affix{Key: p[0].(string), Value: int(p[1].(float64))})
	}
	return out
}

// oldClassGear is CW (index.html lines 1021-1022): the weapons each class
// can use (one per weapon set) and the armor it can wear.
var oldClassGear = map[string]ClassGear{
	"Warrior":       {Weapons: []WeaponSkill{{"sword", "P"}, {"axe", "S"}}, Armor: []string{"M", "H"}},
	"Tank":          {Weapons: []WeaponSkill{{"axe", "P"}, {"lance", "S"}}, Armor: []string{"M", "H"}},
	"Mage":          {Weapons: []WeaponSkill{{"wand", "P"}, {"staff", "S"}}, Armor: []string{"L", "M"}},
	"Healer":        {Weapons: []WeaponSkill{{"staff", "P"}, {"wand", "S"}}, Armor: []string{"L", "M"}},
	"Ranger":        {Weapons: []WeaponSkill{{"bow", "P"}, {"sword", "S"}}, Armor: []string{"L", "M"}},
	"Assassin":      {Weapons: []WeaponSkill{{"dagger", "P"}, {"dagger", "P"}}, Armor: []string{"L", "M"}, Dual: true},
	"Gunslinger":    {Weapons: []WeaponSkill{{"gun", "P"}, {"gun", "P"}}, Armor: []string{"L", "M"}, Dual: true},
	"Witch":         {Weapons: []WeaponSkill{{"staff", "P"}, {"dagger", "S"}}, Armor: []string{"L", "M"}},
	"Knight":        {Weapons: []WeaponSkill{{"lance", "P"}, {"sword", "S"}}, Armor: []string{"M", "H"}},
	"Dragon Knight": {Weapons: []WeaponSkill{{"lance", "P"}, {"axe", "P2"}, {"sword", "S"}}, Armor: []string{"M", "H"}},
}

// oldClassAffixes is CAF (index.html lines 1197-1201): the bonuses loot for
// each class can roll.
var oldClassAffixes = map[string][]string{
	"Warrior":       {"dmg", "phys", "crit", "critDmg", "leech", "bleedCh", "bossDmg", "hit", "hpPct", "defPct", "dr", "thorns", "regen", "fire", "light", "elec"},
	"Tank":          {"dmg", "phys", "crit", "hpPct", "defPct", "dr", "thorns", "regen", "leech", "hit", "bossDmg", "dark", "ice", "fire"},
	"Knight":        {"dmg", "phys", "crit", "critDmg", "leech", "bossDmg", "hit", "hpPct", "defPct", "dr", "thorns", "light", "dark", "elec"},
	"Dragon Knight": {"dmg", "phys", "crit", "critDmg", "leech", "bossDmg", "hit", "addFire", "fire", "ice", "elec", "hpPct", "defPct", "dr", "thorns"},
	"Mage":          {"dmg", "fire", "ice", "elec", "addFire", "addElec", "crit", "critDmg", "hit", "hpPct", "evade", "dr", "regen"},
	"Healer":        {"healPct", "regen", "light", "elec", "dark", "dmg", "hpPct", "defPct", "dr", "evade", "hit"},
	"Witch":         {"dmg", "dark", "poisonCh", "fire", "ice", "crit", "critDmg", "hit", "hpPct", "evade", "dr", "regen", "leech"},
	"Ranger":        {"dmg", "phys", "crit", "critDmg", "hit", "bossDmg", "bleedCh", "poisonCh", "addElec", "ice", "dark", "elec", "evade", "hpPct", "dr"},
	"Assassin":      {"dmg", "phys", "crit", "critDmg", "hit", "poisonCh", "bleedCh", "leech", "evade", "dark", "elec", "fire", "hpPct", "bossDmg"},
	"Gunslinger":    {"dmg", "phys", "crit", "critDmg", "hit", "bossDmg", "addFire", "addElec", "bleedCh", "fire", "dark", "elec", "evade", "hpPct", "dr"},
}

// odds is the old game's two-way roll: rarity lo with chance p, else hi.
func odds(p float64, lo, hi int) []Odds { return []Odds{{p, lo}, {1, hi}} }
