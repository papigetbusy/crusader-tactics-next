package content

import "slices"

// ScheduleLevels are the levels every class learns a skill at (approved
// 2026-10-10): two normal skills and an ultimate at Lv 1, a normal skill
// every 3 levels, a second ultimate at Lv 20 (the ascension's comes at 30).
var ScheduleLevels = []int{1, 1, 1, 3, 6, 9, 12, 15, 18, 20, 21, 24, 27}

// heroSchedule is each class's skills in ScheduleLevels order. The third
// and the tenth are the ultimates.
var heroSchedule = map[string][]string{
	"Warrior": {"Slash", "Rend", "Power Strike", "Whirlwind", "Crescent Wave", "Shoulder Check", "Execution",
		"Riposte Stance", "Leaping Cleave", "Dragon Slash", "Sunder", "Battle Trance", "Momentum"},
	"Tank": {"Bash", "Shield Bash", "Crushing Blow", "Ground Pound", "Ironhide", "Hook Chain", "Bodyguard",
		"Fortress Slam", "Iron Shell", "Titan's Wrath", "Taunting Roar", "Aftershock", "Unstoppable"},
	"Mage": {"Fireball", "Frost Bolt", "Meteor", "Thunder", "Ice Lance", "Chain Lightning", "Blizzard",
		"Blink", "Mana Burn", "Arcane Nova", "Static Field", "Gravity Well", "Echo"},
	"Healer": {"Heal Ally", "Mend Self", "Divine Grace", "Radiant Burst", "Blessing", "Cleanse", "Holy Siphon",
		"Mass Heal", "Prism Aura", "Last Light", "Sacred Flame", "Resurrection", "Renewal"},
	"Ranger": {"Shot", "Snaring Shot", "Power Shot", "Multishot", "Poison Arrow", "Pinning Shot", "Piercing Shot",
		"Steady Aim", "Disengage", "Rain of Arrows", "Hunter's Mark", "Concussive Arrow", "Flaming Volley"},
	"Assassin": {"Stab", "Venom Edge", "Assassinate", "Ambush", "Fan of Knives", "Shadowstep", "Smoke Bomb",
		"Garrote", "Bait and Switch", "Death Blossom", "Crippling Cut", "Open Veins", "Twin Fangs"},
	"Gunslinger": {"Quick Shot", "Double Tap", "Hollow Point", "Ricochet", "Flash Bang", "Scatter Shot", "Boot Shot",
		"Disarming Shot", "Roll Out", "Bullet Hell", "Powder Keg", "Lucky Coin", "High Noon"},
	"Knight": {"Lance Strike", "Charge", "Grand Charge", "Lance Break", "Running Thrust", "Valor", "Trample",
		"Shield of Faith", "Unhorse", "Cavalry Breaker", "Lance Wall", "Rescue", "Relentless Charge"},
	"Dragon Knight": {"Spear Thrust", "Dragon Claw", "Impale", "Dragonfire", "Wing Gust", "Sky Dive", "Tail Sweep",
		"Dragon Scales", "Wyrm Roar", "Dragon's Fury", "Ember Trail", "Skewer", "Wyrmbond"},
	"Witch": {"Toxic Bolt", "Hex", "Witch's Brew", "Miasma", "Curse of Frailty", "Plague Swarm", "Voodoo Link",
		"Wither", "Puppet Strings", "Hex Storm", "Soul Siphon", "Contagion", "Bog of Despair"},
}

// monsterSchedule is each monster type's skills in ScheduleLevels order:
// hero skills of the same name, or the type's own old monster skills.
var monsterSchedule = map[string][]string{
	"Warrior": {"Slash", "Rend", "Power Strike", "Whirlwind", "Savage Claw", "Shoulder Check", "Crescent Wave",
		"Riposte Stance", "Leaping Cleave", "Rampage", "Sunder", "Battle Trance", "Momentum"},
	"Tank": {"Bash", "Shield Bash", "Crushing Blow", "Ground Pound", "Shell Slam", "Hook Chain", "Ironhide",
		"Fortress Slam", "Iron Shell", "Stone Quake", "Taunting Roar", "Aftershock", "Unstoppable"},
	"Mage": {"Fireball", "Frost Bolt", "Meteor", "Thunder", "Ice Lance", "Chain Lightning", "Greater Spell",
		"Blizzard", "Mana Burn", "Cataclysm", "Static Field", "Gravity Well", "Echo"},
	"Ranger": {"Shot", "Snaring Shot", "Power Shot", "Multishot", "Venom Spit", "Pinning Shot", "Piercing Shot",
		"Disengage", "Hunter's Mark", "Barrage", "Concussive Arrow", "Flaming Volley", "Steady Aim"},
	"Healer": {"Spore Bolt", "Heal Ally", "Divine Grace", "Mend Self", "Radiant Burst", "Mass Mend", "Cleanse",
		"Toxic Spores", "Holy Siphon", "Last Light", "Blessing", "Renewal", "Resurrection"},
}

// newStatuses are the timed statuses the new skills use; they count down at
// the end of their owner's side's turn like the old ones.
var newStatuses = []string{"riposte", "trance", "hmark", "silence", "slow", "wither", "link", "taunted",
	"guarded", "cover", "dscale", "lwall", "wyrm", "lastlight"}

// isUltimate is the schedule position of the two ultimates.
func isUltimate(i int) bool { return i == 2 || i == 9 }

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-10",
		What: "Every class learns on the same schedule: two normal skills and an ultimate at Lv 1, a normal skill " +
			"at Lv 3, 6, 9, 12, 15, 18, 21, 24 and 27, a second ultimate at Lv 20 and the ascension ultimate at 30. " +
			"53 new skills and the Healer's Last Light were added; Execution, Fortress Slam and Resurrection became " +
			"normal skills; Thunder moved to Lv 3. Monster types learn on the same schedule and fight with their " +
			"4 newest normal skills, newest ultimate and Strike (before: a fixed list plus 2-3 level skills; elites got " +
			"no level skills). Monster healers can bring back fallen monsters.",
		Apply: applySchedule,
	})
}

func applySchedule(c *Content) {
	for _, s := range newSkills() {
		c.Skills[s.Name] = s
	}
	for class, names := range heroSchedule {
		for i, n := range names {
			s := c.Skills[n]
			s.Class, s.Level, s.Kind = class, ScheduleLevels[i], "n"
			if isUltimate(i) {
				s.Kind = "u"
			}
			c.Skills[n] = s
		}
	}
	c.SkillOrder = scheduleOrder(c.SkillOrder)
	for _, q := range newStatuses {
		if !slices.Contains(c.Turns.CountdownAtEnd, q) {
			c.Turns.CountdownAtEnd = append(c.Turns.CountdownAtEnd, q)
		}
	}
	for class, names := range monsterSchedule {
		cl := c.Classes[class]
		cl.Schedule = nil
		for i, n := range names {
			s, ok := c.Skills[n]
			for _, ls := range cl.LevelSkills {
				if ls.Skill.Name == n {
					s, ok = ls.Skill, true
				}
			}
			if !ok {
				continue
			}
			s.Level, s.Kind = ScheduleLevels[i], "n"
			if isUltimate(i) {
				s.Kind = "u"
			}
			cl.Schedule = append(cl.Schedule, LevelSkill{Level: ScheduleLevels[i], Skill: s})
		}
		c.Classes[class] = cl
	}
	c.Combat.MonstersRevive = true
}

// scheduleOrder puts each scheduled class's skills together in learning
// order, where the class's first skill used to be; other skills keep their
// place (the order decides starting skills and the order of learning).
func scheduleOrder(old []string) []string {
	var out []string
	done := map[string]bool{}
	scheduled := map[string]bool{}
	for _, names := range heroSchedule {
		for _, n := range names {
			scheduled[n] = true
		}
	}
	for _, n := range old {
		if done[n] {
			continue
		}
		if !scheduled[n] {
			out, done[n] = append(out, n), true
			continue
		}
		for _, names := range heroSchedule {
			if slices.Contains(names, n) {
				for _, m := range names {
					if !done[m] {
						out, done[m] = append(out, m), true
					}
				}
			}
		}
	}
	return out
}
