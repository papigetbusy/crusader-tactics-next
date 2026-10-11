package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// half is dice that always roll the middle: every hit lands at the middle
// damage roll and no ordinary crit happens.
type half struct{}

func (half) Roll(string) float64 { return 0.5 }

// open7 is an empty 7x7 plain.
func open7() (*content.Content, *state.Battle) {
	rows := []string{"ggggggg", "ggggggg", "ggggggg", "ggggggg", "ggggggg", "ggggggg", "ggggggg"}
	return content.MustDefault(), &state.Battle{N: 7, Ground: rows, NextID: 50}
}

// add puts a unit with 100 HP, 20 attack, 4 defense and 4 Move on the board.
func add(b *state.Battle, id, team int, class string, x, y int) *state.Unit {
	u := &state.Unit{ID: id, Team: team, Class: class, Level: 30, X: x, Y: y, HP: 100, MaxHP: 100,
		Atk: 20, Def: 4, Mov: 4, Status: map[string]int{}}
	b.Units = append(b.Units, u)
	return u
}

// cast makes u use the named skill with this pick and returns the events.
func cast(t *testing.T, c *content.Content, b *state.Battle, u *state.Unit, name string, p Pick) []action.Event {
	t.Helper()
	s, ok := c.Skills[name]
	if !ok {
		t.Fatalf("no skill %q", name)
	}
	u.Skills, u.Cooldowns, u.Done = []content.Skill{s}, []int{0}, false
	var evs []action.Event
	if err := UseAimed(c, b, half{}, u, 0, p, func(e action.Event) { evs = append(evs, e) }); err != nil {
		t.Fatal(err)
	}
	return evs
}

// dealt is the damage u's hits did to v in these events.
func dealt(evs []action.Event, u, v *state.Unit) int {
	n := 0
	for _, e := range evs {
		if e.Kind == action.Hit && e.Unit == u.ID && e.Target == v.ID {
			n += e.Amount
		}
	}
	return n
}

func at(u *state.Unit, x, y int) bool { return u.X == x && u.Y == y }

func tile(x, y int) []state.Tile { return []state.Tile{{X: x, Y: y}} }

func TestWarriorAndTankSkills(t *testing.T) {
	t.Run("Shoulder Check shoves the target 2 tiles back", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Warrior", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		evs := cast(t, c, b, h, "Shoulder Check", Pick{Target: f})
		if !at(f, 3, 0) || f.Status["stun"] != 0 || dealt(evs, h, f) != 14 {
			t.Fatalf("foe at %d,%d, %d damage, %v", f.X, f.Y, dealt(evs, h, f), f.Status)
		}
	})

	t.Run("A Shoulder Check that slams the target into a wall does 50% more and stuns it", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Warrior", 3, 2), add(b, 2, state.Enemies, "Warrior", 3, 1)
		evs := cast(t, c, b, h, "Shoulder Check", Pick{Target: f})
		if !at(f, 3, 0) || f.Status["stun"] != 1 || dealt(evs, h, f) != 21 {
			t.Fatalf("foe at %d,%d, %d damage, %v", f.X, f.Y, dealt(evs, h, f), f.Status)
		}
	})

	t.Run("Riposte Stance strikes back at an enemy that hits the Warrior in melee", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Warrior", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		cast(t, c, b, h, "Riposte Stance", Pick{Target: h})
		Hurt(c, b, half{}, Attack{By: f, Skill: plain(1), Target: h}, nop)
		if h.Status["riposte"] != 2 || f.HP >= 100 {
			t.Fatalf("stance %d, foe HP %d", h.Status["riposte"], f.HP)
		}
	})

	t.Run("Leaping Cleave leaps next to the target, then hits it and every enemy next to the Warrior", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Warrior", 3, 3)
		f, g, far := add(b, 2, state.Enemies, "Warrior", 3, 0), add(b, 3, state.Enemies, "Warrior", 2, 1), add(b, 4, state.Enemies, "Warrior", 6, 6)
		evs := cast(t, c, b, h, "Leaping Cleave", Pick{Target: f})
		if !at(h, 3, 1) || dealt(evs, h, f) == 0 || dealt(evs, h, g) == 0 || far.HP != 100 {
			t.Fatalf("hero at %d,%d; target %d, neighbour %d, far %d", h.X, h.Y, f.HP, g.HP, far.HP)
		}
	})

	t.Run("Sunder strips the target's buffs and shield", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Warrior", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		f.Status["grd"], f.Status["rally"], f.Shield = 2, 2, 10
		cast(t, c, b, h, "Sunder", Pick{Target: f})
		if f.Status["grd"]+f.Status["rally"]+f.Shield != 0 {
			t.Fatalf("%v shield %d", f.Status, f.Shield)
		}
	})

	t.Run("Battle Trance gives +25% damage and +1 Move, but the Warrior takes 15% more damage", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Warrior", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 2)
		f.Atk = 40
		cast(t, c, b, h, "Battle Trance", Pick{Target: h})
		dealtOut, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1)
		taken, _ := Damage(c, b, Attack{By: f, Skill: plain(1), Target: h}, 1)
		if dealtOut != 21 || taken != 41 || board.Budget(b, h) != 10 {
			t.Fatalf("deals %d (want 21), takes %d (want 41), budget %d", dealtOut, taken, board.Budget(b, h))
		}
	})

	t.Run("A Momentum kill lets the Warrior act once more, once per turn", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Warrior", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 2), add(b, 3, state.Enemies, "Warrior", 2, 3)
		f.HP, g.HP = 1, 1
		cast(t, c, b, h, "Momentum", Pick{Target: f})
		if h.Done || !h.MomentumUsed {
			t.Fatal("the first kill should give another action")
		}
		cast(t, c, b, h, "Momentum", Pick{Target: g})
		if !h.Done {
			t.Fatal("only once per turn")
		}
	})

	t.Run("Hook Chain drags the target to the tile next to the Tank", func(t *testing.T) {
		c, b := open7()
		h, f := add(b, 1, state.Heroes, "Tank", 3, 3), add(b, 2, state.Enemies, "Warrior", 3, 0)
		cast(t, c, b, h, "Hook Chain", Pick{Target: f})
		if !at(f, 3, 2) {
			t.Fatalf("foe at %d,%d", f.X, f.Y)
		}
	})

	t.Run("Bodyguard makes the Tank take hits aimed at the ally, for 70% of the damage", func(t *testing.T) {
		c, b := open7()
		h, ally, f := add(b, 1, state.Heroes, "Tank", 3, 3), add(b, 2, state.Heroes, "Mage", 3, 5), add(b, 3, state.Enemies, "Warrior", 3, 6)
		cast(t, c, b, h, "Bodyguard", Pick{Target: ally})
		r := Hurt(c, b, half{}, Attack{By: f, Skill: plain(1), Target: ally}, nop)
		if r.Target != h || ally.HP != 100 || r.Damage != 11 {
			t.Fatalf("hit %d on unit %d", r.Damage, r.Target.ID)
		}
	})

	t.Run("Iron Shell gives the Tank a shield worth 25% of its max HP", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Tank", 3, 3)
		cast(t, c, b, h, "Iron Shell", Pick{Target: h})
		if h.Shield != 25 {
			t.Fatalf("shield %d", h.Shield)
		}
	})

	t.Run("Taunting Roar taunts every enemy within 3 tiles", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Tank", 3, 3)
		near, edge, far := add(b, 2, state.Enemies, "Warrior", 3, 1), add(b, 3, state.Enemies, "Warrior", 3, 6), add(b, 4, state.Enemies, "Warrior", 0, 0)
		cast(t, c, b, h, "Taunting Roar", Pick{Target: h})
		if near.Status["taunted"] != 1 || near.TauntedBy != 1 || edge.Status["taunted"] != 1 || far.Status["taunted"] != 0 {
			t.Fatal("wrong enemies taunted")
		}
	})

	t.Run("A taunted unit's attacks can only go at the unit that taunted it", func(t *testing.T) {
		c, b := open7()
		m := add(b, 1, state.Heroes, "Mage", 3, 5)
		roarer, other := add(b, 2, state.Enemies, "Tank", 3, 3), add(b, 3, state.Enemies, "Warrior", 3, 1)
		m.Status["taunted"], m.TauntedBy = 1, roarer.ID
		ts := Targets(c, b, m, c.Skills["Fireball"])
		if len(ts) != 1 || ts[0] != roarer || other == nil {
			t.Fatalf("targets %v", ts)
		}
	})

	t.Run("Aftershock snares what it hits and leaves half its damage for the next turn's quake", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Tank", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 2), add(b, 3, state.Enemies, "Warrior", 2, 2)
		evs := cast(t, c, b, h, "Aftershock", Pick{Target: f})
		for _, z := range []*state.Unit{f, g} {
			if d := dealt(evs, h, z); d == 0 || z.Quake != int(Round(float64(d)*0.5)) || z.Status["snare"] != 1 {
				t.Fatalf("unit %d: %d damage, quake %d, %v", z.ID, d, z.Quake, z.Status)
			}
		}
	})

	t.Run("Unstoppable hits every enemy in the line and shoves each aside", func(t *testing.T) {
		c, b := open7()
		h := add(b, 1, state.Heroes, "Tank", 3, 3)
		f, g := add(b, 2, state.Enemies, "Warrior", 3, 2), add(b, 3, state.Enemies, "Warrior", 3, 1)
		evs := cast(t, c, b, h, "Unstoppable", Pick{Tiles: tile(3, 0)})
		if !at(h, 3, 0) || dealt(evs, h, f) == 0 || dealt(evs, h, g) == 0 || f.X == 3 || g.X == 3 {
			t.Fatalf("tank %d,%d; foes %d,%d and %d,%d", h.X, h.Y, f.X, f.Y, g.X, g.Y)
		}
	})
}
