package content

// Change is one approved difference from the old game's data. The live data
// files already include every change; the data test checks that the live
// files equal the old tables plus exactly these changes, so nothing differs
// by accident.
type Change struct {
	Date  string
	What  string
	Apply func(*Content)
}

// Changes are the approved data changes, oldest first.
var Changes = []Change{
	{
		Date: "2026-10-10",
		What: "Monsters use the same version of a skill as heroes (only the animation differs). " +
			"Before, six differed: monster Fireball, Thunder and Frost Bolt reached 1-3 tiles instead of 2-6, " +
			"and monster Shot, Snaring Shot and Power Shot reached 7 tiles instead of 6.",
		Apply: func(c *Content) {
			for name, cl := range c.Classes {
				for i, s := range cl.Skills {
					if h, ok := c.Skills[s.Name]; ok {
						cl.Skills[i] = h
					}
				}
				for i, ls := range cl.LevelSkills {
					if h, ok := c.Skills[ls.Skill.Name]; ok {
						cl.LevelSkills[i].Skill = h
					}
				}
				c.Classes[name] = cl
			}
		},
	},
}

func init() {
	Changes = append(Changes, Change{
		Date:  "2026-10-10",
		What:  "A fallen hero can be revived however they died. Before, heroes killed by poison or bleeding could not be revived.",
		Apply: func(c *Content) { c.Combat.ReviveAnyDeath = true },
	})
}

// ApplyChanges turns the old game's data into today's data.
func ApplyChanges(c *Content) {
	for _, ch := range Changes {
		ch.Apply(c)
	}
}
