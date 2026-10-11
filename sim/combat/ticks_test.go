package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestTicks(t *testing.T) {
	c := content.MustDefault()
	t.Run("Bleed, burn and poison hurt a share of max HP each turn, never less than 3", func(t *testing.T) {
		u := &state.Unit{MaxHP: 100}
		if Tick(c, u, "bleed") != 6 || Tick(c, u, "burn") != 6 || Tick(c, u, "poison") != 7 {
			t.Fatalf("bleed %d, burn %d, poison %d", Tick(c, u, "bleed"), Tick(c, u, "burn"), Tick(c, u, "poison"))
		}
		if small := (&state.Unit{MaxHP: 20}); Tick(c, small, "poison") != 3 {
			t.Fatalf("a small unit's poison: %d", Tick(c, small, "poison"))
		}
	})
	t.Run("A heavy bleed hurts twice as much", func(t *testing.T) {
		if u := (&state.Unit{MaxHP: 100, HeavyBleed: true}); Tick(c, u, "bleed") != 12 {
			t.Fatalf("heavy bleed %d", Tick(c, u, "bleed"))
		}
	})
	t.Run("Regen heals 12% of max HP each turn; other statuses don't tick", func(t *testing.T) {
		u := &state.Unit{MaxHP: 100}
		if Tick(c, u, "regen") != 12 || Tick(c, u, "stun") != 0 {
			t.Fatalf("regen %d, stun %d", Tick(c, u, "regen"), Tick(c, u, "stun"))
		}
	})
}
