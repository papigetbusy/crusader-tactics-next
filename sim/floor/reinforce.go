package floor

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Roll labels, so a failed old-game check can say which roll was out of order.
const (
	RollTile = "reinforcement tile"
	RollKind = "reinforcement kind"
)

// Reinforce brings monsters onto a reinforcement floor at the start of the
// enemy turn on the reinforcement turns (old reinforce, index.html line
// 1209): up to two normal monsters, never healers, at the level of the
// strongest monster, on free tiles of the top two rows.
func Reinforce(c *content.Content, b *state.Battle, dice rng.Dice, emit action.Emit) {
	f := c.Floors
	if !b.Reinforce || !slices.Contains(f.ReinforceTurns, b.Turn) {
		return
	}
	level := 1
	for _, u := range b.Units {
		if u.Team == state.Enemies {
			level = max(level, u.Level)
		}
	}
	n := 0
	for k := 0; k < 30 && n < f.ReinforceMax; k++ {
		x := rng.IntN(dice, RollTile, b.N)
		y := rng.IntN(dice, RollTile, 2)
		if b.Heights[y][x] < 0 || b.UnitAt(x, y) != nil {
			continue
		}
		var m content.Monster
		for {
			m = c.Monsters[rng.IntN(dice, RollKind, len(c.Monsters))]
			if m.Class != "Healer" {
				break
			}
		}
		u := Monster(c, m, level, dice)
		u.ID, u.X, u.Y = b.NextID, x, y
		b.NextID++
		b.Units = append(b.Units, u)
		n++
	}
	if n > 0 {
		emit(action.Event{Kind: action.Reinforced, Amount: n})
	}
}
