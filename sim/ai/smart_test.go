package ai

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// pick returns who monster e attacks with its first skill on floor fl.
func pick(fl int, boss bool, heroes ...*state.Unit) *state.Unit {
	c, b, e := open("Warrior")
	b.Floor = fl
	e.Boss = boss
	e.Element = "fire"
	b.Units = append(b.Units, heroes...)
	return ChooseTarget(c, b, e, e.Skills[0])
}

func TestSmartEnemies(t *testing.T) {
	sturdy := func(id, x, y, hp int) *state.Unit { h := hero(id, x, y, hp); h.Def = 60; return h }
	t.Run("Below floor 20 monsters attack the weakest hero in reach, as in the old game", func(t *testing.T) {
		if v := pick(19, false, hero(1, 1, 0, 9), sturdy(2, 0, 1, 7)); v == nil || v.ID != 2 {
			t.Fatalf("picked %v", v)
		}
	})
	t.Run("From floor 20 a monster finishes off a hero it can defeat this turn", func(t *testing.T) {
		if v := pick(20, false, hero(1, 1, 0, 9), sturdy(2, 0, 1, 7)); v == nil || v.ID != 1 {
			t.Fatalf("picked %v", v)
		}
	})
	t.Run("From floor 20 a monster goes for a healer over a weaker fighter", func(t *testing.T) {
		h := hero(2, 0, 1, 45)
		h.Class = "Healer"
		if v := pick(20, false, hero(1, 1, 0, 40), h); v == nil || v.ID != 2 {
			t.Fatalf("picked %v", v)
		}
	})
	t.Run("From floor 20 a monster goes for a hero weak to its element", func(t *testing.T) {
		c, b, e := open("Mage")
		b.Floor = 20
		e.Element = "fire"
		slot := -1
		for i, s := range e.Skills {
			if s.Element != "" && !s.Heal {
				slot = i
			}
		}
		if slot < 0 {
			t.Fatal("the Mage monster has no spell")
		}
		weak := hero(2, 0, 3, 45)
		weak.Element = "ice"
		b.Units = append(b.Units, hero(1, 3, 0, 40), weak)
		if v := ChooseTarget(c, b, e, e.Skills[slot]); v == nil || v.ID != 2 {
			t.Fatalf("picked %v", v)
		}
	})
	t.Run("Bosses and elites think smarter from floor 10", func(t *testing.T) {
		if v := pick(10, true, hero(1, 1, 0, 9), sturdy(2, 0, 1, 7)); v == nil || v.ID != 1 {
			t.Fatalf("picked %v", v)
		}
		if v := pick(9, true, hero(1, 1, 0, 9), sturdy(2, 0, 1, 7)); v == nil || v.ID != 2 {
			t.Fatalf("below floor 10 picked %v", v)
		}
	})
	t.Run("From floor 50 a monster that can't attack avoids ending close to two heroes", func(t *testing.T) {
		near := func(fl int) int {
			c, b, e := open("Warrior")
			e.Mov = 2
			b.Floor = fl
			b.Units = append(b.Units, hero(1, 4, 0, 50), hero(2, 2, 2, 50))
			tile, _ := ChooseTile(c, b, e, e.Skills[0])
			n := 0
			for _, h := range b.Units[1:] {
				if abs(h.X-tile.X)+abs(h.Y-tile.Y) <= 2 {
					n++
				}
			}
			return n
		}
		if near(49) != 2 || near(50) != 1 {
			t.Fatalf("floor 49: %d heroes close, floor 50: %d", near(49), near(50))
		}
	})
}
