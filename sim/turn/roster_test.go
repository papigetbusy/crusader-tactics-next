package turn

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/ai"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestNewMonstersFight(t *testing.T) {
	c := content.MustDefault()
	for _, m := range []struct {
		name        string
		boss, elite bool
	}{
		{"Bone Hydra", true, false}, {"Fallen Seraph", true, false}, {"Brood Matron", true, false}, {"Ash Colossus", true, false},
		{"Crowned Slime", false, true}, {"Goblin Captain", false, true}, {"Grave Knight", false, true},
		{"Dire Wolf", false, true}, {"Iron Minotaur", false, true}, {"Blight Reaper", false, true},
	} {
		t.Run("The new "+map[bool]string{true: "boss", false: "elite"}[m.boss]+" "+m.name+" walks up to a hero and attacks on its turn", func(t *testing.T) {
			rows := []string{"ggggggg", "ggggggg", "ggggggg", "ggggggg", "ggggggg", "ggggggg", "ggggggg"}
			b := &state.Battle{N: 7, Ground: rows, Side: state.Enemies, NextID: 10}
			h := &state.Unit{ID: 1, Team: state.Heroes, Class: "Warrior", Level: 30, X: 3, Y: 6, HP: 999, MaxHP: 999, Atk: 20, Def: 5, Mov: 4, Status: map[string]int{}}
			e := floor.Enemy(c, m.name, 30, "dark", m.boss, m.elite)
			e.ID, e.X, e.Y = 2, 3, 2
			b.Units = []*state.Unit{h, e}
			hit := false
			for range 3 {
				if err := ai.Turn(c, b, rng.New(5, 5), e, func(ev action.Event) {
					if (ev.Kind == action.Hit || ev.Kind == action.Missed) && ev.Unit == e.ID {
						hit = true
					}
				}); err != nil {
					t.Fatal(err)
				}
				for i := range e.Cooldowns {
					e.Cooldowns[i] = 0
				}
				e.Done, e.Moved, e.Acted = false, false, false
				if hit {
					break
				}
			}
			if !hit {
				t.Fatalf("%s never attacked (stood at %d,%d)", m.name, e.X, e.Y)
			}
		})
	}
}
