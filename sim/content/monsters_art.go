package content

// The owner's monster redraw (2026-10-17). The pictures are only art. These
// ten fighters are new data: four bosses after the old six, and six elites.
// The huge pit lord is not one of them.
func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-17",
		What: "Four bosses join the rotation after the old six (Bone Hydra, Fallen Seraph, Brood Matron, Ash Colossus; " +
			"Fallen Seraph flies) and six elites join the elite list (Crowned Slime, Goblin Captain, Grave Knight, " +
			"Dire Wolf, Iron Minotaur, Blight Reaper). The same floor seed can roll a different elite, and boss floors " +
			"after the old six show one of the new bosses before the list repeats.",
		Apply: addMonsterRoster,
	})
}

func sk(name string, power float64, min, max, cd, area int, magic bool, fx map[string]int) Skill {
	return Skill{Name: name, Power: power, MinRange: min, MaxRange: max, Cooldown: cd, Area: area, Magic: magic, Effects: fx}
}

func addMonsterRoster(c *Content) {
	put := func(cl Class) { c.Classes[cl.Name] = cl }
	put(Class{Name: "Ash Colossus", HP: 110, Atk: 14, Def: 11, Mov: 3, Color: "#5a4030", Skills: []Skill{
		sk("Hammer Fall", 1.4, 1, 1, 0, 0, false, nil),
		sk("Ash Storm", 1, 1, 3, 3, 1, true, nil),
	}})
	put(Class{Name: "Blight Reaper", HP: 46, Atk: 12, Def: 4, Mov: 5, Color: "#3a5a32", Skills: []Skill{
		sk("Blight Bolt", 1.1, 1, 3, 0, 0, true, nil),
		sk("Scythe Sweep", 1.3, 1, 1, 2, 0, false, nil),
	}})
	put(Class{Name: "Bone Hydra", HP: 104, Atk: 15, Def: 7, Mov: 4, Color: "#c8d0c0", Skills: []Skill{
		sk("Triple Bite", 1.2, 1, 1, 0, 0, false, nil),
		sk("Venom Breath", 0.9, 1, 3, 2, 1, false, map[string]int{"bleed": 3}),
	}})
	put(Class{Name: "Brood Matron", HP: 96, Atk: 13, Def: 8, Mov: 4, Color: "#3a2030", Skills: []Skill{
		sk("Leg Spike", 1.2, 1, 1, 0, 0, false, nil),
		sk("Web", 0.7, 1, 3, 2, 0, false, map[string]int{"snare": 2}),
	}})
	put(Class{Name: "Crowned Slime", HP: 48, Atk: 10, Def: 6, Mov: 3, Color: "#6a3a8a", Skills: []Skill{
		sk("Crown Slam", 1.2, 1, 1, 0, 0, false, nil),
		sk("Acid Bite", 1.1, 1, 1, 2, 0, false, nil),
	}})
	put(Class{Name: "Dire Wolf", HP: 40, Atk: 13, Def: 3, Mov: 6, Color: "#2a2a2a", Skills: []Skill{
		sk("Twin Bite", 1.2, 1, 1, 0, 0, false, nil),
		sk("Pounce", 1.3, 1, 2, 2, 0, false, nil),
	}})
	put(Class{Name: "Fallen Seraph", HP: 98, Atk: 16, Def: 6, Mov: 5, Color: "#d0c8a0", Skills: []Skill{
		sk("Spear of Ash", 1.3, 1, 2, 0, 0, false, nil),
		sk("Fallen Light", 1, 1, 4, 3, 1, true, nil),
	}})
	put(Class{Name: "Goblin Captain", HP: 42, Atk: 12, Def: 4, Mov: 5, Color: "#4a6a32", Skills: []Skill{
		sk("Jagged Cut", 1.2, 1, 1, 0, 0, false, nil),
		sk("War Chop", 1, 1, 1, 2, 0, false, nil),
	}})
	put(Class{Name: "Grave Knight", HP: 52, Atk: 11, Def: 9, Mov: 3, Color: "#6a2030", Skills: []Skill{
		sk("Grave Slash", 1.2, 1, 1, 0, 0, false, nil),
		sk("Shield Crash", 0.8, 1, 1, 3, 0, false, map[string]int{"stun": 1}),
	}})
	put(Class{Name: "Iron Minotaur", HP: 58, Atk: 13, Def: 8, Mov: 4, Color: "#6a5a48", Skills: []Skill{
		sk("Chain Axe", 1.3, 1, 1, 0, 0, false, nil),
		sk("Horn Rush", 1.1, 1, 1, 2, 0, false, map[string]int{"stun": 1}),
	}})
	c.Floors.Bosses = append(c.Floors.Bosses,
		Named{"Bone Hydra", "dark"},
		Named{"Fallen Seraph", "light"},
		Named{"Brood Matron", "dark"},
		Named{"Ash Colossus", "fire"},
	)
	c.Floors.Elites = append(c.Floors.Elites,
		Named{"Crowned Slime", "dark"},
		Named{"Goblin Captain", "dark"},
		Named{"Grave Knight", "dark"},
		Named{"Dire Wolf", "dark"},
		Named{"Iron Minotaur", "fire"},
		Named{"Blight Reaper", "dark"},
	)
	c.Flying["Fallen Seraph"] = true
}
