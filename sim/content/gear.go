package content

import "fmt"

// Gear slots. A hero has up to three weapon slots (one per weapon set; dual
// wielders hold two weapons in one set), a helm, armor, boots and a charm.
const (
	SlotWeapon = "w" // an item's slot when it is a weapon (it goes in w, w2 or w3)
	SlotHelm   = "h"
	SlotArmor  = "a"
	SlotBoots  = "b"
	SlotCharm  = "c"
)

// WeaponSlots are the hero's weapon slots, in set order (old WSLOT).
var WeaponSlots = []string{"w", "w2", "w3"}

// Affix is one bonus on an item, such as "crit 8" (8% crit chance). The keys
// and what they do are listed in Gear.Affixes.
type Affix struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

// Item is one piece of gear (old item objects: sl, r, n, wt, at, atk, def,
// hp, mov, af, set, uq).
type Item struct {
	Name       string  `json:"name"`
	Slot       string  `json:"slot"`   // w, h, a, b or c
	Rarity     int     `json:"rarity"` // 0 common, 1 uncommon, 2 rare, 3 legendary
	WeaponType string  `json:"weaponType,omitempty"`
	Weight     string  `json:"weight,omitempty"` // armor: L light, M medium, H heavy
	Atk        int     `json:"atk,omitempty"`
	Def        int     `json:"def,omitempty"`
	HP         int     `json:"hp,omitempty"`
	Mov        int     `json:"mov,omitempty"`
	Affixes    []Affix `json:"affixes,omitempty"`
	Set        string  `json:"set,omitempty"` // armor set key (pieces of one set give set bonuses)
	Unique     bool    `json:"unique,omitempty"`
}

// WeaponSkill is one weapon type a class can use and how well: P (primary),
// P2 (a second primary) or S (secondary). It also names a weapon set slot.
type WeaponSkill struct {
	Type string `json:"type"`
	Role string `json:"role"`
}

// ClassGear is what a class can carry (old CW).
type ClassGear struct {
	Weapons []WeaponSkill `json:"weapons"` // one per weapon slot, in order
	Armor   []string      `json:"armor"`   // armor weights it can wear
	Dual    bool          `json:"dual,omitempty"`
}

// RankStep is how a weapon rank grows with level: Start at level 1, one rank
// more at each of Levels.
type RankStep struct {
	Start  int   `json:"start"`
	Levels []int `json:"levels"`
}

// GearSet is a named armor set: bonuses for wearing two and three pieces.
type GearSet struct {
	Name  string  `json:"name"`
	Two   []Affix `json:"two"`
	Three []Affix `json:"three"`
}

// AffixKind is one kind of item bonus and the range it rolls in for each
// rarity (old AFX).
type AffixKind struct {
	Key           string   `json:"key"`
	Text          string   `json:"text"`
	Ranges        [][2]int `json:"ranges"` // per rarity: lowest and highest roll (0-0: never)
	Slots         string   `json:"slots"`  // the slots it can roll on (w h a b c)
	LegendaryOnly bool     `json:"legendaryOnly,omitempty"`
}

// Unique is a named legendary item with fixed bonuses (old UNQ).
type Unique struct {
	Name       string  `json:"name"`
	Slot       string  `json:"slot"`
	WeaponType string  `json:"weaponType,omitempty"`
	Weight     string  `json:"weight,omitempty"`
	Set        string  `json:"set,omitempty"`
	Affixes    []Affix `json:"affixes"`
	Lore       string  `json:"lore,omitempty"`
}

// WeaponBase is a weapon name for a type and rarity (old WPN).
type WeaponBase struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rarity int    `json:"rarity"`
}

// Rarity is one rarity: its name and how much it multiplies an item's power.
type Rarity struct {
	Name  string `json:"name"`
	Power int    `json:"power"`
	Color string `json:"color"`
}

// Gear is every gear number (data/gear.json).
type Gear struct {
	About      string               `json:"_about"`
	Ranks      []string             `json:"ranks"`      // E, D, C, B, A, S
	RankDamage []float64            `json:"rankDamage"` // damage multiplier per rank (old RKD)
	RankHit    []int                `json:"rankHit"`    // hit chance bonus per rank (old RKH)
	RarityRank []int                `json:"rarityRank"` // weapon rank needed for each rarity (old RREQ)
	RankSteps  map[string]RankStep  `json:"rankSteps"`  // by role (old rankIdx)
	Classes    map[string]ClassGear `json:"classes"`
	// WeaponCategory: m melee, r ranged, g magic (old WCAT). A class's
	// skills only work with a weapon of its first weapon's category.
	WeaponCategory map[string]string `json:"weaponCategory"`
	WeaponNames    map[string]string `json:"weaponNames"` // sword -> Sword...
	StaffHeal      float64           `json:"staffHeal"`   // share of a staff or wand's attack added to heals
	MaxReduction   int               `json:"maxDamageReduction"`
	DoubleStrike   float64           `json:"doubleStrike"` // power of the extra hit from "strike twice"
	AddedElements  map[string]string `json:"addedElements"`
	StatusChances  map[string]Affix  `json:"statusChances"` // affix -> status and its turns

	Rarities     []Rarity                        `json:"rarities"`
	Weapons      []WeaponBase                    `json:"weapons"`
	ArmorNames   map[string]map[string][2]string `json:"armorNames"` // weight -> slot -> [common/uncommon, rare+]
	CharmNames   []string                        `json:"charmNames"`
	Affixes      []AffixKind                     `json:"affixes"`
	Uniques      []Unique                        `json:"uniques"`
	Sets         map[string]GearSet              `json:"sets"`
	ArmorWeights map[string]ArmorWeight          `json:"armorWeights"`
	Loot         Loot                            `json:"loot"`
}

// ArmorWeight scales a piece of armor's DEF and HP by its weight.
type ArmorWeight struct {
	Def float64 `json:"def"`
	HP  float64 `json:"hp"`
}

// Category is a weapon type's category (m, r or g); unknown types are melee.
func (g Gear) Category(weaponType string) string {
	if c, ok := g.WeaponCategory[weaponType]; ok {
		return c
	}
	return "m"
}

// Rank is a hero's rank with a weapon role at a level (old rankIdx).
func (g Gear) Rank(role string, level int) int {
	st := g.RankSteps[role]
	r := st.Start
	for _, l := range st.Levels {
		if level >= l {
			r++
		}
	}
	return r
}

func (g Gear) check() error {
	n := len(g.Ranks)
	if n == 0 || len(g.RankDamage) != n || len(g.RankHit) != n {
		return fmt.Errorf("gear.json: ranks, rankDamage and rankHit need one entry per rank")
	}
	if len(g.RarityRank) != len(g.Rarities) {
		return fmt.Errorf("gear.json: rarityRank needs one entry per rarity")
	}
	for class, cg := range g.Classes {
		for _, w := range cg.Weapons {
			if _, ok := g.RankSteps[w.Role]; !ok {
				return fmt.Errorf("gear.json, %s: weapon role %q has no rankSteps", class, w.Role)
			}
		}
	}
	for _, a := range g.Affixes {
		if len(a.Ranges) != len(g.Rarities) {
			return fmt.Errorf("gear.json, affix %q: needs one range per rarity", a.Key)
		}
	}
	return nil
}

// LootBand is the rarity range loot rolls in up to a floor: Low with
// chance LowChance, else High (old bandRoll).
type LootBand struct {
	UpToFloor int     `json:"upToFloor"` // 0: every floor after the bands before
	LowChance float64 `json:"lowChance"`
	Low       int     `json:"low"`
	High      int     `json:"high"`
}

// Loot is how items drop: from monsters, chests and bounties (old drop,
// lootFor, lootPick, genItem, rollAffixes, openChest).
type Loot struct {
	Power         float64             `json:"power"` // item power per floor (old TUNE.loot)
	Bands         []LootBand          `json:"bands"`
	DropChance    float64             `json:"dropChance"`    // a normal monster drops something
	BossLegendary float64             `json:"bossLegendary"` // a boss drops a legendary (once per floor)
	EliteBetter   float64             `json:"eliteBetter"`   // an elite's drop is one rarity higher
	ForKiller     float64             `json:"forKiller"`     // loot is for the hero who made the kill (else the least geared hero)
	EmptySlot     float64             `json:"emptySlot"`     // loot fills an empty slot
	MainWeapon    float64             `json:"mainWeapon"`    // a weapon is of the class's first weapon type
	Chests        map[string][2]int   `json:"chests"`        // chest kind -> rarity [usual, lucky]
	ChestUsual    float64             `json:"chestUsual"`    // chance of the usual rarity
	ClassAffixes  map[string][]string `json:"classAffixes"`  // the bonuses each class's loot can roll
	// LegendaryFromLevel: legendaries only drop once a hero has reached this
	// level (0: any time). Until then a legendary roll gives a rare item.
	LegendaryFromLevel int `json:"legendaryFromLevel"`

	GoldBase     float64 `json:"goldBase"`
	GoldPerLevel float64 `json:"goldPerLevel"`
	GoldBoss     float64 `json:"goldBoss"`
	GoldElite    float64 `json:"goldElite"`
	GoldBounty   float64 `json:"goldBounty"`   // Bounty Hunter
	GoldWanted   int     `json:"goldWanted"`   // a Wanted Poster target
	GoldReplayed float64 `json:"goldReplayed"` // a floor cleared before
}
