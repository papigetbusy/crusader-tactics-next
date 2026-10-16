package combat

import "github.com/papigetbusy/crusader-tactics-next/sim/state"

// FallenOf is the list of a side's fallen, the ones a revive can bring back:
// the heroes' in single player, each player's own in versus.
func FallenOf(b *state.Battle, team int) *[]*state.Unit {
	if team == state.Heroes {
		return &b.Fallen
	}
	return &b.FallenFoes
}
