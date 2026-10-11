package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestGunslingerKnightDragonWitchSkills(t *testing.T) {
	t.Run("Boot Shot kicks an enemy next to the Gunslinger 2 tiles away", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Gunslinger", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Boot Shot", Pick{Target: f})
		if !at(f, 3, 0) || f.HP == 100 {
			t.Fatalf("foe at %d,%d", f.X, f.Y)
		}
	})

	t.Run("Disarming Shot makes the target deal 30% less damage for 2 turns", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Gunslinger", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Disarming Shot", Pick{Target: f})
		if f.Status["weak"] != 2 {
			t.Fatalf("%v", f.Status)
		}
	})

	t.Run("Roll Out rolls to a tile, then shoots from there", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Gunslinger", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		p := Pick{Target: f, Tiles: tile(3, 5)}
		if !PickOK(c, b, h, c.Skills["Roll Out"], p) {
			t.Fatal("should be allowed")
		}
		cast(t, c, b, h, "Roll Out", p)
		if !at(h, 3, 5) || f.HP == 100 {
			t.Fatalf("at %d,%d", h.X, h.Y)
		}
	})

	t.Run("A Powder Keg blows up at the start of the next enemy turn, hitting everything within 1 tile", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Gunslinger", 3, 3)
		f, far := add(b, 2, state.Enemies, "Warrior", 3, 1), add(b, 3, state.Enemies, "Warrior", 6, 6)
		cast(t, c, b, h, "Powder Keg", Pick{Tiles: tile(3, 0)})
		if f.HP != 100 || len(b.Kegs) != 1 {
			t.Fatal("the keg should wait")
		}
		Kegs(c, b, half{}, state.Enemies, nop)
		if f.HP == 100 || far.HP != 100 || len(b.Kegs) != 0 {
			t.Fatalf("near %d far %d", f.HP, far.HP)
		}
	})

	t.Run("Lucky Coin makes the next 2 shots critical hits, without using up the turn", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Gunslinger", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Lucky Coin", Pick{Target: h})
		r := Hurt(c, b, half{}, Attack{By: h, Skill: c.Skills["Quick Shot"], Target: f}, nop)
		if h.Done || !r.Crit || h.Status["coin"] != 1 {
			t.Fatalf("done %v crit %v coins %d", h.Done, r.Crit, h.Status["coin"])
		}
	})

	t.Run("High Noon hits 10% harder for every enemy in range", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Gunslinger", 3, 3)
		f := add(b, 2, state.Enemies, "Warrior", 3, 0)
		add(b, 3, state.Enemies, "Warrior", 0, 3)
		add(b, 4, state.Enemies, "Warrior", 6, 3)
		if dm, _ := Damage(c, b, Attack{By: h, Skill: c.Skills["High Noon"], Target: f}, 1); dm != 30 {
			t.Fatalf("damage %d, want (20x1.3x1.3)-4 = 30", dm)
		}
	})

	t.Run("Trample rides straight through and hits every enemy passed", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Knight", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 2), add(b, 3, state.Enemies, "Warrior", 3, 1)
		evs := cast(t, c, b, h, "Trample", Pick{Tiles: tile(3, 0)})
		if !at(h, 3, 0) || dealt(evs, h, f) == 0 || dealt(evs, h, g) == 0 || !at(f, 3, 2) {
			t.Fatalf("knight %d,%d", h.X, h.Y)
		}
	})

	t.Run("Shield of Faith gives an ally a shield worth 20% of the Knight's max HP", func(t *testing.T) {
		c, b := open7()
		h, ally := add(b, 1, state.Heroes, "Knight", 3, 3), add(b, 2, state.Heroes, "Mage", 3, 5)
		cast(t, c, b, h, "Shield of Faith", Pick{Target: ally})
		if ally.Shield != 20 {
			t.Fatalf("shield %d", ally.Shield)
		}
	})

	t.Run("Unhorse knocks the target 1 tile back and roots it", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Unhorse", Pick{Target: f})
		if !at(f, 3, 1) || f.Status["snare"] != 1 {
			t.Fatalf("foe at %d,%d %v", f.X, f.Y, f.Status)
		}
	})

	t.Run("Lance Wall strikes an enemy that walks up next to the Knight", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Lance Wall", Pick{Target: h})
		LanceWall(c, b, half{}, f, []board.Tile{{X: 3, Y: 0}, {X: 3, Y: 2}}, nop)
		if f.HP == 100 {
			t.Fatal("no strike")
		}
	})

	t.Run("Rescue pulls an ally next to the Knight, and it takes 20% less damage", func(t *testing.T) {
		c, b := open7()
		h, ally := add(b, 1, state.Heroes, "Knight", 3, 3), add(b, 2, state.Heroes, "Mage", 3, 6)
		f := add(b, 3, state.Enemies, "Warrior", 0, 6)
		cast(t, c, b, h, "Rescue", Pick{Target: ally})
		if dm, _ := Damage(c, b, Attack{By: f, Skill: plain(1), Target: ally}, 1); !at(ally, 3, 4) || dm != 13 {
			t.Fatalf("ally at %d,%d takes %d", ally.X, ally.Y, dm)
		}
	})

	t.Run("Relentless Charge lets the Knight walk on with the Move it has left", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		h.TilesMoved = 1
		cast(t, c, b, h, "Relentless Charge", Pick{Target: f})
		if h.Done || h.Moved || !h.Acted || board.Budget(b, h) != 6 {
			t.Fatalf("done %v moved %v budget %d", h.Done, h.Moved, board.Budget(b, h))
		}
	})

	t.Run("Tail Sweep hits every enemy next to the Dragon Knight and shoves each 1 tile away", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Dragon Knight", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 2), add(b, 3, state.Enemies, "Warrior", 2, 3)
		cast(t, c, b, h, "Tail Sweep", Pick{Target: h})
		if !at(f, 3, 1) || !at(g, 1, 3) || f.HP == 100 || g.HP == 100 {
			t.Fatal("not swept")
		}
	})

	t.Run("Dragon Scales cut spell damage by 30% and throw back 20% of melee damage", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Dragon Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		f.Atk = 40
		cast(t, c, b, h, "Dragon Scales", Pick{Target: h})
		spell := plain(1)
		spell.Magic = true
		if dm, _ := Damage(c, b, Attack{By: f, Skill: spell, Target: h}, 1); dm != 27 {
			t.Fatalf("spell damage %d, want (40-2)x0.7 = 27", dm)
		}
		r := Hurt(c, b, half{}, Attack{By: f, Skill: plain(1), Target: h}, nop)
		if f.HP != 100-int(Round(float64(r.Damage)*0.2)) {
			t.Fatalf("thrown back: foe HP %d after a %d hit", f.HP, r.Damage)
		}
	})

	t.Run("Wyrm Roar pushes enemies within 2 tiles away and weakens them", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Dragon Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 1)
		cast(t, c, b, h, "Wyrm Roar", Pick{Target: h})
		if !at(f, 3, 0) || f.Status["weak"] != 2 || f.HP != 100 {
			t.Fatalf("foe at %d,%d %v", f.X, f.Y, f.Status)
		}
	})

	t.Run("Ember Trail flies along a line and sets every tile passed on fire", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Dragon Knight", 3, 3)
		add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Ember Trail", Pick{Tiles: tile(3, 0)})
		if !at(h, 3, 0) || len(b.Fields) != 3 {
			t.Fatalf("at %d,%d, %d fires", h.X, h.Y, len(b.Fields))
		}
	})

	t.Run("Skewer lances a target 2 tiles away and drags it next to the Dragon Knight", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Dragon Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 1)
		cast(t, c, b, h, "Skewer", Pick{Target: f})
		if !at(f, 3, 2) {
			t.Fatalf("foe at %d,%d", f.X, f.Y)
		}
	})

	t.Run("Wyrmbond's dragon claws an enemy next to the rider at the end of each turn", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Dragon Knight", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Wyrmbond", Pick{Target: h})
		WyrmClaws(c, b, half{}, state.Heroes, nop)
		if h.Status["wyrm"] != 3 || f.HP == 100 {
			t.Fatal("no claw")
		}
	})

	t.Run("Voodoo Link passes half of the damage one linked enemy takes to the other", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Witch", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 0), add(b, 3, state.Enemies, "Warrior", 0, 3)
		cast(t, c, b, h, "Voodoo Link", Pick{Target: f, Second: g})
		r := Hurt(c, b, half{}, Attack{By: h, Skill: plain(1), Target: f}, nop)
		if g.HP != 100-r.Damage/2 || f.Status["link"] != 3 {
			t.Fatalf("linked foe HP %d after a %d hit", g.HP, r.Damage)
		}
	})

	t.Run("Wither halves healing and stops buffs", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Witch", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Wither", Pick{Target: f})
		f.HP = 50
		heal(f, 20)
		giveBuffs(f, map[string]int{"rally": 2}, nop)
		if f.HP != 60 || f.Status["rally"] != 0 {
			t.Fatalf("HP %d %v", f.HP, f.Status)
		}
	})

	t.Run("Puppet Strings moves an enemy up to 2 tiles wherever the Witch chooses", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Witch", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 1)
		cast(t, c, b, h, "Puppet Strings", Pick{Target: f, Tiles: tile(5, 1)})
		if !at(f, 5, 1) || PickOK(c, b, h, c.Skills["Puppet Strings"], Pick{Target: f, Tiles: tile(0, 0)}) {
			t.Fatal("wrong spot")
		}
	})

	t.Run("Soul Siphon heals the Witch for half the damage", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Witch", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		h.HP = 50
		evs := cast(t, c, b, h, "Soul Siphon", Pick{Target: f})
		if h.HP != 50+int(Round(float64(dealt(evs, h, f))*0.5)) {
			t.Fatalf("HP %d after %d", h.HP, dealt(evs, h, f))
		}
	})

	t.Run("Contagion copies the target's curses and poison onto enemies within 2 tiles of it", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Witch", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 0), add(b, 3, state.Enemies, "Warrior", 5, 0)
		f.Status["poison"], f.Status["weak"] = 3, 2
		cast(t, c, b, h, "Contagion", Pick{Target: f})
		if g.Status["poison"] != 3 || g.Status["weak"] != 2 || f.HP != 100 {
			t.Fatalf("%v", g.Status)
		}
	})

	t.Run("Bog of Despair snares and poisons enemies standing in the swamp", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Witch", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 1)
		cast(t, c, b, h, "Bog of Despair", Pick{Tiles: tile(3, 0)})
		FieldTick(b, state.Heroes)
		if f.Status["snare"] != 1 || f.Status["poison"] != 2 {
			t.Fatalf("%v", f.Status)
		}
	})
}
