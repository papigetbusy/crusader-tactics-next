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

// awayForfeit is how long a seated phone can be gone before the other player
// wins. A refresh reconnects inside this. A locked phone does not.
const awayForfeit = 20 * time.Second

// Away is a seated player whose phone has been gone past awayForfeit. They
// lose the match. Watchers do not end it.
func (r *Room) Away(now time.Time) (string, []action.Event, bool) {
	if !r.playing() {
		return "", nil, false
	}
	for _, id := range r.Seats {
		m := r.member(id)
		if m == nil || m.Gone.IsZero() || now.Sub(m.Gone) <= awayForfeit {
			continue
		}
		ev, err := r.Forfeit(m.ID, now)
		if err != nil {
			return "", nil, false
		}
		return m.ID, ev, true
	}
	return "", nil, false
}
