package content

// Ability is a class ability: a free action with a cooldown (old AB2 table,
// index.html lines 1129-1190). Most put statuses on the user, with numbers
// that tune them; a few also do something special, named by Effect.
type Ability struct {
	Name     string             `json:"name"`
	Cooldown int                `json:"cooldown"`
	Text     string             `json:"text"`
	Statuses map[string]int     `json:"statuses,omitempty"` // put on the user (replacing what was there)
	Params   map[string]float64 `json:"params,omitempty"`   // numbers the statuses use (old names: dblN, imhM...)
	Effect   string             `json:"effect,omitempty"`   // one of the special effects (see combat/ability.go)
	Amount   float64            `json:"amount,omitempty"`   // the effect's number, if it has one
	Target   *AbilityTarget     `json:"target,omitempty"`   // who or where it is aimed at, if anywhere
	Free     bool               `json:"free,omitempty"`     // usable even when finished, and has no cooldown
	Still    bool               `json:"still,omitempty"`    // only before moving this turn
}

// AbilityTarget is what an aimed ability needs.
type AbilityTarget struct {
	Kind  string `json:"kind"`            // "unit" or "tile"
	Range int    `json:"range"`           // up to this many tiles away
	Min   int    `json:"min,omitempty"`   // at least this many tiles away (tiles)
	N     int    `json:"n,omitempty"`     // how many tiles to pick
	Foe   bool   `json:"foe,omitempty"`   // an enemy (units)
	Empty bool   `json:"empty,omitempty"` // only tiles nobody stands on
}

// ClassAbility is a class's ability at each tier (Lv 1, 5, 20) and the
// ascensions' own versions.
type ClassAbility struct {
	Tiers      []Ability          `json:"tiers"`
	Ascensions map[string]Ability `json:"ascensions"`
}

// The tier levels.
var abilityTierLevels = []int{1, 5, 20}

// AbilityFor is the ability a hero of this class, level and ascension uses
// now (old abOf, line 1191). A Storm Dragoon in the sky uses Crash Down.
func (c *Content) AbilityFor(class string, level int, ascension string, inSky bool) (Ability, bool) {
	ca, ok := c.Abilities[class]
	if !ok {
		return Ability{}, false
	}
	if inSky {
		return c.CrashDown, true
	}
	if a, ok := ca.Ascensions[ascension]; ok {
		return a, true
	}
	t := 0
	for i, lv := range abilityTierLevels {
		if level >= lv {
			t = i
		}
	}
	return ca.Tiers[t], true
}

type p = map[string]float64
type st = map[string]int

func unit(r int, foe bool) *AbilityTarget { return &AbilityTarget{Kind: "unit", Range: r, Foe: foe} }
func tile(r, n int, empty bool) *AbilityTarget {
	return &AbilityTarget{Kind: "tile", Range: r, N: n, Empty: empty}
}

// oldCrashDown is the Storm Dragoon's landing from Sky Lord (old SKYC, line 1129).
var oldCrashDown = Ability{Name: "Crash Down", Text: "Land on any open tile within 8: 2.5× damage to every enemy within 2 tiles, and shock them.",
	Effect: "crashDown", Amount: 2.5, Target: tile(8, 1, true), Free: true}

// oldAbilities is the old AB2 table, transcribed: the old game stored each
// ability as a small program, so the effects are named here instead.
var oldAbilities = map[string]ClassAbility{
	"Warrior": {Tiers: []Ability{
		{Name: "Double Strike", Cooldown: 3, Text: "Your next skill (not ultimates) hits twice.", Statuses: st{"dbl": 2}, Params: p{"dblN": 2, "dblSure": 0, "dblB": 0, "dblC": 0}},
		{Name: "Double Strike+", Cooldown: 3, Text: "Next skill hits twice; the second hit can't miss and deals +15%.", Statuses: st{"dbl": 2}, Params: p{"dblN": 2, "dblSure": 1, "dblB": 0.15, "dblC": 0}},
		{Name: "Triple Strike", Cooldown: 3, Text: "Next skill hits three times; the extra hits can't miss.", Statuses: st{"dbl": 2}, Params: p{"dblN": 3, "dblSure": 1, "dblB": 0, "dblC": 0}},
	}, Ascensions: map[string]Ability{
		"Duelist":   {Name: "Flurry of Blades", Cooldown: 3, Text: "Next skill hits four times, extra hits can't miss, +15% crit chance per hit, and Riposte counters hit twice this turn.", Statuses: st{"dbl": 2, "flurry": 2}, Params: p{"dblN": 4, "dblSure": 1, "dblB": 0, "dblC": 15}},
		"Berserker": {Name: "Blood Frenzy", Cooldown: 5, Text: "Spend 20% of your HP: for 3 turns, +50% damage and 25% life steal. Each kill adds a turn.", Statuses: st{"frenzy": 3}, Effect: "payHP", Amount: 0.2},
		"Paladin":   {Name: "Oath of Protection", Cooldown: 5, Text: "For 2 turns, enemies within 4 tiles must attack you. You take 30% less damage and heal 10% of max HP each time you're hit.", Statuses: st{"taunt": 2}},
	}},
	"Tank": {Tiers: []Ability{
		{Name: "Iron Maiden", Cooldown: 4, Text: "Next attack adds your full DEF to its damage. +25% DEF for 3 turns.", Statuses: st{"imh": 2, "ironm": 3}, Params: p{"imhM": 1, "imhStun": 0, "imhCl": 0, "ironP": 0.25}},
		{Name: "Iron Maiden+", Cooldown: 4, Text: "Next attack adds 1.5× your DEF. +30% DEF for 3 turns.", Statuses: st{"imh": 2, "ironm": 3}, Params: p{"imhM": 1.5, "imhStun": 0, "imhCl": 0, "ironP": 0.3}},
		{Name: "Iron Maiden II", Cooldown: 4, Text: "Next attack adds 2× your DEF and stuns. +35% DEF for 3 turns.", Statuses: st{"imh": 2, "ironm": 3}, Params: p{"imhM": 2, "imhStun": 1, "imhCl": 0, "ironP": 0.35}},
	}, Ascensions: map[string]Ability{
		"Juggernaut": {Name: "Unbroken Maiden", Cooldown: 4, Text: "Next attack adds 2.5× your DEF, stuns, and hits every adjacent enemy. +50% DEF for 3 turns, and the first killing blow on you leaves you at 1 HP.", Statuses: st{"imh": 2, "ironm": 3, "unb": 3}, Params: p{"imhM": 2.5, "imhStun": 1, "imhCl": 1, "ironP": 0.5}, Effect: "resetUnbroken"},
		"Guardian":   {Name: "Frost Bulwark", Cooldown: 5, Text: "For 3 turns, allies within 2 tiles of you get +30% DEF, and any enemy that hits one of them is rooted.", Statuses: st{"bulwark": 3}},
		"Warlord":    {Name: "War Banner", Cooldown: 5, Text: "Plant a banner on a tile within 3: for 3 turns, allies within 3 tiles of it get +20% damage and +1 move, and enemies in range burn each turn.", Effect: "banner", Target: tile(3, 1, false)},
	}},
	"Mage": {Tiers: []Ability{
		{Name: "Inner Element", Cooldown: 3, Text: "Next attack +30%, can't miss, and always applies its element's status.", Statuses: st{"inner": 2}, Params: p{"inP": 0.3, "inExt": 0, "inArea": 0, "inN": 1}},
		{Name: "Inner Element+", Cooldown: 3, Text: "Next attack +40%, can't miss, and its status lasts 1 turn longer.", Statuses: st{"inner": 2}, Params: p{"inP": 0.4, "inExt": 1, "inArea": 0, "inN": 1}},
		{Name: "Inner Element II", Cooldown: 3, Text: "Next attack +50%, can't miss, longer status, and its area grows 1 tile (single-target spells splash to adjacent enemies).", Statuses: st{"inner": 2}, Params: p{"inP": 0.5, "inExt": 1, "inArea": 1, "inN": 1}},
	}, Ascensions: map[string]Ability{
		"Stormcaller": {Name: "Elemental Overload", Cooldown: 3, Text: "Your next 2 spells get +60%, can't miss, longer status and +1 tile area.", Statuses: st{"inner": 3}, Params: p{"inP": 0.6, "inExt": 1, "inArea": 1, "inN": 2}},
		"Pyromancer":  {Name: "Firestorm Field", Cooldown: 5, Text: "Set a 3×3 area within 5 tiles ablaze for 3 turns: enemies in it burn every turn, and you deal +20% fire damage while it lasts.", Statuses: st{"fstorm": 3}, Effect: "fireField", Target: tile(5, 1, false)},
		"Cryomancer":  {Name: "Absolute Freeze", Cooldown: 5, Text: "Freeze one enemy within 6 tiles for 2 turns: it can't act, takes +50% damage from your party, and can't be healed.", Effect: "freeze", Target: unit(6, true)},
	}},
	"Healer": {Tiers: []Ability{
		{Name: "Enlightenment", Cooldown: 4, Text: "2 turns: healing +30%, and every heal reaches the whole party.", Statuses: st{"enl": 2}, Params: p{"enlP": 0.3, "enlC": 0}},
		{Name: "Enlightenment+", Cooldown: 4, Text: "2 turns: healing +40%, party heals, and each heal removes one debuff.", Statuses: st{"enl": 2}, Params: p{"enlP": 0.4, "enlC": 1}},
		{Name: "Enlightenment II", Cooldown: 4, Text: "3 turns: healing +50%, party heals, and each heal removes all debuffs.", Statuses: st{"enl": 3}, Params: p{"enlP": 0.5, "enlC": 2}},
	}, Ascensions: map[string]Ability{
		"Saint":         {Name: "Divine Choir", Cooldown: 4, Text: "3 turns: healing +60%, party heals that cleanse everything, overheal shields doubled, and one fallen ally is revived right away.", Statuses: st{"enl": 3, "choir": 3}, Params: p{"enlP": 0.6, "enlC": 2}, Effect: "reviveOne"},
		"Oracle":        {Name: "Time Shift", Cooldown: 5, Text: "One ally acts again this turn (move and action refreshed) and gains Haste.", Effect: "timeShift", Target: unit(99, false)},
		"Plague Doctor": {Name: "Plague Bloom", Cooldown: 5, Text: "3 turns: each heal also deals half its amount as dark damage to every enemy within 2 tiles of the healed ally, and poisons them.", Statuses: st{"pbloom": 3}},
	}},
	"Ranger": {Tiers: []Ability{
		{Name: "Hunter's Eye", Cooldown: 3, Text: "This turn: +3 range, and your next skill (not ultimates) hits up to 3 enemies at +30%.", Statuses: st{"heye": 1}, Params: p{"heR": 3, "heN": 3, "heP": 0.3, "heI": 0, "heK": 0}},
		{Name: "Hunter's Eye+", Cooldown: 3, Text: "This turn: +3 range, next skill hits up to 4 enemies at +35%.", Statuses: st{"heye": 1}, Params: p{"heR": 3, "heN": 4, "heP": 0.35, "heI": 0, "heK": 0}},
		{Name: "Hunter's Eye II", Cooldown: 3, Text: "This turn: +4 range, next skill hits up to 5 enemies at +40% and ignores 30% DEF.", Statuses: st{"heye": 1}, Params: p{"heR": 4, "heN": 5, "heP": 0.4, "heI": 0.3, "heK": 0}},
	}, Ascensions: map[string]Ability{
		"Tempest Archer": {Name: "Storm Eye", Cooldown: 4, Text: "For 2 turns: +5 range, and each skill hits every enemy in range (up to 6) at +50%, ignoring 30% DEF.", Statuses: st{"heye": 2}, Params: p{"heR": 5, "heN": 6, "heP": 0.5, "heI": 0.3, "heK": 1}},
		"Sniper":         {Name: "Deadeye Focus", Cooldown: 5, Text: "Stay still this turn: your next shot can hit any enemy on the map, always crits for 3× damage, and ignores all DEF.", Statuses: st{"deadeye": 2}, Effect: "standStill", Still: true},
		"Trapper":        {Name: "Snare Field", Cooldown: 4, Text: "Hide 3 traps on tiles within 6. An enemy that steps on one is rooted and poisoned for 3 turns and takes +30% damage from you.", Effect: "traps", Target: tile(6, 3, true)},
	}},
	"Assassin": {Tiers: []Ability{
		{Name: "Bloodlust", Cooldown: 5, Text: "3 turns: cloaked even while attacking, +3 move, +50% crit chance.", Statuses: st{"blood": 3}, Params: p{"blM": 3, "blC": 50, "blD": 0, "blR": 0, "blX": 0}},
		{Name: "Bloodlust+", Cooldown: 5, Text: "3 turns: cloaked, +3 move, +60% crit chance, and crits deal +25% more.", Statuses: st{"blood": 3}, Params: p{"blM": 3, "blC": 60, "blD": 25, "blR": 0, "blX": 0}},
		{Name: "Bloodlust II", Cooldown: 5, Text: "4 turns: cloaked, +4 move, +70% crit chance, crits +25%, and each kill recharges Cloak.", Statuses: st{"blood": 4}, Params: p{"blM": 4, "blC": 70, "blD": 25, "blR": 1, "blX": 0}},
	}, Ascensions: map[string]Ability{
		"Blade Dancer": {Name: "Crimson Bloodlust", Cooldown: 5, Text: "4 turns: cloaked, +5 move, every hit crits, crits +50%, Cloak recharges on kills, and each kill adds a turn.", Statuses: st{"blood": 4}, Params: p{"blM": 5, "blC": 100, "blD": 50, "blR": 1, "blX": 1}},
		"Phantom":      {Name: "Shadow Clones", Cooldown: 5, Text: "Summon 2 decoy clones next to you for 3 turns. Enemies prefer to attack them, and a clone that's destroyed explodes.", Effect: "clones"},
		"Viper":        {Name: "Death Mark", Cooldown: 4, Text: "Mark one enemy for 3 turns: it takes +50% damage from your party, and when it dies its poison spreads to every enemy within 2 tiles.", Effect: "deathMark", Target: unit(8, true)},
	}},
	"Gunslinger": {Tiers: []Ability{
		{Name: "Quick Draw", Cooldown: 3, Text: "This turn, fire two different skills, one from each gun, at one target or two.", Statuses: st{"qd": 1}, Params: p{"qdN": 2, "qdSame": 0, "qdP": 0, "qd2P": 0}},
		{Name: "Quick Draw+", Cooldown: 3, Text: "Two different skills this turn, and the second deals +25%.", Statuses: st{"qd": 1}, Params: p{"qdN": 2, "qdSame": 0, "qdP": 0, "qd2P": 0.25}},
		{Name: "Quick Draw II", Cooldown: 3, Text: "Two skills this turn (the same skill is allowed), both +20%.", Statuses: st{"qd": 1}, Params: p{"qdN": 2, "qdSame": 1, "qdP": 0.2, "qd2P": 0}},
	}, Ascensions: map[string]Ability{
		"Outlaw":        {Name: "Dead Man's Hand", Cooldown: 3, Text: "Three skills this turn (repeats allowed), each +25%, at any mix of targets.", Statuses: st{"qd": 1}, Params: p{"qdN": 3, "qdSame": 1, "qdP": 0.25, "qd2P": 0}},
		"Bounty Hunter": {Name: "Wanted Poster", Cooldown: 4, Text: "Mark one enemy until it dies: your shots at it can't miss and deal +50%. Killing it recharges Wanted Poster and gives triple gold and a guaranteed item.", Effect: "wanted", Target: unit(99, true)},
		"Trickshot":     {Name: "Overwatch", Cooldown: 4, Text: "Ends your turn aiming a 4-tile cone. During the enemy turn you shoot every enemy that moves inside it, at 70% damage.", Effect: "overwatch", Target: &AbilityTarget{Kind: "tile", Range: 1, Min: 1, N: 1}},
	}},
	"Dragon Knight": {Tiers: []Ability{
		{Name: "Beast Mode", Cooldown: 5, Text: "3 turns: +7 move, attacks add 25% of their damage as fire, 35% chance to burn for 3 turns.", Statuses: st{"beast": 3}, Params: p{"bmM": 7, "bmF": 0.25, "bmB": 0.35, "bmS": 0, "bmT": 0}},
		{Name: "Beast Mode+", Cooldown: 5, Text: "3 turns: +7 move, 30% added fire, 45% burn chance.", Statuses: st{"beast": 3}, Params: p{"bmM": 7, "bmF": 0.3, "bmB": 0.45, "bmS": 0, "bmT": 0}},
		{Name: "Beast Mode II", Cooldown: 5, Text: "4 turns: +7 move, 35% added fire, 55% burn chance, and attacks splash 50% to adjacent enemies.", Statuses: st{"beast": 4}, Params: p{"bmM": 7, "bmF": 0.35, "bmB": 0.55, "bmS": 0.5, "bmT": 0}},
	}, Ascensions: map[string]Ability{
		"Inferno Rider": {Name: "Inferno Beast", Cooldown: 5, Text: "4 turns: +8 move, 50% added fire, attacks always burn, splash, and every tile you end on burns for 2 turns.", Statuses: st{"beast": 4}, Params: p{"bmM": 8, "bmF": 0.5, "bmB": 1, "bmS": 0.5, "bmT": 1}},
		"Frost Wyrm":    {Name: "Glacial Scales", Cooldown: 5, Text: "3 turns: ranged attacks can't hurt you, melee damage to you is cut by 40%, and melee attackers are rooted.", Statuses: st{"scales": 3}},
		"Storm Dragoon": {Name: "Sky Lord", Cooldown: 5, Text: "Take off (ends your turn): untargetable until your next turn, then Crash Down on any open tile within 8 for 2.5× damage to everything within 2 tiles, shocking them.", Statuses: st{"sky": 2}, Effect: "skyLord"},
	}},
	"Witch": {Tiers: []Ability{
		{Name: "Poison Ivy", Cooldown: 4, Text: "2 turns: every hit makes the target Immobile for 2 turns and applies Ivy (−30% ATK and DEF), stacking.", Statuses: st{"ivyA": 2}, Params: p{"ivP": 0.3, "ivPo": 0, "ivS": 0}},
		{Name: "Poison Ivy+", Cooldown: 4, Text: "3 turns: Immobile and Ivy, and Ivy also poisons.", Statuses: st{"ivyA": 3}, Params: p{"ivP": 0.3, "ivPo": 1, "ivS": 0}},
		{Name: "Poison Ivy II", Cooldown: 4, Text: "3 turns: Immobile, poisoning Ivy at −40%, and each hit spreads Ivy to one enemy next to the target.", Statuses: st{"ivyA": 3}, Params: p{"ivP": 0.4, "ivPo": 1, "ivS": 1}},
	}, Ascensions: map[string]Ability{
		"Hex Queen":   {Name: "Queen's Garden", Cooldown: 4, Text: "3 turns: Immobile, poisoning Ivy at −50%, spreading to every enemy next to the target, and all your debuffs last 1 turn longer.", Statuses: st{"ivyA": 3}, Params: p{"ivP": 0.5, "ivPo": 1, "ivS": 9}},
		"Necromancer": {Name: "Grave Bloom", Cooldown: 5, Text: "Raise a skeleton for every monster killed on this floor. For 3 turns every new kill rises too. Skeletons fight on their own for the rest of the floor and explode when destroyed, poisoning and applying Ivy to adjacent enemies.", Statuses: st{"gbloom": 3}, Effect: "graveBloom"},
		"Blood Witch": {Name: "Blood Pact", Cooldown: 5, Text: "3 turns: move to any open tile, heal for 20% of all damage you deal, and every hit causes Heavy Bleed (5 turns, double damage) whose damage heals you.", Statuses: st{"vamp": 3}},
	}},
	"Knight": {Tiers: []Ability{
		{Name: "Vow of Honor", Cooldown: 5, Text: "3 turns: +30% attack, +5 move.", Statuses: st{"vow": 3}, Params: p{"vwP": 0.3, "vwM": 5, "vwC": 0, "vwA": 0}},
		{Name: "Vow of Honor+", Cooldown: 5, Text: "3 turns: +35% attack, +5 move, and charge bonus +50%.", Statuses: st{"vow": 3}, Params: p{"vwP": 0.35, "vwM": 5, "vwC": 0.5, "vwA": 0}},
		{Name: "Vow of Honor II", Cooldown: 5, Text: "3 turns: +40% attack, +6 move, charge bonus +50%, and allies next to you get +15% attack.", Statuses: st{"vow": 3}, Params: p{"vwP": 0.4, "vwM": 6, "vwC": 0.5, "vwA": 0.15}},
	}, Ascensions: map[string]Ability{
		"Storm Rider":  {Name: "Thunderous Oath", Cooldown: 5, Text: "4 turns: +50% attack, +7 move, charge bonus doubled, and adjacent allies get +20% attack.", Statuses: st{"vow": 4}, Params: p{"vwP": 0.5, "vwM": 7, "vwC": 1, "vwA": 0.2}},
		"Templar":      {Name: "Sacred Vow", Cooldown: 5, Text: "3 turns: half of all damage aimed at allies within 2 tiles hits you instead, you heal 10% of max HP each turn, and those allies get +20% DEF.", Statuses: st{"svow": 3}},
		"Black Knight": {Name: "Oath of Ruin", Cooldown: 5, Text: "3 turns: every kill gives you a full extra turn (move and attack) and stuns enemies within 2 tiles.", Statuses: st{"ruin": 3}},
	}},
}
