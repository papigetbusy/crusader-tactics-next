package main

import (
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
)

// ClockView is the server's clock, sent about once a second during a match.
// It is not a move. The phone only displays it.
type ClockView struct {
	Type    string `json:"type"` // "clock"
	Seq     int    `json:"seq"`
	ThinkMs int    `json:"thinkMs"`
	Side    int    `json:"side"`
	BankMs  [2]int `json:"bankMs"`
}

// chargeMs is the time to charge, after the free seconds that follow a move.
func chargeMs(last, now time.Time, grace time.Duration) int {
	d := now.Sub(last) - grace
	if d < 0 {
		return 0
	}
	return int(d / time.Millisecond)
}

func (r *Room) think(now time.Time) int {
	if !r.playing() {
		return 0
	}
	return r.charge(now)
}

// charge is the time used since the last accepted action. The free seconds
// only follow a turn starting or a real turn action (attack, ability, wait,
// defend, potion, end turn). After a move, an undo, a weapon swap or a gear
// change every second counts, so repeating those can't stop the clock
// (owner, 2026-10-16: the move-and-undo exploit).
func (r *Room) charge(now time.Time) int {
	if r.Free {
		return chargeMs(r.Last, now, r.grace)
	}
	return chargeMs(r.Last, now, 0)
}

// earnsFree is whether an action is followed by the free seconds.
func earnsFree(k action.Kind) bool {
	switch k {
	case action.Attack, action.Ability, action.Wait, action.Defend, action.Potion, action.EndTurn, action.Clock:
		return true
	}
	return false
}
