package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// power gives a unit these bonuses, as if from its class legendary.
func power(u *state.Unit, kv ...any) {
	u.Affix = map[string]int{}
	for i := 0; i < len(kv); i += 2 {
		u.Affix[kv[i].(string)] = kv[i+1].(int)
	}
}

func dm(c *content.Content, b *state.Battle, a Attack) int {
	n, _ := Damage(c, b, a, 1)
	return n
}

func record(evs *[]action.Event) action.Emit {
	return func(e action.Event) { *evs = append(*evs, e) }
}

func TestClassLegendaryPowers(t *testing.T) {
	t.Run("Battle Rhythm: each hit landed this turn adds 8% damage, up to 3 hits, and resets next turn", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Atk, f.Def = 105, 5
		power(h, "rhythm", 8)
		base := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		h.Rhythm = 1
		one := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		h.Rhythm = 3
		three := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		h.Rhythm = 9
		capped := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		if base != 100 || one != 108 || three != 125 || capped != three {
			t.Fatalf("0 hits %d, 1 hit %d, 3 hits %d, 9 hits %d", base, one, three, capped)
		}
		h.Rhythm = 0
		f.HP, f.MaxHP = 999, 999
		Hurt(c, b, hits(), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if h.Rhythm != 1 {
			t.Fatalf("after a hit: %d", h.Rhythm)
		}
		StartTurnPowers(h)
		if h.Rhythm != 0 {
			t.Fatalf("next turn: %d", h.Rhythm)
		}
	})

	t.Run("Unbowed: below 30% HP the hero takes 25% less damage", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Atk, h.Def = 44, 4
		power(h, "unbowed", 25)
		if n := dm(c, b, Attack{By: f, Skill: plain(1), Target: h}); n != 40 {
			t.Fatalf("at full HP: %d", n)
		}
		h.HP = 8 // under 30% of 30
		if n := dm(c, b, Attack{By: f, Skill: plain(1), Target: h}); n != 30 {
			t.Fatalf("below 30%%: %d", n)
		}
	})

	t.Run("Bulwark Aura: allies next to the Tank take 15% less damage, the Tank itself does not", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Warrior", 3, 3)
		tank := add(b, 2, state.Heroes, "Tank", 2, 3)
		f := add(b, 3, state.Enemies, "Orc Brute", 3, 4)
		f.Atk = 104
		power(tank, "aura", 15)
		if n := dm(c, b, Attack{By: f, Skill: plain(1), Target: h}); n != 85 {
			t.Fatalf("the ally took %d", n)
		}
		if n := dm(c, b, Attack{By: f, Skill: plain(1), Target: tank}); n != 100 {
			t.Fatalf("the Tank took %d", n)
		}
	})

	t.Run("Retaliate: a hit can stun the attacker through its next turn", func(t *testing.T) {
		c, b, h, f := arena('g')
		b.Side = state.Enemies // the monster attacks on its own turn
		power(h, "retaliate", 30)
		h.HP, h.MaxHP = 99, 99
		Hurt(c, b, rolls(0, 0.5, 0.9), Attack{By: f, Skill: plain(1), Target: h}, nop)
		if f.Status["stun"] != 0 {
			t.Fatalf("a high roll stunned: %v", f.Status)
		}
		Hurt(c, b, rolls(0, 0.5, 0.1), Attack{By: f, Skill: plain(1), Target: h}, nop)
		if f.Status["stun"] != 2 {
			t.Fatalf("stun %d (it must outlast the end of the monster's own turn)", f.Status["stun"])
		}
	})

	t.Run("Charge: each tile moved this turn adds 5% damage, up to 30%", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Atk, f.Def = 105, 5
		power(h, "charge", 5)
		h.TilesMoved = 2
		two := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		h.TilesMoved = 6
		six := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		h.TilesMoved = 9
		if two <= 100 || six <= two || dm(c, b, Attack{By: h, Skill: plain(1), Target: f}) != six {
			t.Fatalf("2 tiles %d, 6 tiles %d", two, six)
		}
	})

	t.Run("Dragonblood: 15% of a hit is added as fire, and the hit burns for 2 turns", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 999, 999
		power(h, "dragonblood", 15)
		var evs []action.Event
		Hurt(c, b, hits(), Attack{By: h, Skill: plain(1), Target: f}, record(&evs))
		fire := 0
		for _, e := range evs {
			if e.Kind == action.Hit && e.Skill == "fire" {
				fire += e.Amount
			}
		}
		if fire == 0 || f.Status["burn"] < 2 {
			t.Fatalf("fire %d, statuses %v", fire, f.Status)
		}
	})

	t.Run("Wyrm's Hide: the Dragon Knight can't be burned or frozen", func(t *testing.T) {
		c, b, h, f := arena('g')
		power(h, "wyrmhide", 1)
		s := plain(1)
		s.Effects = map[string]int{"burn": 3, "frozen": 2, "poison": 2}
		put := applyStatuses(c, b, hits(), Attack{By: f, Skill: s, Target: h}, h)
		if put["burn"] != 0 || put["frozen"] != 0 || put["poison"] != 2 {
			t.Fatalf("put %v", put)
		}
	})

	t.Run("Eagle Eye: ranged skills reach 1 tile further, melee ones do not", func(t *testing.T) {
		u := &state.Unit{}
		power(u, "eagle", 1)
		if u.Reach(6) != 1 || u.Reach(1) != 0 {
			t.Fatalf("ranged +%d, melee +%d", u.Reach(6), u.Reach(1))
		}
	})

	t.Run("Pinning Shot: a hit can root the target for 1 turn", func(t *testing.T) {
		c, b, h, f := arena('g')
		power(h, "pin", 25)
		if put := applyStatuses(c, b, rolls(0.9), Attack{By: h, Skill: plain(1), Target: f}, f); put["snare"] != 0 {
			t.Fatalf("a high roll rooted: %v", put)
		}
		if put := applyStatuses(c, b, rolls(0.1), Attack{By: h, Skill: plain(1), Target: f}, f); put["snare"] != 1 {
			t.Fatalf("put %v", put)
		}
	})

	t.Run("Quick Hands: the first attack each turn does not end the turn, the second does", func(t *testing.T) {
		c, b := open7()
		g := add(b, 1, state.Heroes, "Gunslinger", 3, 3)
		f := add(b, 2, state.Enemies, "Orc Brute", 3, 4)
		f.HP, f.MaxHP = 999, 999
		power(g, "quick", 1)
		cast(t, c, b, g, "Slash", Pick{Target: f})
		if g.Done || !g.QuickHands {
			t.Fatal("the first attack ended the turn")
		}
		cast(t, c, b, g, "Slash", Pick{Target: f})
		if !g.Done {
			t.Fatal("the second attack kept the turn going")
		}
	})

	t.Run("Ricochet: 20% of a hit bounces to the nearest other enemy within 2 tiles", func(t *testing.T) {
		c, b := open7()
		g := add(b, 1, state.Heroes, "Gunslinger", 3, 3)
		f := add(b, 2, state.Enemies, "Orc Brute", 3, 4)
		near := add(b, 3, state.Enemies, "Orc Brute", 3, 6)
		far := add(b, 4, state.Enemies, "Orc Brute", 0, 0)
		power(g, "ricochet", 20)
		evs := cast(t, c, b, g, "Slash", Pick{Target: f})
		main, bounce := dealt(evs, g, f), dealt(evs, g, near)
		if bounce == 0 || bounce != int(Round(float64(main)*0.2)) || dealt(evs, g, far) != 0 {
			t.Fatalf("main %d, bounce %d, far %d", main, bounce, dealt(evs, g, far))
		}
	})

	t.Run("Ambush: 30% more damage to an enemy at full HP", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Atk, f.Def = 105, 5
		power(h, "ambush", 30)
		full := dm(c, b, Attack{By: h, Skill: plain(1), Target: f})
		f.HP--
		if hurt := dm(c, b, Attack{By: h, Skill: plain(1), Target: f}); full <= hurt || hurt != 100 {
			t.Fatalf("full HP %d, hurt %d", full, hurt)
		}
	})

	t.Run("Shadowstep: a kill cloaks the Assassin until its next turn, without touching the Cloak skill", func(t *testing.T) {
		c, b, h, f := arena('g')
		b.Side = state.Heroes
		f.HP = 1
		power(h, "shadowstep", 1)
		Hurt(c, b, hits(), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if h.Status["cloak"] != 2 || !h.Shadowstep {
			t.Fatalf("cloak %d", h.Status["cloak"])
		}
		StartTurnPowers(h)
		if h.Status["cloak"] != 0 || h.Shadowstep {
			t.Fatalf("next turn: cloak %d", h.Status["cloak"])
		}
	})

	t.Run("Arcane Echo: a spell can hit its target a second time", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 999, 999
		power(h, "echo", 20)
		s := plain(1)
		s.Magic = true
		var evs []action.Event
		Hurt(c, b, rolls(0, 0.5, 0.1, 0, 0.5), Attack{By: h, Skill: s, Target: f}, record(&evs))
		n := 0
		for _, e := range evs {
			if e.Kind == action.Hit && e.Unit == h.ID && e.Target == f.ID {
				n++
			}
		}
		if n != 2 {
			t.Fatalf("%d hits", n)
		}
	})

	t.Run("Overflow: every skill cooldown is 1 turn shorter", func(t *testing.T) {
		c, b := open7()
		m := add(b, 1, state.Heroes, "Mage", 3, 3)
		f := add(b, 2, state.Enemies, "Orc Brute", 3, 5)
		f.HP, f.MaxHP = 999, 999
		power(m, "overflow", 1)
		cast(t, c, b, m, "Meteor", Pick{Target: f})
		if m.Cooldowns[0] != c.Skills["Meteor"].Cooldown-1 {
			t.Fatalf("cooldown %d", m.Cooldowns[0])
		}
	})

	t.Run("Sanctuary: a heal also shields the ally for 20% of their max HP", func(t *testing.T) {
		c, b := open7()
		hl := add(b, 1, state.Heroes, "Healer", 3, 3)
		w := add(b, 2, state.Heroes, "Warrior", 3, 4)
		w.HP = 50
		power(hl, "sanctuary", 20)
		healSkill(c, b, hits(), hl, c.Skills["Heal Ally"], w, nop)
		if w.Shield != 20 {
			t.Fatalf("shield %d", w.Shield)
		}
	})

	t.Run("Lifeline: once per battle, an ally who would fall stays at 1 HP", func(t *testing.T) {
		c, b := open7()
		hl := add(b, 1, state.Heroes, "Healer", 0, 0)
		w := add(b, 2, state.Heroes, "Warrior", 3, 3)
		f := add(b, 3, state.Enemies, "Orc Brute", 3, 4)
		power(hl, "lifeline", 1)
		w.HP, f.Atk = 5, 200
		Hurt(c, b, hits(), Attack{By: f, Skill: plain(1), Target: w}, nop)
		if w.HP != 1 || b.Unit(w.ID) == nil {
			t.Fatalf("HP %d", w.HP)
		}
		Hurt(c, b, hits(), Attack{By: f, Skill: plain(1), Target: w}, nop)
		if b.Unit(w.ID) != nil {
			t.Fatal("Lifeline saved the ally twice")
		}
	})

	t.Run("Hex Mastery: statuses the Witch causes last 1 turn longer", func(t *testing.T) {
		c, b, h, f := arena('g')
		power(h, "hexMastery", 1)
		s := plain(1)
		s.Effects = map[string]int{"poison": 2}
		if put := applyStatuses(c, b, hits(), Attack{By: h, Skill: s, Target: f}, f); put["poison"] != 3 {
			t.Fatalf("put %v", put)
		}
	})

	t.Run("Soul Siphon: each status the Witch causes heals her 5% of her max HP", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.HP, h.MaxHP = 50, 100
		power(h, "siphon", 5)
		s := plain(1)
		s.Effects = map[string]int{"poison": 2, "weak": 2}
		curse(c, b, hits(), h, s, f, nop)
		if h.HP != 60 {
			t.Fatalf("HP %d", h.HP)
		}
	})
}
