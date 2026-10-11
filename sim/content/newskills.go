package content

// newSkills are the 54 skills the owner approved on 2026-10-10 for the even
// skill schedule (docs/skill-plan.md): 53 normal skills and the Healer's new
// Lv 20 ultimate, Last Light. Their levels come from heroSchedule.
func newSkills() []Skill {
	foe := func(name, class string, pw float64, lo, hi, cd int, text string) Skill {
		return Skill{Name: name, Class: class, Kind: "n", Power: pw, MinRange: lo, MaxRange: hi, Cooldown: cd, Text: text}
	}
	self := func(name, class string, cd int, buffs map[string]int, text string) Skill {
		return Skill{Name: name, Class: class, Kind: "n", Aim: AimSelf, Cooldown: cd, Buffs: buffs, Text: text}
	}
	with := func(s Skill, f func(*Skill)) Skill { f(&s); return s }
	return []Skill{
		// Warrior
		with(foe("Shoulder Check", "Warrior", 0.9, 1, 1, 2, "Shoves the target 2 tiles back. If it slams into a wall or another unit it takes +50% and is stunned 1 turn."),
			func(s *Skill) { s.Push, s.Slam = 2, true }),
		self("Riposte Stance", "Warrior", 4, map[string]int{"riposte": 2}, "Until your next turn, strike back at every enemy that hits you in melee (power 0.8)."),
		with(foe("Leaping Cleave", "Warrior", 1.1, 1, 4, 3, "Leap up to 3 tiles to land next to the target, then hit it and every enemy next to you."),
			func(s *Skill) { s.Leap, s.Area, s.Around = 3, 1, true }),
		with(foe("Sunder", "Warrior", 1.2, 1, 1, 3, "Strips every buff and shield from the target."), func(s *Skill) { s.Dispel = true }),
		self("Battle Trance", "Warrior", 4, map[string]int{"trance": 2}, "For 2 turns: +25% damage and +1 Move, but you take 15% more damage."),
		with(foe("Momentum", "Warrior", 1, 1, 1, 3, "If it defeats the target you may act once more this turn (once per turn)."), func(s *Skill) { s.Trick = "momentum" }),
		// Tank
		with(foe("Hook Chain", "Tank", 0.7, 2, 3, 2, "Throw a chain 2-3 tiles and drag the target to the tile next to you."), func(s *Skill) { s.Pull = true }),
		{Name: "Bodyguard", Class: "Tank", Kind: "n", Aim: AimAlly, MinRange: 1, MaxRange: 3, Cooldown: 4, Trick: "bodyguard",
			Text: "Pick an ally within 3 tiles. For 2 turns, hits aimed at them hit you instead, for 70% of the damage."},
		{Name: "Iron Shell", Class: "Tank", Kind: "n", Aim: AimSelf, Cooldown: 4, Shield: 0.25, Text: "Gain a shield worth 25% of your max HP."},
		{Name: "Taunting Roar", Class: "Tank", Kind: "n", Aim: AimSelf, Area: 3, Cooldown: 4, Trick: "taunt",
			Text: "Every enemy within 3 tiles must attack you on its next turn."},
		with(foe("Aftershock", "Tank", 1, 1, 1, 3, "Hits the target and its neighbors and snares them; at the start of the next enemy turn the ground shakes again for half damage."),
			func(s *Skill) { s.Area, s.Effects, s.Trick = 1, map[string]int{"snare": 1}, "aftershock" }),
		{Name: "Unstoppable", Class: "Tank", Kind: "n", Aim: AimLine, Power: 0.9, MaxRange: 4, Cooldown: 4, Trick: "unstoppable",
			Text: "Bulldoze up to 4 tiles in a straight line. Every enemy in the way is hit and shoved to the side."},
		// Mage
		{Name: "Blink", Class: "Mage", Kind: "n", Aim: AimTile, MinRange: 1, MaxRange: 4, Cooldown: 4, Free: true, Trick: "blink",
			Text: "Teleport to an open tile up to 4 away. You can still attack this turn."},
		with(foe("Mana Burn", "Mage", 0.8, 2, 6, 3, "Arcane hit that adds 1 turn to every cooldown the target has."),
			func(s *Skill) { s.Magic, s.Trick = true, "manaBurn" }),
		with(foe("Static Field", "Mage", 0.6, 2, 6, 3, "Lightning within 2 tiles of the target; those tiles stay charged for 2 turns and shock any enemy that ends its turn on them."),
			func(s *Skill) { s.Element, s.Area, s.Field, s.FieldRange, s.FieldTurns = "elec", 2, "static", 2, 2 }),
		with(foe("Gravity Well", "Mage", 0.7, 2, 6, 3, "Arcane hit on an area that pulls every enemy within 2 tiles one step toward the center."),
			func(s *Skill) { s.Magic, s.Area, s.Gather = true, 1, 2 }),
		{Name: "Echo", Class: "Mage", Kind: "n", Aim: AimEcho, Cooldown: 4, Trick: "echo",
			Text: "Cast the last spell you used again at 60% power on a new target, without starting its cooldown."},
		// Healer
		{Name: "Last Light", Class: "Healer", Kind: "u", Aim: AimSelf, Support: true, Area: 3, Cooldown: 6, Buffs: map[string]int{"lastlight": 2}, Trick: "lastLight",
			Text: "You and allies within 3 tiles are healed 30% and, for 2 turns, a blow that would defeat them leaves them at 1 HP instead."},
		// Ranger
		{Name: "Steady Aim", Class: "Ranger", Kind: "n", Aim: AimSelf, Cooldown: 3, Free: true, Still: true, Buffs: map[string]int{"aim": 1},
			Text: "Stand still this turn: your next shot gets +2 range and cannot miss."},
		with(foe("Disengage", "Ranger", 0.8, 2, 6, 3, "Shoot, then leap 2 tiles directly away from the target."), func(s *Skill) { s.Away = 2 }),
		with(foe("Hunter's Mark", "Ranger", 0.5, 2, 6, 3, "Marks the target for 3 turns: everyone on your side deals it +20% damage."),
			func(s *Skill) { s.Effects = map[string]int{"hmark": 3} }),
		with(foe("Concussive Arrow", "Ranger", 0.9, 2, 6, 2, "Knocks the target 2 tiles back."), func(s *Skill) { s.Push = 2 }),
		with(foe("Flaming Volley", "Ranger", 0.7, 2, 6, 3, "Fire arrows on the target and its neighbors; the ground burns for 2 turns."),
			func(s *Skill) {
				s.Element, s.Physical, s.Area, s.Field, s.FieldRange, s.FieldTurns = "fire", true, 1, "fire", 1, 2
			}),
		// Assassin
		with(foe("Shadowstep", "Assassin", 1, 1, 4, 3, "Teleport to the far side of an enemy within 4 tiles and stab it in the back."), func(s *Skill) { s.Behind = true }),
		with(foe("Garrote", "Assassin", 0.7, 1, 1, 3, "Silences the target for 2 turns: it can only use its most basic attack."),
			func(s *Skill) { s.Effects = map[string]int{"silence": 2} }),
		{Name: "Bait and Switch", Class: "Assassin", Kind: "n", Aim: AimAny, MinRange: 1, MaxRange: 3, Cooldown: 3, Trick: "swap",
			Text: "Swap places with any unit, friend or foe, within 3 tiles."},
		with(foe("Crippling Cut", "Assassin", 0.9, 1, 1, 2, "Target bleeds 2 turns and loses 2 Move for 2 turns."),
			func(s *Skill) { s.Effects = map[string]int{"bleed": 2, "slow": 2} }),
		with(foe("Open Veins", "Assassin", 0.6, 1, 1, 3, "Every bleed and poison on the target ticks right now, and their turns left double."),
			func(s *Skill) { s.Trick = "openVeins" }),
		with(foe("Twin Fangs", "Assassin", 0.9, 1, 1, 2, "Strike two different enemies next to you."), func(s *Skill) { s.ExtraTargets = 2 }),
		// Gunslinger
		with(foe("Boot Shot", "Gunslinger", 0.9, 1, 1, 2, "Kick an enemy next to you 2 tiles away and fire as it flies."), func(s *Skill) { s.Push = 2 }),
		with(foe("Disarming Shot", "Gunslinger", 0.8, 2, 6, 3, "Shoots the weapon hand: target deals 30% less damage for 2 turns."),
			func(s *Skill) { s.Effects = map[string]int{"weak": 2} }),
		with(foe("Roll Out", "Gunslinger", 0.9, 2, 6, 3, "Dive-roll up to 3 tiles, then shoot."), func(s *Skill) { s.Aim, s.Dash = AimRoll, 3 }),
		{Name: "Powder Keg", Class: "Gunslinger", Kind: "n", Aim: AimTile, Power: 1.6, MinRange: 1, MaxRange: 4, Cooldown: 4, Element: "fire", Physical: true, Area: 1, Trick: "keg",
			Text: "Toss a keg on a tile up to 4 away. It blows up at the start of the next enemy turn: fire on everything within 1 tile."},
		{Name: "Lucky Coin", Class: "Gunslinger", Kind: "n", Aim: AimSelf, Cooldown: 5, Free: true, Buffs: map[string]int{"coin": 2},
			Text: "Your next 2 shots are sure critical hits."},
		with(foe("High Noon", "Gunslinger", 1.3, 2, 6, 3, "+10% damage for every enemy in your range (max +60%)."), func(s *Skill) { s.Trick = "highNoon" }),
		// Knight
		{Name: "Trample", Class: "Knight", Kind: "n", Aim: AimLine, Power: 0.8, MaxRange: 4, Cooldown: 3, Trick: "trample",
			Text: "Ride straight through up to 3 tiles of enemies to an open tile, hitting each one."},
		{Name: "Shield of Faith", Class: "Knight", Kind: "n", Aim: AimAlly, MinRange: 1, MaxRange: 3, Cooldown: 3, Shield: 0.2,
			Text: "Give an ally within 3 tiles a shield worth 20% of your max HP."},
		with(foe("Unhorse", "Knight", 1.2, 1, 1, 3, "Knocks the target 1 tile back and it can't move next turn."),
			func(s *Skill) { s.Push, s.Effects = 1, map[string]int{"snare": 1} }),
		{Name: "Lance Wall", Class: "Knight", Kind: "n", Aim: AimSelf, Power: 0.9, Cooldown: 4, Buffs: map[string]int{"lwall": 2},
			Text: "Until your next turn, any enemy that steps next to you is struck first (power 0.9)."},
		{Name: "Rescue", Class: "Knight", Kind: "n", Aim: AimAlly, MinRange: 2, MaxRange: 4, Cooldown: 4, Trick: "rescue",
			Text: "Pull an ally within 4 tiles to a tile next to you; they take 20% less damage for 1 turn."},
		with(foe("Relentless Charge", "Knight", 1.3, 1, 1, 3, "A charge (charge damage); afterwards you may keep moving with the Move you have left."),
			func(s *Skill) { s.Charge, s.Trick = true, "relentless" }),
		// Dragon Knight
		{Name: "Tail Sweep", Class: "Dragon Knight", Kind: "n", Aim: AimSelf, Power: 0.8, Area: 1, Around: true, Push: 1, Cooldown: 2,
			Text: "Hits every enemy next to you and shoves each 1 tile away."},
		self("Dragon Scales", "Dragon Knight", 4, map[string]int{"dscale": 2}, "For 2 turns take 30% less spell damage and reflect 20% of melee damage back."),
		{Name: "Wyrm Roar", Class: "Dragon Knight", Kind: "n", Aim: AimSelf, Area: 2, Around: true, Push: 1, NoDamage: true, Cooldown: 4, Effects: map[string]int{"weak": 2},
			Text: "Enemies within 2 tiles are pushed 1 tile away and deal 30% less damage for 2 turns."},
		{Name: "Ember Trail", Class: "Dragon Knight", Kind: "n", Aim: AimLine, MaxRange: 4, Cooldown: 3, Trick: "emberTrail",
			Text: "Fly up to 4 tiles in a line; every tile you pass over burns for 2 turns."},
		with(foe("Skewer", "Dragon Knight", 1.2, 2, 2, 3, "Lance a target 2 tiles away and drag it next to you."), func(s *Skill) { s.Pull = true }),
		{Name: "Wyrmbond", Class: "Dragon Knight", Kind: "n", Aim: AimSelf, Power: 0.9, Cooldown: 5, Buffs: map[string]int{"wyrm": 3},
			Text: "For 3 turns your dragon claws an enemy next to you at the end of each of your turns (power 0.9)."},
		// Witch
		{Name: "Voodoo Link", Class: "Witch", Kind: "n", Aim: AimTwo, MinRange: 2, MaxRange: 6, Cooldown: 4, Trick: "link",
			Text: "Bind two enemies in range for 3 turns: half of any damage one takes also hits the other."},
		with(foe("Wither", "Witch", 0.6, 2, 6, 3, "For 3 turns the target heals only half and can't gain buffs."),
			func(s *Skill) { s.Element, s.Effects = "dark", map[string]int{"wither": 3} }),
		{Name: "Puppet Strings", Class: "Witch", Kind: "n", Aim: AimShove, MinRange: 1, MaxRange: 4, Cooldown: 4, Trick: "puppet",
			Text: "Move an enemy within 4 tiles up to 2 tiles wherever you choose."},
		with(foe("Soul Siphon", "Witch", 1, 2, 6, 2, "Dark hit that heals you for half the damage."), func(s *Skill) { s.Element, s.Drain = "dark", 0.5 }),
		with(foe("Contagion", "Witch", 0, 2, 6, 3, "Copies every curse and poison on the target onto enemies within 2 tiles of it."),
			func(s *Skill) { s.NoDamage, s.Trick = true, "contagion" }),
		{Name: "Bog of Despair", Class: "Witch", Kind: "n", Aim: AimTile, MinRange: 2, MaxRange: 6, Cooldown: 4, Field: "swamp", FieldRange: 1, FieldTurns: 3,
			Text: "A swamp 1 tile around a spot for 3 turns: enemies inside are snared and poisoned."},
	}
}
