package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Tick is how much a damage-over-time status (bleed, burn, poison) hurts a
// unit each time it ticks, or how much regen heals; 0 for other statuses
// (old turn start, line 1139). Bleed is doubled by a heavy bleed. The turn
// start, Open Veins and the screen's status badges all use this one formula.
func Tick(c *content.Content, u *state.Unit, status string) int {
	t := c.Turns
	if status == "regen" {
		return int(Round(float64(u.MaxHP) * t.RegenShare))
	}
	share, ok := t.DamageOverTime[status]
	if !ok {
		return 0
	}
	hurt := float64(u.MaxHP) * share
	if status == "bleed" && u.HeavyBleed {
		hurt *= 2
	}
	return max(t.DamageOverTimeMin, int(Round(hurt)))
}
