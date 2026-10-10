package ai

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// open is a 7x7 plain with a monster in one corner.
func open(class string) (*content.Content, *state.Battle, *state.Unit) {
	c := content.MustDefault()
	rows := make([]string, 7)
	for i := range rows {
		rows[i] = "ggggggg"
	}
	e := &state.Unit{ID: 10, Team: state.Enemies, Class: class, Mov: 3, HP: 30, MaxHP: 30, Atk: 10, Status: map[string]int{}}
	e.Skills = c.MonsterSkills(class, 1, false)
	e.Cooldowns = make([]int, len(e.Skills))
	b := &state.Battle{N: 7, Ground: rows, Units: []*state.Unit{e}}
	return c, b, e
}

func hero(id, x, y, hp int) *state.Unit {
	return &state.Unit{ID: id, Team: state.Heroes, Class: "Warrior", X: x, Y: y, HP: hp, MaxHP: 50, Status: map[string]int{}}
}

func TestEnemyAI(t *testing.T) {
	t.Run("A melee monster walks next to the nearest hero", func(t *testing.T) {
		c, b, e := open("Warrior")
		b.Units = append(b.Units, hero(1, 3, 0, 50))
		tile, ok := ChooseTile(c, b, e, e.Skills[0])
		if !ok || abs(tile.X-3)+abs(tile.Y) != 1 {
			t.Fatalf("went to %v", tile)
		}
	})

	t.Run("A ranged monster stops at range instead of walking up close", func(t *testing.T) {
		c, b, e := open("Ranger")
		b.Units = append(b.Units, hero(1, 4, 0, 50))
		tile, _ := ChooseTile(c, b, e, e.Skills[0])
		if d := abs(tile.X-4) + abs(tile.Y); d < e.Skills[0].MinRange || d > e.Skills[0].MaxRange {
			t.Fatalf("stood %d tiles away", d)
		}
	})

	t.Run("A snared monster does not move", func(t *testing.T) {
		c, b, e := open("Warrior")
		e.Status["snare"] = 1
		b.Units = append(b.Units, hero(1, 3, 0, 50))
		if _, ok := ChooseTile(c, b, e, e.Skills[0]); ok {
			t.Fatal("snared monster moved")
		}
	})

	t.Run("Monsters attack the weakest hero in range", func(t *testing.T) {
		c, b, e := open("Warrior")
		b.Units = append(b.Units, hero(1, 1, 0, 40), hero(2, 0, 1, 12))
		if v := ChooseTarget(c, b, e, e.Skills[0]); v == nil || v.ID != 2 {
			t.Fatalf("picked %v", v)
		}
	})

	t.Run("A healer heals when an ally is badly hurt", func(t *testing.T) {
		c, b, e := open("Healer")
		ally := &state.Unit{ID: 11, Team: state.Enemies, Class: "Warrior", X: 1, HP: 5, MaxHP: 30, Status: map[string]int{}}
		b.Units = append(b.Units, ally, hero(1, 6, 6, 50))
		slot := ChooseSkill(c, b, rng.New(1, 1), e)
		if slot < 0 || !e.Skills[slot].Heal {
			t.Fatalf("healer chose slot %d", slot)
		}
	})

	t.Run("Bosses pick any ready skill at random", func(t *testing.T) {
		c, b, e := open("Lich King")
		e.Boss = true
		e.Skills = c.MonsterSkills("Lich King", 1, true)
		e.Cooldowns = make([]int, len(e.Skills))
		seen := map[int]bool{}
		d := rng.New(3, 1)
		for i := 0; i < 200; i++ {
			seen[ChooseSkill(c, b, d, e)] = true
		}
		if len(seen) != len(e.Skills) {
			t.Fatalf("boss used %d of %d skills", len(seen), len(e.Skills))
		}
	})
}
