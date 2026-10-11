package turn

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestGearAtTurnStart(t *testing.T) {
	t.Run("Regeneration gear heals its share of max HP at the start of each of the hero's turns", func(t *testing.T) {
		c, b, h, _ := field("ggg")
		h.Affix = map[string]int{"regen": 4}
		Start(c, b, rng.New(1, 1), state.Heroes, quiet)
		if h.HP != 54 {
			t.Fatalf("hp %d, want 54", h.HP)
		}
	})
}
