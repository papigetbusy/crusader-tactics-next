package content

import (
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestData(t *testing.T) {
	t.Run("Every data file loads and passes its checks", func(t *testing.T) {
		if _, err := Default(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("The old game's saved tables load", func(t *testing.T) {
		if _, err := FromOld(os.DirFS("../../testdata/oldgame/tables")); err != nil {
			t.Fatal(err)
		}
	})

	// This is the seed of the "Changes from the old game" page: when a number is
	// changed on purpose, this test is updated to list the change.
	t.Run("The live data files equal the old game's tables plus the approved changes, and nothing else", func(t *testing.T) {
		live := MustDefault()
		old, err := FromOld(os.DirFS("../../testdata/oldgame/tables"))
		if err != nil {
			t.Fatal(err)
		}
		ApplyChanges(old)
		if !reflect.DeepEqual(live, old) {
			t.Fatalf("live data differs from the old game plus approved changes:\nlive: %+v\nwant: %+v", live, old)
		}
	})

	t.Run("A data file with a broken number is rejected with a plain message", func(t *testing.T) {
		c := MustDefault()
		bad := c.Terrain['f']
		bad.MoveCost = 0
		c.Terrain['f'] = bad
		if err := c.Check(); err == nil {
			t.Fatal("expected the check to complain")
		}
	})
}

func TestApprovedChanges(t *testing.T) {
	t.Run("Monsters use the same version of a skill as heroes", func(t *testing.T) {
		c := MustDefault()
		for _, cl := range c.Classes {
			for _, s := range cl.Skills {
				if h, ok := c.Skills[s.Name]; ok && (h.MinRange != s.MinRange || h.MaxRange != s.MaxRange || h.Power != s.Power) {
					t.Errorf("%s's %s differs from the hero version", cl.Name, s.Name)
				}
			}
		}
	})
}

func TestSkillSchedule(t *testing.T) {
	c := MustDefault()
	classes := []string{"Warrior", "Tank", "Mage", "Healer", "Ranger", "Assassin", "Gunslinger", "Knight", "Dragon Knight", "Witch"}

	t.Run("Every class learns 13 skills at the same levels: 2 normal and an ultimate at Lv 1, a normal every 3 levels, an ultimate at Lv 20", func(t *testing.T) {
		for _, class := range classes {
			var levels []int
			var ults []int
			for _, n := range c.SkillOrder {
				if s := c.Skills[n]; s.Class == class && s.Ascension == "" {
					levels = append(levels, s.Level)
					if s.Kind == "u" {
						ults = append(ults, s.Level)
					}
				}
			}
			if !slices.Equal(levels, ScheduleLevels) || !slices.Equal(ults, []int{1, 20}) {
				t.Errorf("%s learns at %v (ultimates %v)", class, levels, ults)
			}
		}
	})

	t.Run("Resurrection is a normal skill at Lv 24, and the Healer's Lv 20 ultimate is Last Light", func(t *testing.T) {
		r, l := c.Skills["Resurrection"], c.Skills["Last Light"]
		if r.Kind != "n" || r.Level != 24 || l.Kind != "u" || l.Level != 20 || l.Class != "Healer" {
			t.Fatalf("Resurrection %s %d, Last Light %s %d", r.Kind, r.Level, l.Kind, l.Level)
		}
	})

	t.Run("A monster fights with its 4 newest normal skills, its newest ultimate and Strike", func(t *testing.T) {
		names := func(ss []Skill) []string {
			var out []string
			for _, s := range ss {
				out = append(out, s.Name)
			}
			return out
		}
		if got := names(c.MonsterSkills("Warrior", 1, false)); !slices.Equal(got, []string{"Slash", "Rend", "Power Strike", "Strike"}) {
			t.Errorf("Lv 1: %v", got)
		}
		want := []string{"Riposte Stance", "Leaping Cleave", "Sunder", "Battle Trance", "Rampage", "Strike"}
		if got := names(c.MonsterSkills("Warrior", 25, true)); !slices.Equal(got, want) {
			t.Errorf("Lv 25 elite: %v", got)
		}
	})

	t.Run("Every monster type learns on the same schedule", func(t *testing.T) {
		for _, class := range []string{"Warrior", "Tank", "Mage", "Ranger", "Healer"} {
			var levels []int
			for _, ls := range c.Classes[class].Schedule {
				levels = append(levels, ls.Level)
			}
			if !slices.Equal(levels, ScheduleLevels) {
				t.Errorf("%s monsters learn at %v", class, levels)
			}
		}
	})
}
