package stats

import (
	"slices"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// built is a hero with the build a new hero of its level starts with.
func built(c *content.Content, class string, level int) (*state.Battle, *state.Unit) {
	u := hero(c, 1, class, level)
	u.Learned, u.Equipped, u.Ultimate = c.StartingBuild(class, level)
	u.Skills = c.Bar(class, u.Equipped, u.Ultimate)
	u.Cooldowns = make([]int, len(u.Skills))
	return &state.Battle{Units: []*state.Unit{u}}, u
}

func none(action.Event) {}

func TestChoices(t *testing.T) {
	c := content.MustDefault()

	t.Run("A new hero's bar is its first four skills, its newest ultimate, Strike, and Cloak for Assassins", func(t *testing.T) {
		_, u := built(c, "Assassin", 20)
		var names []string
		for _, s := range u.Skills {
			names = append(names, s.Name)
		}
		want := []string{"Stab", "Venom Edge", "Ambush", "Fan of Knives", "Death Blossom", "Strike", "Cloak"}
		if !slices.Equal(names, want) {
			t.Fatalf("bar %v", names)
		}
	})

	t.Run("Levelling up to a skill's level offers it to the hero", func(t *testing.T) {
		b, u := built(c, "Warrior", 2)
		Gain(c, b, u, 24, none) // 12 XP: level 2 needs 12
		if u.Level != 3 || len(b.Decisions) != 1 || b.Decisions[0].Skill != "Whirlwind" {
			t.Fatalf("level %d decisions %v", u.Level, b.Decisions)
		}
	})

	t.Run("A new skill goes into an empty slot by itself", func(t *testing.T) {
		b, u := built(c, "Warrior", 3) // Slash, Rend: two slots free
		b.Decisions = []state.Decision{{Kind: state.DecideLearn, Unit: u.ID, Skill: "Whirlwind"}}
		if o := Options(c, b); !slices.Equal(o, []string{ChooseOK}) {
			t.Fatalf("options %v", o)
		}
		if err := Decide(c, b, ChooseOK, 0, none); err != nil || u.Equipped[len(u.Equipped)-1] != "Whirlwind" || u.Skills[2].Name != "Whirlwind" {
			t.Fatalf("err %v equipped %v", err, u.Equipped)
		}
	})

	t.Run("With four skills equipped, a new skill replaces one of them or waits in the library", func(t *testing.T) {
		b, u := built(c, "Healer", 6)
		b.Decisions = []state.Decision{{Kind: state.DecideLearn, Unit: u.ID, Skill: "Holy Siphon"}, {Kind: state.DecideLearn, Unit: u.ID, Skill: "Mass Heal"}}
		if o := Options(c, b); !slices.Equal(o, []string{ChooseReplace, ChooseKeep}) {
			t.Fatalf("options %v", o)
		}
		_ = Decide(c, b, ChooseReplace, 1, none)
		_ = Decide(c, b, ChooseKeep, 0, none)
		if u.Equipped[1] != "Holy Siphon" || slices.Contains(u.Equipped, "Mass Heal") || !slices.Contains(u.Learned, "Mass Heal") {
			t.Fatalf("equipped %v learned %v", u.Equipped, u.Learned)
		}
	})

	t.Run("A cooldown follows its skill when the bar changes", func(t *testing.T) {
		b, u := built(c, "Healer", 6)
		u.Cooldowns[3] = 2 // the fourth skill
		fourth := u.Skills[3].Name
		b.Decisions = []state.Decision{{Kind: state.DecideLearn, Unit: u.ID, Skill: "Holy Siphon"}}
		_ = Decide(c, b, ChooseReplace, 0, none)
		if u.Skills[3].Name != fourth || u.Cooldowns[3] != 2 || u.Cooldowns[0] != 0 {
			t.Fatalf("cooldowns %v", u.Cooldowns)
		}
	})

	t.Run("At level 30 a hero chooses one of three ascensions, which teaches its ultimate", func(t *testing.T) {
		b, u := built(c, "Warrior", 29)
		u.XP = XPToNext(c, 29) - 1
		Gain(c, b, u, 2, none)
		if len(b.Decisions) == 0 || b.Decisions[len(b.Decisions)-1].Kind != state.DecideAscend {
			t.Fatalf("decisions %v", b.Decisions)
		}
		b.Decisions = b.Decisions[len(b.Decisions)-1:]
		if o := Options(c, b); len(o) != 3 {
			t.Fatalf("options %v", o)
		}
		if err := Decide(c, b, "Berserker", 0, none); err != nil {
			t.Fatal(err)
		}
		if b.Decisions[0].Skill != "Blood Rage" {
			t.Fatalf("decisions %v", b.Decisions)
		}
		_ = Decide(c, b, ChooseReplace, 0, none)
		if u.Ultimate != "Blood Rage" {
			t.Fatalf("ultimate %s", u.Ultimate)
		}
	})

	t.Run("An ascension changes the hero's numbers, element and range", func(t *testing.T) {
		before := Hero(c, "Ranger", 32, "")
		after := Hero(c, "Ranger", 32, "Sniper")
		if after.MaxHP <= before.MaxHP {
			t.Fatalf("hp %d vs %d", after.MaxHP, before.MaxHP)
		}
		b, u := built(c, "Ranger", 32)
		u.Ascension = "Sniper"
		Refresh(c, b, u)
		if u.Element != "ice" || u.RangeBonus != 2 {
			t.Fatalf("element %q range %d", u.Element, u.RangeBonus)
		}
	})
}
