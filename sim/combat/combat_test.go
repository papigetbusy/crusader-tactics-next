package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// arena puts a hero and a monster next to each other on open plains.
func arena(ground byte) (*content.Content, *state.Battle, *state.Unit, *state.Unit) {
	c := content.MustDefault()
	row := []byte("ggg")
	row[1] = ground
	b := &state.Battle{N: 3, Ground: []string{string(row), "ggg", "ggg"}}
	hero := &state.Unit{ID: 1, Team: state.Heroes, Class: "Warrior", X: 0, HP: 30, MaxHP: 30, Atk: 20, Def: 4, Status: map[string]int{}}
	foe := &state.Unit{ID: 2, Team: state.Enemies, Class: "Warrior", X: 1, HP: 40, MaxHP: 40, Atk: 10, Def: 5, Status: map[string]int{}}
	b.Units = []*state.Unit{hero, foe}
	return c, b, hero, foe
}

func plain(power float64) content.Skill {
	return content.Skill{Name: "Test Cut", Power: power, MinRange: 1, MaxRange: 1}
}

// rolls makes dice that return exactly these fractions.
func rolls(fs ...float64) *rng.Feed {
	var v []uint32
	for _, f := range fs {
		v = append(v, uint32(f*4294967296.0))
	}
	return rng.NewFeed(v)
}

func TestCombat(t *testing.T) {
	t.Run("Damage is attack times power minus defense, at the middle damage roll", func(t *testing.T) {
		c, b, h, f := arena('g')
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 15 {
			t.Fatalf("20 attack vs 5 defense: got %d, want 15", dm)
		}
	})

	t.Run("A hit can never do less than 1 damage", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Def = 99
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 1 {
			t.Fatalf("got %d", dm)
		}
	})

	t.Run("Magic only faces half of the target's defense", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Def = 10
		s := plain(1)
		s.Magic = true
		if dm, _ := Damage(c, b, Attack{By: h, Skill: s, Target: f}, 1); dm != 15 {
			t.Fatalf("got %d, want 20-5=15", dm)
		}
	})

	t.Run("Defending halves the damage taken", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Defending = true
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 8 {
			t.Fatalf("got %d, want round(7.5)=8", dm)
		}
	})

	t.Run("Standing in a forest cuts damage taken by 15%", func(t *testing.T) {
		c, b, h, f := arena('f')
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 13 {
			t.Fatalf("got %d, want round(12.75)=13", dm)
		}
	})

	t.Run("Fire beats ice: 50% more damage, and ice resists fire hits back", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Element = "ice"
		s := plain(1)
		s.Element = "fire"
		s.Physical = true // a weapon art, so defense counts in full
		dm, m := Damage(c, b, Attack{By: h, Skill: s, Target: f}, 1)
		if m != 1.5 || dm != 23 {
			t.Fatalf("got %d x%v", dm, m)
		}
	})

	t.Run("Sword beats lance: more damage and 15 more hit chance", func(t *testing.T) {
		c, b, _, f := arena('g')
		f.Weapon = "lance"
		foe2 := &state.Unit{ID: 3, Team: state.Enemies, Weapon: "sword", Atk: 20, HP: 10, MaxHP: 10, Status: map[string]int{}}
		target := &state.Unit{ID: 4, Team: state.Heroes, Weapon: "lance", Def: 5, X: 1, HP: 10, MaxHP: 10, Status: map[string]int{}}
		_, m := Damage(c, b, Attack{By: foe2, Skill: plain(1), Target: target}, 1)
		if m != 1.5 || HitChance(c, Attack{By: foe2, Skill: plain(1), Target: target}) != 100 {
			t.Fatalf("multiplier %v, hit %d", m, HitChance(c, Attack{By: foe2, Skill: plain(1), Target: target}))
		}
	})

	t.Run("Bosses under half HP hit 30% harder", func(t *testing.T) {
		c, b, _, f := arena('g')
		boss := &state.Unit{ID: 9, Team: state.Enemies, Boss: true, Atk: 20, HP: 10, MaxHP: 30, Status: map[string]int{}}
		target := &state.Unit{ID: 8, Team: state.Heroes, Def: 6, X: 1, HP: 40, MaxHP: 40, Status: map[string]int{}}
		_ = f
		if dm, _ := Damage(c, b, Attack{By: boss, Skill: plain(1), Target: target}, 1); dm != 20 {
			t.Fatalf("got %d, want 26-6=20", dm)
		}
	})

	t.Run("A hero's skill reaches level 2 after 8 uses and hits 5% harder", func(t *testing.T) {
		c, _, h, _ := arena('g')
		h.SkillUses = map[string]int{"Test Cut": 7}
		if SkillLevel(c, h, "Test Cut") != 1 {
			t.Fatal("7 uses should still be level 1")
		}
		h.SkillUses["Test Cut"] = 8
		if SkillLevel(c, h, "Test Cut") != 2 {
			t.Fatal("8 uses should be level 2")
		}
	})

	t.Run("Heals, buffs and revives never miss", func(t *testing.T) {
		c, _, h, f := arena('g')
		for _, s := range []content.Skill{{Heal: true}, {Support: true}, {Revive: true}} {
			if HitChance(c, Attack{By: h, Skill: s, Target: f}) != 100 {
				t.Fatalf("%+v can miss", s)
			}
		}
	})

	t.Run("Hit chance is 95% for a hero's melee attack and 90% for a monster's", func(t *testing.T) {
		c, _, h, f := arena('g')
		if HitChance(c, Attack{By: h, Skill: plain(1), Target: f}) != 95 || HitChance(c, Attack{By: f, Skill: plain(1), Target: h}) != 90 {
			t.Fatal("base hit chances wrong")
		}
	})

	t.Run("A miss uses one roll and a hit uses two", func(t *testing.T) {
		c, b, h, f := arena('g')
		miss := rolls(0.99)
		if r := Hurt(c, b, miss, Attack{By: h, Skill: plain(1), Target: f}, false); !r.Missed || miss.Used() != 1 {
			t.Fatalf("miss: %+v, rolls %d", r, miss.Used())
		}
		hit := rolls(0.0, 0.5)
		if r := Hurt(c, b, hit, Attack{By: h, Skill: plain(1), Target: f}, false); r.Missed || hit.Used() != 2 || r.Damage != 15 {
			t.Fatalf("hit: %+v, rolls %d", r, hit.Used())
		}
	})

	t.Run("Shields soak damage before HP", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Shield = 10
		r := Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: plain(1), Target: f}, false)
		if f.Shield != 0 || r.HPAfter != 35 {
			t.Fatalf("shield %d, hp %d", f.Shield, r.HPAfter)
		}
	})

	t.Run("A unit brought to 0 HP leaves the board", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP = 5
		r := Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: plain(1), Target: f}, false)
		if !r.Killed || b.Unit(f.ID) != nil {
			t.Fatal("dead unit still on the board")
		}
	})

	t.Run("A skill's status lands on a surviving target and keeps the longer duration", func(t *testing.T) {
		c, b, h, f := arena('g')
		s := plain(0.1)
		s.Effects = map[string]int{"bleed": 3, "stun": 1}
		f.Status["bleed"] = 5
		Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: s, Target: f}, false)
		if f.Status["bleed"] != 5 || f.Status["stun"] != 1 {
			t.Fatalf("statuses %v", f.Status)
		}
	})

	t.Run("Lightning always shocks the target", func(t *testing.T) {
		c, b, h, f := arena('g')
		s := plain(0.1)
		s.Element = "elec"
		Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: s, Target: f}, false)
		if f.Status["shock"] != 1 {
			t.Fatalf("statuses %v", f.Status)
		}
	})

	t.Run("The forecast uses the same math as the real hit", func(t *testing.T) {
		c, b, h, f := arena('g')
		fc := Preview(c, b, Attack{By: h, Skill: plain(1.3), Target: f})
		low := Hurt(c, b, rolls(0, 0), Attack{By: h, Skill: plain(1.3), Target: f}, false)
		f.HP = f.MaxHP
		high := Hurt(c, b, rolls(0, 0.9999999), Attack{By: h, Skill: plain(1.3), Target: f}, false)
		if low.Damage != fc.Low || high.Damage != fc.High {
			t.Fatalf("forecast %d-%d, real %d and %d", fc.Low, fc.High, low.Damage, high.Damage)
		}
	})

	t.Run("Halves round up, like the old game", func(t *testing.T) {
		if Round(2.5) != 3 || Round(-2.5) != -2 || Round(2.4999) != 2 {
			t.Fatal("rounding differs from JavaScript")
		}
	})
}
