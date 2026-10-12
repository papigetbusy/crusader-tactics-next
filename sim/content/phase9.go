package content

// Phase 9 changes (owner, 2026-10-11; design-lock 20, docs/phase9-proposal.md).
func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-11",
		What: "Fire Emblem counterattacks: a unit hit by a single-target attack that survives strikes back once with " +
			"a full Strike if the attacker is within its Strike range. Not when stunned, frozen, rooted or cloaked, and " +
			"not against area skills, ultimates, heals or buffs. Before, only a Duelist (and Riposte Stance) struck back.",
		Apply: func(c *Content) { c.Combat.Counterattacks, c.Combat.CounterPower = true, 1 },
	})
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-11",
		What: "Smarter enemies as you climb: from floor 20 (bosses and elites from floor 10) a monster prefers a hero " +
			"it can defeat this turn, then healers, then heroes weak to its attack; from floor 50 it avoids ending its " +
			"move within 2 tiles of two or more heroes when it can't attack. Before, monsters always hit the hero with " +
			"the lowest HP in reach.",
		Apply: func(c *Content) {
			c.AI.SmartFrom, c.AI.SmartBossFrom, c.AI.CautiousFrom = 20, 10, 50
			c.AI.KillWeight, c.AI.HealerWeight, c.AI.WeakWeight = 100, 40, 30
			c.AI.TileWeight, c.AI.CautiousPenalty = 0.02, 5
		},
	})
}
