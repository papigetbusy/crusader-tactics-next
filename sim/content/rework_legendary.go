package content

// The legendary rework, approved by the owner on 2026-10-17 (see
// docs/legendary-proposal.md; picture picks and answers at 04:53). It runs
// after the purple tier (this file's name sorts after purple.go, and Go runs
// a package's files in name order).

// legendPowers are the class legendaries' unique powers and their 4-piece
// class bonuses: bonus kinds that never roll on ordinary items.
var legendPowers = []AffixKind{
	{Key: "rhythm", Text: "Battle Rhythm: each hit you land gives +%d%% damage until your next turn (up to 3 times)"},
	{Key: "unbowed", Text: "Unbowed: below 30%% HP you take %d%% less damage"},
	{Key: "aura", Text: "Bulwark Aura: allies next to you take %d%% less damage"},
	{Key: "retaliate", Text: "Retaliate: when hit, %d%% chance to stun the attacker for 1 turn"},
	{Key: "charge", Text: "Charge: +%d%% damage for every tile you moved this turn (up to +30%%)"},
	{Key: "valor", Text: "Valor: +%d move"},
	{Key: "dragonblood", Text: "Dragonblood: %d%% of your damage is added as fire, and your hits burn for 2 turns"},
	{Key: "wyrmhide", Text: "Wyrm's Hide: you can't be burned or frozen"},
	{Key: "eagle", Text: "Eagle Eye: +%d range"},
	{Key: "pin", Text: "Pinning Shot: %d%% chance a hit roots the target for 1 turn"},
	{Key: "quick", Text: "Quick Hands: once per turn, your first attack doesn't end your turn"},
	{Key: "ricochet", Text: "Ricochet: %d%% of your damage bounces to another enemy within 2 tiles"},
	{Key: "ambush", Text: "Ambush: +%d%% damage to an enemy at full HP"},
	{Key: "shadowstep", Text: "Shadowstep: after a kill you are cloaked until your next turn"},
	{Key: "echo", Text: "Arcane Echo: %d%% chance a spell hits each target a second time"},
	{Key: "overflow", Text: "Overflow: -%d turn on every skill cooldown"},
	{Key: "sanctuary", Text: "Sanctuary: your heals also shield the ally for %d%% of their max HP"},
	{Key: "lifeline", Text: "Lifeline: once per battle, an ally who would fall stays at 1 HP"},
	{Key: "hexMastery", Text: "Hex Mastery: the statuses you cause last %d turn longer"},
	{Key: "siphon", Text: "Soul Siphon: each status you cause heals you %d%% of your max HP"},
	// 4-piece class bonuses.
	{Key: "killReset", Text: "Set of 4: kills reset your class ability's cooldown"},
	{Key: "counterBoost", Text: "Set of 4: your counterattacks deal +%d%% damage"},
	{Key: "firstWard", Text: "Set of 4: the first hit you take each battle does nothing"},
	{Key: "breathWide", Text: "Set of 4: your breath and dive skills hit %d tile wider"},
	{Key: "farShot", Text: "Set of 4: +%d%% damage to targets 4 or more tiles away"},
	{Key: "critRefund", Text: "Set of 4: your crits take %d turn off every cooldown"},
	{Key: "opener", Text: "Set of 4: your first attack each battle always crits"},
	{Key: "wideSpells", Text: "Set of 4: your area spells are %d tile wider"},
	{Key: "cleanseHeal", Text: "Set of 4: your heals also remove one bad status"},
	{Key: "hexSpread", Text: "Set of 4: your curses and poisons also spread to one enemy next to the target"},
}

// powerNumbers are the powers' fixed numbers.
var powerNumbers = map[string]float64{
	"rhythmStacks":  3,  // Battle Rhythm stacks at most 3 times
	"unbowedBelow":  30, // Unbowed: below this % of max HP
	"chargeCap":     30, // Charge: at most +30%
	"dragonBurn":    2,  // Dragonblood: burn turns
	"ricochetRange": 2,  // Ricochet: tiles from the target
	"pinTurns":      1,  // Pinning Shot: root turns
	"retaliateStun": 1,  // Retaliate: stun turns
	"farTiles":      4,  // Ranger set of 4: from this distance
}

func ax(kv ...any) []Affix {
	var out []Affix
	for i := 0; i < len(kv); i += 2 {
		out = append(out, Affix{Key: kv[i].(string), Value: kv[i+1].(int)})
	}
	return out
}

// classLegendaries are the ten class weapons: 2 unique powers and 3
// ordinary bonuses each, and a 4-piece class bonus. The pictures are the
// owner's picks from Grok's sheets.
var classLegendaries = []Unique{
	{Name: "Bloodcrescent", Slot: "w", WeaponType: "sword", Class: "Warrior",
		Affixes: ax("rhythm", 8, "unbowed", 25, "phys", 30, "crit", 12, "leech", 10), Full: ax("killReset", 1),
		Lore: "A dragon's claw, ground into a blade. It grows hungrier with every swing."},
	{Name: "Voidreaver", Slot: "w", WeaponType: "axe", Class: "Tank",
		Affixes: ax("aura", 15, "retaliate", 30, "dark", 30, "hpPct", 20, "thorns", 20), Full: ax("counterBoost", 50),
		Lore: "Each swing drinks from the dark between stars."},
	{Name: "Frostspire", Slot: "w", WeaponType: "lance", Class: "Knight",
		Affixes: ax("charge", 5, "valor", 1, "dmg", 25, "hit", 15, "defPct", 20), Full: ax("firstWard", 1),
		Lore: "Its tip was cut from a glacier that never melts."},
	{Name: "Dragonbone Lance", Slot: "w", WeaponType: "lance", Class: "Dragon Knight",
		Affixes: ax("dragonblood", 15, "wyrmhide", 1, "phys", 25, "crit", 10, "hpPct", 15), Full: ax("breathWide", 1),
		Lore: "Carved from the spine of the first dragon. The ruby still beats."},
	{Name: "Silverwing", Slot: "w", WeaponType: "bow", Class: "Ranger",
		Affixes: ax("eagle", 1, "pin", 25, "crit", 15, "critDmg", 50, "bossDmg", 30), Full: ax("farShot", 20),
		Lore: "Strung with an angel's hair; its arrows never tire."},
	{Name: "Cinderhowl", Slot: "w", WeaponType: "gun", Class: "Gunslinger",
		Affixes: ax("quick", 1, "ricochet", 20, "fire", 30, "crit", 12, "critDmg", 40), Full: ax("critRefund", 1),
		Lore: "Molten iron runs in its cracks. It is never cool enough to holster."},
	{Name: "Nightfang", Slot: "w", WeaponType: "dagger", Class: "Assassin",
		Affixes: ax("ambush", 30, "shadowstep", 1, "poisonCh", 30, "crit", 20, "bleedCh", 25), Full: ax("opener", 1),
		Lore: "It bites twice, and the second bite is poison."},
	{Name: "Umbral Scepter", Slot: "w", WeaponType: "wand", Class: "Mage",
		Affixes: ax("echo", 20, "overflow", 1, "dmg", 25, "crit", 12, "critDmg", 50), Full: ax("wideSpells", 1),
		Lore: "Two shadows hold its orb, and every spell is cast twice in their whispers."},
	{Name: "Staff of Eternity", Slot: "w", WeaponType: "staff", Class: "Healer",
		Affixes: ax("sanctuary", 20, "lifeline", 1, "healPct", 40, "regen", 6, "light", 20), Full: ax("cleanseHeal", 1),
		Lore: "Time bends gently around its bearer."},
	{Name: "Hexflame Crook", Slot: "w", WeaponType: "staff", Class: "Witch",
		Affixes: ax("hexMastery", 1, "siphon", 5, "dark", 30, "poisonCh", 25, "dmg", 15), Full: ax("hexSpread", 1),
		Lore: "A green fire that burns curses into whatever it touches."},
}

// legendExtra are the bonuses the legendary armor and charms gain so that
// every legendary has 5, one more than an epic.
var legendExtra = map[string][]Affix{
	"Dawnguard Helm":      ax("dr", 6, "thorns", 15, "regen", 3),
	"Aegis of Dawn":       ax("hpPct", 15, "thorns", 20, "regen", 3),
	"Dawnstride Greaves":  ax("hpPct", 10, "dr", 6, "regen", 3),
	"Stormweave Coif":     ax("hpPct", 10, "dmg", 10, "hit", 10),
	"Stormweave Mail":     ax("crit", 6, "dmg", 10),
	"Stormweave Treads":   ax("crit", 6, "hpPct", 10, "dmg", 8),
	"Nightbloom Hood":     ax("crit", 8, "dmg", 10, "hpPct", 10),
	"Nightbloom Robe":     ax("dmg", 10, "leech", 6),
	"Nightbloom Slippers": ax("crit", 6, "dmg", 8, "hpPct", 10),
	"Heart of the Stair":  ax("dr", 6, "crit", 8),
	"Phoenix Feather":     ax("hpPct", 15, "ward", 20),
}

// newSets are the set bonuses: the epic sets by weight (new) and the
// legendary sets, about double the epic ones.
var newSets = map[string]GearSet{
	"eH": {Name: "Vigil (heavy epic set)", Two: ax("dr", 5), Three: ax("hpPct", 10, "thorns", 10)},
	"eM": {Name: "Wolfsteel (medium epic set)", Two: ax("crit", 5), Three: ax("dmg", 8, "movA", 1)},
	"eL": {Name: "Shadowweave (light epic set)", Two: ax("evade", 5), Three: ax("dmg", 8, "regen", 2)},
	"H":  {Name: "Aegis of Dawn (heavy set)", Two: ax("dr", 10), Three: ax("hpPct", 20, "thorns", 20)},
	"M":  {Name: "Stormweave (medium set)", Two: ax("crit", 10), Three: ax("dmg", 16, "movA", 1)},
	"L":  {Name: "Nightbloom (light set)", Two: ax("evade", 10), Three: ax("dmg", 16, "regen", 4)},
}

// reworkLegendaries applies the approved rework to the data.
func reworkLegendaries(c *Content) {
	g := &c.Gear
	g.BonusCounts = []int{1, 2, 3, 3, 5} // exact counts; an epic adds its own epic bonus (4 in all)
	g.PowerNumbers = powerNumbers
	for _, p := range legendPowers {
		p.Ranges, p.Slots, p.LegendaryOnly, p.Power = make([][2]int, len(g.Rarities)), "w", true, true
		g.Affixes = append(g.Affixes, p)
	}
	var kept []Unique
	for _, u := range g.Uniques {
		if u.Slot == SlotWeapon {
			continue // the old one-per-weapon-type legendaries leave the loot table
		}
		u.Affixes = append(u.Affixes, legendExtra[u.Name]...)
		kept = append(kept, u)
	}
	g.Uniques = append(classLegendaries, kept...)
	var bases []WeaponBase
	for _, w := range g.Weapons {
		if w.Rarity != g.Legendary() {
			bases = append(bases, w)
		}
	}
	for _, u := range classLegendaries {
		bases = append(bases, WeaponBase{Name: u.Name, Type: u.WeaponType, Rarity: g.Legendary()})
	}
	g.Weapons = bases
	for k, s := range newSets {
		g.Sets[k] = s
	}
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-17",
		What: "Legendary rework: one legendary weapon per class (only that class can use it), each with 2 unique class " +
			"powers, 3 ordinary bonuses and a 4-piece class bonus (with all 3 pieces of one legendary armor set). Every " +
			"legendary has 5 bonuses. Items roll exactly 1 (common), 2 (uncommon), 3 (rare) or 3 + the epic bonus (epic) " +
			"bonuses. Epic armor gets 2- and 3-piece set bonuses by weight; legendary set bonuses are about double. " +
			"Before, the 8 legendary weapons went by weapon type, items rolled 0-1, 1-2 or 2-3 bonuses, and epics had no sets.",
		Apply: reworkLegendaries,
	})
}
