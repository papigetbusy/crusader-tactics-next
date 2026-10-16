package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
)

// prepared is duel() with a Lv 27 Warrior who has learned every skill up to
// Lv 27 and carries the first four.
func prepared() *Game {
	g := duel()
	h := g.B.Unit(1)
	h.Level = 27
	h.Learned, h.Equipped, h.Ultimate = g.C.StartingBuild("Warrior", 27)
	h.Skills = g.C.Bar("Warrior", h.Equipped, h.Ultimate)
	h.Cooldowns = make([]int, len(h.Skills))
	return g
}

// spare is the hero's first learned normal skill and ultimate not carried.
func spare(g *Game) (normal, ult string) {
	for _, s := range g.library(g.B.Unit(1)) {
		if s.Ultimate && ult == "" {
			ult = s.Name
		} else if !s.Ultimate && normal == "" {
			normal = s.Name
		}
	}
	return
}

func TestBringSkills(t *testing.T) {
	t.Run("A hero can change any number of its skills, and its ultimate, in one go", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		n, u := spare(g)
		bar := []string{h.Equipped[0], n, h.Equipped[2]}
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Bar: bar, Choice: u}); err != nil {
			t.Fatal(err)
		}
		if len(h.Equipped) != 3 || h.Equipped[1] != n || h.Ultimate != u || h.Skills[1].Name != n || h.Done {
			t.Fatalf("bar %v, ultimate %s, done %v", h.Equipped, h.Ultimate, h.Done)
		}
	})

	t.Run("Skills can be changed only once per floor", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		n, _ := spare(g)
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Bar: []string{n}}); err != nil {
			t.Fatal(err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Bar: []string{h.Learned[0]}}); code(err) != "swapped" {
			t.Fatalf("second change: %v", err)
		}
	})

	t.Run("A skill kept on the bar keeps its cooldown; a skill brought in starts ready", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		h.Cooldowns[0], h.Cooldowns[1] = 2, 3
		n, _ := spare(g)
		if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Bar: []string{h.Equipped[0], n}}); err != nil {
			t.Fatal(err)
		}
		if h.Cooldowns[0] != 2 || h.Cooldowns[1] != 0 {
			t.Fatalf("cooldowns %v", h.Cooldowns)
		}
	})

	t.Run("Only learned skills can be brought: 1 to 4 normal skills, and an ultimate only in the ultimate slot", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		_, u := spare(g)
		for want, bar := range map[string][]string{
			"not-learned": {"Fireball"},
			"ultimate":    {u},
			"bar-size":    {h.Learned[0], h.Learned[1], h.Learned[3], h.Learned[4], h.Learned[5]},
			"twice":       {h.Equipped[0], h.Equipped[0]},
		} {
			if _, err := g.Apply(action.Action{Kind: action.Bring, Unit: 1, Bar: bar}); code(err) != want {
				t.Errorf("%v: %v, want %s", bar, err, want)
			}
		}
	})
}

func TestSkillDetails(t *testing.T) {
	t.Run("Every skill a hero knows tells the screen its range, power, cooldown and level", func(t *testing.T) {
		g := prepared()
		h := g.B.Unit(1)
		lib := g.library(h)
		if len(lib) == 0 {
			t.Fatal("no learned skills off the bar")
		}
		for _, l := range lib {
			s, _ := g.C.SkillNamed(l.Name)
			if l.MaxRange != s.MaxRange || l.Power != s.Power || l.Cooldown != s.Cooldown || l.Level != s.Level {
				t.Fatalf("%s: view %+v, skill range %d power %v cooldown %d level %d", l.Name, l, s.MaxRange, s.Power, s.Cooldown, s.Level)
			}
		}
		for _, u := range g.View().Units {
			if u.ID != 1 {
				continue
			}
			if ab, _ := g.C.AbilityFor(h.Class, h.Level, h.Ascension, false); u.Ability == nil || u.Ability.Cooldown != ab.Cooldown {
				t.Fatalf("ability view %+v, cooldown %d", u.Ability, ab.Cooldown)
			}
			for i, sv := range u.Skills {
				if sv.Cooldown != h.Skills[i].Cooldown || sv.Level != h.Skills[i].Level {
					t.Fatalf("%s: view cooldown %d level %d", sv.Name, sv.Cooldown, sv.Level)
				}
			}
		}
	})
}
