package turn

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestLifelineAtTurnStart(t *testing.T) {
	t.Run("Lifeline also keeps an ally at 1 HP who would die of poison at the start of their turn, once per battle", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		healer := &state.Unit{ID: 3, Team: state.Heroes, Class: "Healer", X: 1, HP: 50, MaxHP: 100, Status: map[string]int{},
			Affix: map[string]int{"lifeline": 1}}
		b.Units = append(b.Units, healer)
		h.HP, h.Status["poison"] = 2, 3
		Start(c, b, rng.New(1, 1), state.Heroes, quiet)
		if h.HP != 1 || b.Unit(h.ID) == nil {
			t.Fatalf("HP %d, still on the board: %v", h.HP, b.Unit(h.ID) != nil)
		}
		Start(c, b, rng.New(1, 1), state.Heroes, quiet)
		if b.Unit(h.ID) != nil {
			t.Fatal("Lifeline saved the ally twice")
		}
	})
}
