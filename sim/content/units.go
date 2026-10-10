package content

// HeroSkills is the skill bar a hero of this class and level starts a fight
// with (old initP and heroSks, index.html lines 436-439): the first four
// normal skills learned by this level in table order, the newest ultimate,
// the basic Strike, and Cloak for Assassins.
func (c *Content) HeroSkills(class string, level int) []Skill {
	var normal, ult []Skill
	for _, name := range c.SkillOrder {
		s := c.Skills[name]
		if s.Class != class || s.Level > level || s.Ascension != "" {
			continue
		}
		switch s.Kind {
		case "n":
			normal = append(normal, s)
		case "u":
			ult = append(ult, s)
		}
	}
	out := normal[:min(4, len(normal))]
	out = append([]Skill(nil), out...)
	if len(ult) > 0 {
		out = append(out, ult[len(ult)-1])
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
