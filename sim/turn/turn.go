// Package turn runs the rhythm of a battle: the start and end of each side's
// turn, the enemy phase, and win or loss (old startPhase, endPhase,
// enemyPhase and check, index.html lines 1735-1752).
package turn

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/ai"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func members(b *state.Battle, side int) []*state.Unit {
	var out []*state.Unit
	for _, u := range b.Units {
		if u.Team == side {
			out = append(out, u)
		}
	}
	return out
}

// Start begins a side's turn: cooldowns tick down, burn, bleed and poison
// hurt, forts heal, flags reset, stunned units skip. Units brought to 0 HP die.
func Start(c *content.Content, b *state.Battle, side int, emit action.Emit) {
	t := c.Turns
	dot := func(u *state.Unit, status string) {
		if u.Status[status] <= 0 {
			return
		}
		n := max(t.DamageOverTimeMin, int(combat.Round(float64(u.MaxHP)*t.DamageOverTime[status])))
		u.HP -= n
		u.Status[status]--
		emit(action.Event{Kind: action.Ticked, Unit: u.ID, Amount: n, Statuses: map[string]int{status: u.Status[status]}})
	}
	b.Side = side
	emit(action.Event{Kind: action.TurnStart, Side: side, Turn: b.Turn})
	for _, u := range members(b, side) {
		for i, cd := range u.Cooldowns {
			if cd > 0 && cd < 99 {
				u.Cooldowns[i] = cd - 1
			}
		}
		dot(u, "burn")
		u.Defending, u.Done, u.Moved, u.Acted, u.TilesMoved = false, false, false, false, 0
		dot(u, "bleed")
		dot(u, "poison")
		if tr, ok := c.Terrain[board.GroundAt(b, u.X, u.Y)]; ok && tr.Heal > 0 && u.HP > 0 && u.HP < u.MaxHP {
			u.HP = min(u.MaxHP, u.HP+int(combat.Round(float64(u.MaxHP)*tr.Heal)))
		}
		if u.Status["regen"] > 0 && u.HP > 0 {
			u.HP = min(u.MaxHP, u.HP+int(combat.Round(float64(u.MaxHP)*t.RegenShare)))
		}
		for _, q := range t.Stunned {
			if u.Status[q] > 0 {
				u.Done = true
			}
		}
	}
	for _, u := range slices.Clone(b.Units) {
		if u.HP <= 0 {
			b.Remove(u.ID)
			emit(action.Event{Kind: action.Died, Unit: u.ID})
		}
	}
	Check(b, emit)
}

// End finishes a side's turn: timed statuses count down by one.
func End(c *content.Content, b *state.Battle, side int, emit action.Emit) {
	for _, u := range members(b, side) {
		for _, q := range c.Turns.CountdownAtEnd {
			if u.Status[q] > 0 {
				u.Status[q]--
			}
		}
	}
	emit(action.Event{Kind: action.TurnEnd, Side: side})
}

// Check decides the battle once one side is gone.
func Check(b *state.Battle, emit action.Emit) {
	if b.Outcome != state.Ongoing {
		return
	}
	if len(members(b, state.Enemies)) == 0 {
		b.Outcome = state.Won
		emit(action.Event{Kind: action.Won})
	} else if len(members(b, state.Heroes)) == 0 {
		b.Outcome = state.Lost
		emit(action.Event{Kind: action.Lost})
	}
}

// EndTurn is the player ending their turn: the heroes' turn ends, every
// monster acts in board order, then a new turn starts for the heroes.
func EndTurn(c *content.Content, b *state.Battle, dice rng.Dice, emit action.Emit) error {
	End(c, b, state.Heroes, emit)
	Start(c, b, state.Enemies, emit)
	for _, e := range members(b, state.Enemies) {
		if b.Unit(e.ID) == nil || e.Done {
			continue
		}
		if err := ai.Turn(c, b, dice, e, emit); err != nil {
			return err
		}
		Check(b, emit)
		if b.Outcome == state.Lost {
			return nil
		}
	}
	End(c, b, state.Enemies, emit)
	b.Turn++
	Start(c, b, state.Heroes, emit)
	return nil
}
