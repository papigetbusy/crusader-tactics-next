package combat

import "testing"

func TestPurpleBonuses(t *testing.T) {
	t.Run("Executioner's gear hits 25% harder against a target below 30% of its max HP", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Def = 0
		h.Affix = map[string]int{"execute": 25}
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 20 {
			t.Fatalf("against a healthy target: %d", dm)
		}
		f.HP = 11 // under 30% of 40
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 25 {
			t.Fatalf("against a wounded target: %d", dm)
		}
	})

	t.Run("Warding gear shrugs off each harmful status 25% of the time", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Affix = map[string]int{"ward": 25}
		s := plain(1)
		s.Effects = map[string]int{"bleed": 3, "snare": 2}
		// statuses are checked in name order: bleed (roll 0.1, shrugged off), then snare (0.5, it lands)
		put := applyStatuses(c, b, rolls(0.1, 0.5), Attack{By: h, Skill: s, Target: f}, f)
		if put["bleed"] != 0 || put["snare"] != 2 || f.Status["bleed"] != 0 {
			t.Fatalf("put %v, statuses %v", put, f.Status)
		}
	})
}
