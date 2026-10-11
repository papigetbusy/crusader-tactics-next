package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestMageHealerRangerAssassinSkills(t *testing.T) {
	t.Run("Blink teleports the Mage without using up its turn, but it cannot walk afterwards", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Mage", 3, 3)
		cast(t, c, b, h, "Blink", Pick{Tiles: tile(3, 0)})
		if !at(h, 3, 0) || h.Done || !h.Moved {
			t.Fatalf("at %d,%d done %v moved %v", h.X, h.Y, h.Done, h.Moved)
		}
	})

	t.Run("Mana Burn adds a turn to every cooldown the target has", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Mage", 3, 3), add(b, 2, state.Enemies, "Tank", 3, 0)
		f.Skills, f.Cooldowns = []content.Skill{c.Skills["Bash"], c.Skills["Shield Bash"]}, []int{0, 2}
		cast(t, c, b, h, "Mana Burn", Pick{Target: f})
		if f.Cooldowns[0] != 0 || f.Cooldowns[1] != 3 {
			t.Fatalf("cooldowns %v", f.Cooldowns)
		}
	})

	t.Run("Static Field charges the ground: an enemy ending its turn there is shocked", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Mage", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Static Field", Pick{Target: f})
		f.Status["shock"] = 0
		Charged(b, state.Enemies, nop)
		if len(b.Fields) != 1 || b.Fields[0].Kind != "static" || f.Status["shock"] != 1 {
			t.Fatalf("fields %+v, %v", b.Fields, f.Status)
		}
	})

	t.Run("Gravity Well pulls enemies within 2 tiles one step toward the center", func(t *testing.T) {
		c, b := open7()
		h, f, g := add(b, 1, state.Heroes, "Mage", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0), add(b, 3, state.Enemies, "Warrior", 5, 0)
		cast(t, c, b, h, "Gravity Well", Pick{Target: f})
		if !at(g, 4, 0) || !at(f, 3, 0) {
			t.Fatalf("pulled to %d,%d", g.X, g.Y)
		}
	})

	t.Run("Echo casts the last spell again at 60% power without starting its cooldown", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Mage", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		f.HP, f.MaxHP = 500, 500
		h.Skills, h.Cooldowns = []content.Skill{c.Skills["Meteor"], c.Skills["Echo"]}, []int{0, 0}
		var evs []action.Event
		rec := func(e action.Event) { evs = append(evs, e) }
		_ = UseAimed(c, b, half{}, h, 0, Pick{Target: f}, rec)
		first := dealt(evs, h, f)
		h.Done, h.Cooldowns[0], evs = false, 0, nil
		_ = UseAimed(c, b, half{}, h, 1, Pick{Target: f}, rec)
		weak := c.Skills["Meteor"]
		weak.Power *= 0.6
		want, _ := Damage(c, b, Attack{By: h, Skill: weak, Target: f}, 1)
		if h.LastSpell != "Meteor" || h.Cooldowns[0] != 0 || dealt(evs, h, f) != want || want >= first {
			t.Fatalf("echo %d (want %d, first %d), cooldown %d", dealt(evs, h, f), want, first, h.Cooldowns[0])
		}
	})

	t.Run("Last Light heals 30% and leaves an ally at 1 HP instead of falling", func(t *testing.T) {
		c, b := open7()
		h, ally := add(b, 1, state.Heroes, "Healer", 3, 3), add(b, 2, state.Heroes, "Warrior", 3, 5)
		ally.HP = 50
		cast(t, c, b, h, "Last Light", Pick{Target: h})
		if ally.HP != 80 {
			t.Fatalf("healed to %d", ally.HP)
		}
		f := add(b, 3, state.Enemies, "Warrior", 3, 6)
		Hurt(c, b, half{}, Attack{By: f, Skill: plain(1), Target: ally, Fixed: 500}, nop)
		if ally.HP != 1 || b.Unit(2) == nil {
			t.Fatalf("HP %d", ally.HP)
		}
	})

	t.Run("A monster healer's Resurrection brings back the monster that fell last", func(t *testing.T) {
		c, b := open7()
		h, slime := add(b, 1, state.Heroes, "Warrior", 0, 0), add(b, 2, state.Enemies, "Healer", 3, 3)
		gob := add(b, 3, state.Enemies, "Warrior", 1, 0)
		Hurt(c, b, half{}, Attack{By: h, Skill: plain(1), Target: gob, Fixed: 500}, nop)
		if len(b.FallenFoes) != 1 {
			t.Fatal("a fallen monster should be kept")
		}
		cast(t, c, b, slime, "Resurrection", Pick{Target: slime})
		if b.Unit(3) == nil || Distance(gob, slime) > 2 || gob.HP != 40 {
			t.Fatalf("not back: %+v", gob)
		}
	})

	t.Run("Steady Aim gives the next shot 2 more range and a sure hit, without using up the turn", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Ranger", 0, 0), add(b, 2, state.Enemies, "Warrior", 6, 2)
		h.Skills, h.Cooldowns = []content.Skill{c.Skills["Steady Aim"], c.Skills["Shot"]}, []int{0, 0}
		if len(Targets(c, b, h, h.Skills[1])) != 0 {
			t.Fatal("8 tiles should be out of range before aiming")
		}
		_ = UseAimed(c, b, half{}, h, 0, Pick{Target: h}, nop)
		if h.Done || !h.Moved || len(Targets(c, b, h, h.Skills[1])) != 1 {
			t.Fatalf("done %v moved %v", h.Done, h.Moved)
		}
		_ = UseAimed(c, b, rolls(0.99, 0.5), h, 1, Pick{Target: f}, nop)
		if f.HP == 100 || h.Status["aim"] != 0 {
			t.Fatalf("a 99 roll should still hit; aim %d", h.Status["aim"])
		}
	})

	t.Run("Disengage shoots, then leaps 2 tiles away from the target", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Ranger", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Disengage", Pick{Target: f})
		if !at(h, 3, 5) || f.HP == 100 {
			t.Fatalf("ranger at %d,%d", h.X, h.Y)
		}
	})

	t.Run("Hunter's Mark makes its target take 20% more damage", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Ranger", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Hunter's Mark", Pick{Target: f})
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); f.Status["hmark"] != 3 || dm != 20 {
			t.Fatalf("damage %d (want 20), mark %d", dm, f.Status["hmark"])
		}
	})

	t.Run("Concussive Arrow knocks the target 2 tiles back", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Ranger", 3, 6), add(b, 2, state.Enemies, "Warrior", 3, 3)
		cast(t, c, b, h, "Concussive Arrow", Pick{Target: f})
		if !at(f, 3, 1) {
			t.Fatalf("foe at %d,%d", f.X, f.Y)
		}
	})

	t.Run("Flaming Volley leaves the ground around the target burning for 2 turns", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Ranger", 3, 6), add(b, 2, state.Enemies, "Warrior", 3, 3)
		cast(t, c, b, h, "Flaming Volley", Pick{Target: f})
		FieldTick(b, state.Heroes)
		if len(b.Fields) != 1 || b.Fields[0].Kind != "fire" || f.Status["burn"] != 2 {
			t.Fatalf("fields %+v, %v", b.Fields, f.Status)
		}
	})

	t.Run("Shadowstep teleports behind the target before stabbing it", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Assassin", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 1)
		cast(t, c, b, h, "Shadowstep", Pick{Target: f})
		if !at(h, 3, 0) || f.HP == 100 {
			t.Fatalf("assassin at %d,%d", h.X, h.Y)
		}
	})

	t.Run("Garrote silences: a silenced hero can only use Strike", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Assassin", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Garrote", Pick{Target: f})
		h.Status["silence"] = 2
		if f.Status["silence"] != 2 || len(Targets(c, b, h, c.Skills["Stab"])) != 0 || len(Targets(c, b, h, c.Skills["Strike"])) != 1 {
			t.Fatal("silence did not hold")
		}
	})

	t.Run("Bait and Switch swaps places with any unit within 3 tiles", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Assassin", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 1)
		cast(t, c, b, h, "Bait and Switch", Pick{Target: f})
		if !at(h, 3, 1) || !at(f, 3, 3) {
			t.Fatal("not swapped")
		}
	})

	t.Run("Crippling Cut bleeds the target and takes 2 Move away for 2 turns", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Assassin", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Crippling Cut", Pick{Target: f})
		if f.Status["slow"] != 2 || f.Status["bleed"] != 2 || board.Budget(b, f) != 4 {
			t.Fatalf("%v budget %d", f.Status, board.Budget(b, f))
		}
	})

	t.Run("Open Veins makes bleed and poison hurt right away and doubles their turns", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Assassin", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		f.Status["bleed"], f.Status["poison"] = 2, 1
		evs := cast(t, c, b, h, "Open Veins", Pick{Target: f})
		if f.Status["bleed"] != 4 || f.Status["poison"] != 2 || f.HP != 100-dealt(evs, h, f)-6-7 {
			t.Fatalf("%v HP %d", f.Status, f.HP)
		}
	})

	t.Run("Twin Fangs strikes two different enemies next to the Assassin", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Assassin", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 2), add(b, 3, state.Enemies, "Warrior", 2, 3)
		evs := cast(t, c, b, h, "Twin Fangs", Pick{Target: f})
		if dealt(evs, h, f) == 0 || dealt(evs, h, g) == 0 {
			t.Fatal("both should be hit")
		}
	})
}
