package content

// StartingBuild is what a new hero of this class and level knows (old
// initP, index.html line 436): every skill of its class up to its level,
// the first four normal skills equipped in table order, and the newest
// ultimate in the ultimate slot.
func (c *Content) StartingBuild(class string, level int) (learned, equipped []string, ultimate string) {
	for _, name := range c.SkillOrder {
		s := c.Skills[name]
		if s.Class != class || s.Level > level || s.Ascension != "" {
			continue
		}
		learned = append(learned, name)
		switch s.Kind {
		case "n":
			if len(equipped) < 4 {
				equipped = append(equipped, name)
			}
		case "u":
			ultimate = name
		}
	}
	return learned, equipped, ultimate
}

// Bar is a hero's skill bar (old heroSks, index.html line 440): the four
// equipped skills, the ultimate, the basic Strike, and Cloak for Assassins.
func (c *Content) Bar(class string, equipped []string, ultimate string) []Skill {
	var out []Skill
	for _, n := range equipped {
		if s, ok := c.Skills[n]; ok {
			out = append(out, s)
		}
	}
	if s, ok := c.Skills[ultimate]; ok && ultimate != "" {
		out = append(out, s)
	}
	if s, ok := c.Skills["Strike"]; ok {
		out = append(out, s)
	}
	if class == "Assassin" {
		if s, ok := c.Skills["Cloak"]; ok {
			out = append(out, s)
		}
	}
	return out
}

// HeroSkills is the skill bar of a new hero of this class and level.
func (c *Content) HeroSkills(class string, level int) []Skill {
	_, eq, ult := c.StartingBuild(class, level)
	return c.Bar(class, eq, ult)
}

// MonsterSkills is the skill bar of a monster (old enemySks, line 438):
// its type's skills, plus level-unlocked ones unless it is a boss or elite.
func (c *Content) MonsterSkills(class string, level int, bossOrElite bool) []Skill {
	cl := c.Classes[class]
	out := append([]Skill(nil), cl.Skills...)
	if bossOrElite {
		return out
	}
	for _, ls := range cl.LevelSkills {
		if level >= ls.Level {
			out = append(out, ls.Skill)
		}
	}
	return out
}
