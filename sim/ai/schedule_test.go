package ai

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// fixed makes dice that always return this fraction.
func fixed(f float64, n int) *rng.Feed {
	v := make([]uint32, n)
	for i := range v {
		v[i] = uint32(f * 4294967296.0)
	}
	return rng.NewFeed(v)
}

func TestScheduledMonsters(t *testing.T) {
	t.Run("A silenced monster can only use Strike", func(t *testing.T) {
		c, b, e := open("Warrior")
		e.Skills = c.MonsterSkills("Warrior", 27, false)
		e.Cooldowns = make([]int, len(e.Skills))
		e.Status["silence"] = 2
		b.Units = append(b.Units, hero(1, 0, 1, 50))
		if got := chooseScheduled(c, b, fixed(0, 8), e); got < 0 || !e.Skills[got].Strike {
			t.Fatalf("picked slot %d", got)
		}
	})

	t.Run("A monster healer brings back a fallen monster before anything else", func(t *testing.T) {
		c, b, e := open("Healer")
		e.Level = 27
		e.Skills = c.MonsterSkills("Healer", 27, false)
		e.Cooldowns = make([]int, len(e.Skills))
		b.FallenFoes = []*state.Unit{{ID: 11, Team: state.Enemies, Class: "Warrior", MaxHP: 20, Status: map[string]int{}}}
		b.Units = append(b.Units, hero(1, 3, 3, 50))
		got := chooseScheduled(c, b, fixed(0.99, 8), e)
		if got < 0 || e.Skills[got].Name != "Resurrection" {
			t.Fatalf("picked slot %d", got)
		}
	})

	t.Run("A high-level monster always has Strike to fall back on when its skills are cooling down", func(t *testing.T) {
		c, _, e := open("Warrior")
		e.Skills = c.MonsterSkills("Warrior", 27, false)
		e.Cooldowns = make([]int, len(e.Skills))
		for i, s := range e.Skills {
			if !s.Strike {
				e.Cooldowns[i] = 2
			}
		}
		i := plainAttack(e)
		if i < 0 || e.Skills[i].Cooldown != 0 || e.Skills[i].Friendly() {
			t.Fatalf("plain attack slot %d", i)
		}
	})
}
