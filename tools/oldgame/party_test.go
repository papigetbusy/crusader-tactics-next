package oldgame

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// partyOf lists the heroes' IDs in party order, from a moment when the whole
// party is on the board (the start of a fight). The recordings list each
// party member's level and XP in this order.
func partyOf(s Snap) []int {
	var ids []int
	for _, u := range s.Units {
		if u.Team == state.Heroes && u.Cp == 0 && u.M == nil { // not a companion or a monster ally
			ids = append(ids, u.ID)
		}
	}
	return ids
}

// levelCap is the level heroes stop at on this floor: the floor's monster
// level plus the gap (old capLv, index.html line 1462). The recordings start
// on the best floor reached, and normal monsters and elites have the floor's
// level; a boss is one level higher (newFloor0, line 756).
func levelCap(c *content.Content, s Snap) int {
	low, boss := 0, 0
	for _, u := range s.Units {
		switch {
		case u.Team != state.Enemies:
		case u.Boss != 0:
			boss = u.Lv - 1
		case low == 0 || u.Lv < low:
			low = u.Lv
		}
	}
	if low == 0 {
		low = boss
	}
	return low + c.Progress.CapAboveFloor
}

// heroOf finds a party member in Go, on the board or among the fallen.
func heroOf(b *state.Battle, id int) *state.Unit {
	if u := b.Unit(id); u != nil {
		return u
	}
	for _, u := range b.Fallen {
		if u.ID == id {
			return u
		}
	}
	return nil
}
