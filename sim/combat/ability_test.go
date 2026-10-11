package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func nop(action.Event) {}

// rolls that always hit with a middle damage roll.
func hits() *rng.Feed { return rolls(0, 0.5, 0, 0.5, 0, 0.5, 0, 0.5, 0, 0.5, 0, 0.5) }

func TestAbilities(t *testing.T) {
	t.Run("A class ability grows at levels 5 and 20, and each ascension has its own", func(t *testing.T) {
		c := content.MustDefault()
		for lv, want := range map[int]string{1: "Double Strike", 5: "Double Strike+", 20: "Triple Strike"} {
			if a, _ := c.AbilityFor("Warrior", lv, "", false); a.Name != want {
				t.Fatalf("level %d: %s", lv, a.Name)
			}
		}
		if a, _ := c.AbilityFor("Warrior", 30, "Duelist", false); a.Name != "Flurry of Blades" {
			t.Fatalf("Duelist: %s", a.Name)
		}
	})

	t.Run("Using an ability starts its cooldown but does not end the hero's turn", func(t *testing.T) {
		c, b, h, _ := arena('g')
		ab, _ := c.AbilityFor("Warrior", 1, "", false)
		_ = UseAbility(c, b, hits(), h, ab, nil, nil, nop)
		if h.AbilityCooldown != 3 || h.Done || h.Status["dbl"] != 2 {
			t.Fatalf("cooldown %d done %v statuses %v", h.AbilityCooldown, h.Done, h.Status)
		}
	})

	t.Run("Double Strike makes the next skill hit twice, then wears off", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 400, 400
		ab, _ := c.AbilityFor("Warrior", 1, "", false)
		_ = UseAbility(c, b, hits(), h, ab, nil, nil, nop)
		h.Skills = []content.Skill{plain(1)}
		h.Cooldowns = []int{0}
		var n int
		_ = UseSkill(c, b, hits(), h, 0, f, func(e action.Event) {
			if e.Kind == action.Hit {
				n++
			}
		})
		if n != 2 || h.Status["dbl"] != 0 {
			t.Fatalf("%d hits, dbl %d", n, h.Status["dbl"])
		}
	})

	t.Run("A Guardian takes a hit aimed at an ally next to it, for 70% of the damage", func(t *testing.T) {
		c, b, h, f := arena('g')
		ally := &state.Unit{ID: 3, Team: state.Heroes, Class: "Mage", Level: 1, X: 1, Y: 1, HP: 30, MaxHP: 30, Def: 4, Status: map[string]int{}}
		h.Ascension, h.X, h.Y = "Guardian", 0, 1
		b.Units = append(b.Units, ally)
		plainHit := Hurt(c, &state.Battle{N: 3, Ground: b.Ground, Units: []*state.Unit{f, ally}}, hits(), Attack{By: f, Skill: plain(1), Target: ally}, nop)
		ally.HP = 30
		r := Hurt(c, b, hits(), Attack{By: f, Skill: plain(1), Target: ally}, nop)
		if r.Target != h || ally.HP != 30 || r.Damage >= plainHit.Damage {
			t.Fatalf("hit %d (plain %d) on %d", r.Damage, plainHit.Damage, r.Target.ID)
		}
	})

	t.Run("Shadow Clones are two decoys with 30% of the Phantom's health that monsters go for first", func(t *testing.T) {
		c, b, h, _ := arena('g')
		h.Ascension, h.MaxHP = "Phantom", 40
		ab, _ := c.AbilityFor("Assassin", 30, "Phantom", false)
		_ = UseAbility(c, b, hits(), h, ab, nil, nil, nop)
		clones := 0
		for _, z := range b.Units {
			if z.Decoy && z.MaxHP == 12 && z.Status["clone"] == 3 {
				clones++
			}
		}
		if clones != 2 {
			t.Fatalf("%d clones", clones)
		}
	})

	t.Run("Hidden traps root, poison and mark the monster that stops on them", func(t *testing.T) {
		c, b, _, f := arena('g')
		b.Traps = []state.Trap{{X: f.X, Y: f.Y}}
		Trap(c, b, f, nil, nop)
		if len(b.Traps) != 0 || f.Status["snare"] != 3 || f.Status["poison"] != 3 || f.Status["tmark"] != 3 {
			t.Fatalf("statuses %v", f.Status)
		}
	})

	t.Run("A Duelist strikes back when hit in melee", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Ascension = "Duelist"
		before := f.HP
		Hurt(c, b, hits(), Attack{By: f, Skill: plain(0.2), Target: h}, nop)
		if f.HP >= before {
			t.Fatal("no riposte")
		}
	})

	t.Run("A Juggernaut cannot be stunned, snared or broken", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Ascension = "Juggernaut"
		s := plain(0.1)
		s.Effects = map[string]int{"stun": 1, "snare": 2, "brk": 2, "bleed": 2}
		Hurt(c, b, hits(), Attack{By: f, Skill: s, Target: h}, nop)
		if h.Status["stun"]+h.Status["snare"]+h.Status["brk"] != 0 || h.Status["bleed"] != 2 {
			t.Fatalf("statuses %v", h.Status)
		}
	})
}
