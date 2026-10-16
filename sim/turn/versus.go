package turn

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/ai"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// versusEnd ends the side to move's turn in a 1v1 match: its statuses count
// down, its companions and raised monsters act on their own, then the other
// player's turn starts. A round (Turn) is both players' turns.
func versusEnd(c *content.Content, b *state.Battle, dice rng.Dice, emit action.Emit) error {
	side := b.Side
	End(c, b, dice, side, emit)
	for _, z := range members(b, side) {
		if !z.Companion || b.Unit(z.ID) == nil || z.Status["stun"] > 0 {
			continue
		}
		if err := ai.Turn(c, b, dice, z, emit); err != nil {
			return err
		}
		combat.FlushRaises(c, b, emit)
		Check(b, emit)
		if b.Outcome != state.Ongoing {
			return nil
		}
	}
	next := state.Other(side)
	if next == state.Heroes {
		b.Turn++
	}
	b.Versus.UsedMs = 0
	Start(c, b, dice, next, emit)
	return nil
}

// versusCheck ends a 1v1 match when one player has no real heroes left
// (companions, summons and clones do not count). Won means the first player
// (side 0) won; Lost means the second player did.
func versusCheck(b *state.Battle, emit action.Emit) {
	var left [2]int
	for _, u := range b.Units {
		if u.HP > 0 && !u.Companion && !u.Summoned && !u.Decoy {
			left[u.Team]++
		}
	}
	switch {
	case left[1] == 0:
		b.Outcome = state.Won
		emit(action.Event{Kind: action.Won, Side: state.Heroes})
	case left[0] == 0:
		b.Outcome = state.Lost
		emit(action.Event{Kind: action.Lost, Side: state.Enemies})
	}
}
