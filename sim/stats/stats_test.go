package stats

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func hero(c *content.Content, id int, class string, level int) *state.Unit {
	n := Hero(c, class, level, "")
	return &state.Unit{ID: id, Team: state.Heroes, Class: class, Level: level, HP: n.MaxHP, MaxHP: n.MaxHP,
		Atk: n.Atk, Def: n.Def, Mov: n.Mov, Status: map[string]int{}}
}

func collect(evs *[]action.Event) action.Emit { return func(e action.Event) { *evs = append(*evs, e) } }

func TestGrowth(t *testing.T) {
	c := content.MustDefault()

	t.Run("A hero needs its level times 6 XP to reach the next level", func(t *testing.T) {
		if XPToNext(c, 1) != 6 || XPToNext(c, 12) != 72 {
			t.Fatalf("got %d and %d", XPToNext(c, 1), XPToNext(c, 12))
		}
	})

	t.Run("Experience earned is halved and rounded, but is never less than 1", func(t *testing.T) {
		b := &state.Battle{}
		u := hero(c, 1, "Warrior", 3)
		Gain(c, b, u, 7, func(action.Event) {})
		if u.XP != 4 {
			t.Fatalf("7 XP should give 4, got %d", u.XP)
		}
		Gain(c, b, u, 1, func(action.Event) {})
		if u.XP != 5 {
			t.Fatalf("1 XP should still give 1, got %d", u.XP-4)
		}
	})

	t.Run("Levelling up raises a hero's numbers, and the extra max HP is healed too", func(t *testing.T) {
		b := &state.Battle{}
		u := hero(c, 1, "Mage", 2)
		u.HP = 10
		before := u.MaxHP
		var evs []action.Event
		Gain(c, b, u, 24, collect(&evs)) // 12 XP after halving: level 2 needs 12
		if u.Level != 3 || u.XP != 0 || len(evs) == 0 || evs[0].Kind != action.LevelUp {
			t.Fatalf("level %d xp %d events %v", u.Level, u.XP, evs)
		}
		if u.HP != 10+u.MaxHP-before || u.MaxHP != Hero(c, "Mage", 3, "").MaxHP {
			t.Fatalf("hp %d of %d (was 10 of %d)", u.HP, u.MaxHP, before)
		}
	})

	t.Run("A big gain can level a hero up more than once", func(t *testing.T) {
		u := hero(c, 1, "Tank", 1)
		Gain(c, &state.Battle{}, u, 40, func(action.Event) {}) // 20 XP: 6 for level 2, 12 for level 3
		if u.Level != 3 || u.XP != 2 {
			t.Fatalf("level %d xp %d", u.Level, u.XP)
		}
	})

	t.Run("A hero evolves at level 5 and again at level 20, growing stronger each time", func(t *testing.T) {
		if FormName(c, "Warrior", 4) != "Warrior" || FormName(c, "Warrior", 5) != "Vanguard" || FormName(c, "Warrior", 20) != "Champion" {
			t.Fatal("wrong form names")
		}
		u := hero(c, 1, "Warrior", 4)
		var evs []action.Event
		Gain(c, &state.Battle{}, u, 48, collect(&evs))
		evolved := false
		for _, e := range evs {
			evolved = evolved || (e.Kind == action.Evolved && e.Skill == "Vanguard")
		}
		if !evolved || Hero(c, "Warrior", 5, "").Def != Hero(c, "Warrior", 4, "").Def+1 {
			t.Fatalf("events %v", evs)
		}
	})

	t.Run("A hero at the level cap gains no experience, and reaching the cap empties the bar", func(t *testing.T) {
		b := &state.Battle{LevelCap: 3}
		u := hero(c, 1, "Ranger", 2)
		Gain(c, b, u, 100, func(action.Event) {})
		if u.Level != 3 || u.XP != 0 {
			t.Fatalf("level %d xp %d", u.Level, u.XP)
		}
		Gain(c, b, u, 100, func(action.Event) {})
		if u.Level != 3 || u.XP != 0 {
			t.Fatalf("at the cap: level %d xp %d", u.Level, u.XP)
		}
	})

	t.Run("Kill experience is shared: the killer keeps 60% and the other heroes split 40%", func(t *testing.T) {
		killer, a, z := hero(c, 1, "Warrior", 9), hero(c, 2, "Mage", 9), hero(c, 3, "Healer", 9)
		b := &state.Battle{Units: []*state.Unit{killer, a, z}}
		Share(c, b, killer, 20, func(action.Event) {}) // killer 12 -> 6 XP; others 4 each -> 2 XP
		if killer.XP != 6 || a.XP != 2 || z.XP != 2 {
			t.Fatalf("xp %d %d %d", killer.XP, a.XP, z.XP)
		}
	})

	t.Run("Bosses give 4 times and elites 2 times the experience of a normal monster", func(t *testing.T) {
		m := &state.Unit{Team: state.Enemies, Level: 7}
		normal := KillXP(c, m)
		m.Elite = true
		elite := KillXP(c, m)
		m.Elite, m.Boss = false, true
		if normal != 10 || elite != 20 || KillXP(c, m) != 40 {
			t.Fatalf("normal %d elite %d boss %d", normal, elite, KillXP(c, m))
		}
	})

	t.Run("Monsters and summons never gain experience", func(t *testing.T) {
		m := &state.Unit{Team: state.Enemies, Level: 3}
		s := hero(c, 2, "Warrior", 3)
		s.Summoned = true
		Gain(c, &state.Battle{}, m, 50, func(action.Event) {})
		Gain(c, &state.Battle{}, s, 50, func(action.Event) {})
		if m.XP != 0 || s.XP != 0 {
			t.Fatal("they gained experience")
		}
	})

	t.Run("A hit earns experience for its damage, more if it puts statuses on; a heal earns more for bigger heals, up to a limit", func(t *testing.T) {
		if HitXP(c, 30, false) != 5 || HitXP(c, 30, true) != 7 || HealXP(c, 20) != 5 || HealXP(c, 500) != 10 {
			t.Fatalf("hit %d %d heal %d %d", HitXP(c, 30, false), HitXP(c, 30, true), HealXP(c, 20), HealXP(c, 500))
		}
	})
}
