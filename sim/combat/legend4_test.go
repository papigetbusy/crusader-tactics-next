package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestClassLegendarySetsOfFour(t *testing.T) {
	t.Run("Warrior set of 4: a kill resets the class ability's cooldown", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, h.AbilityCooldown = 1, 3
		power(h, "killReset", 1)
		Hurt(c, b, hits(), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if h.AbilityCooldown != 0 {
			t.Fatalf("cooldown %d", h.AbilityCooldown)
		}
	})

	t.Run("Tank set of 4: counterattacks deal 50% more damage", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Atk, f.Def = 105, 5
		power(h, "counterBoost", 50)
		if a, back := dm(c, b, Attack{By: h, Skill: plain(1), Target: f}), dm(c, b, Attack{By: h, Skill: plain(1), Target: f, Counter: true}); a != 100 || back <= 150 {
			t.Fatalf("attack %d, counter %d", a, back)
		}
	})

	t.Run("Knight set of 4: the first hit taken each battle does nothing", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.HP, h.MaxHP = 99, 99
		power(h, "firstWard", 1)
		if r := Hurt(c, b, hits(), Attack{By: f, Skill: plain(1), Target: h}, nop); !r.Immune || h.HP != 99 {
			t.Fatalf("first hit: %+v, HP %d", r, h.HP)
		}
		if Hurt(c, b, hits(), Attack{By: f, Skill: plain(1), Target: h}, nop); h.HP == 99 {
			t.Fatal("the second hit did nothing too")
		}
	})

	t.Run("Dragon Knight set of 4: breath and dive skills hit 1 tile wider", func(t *testing.T) {
		c, _ := open7()
		u := &state.Unit{}
		power(u, "breathWide", 1)
		if Widen(u, c.Skills["Dragonfire"]).Area != 2 || Widen(u, c.Skills["Sky Dive"]).Area != 1 || Widen(u, c.Skills["Spear Thrust"]).Area != 0 {
			t.Fatal("wrong areas")
		}
	})

	t.Run("Ranger set of 4: 20% more damage to targets 4 or more tiles away", func(t *testing.T) {
		c, b := open7()
		r := add(b, 1, state.Heroes, "Ranger", 3, 0)
		f := add(b, 2, state.Enemies, "Orc Brute", 3, 3)
		r.Atk, f.Def = 104, 4
		power(r, "farShot", 20)
		close := dm(c, b, Attack{By: r, Skill: plain(1), Target: f})
		f.Y = 4
		if far := dm(c, b, Attack{By: r, Skill: plain(1), Target: f}); close != 100 || far <= close {
			t.Fatalf("3 tiles %d, 4 tiles %d", close, far)
		}
	})

	t.Run("Gunslinger set of 4: a crit takes 1 turn off every cooldown", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 999, 999
		power(h, "critRefund", 1, "crit", 100)
		h.Cooldowns, h.AbilityCooldown = []int{3, 0, 99}, 2
		if r := Hurt(c, b, rolls(0, 0.5, 0), Attack{By: h, Skill: plain(1), Target: f}, nop); !r.Crit {
			t.Fatal("no crit")
		}
		if h.Cooldowns[0] != 2 || h.Cooldowns[1] != 0 || h.Cooldowns[2] != 99 || h.AbilityCooldown != 1 {
			t.Fatalf("cooldowns %v, ability %d", h.Cooldowns, h.AbilityCooldown)
		}
	})

	t.Run("Assassin set of 4: the first attack each battle always crits", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 999, 999
		power(h, "opener", 1)
		if r := Hurt(c, b, rolls(0, 0.5, 0.99), Attack{By: h, Skill: plain(1), Target: f}, nop); !r.Crit {
			t.Fatal("the first attack did not crit")
		}
		if r := Hurt(c, b, rolls(0, 0.5, 0.99), Attack{By: h, Skill: plain(1), Target: f}, nop); r.Crit {
			t.Fatal("the second attack crit too")
		}
	})

	t.Run("Mage set of 4: area spells are 1 tile wider, other skills are not", func(t *testing.T) {
		c, _ := open7()
		u := &state.Unit{}
		power(u, "wideSpells", 1)
		if Widen(u, c.Skills["Meteor"]).Area != 2 || Widen(u, c.Skills["Fireball"]).Area != 0 {
			t.Fatal("wrong areas")
		}
	})

	t.Run("Healer set of 4: a heal also removes one bad status", func(t *testing.T) {
		c, b := open7()
		hl := add(b, 1, state.Heroes, "Healer", 3, 3)
		w := add(b, 2, state.Heroes, "Warrior", 3, 4)
		w.HP, w.Status = 50, map[string]int{"bleed": 2, "poison": 2}
		power(hl, "cleanseHeal", 1)
		healSkill(c, b, hits(), hl, c.Skills["Heal Ally"], w, nop)
		if len(w.Status) != 1 || w.Status["bleed"] != 0 {
			t.Fatalf("statuses %v", w.Status)
		}
	})

	t.Run("Witch set of 4: curses and poisons also spread to one enemy next to the target", func(t *testing.T) {
		c, b := open7()
		w := add(b, 1, state.Heroes, "Witch", 3, 0)
		f := add(b, 2, state.Enemies, "Orc Brute", 3, 3)
		g := add(b, 3, state.Enemies, "Orc Brute", 3, 4)
		x := add(b, 4, state.Enemies, "Orc Brute", 2, 3)
		power(w, "hexSpread", 1)
		s := plain(1)
		s.Effects = map[string]int{"poison": 3}
		var evs []action.Event
		curse(c, b, hits(), w, s, f, record(&evs))
		if f.Status["poison"] != 3 || g.Status["poison"]+x.Status["poison"] != 3 {
			t.Fatalf("target %v, neighbours %v %v", f.Status, g.Status, x.Status)
		}
	})
}

func TestStatusImmunityEverywhere(t *testing.T) {
	t.Run("Contagion can't burn a Dragon Knight with Wyrm's Hide, but still burns the others", func(t *testing.T) {
		_, b := open7()
		w := add(b, 1, state.Heroes, "Witch", 0, 0)
		v := add(b, 2, state.Enemies, "Orc Brute", 3, 3)
		dk := add(b, 3, state.Enemies, "Dragon Knight", 3, 4)
		other := add(b, 4, state.Enemies, "Orc Brute", 2, 3)
		power(dk, "wyrmhide", 1)
		v.Status["burn"] = 3
		contagion(b, w, v, nop)
		if dk.Status["burn"] != 0 || other.Status["burn"] != 3 {
			t.Fatalf("Dragon Knight %v, other %v", dk.Status, other.Status)
		}
	})
}
