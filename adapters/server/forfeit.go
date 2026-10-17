package main

import (
	"errors"
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
)

// Forfeit ends the match now. The player who gives up loses, and the other
// player wins. Closing the page does not do this; only an explicit give-up does.
func (r *Room) Forfeit(id string, now time.Time) ([]action.Event, error) {
	if r.member(id) == nil {
		return nil, errNoOne
	}
	if !r.playing() {
		return nil, errNoMatch
	}
	side := r.seatOf(id)
	if side < 0 {
		return nil, errors.New("You are watching.")
	}
	return r.apply(action.Action{Kind: action.Forfeit, Side: side}, now)
}
