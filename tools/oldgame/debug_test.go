package oldgame

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// debugging reports the step chosen with OLDGAME_DEBUG="set:fight:step"
// (fight and step counted from 1): the check then prints what happened.
func debugging(set string, fi, si int) bool {
	return os.Getenv("OLDGAME_DEBUG") == fmt.Sprintf("%s:%d:%d", set, fi+1, si+1)
}

// printUnits prints Go's units before the step being debugged.
func printUnits(g *engine.Game) {
	for _, u := range g.B.Units {
		var sk []string
		for _, x := range u.Skills {
			sk = append(sk, x.Name)
		}
		fmt.Println("before:", u.ID, u.Class, u.Team, u.Level, u.Companion, u.Undead, sk, u.Cooldowns, u.Status, u.Done)
	}
}

// debugStep prints the old game's and Go's events for one step when
// OLDGAME_DEBUG is "set:fight:step" (fight and step counted from 1).
func debugStep(set string, fi, si int, step Step, evs []action.Event) {
	if !debugging(set, fi, si) {
		return
	}
	for _, e := range step.Ev {
		b, _ := json.Marshal(e)
		fmt.Println("old:", string(b))
	}
	for _, e := range evs {
		b, _ := json.Marshal(e)
		fmt.Println(" go:", string(b))
	}
}

// loudDice prints every roll Go asks for (debugging only).
type loudDice struct{ *siteFeed }

func (d loudDice) Roll(label string) (v float64) {
	fmt.Println("roll:", label)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("MISMATCH:", r)
			v = 0.5
		}
	}()
	return d.siteFeed.Roll(label)
}

// skeletonsFromRecording makes Go's Grave Bloom skeletons match the old
// game's. The old game raised each one on a half-second timer after the
// kill, so it appeared whenever the player happened to be by then (even in
// the middle of the enemy turn). Go raises it right after the kill. So the
// check removes skeletons Go raised that the old game had not shown yet, and
// adds the old game's when they appear.
func skeletonsFromRecording(c *content.Content, g *engine.Game, s Snap, seen int) int {
	in := map[int]SnapUnit{}
	for _, su := range s.Units {
		if su.Cp != 0 && su.Team == state.Heroes {
			in[su.ID] = su
		}
	}
	for _, u := range slices.Clone(g.B.Units) {
		if u.Undead && u.Companion && u.ID > seen {
			if _, ok := in[u.ID]; !ok {
				g.B.Remove(u.ID)
			}
		}
	}
	n := 0
	for id, su := range in {
		if g.B.Unit(id) == nil && id > seen {
			g.B.Units = append(g.B.Units, goUnit(c, su))
			n++
		}
	}
	return n
}

// timerRaised reports a step in which the old game's monsters hit a hero-side
// unit that did not exist when the step began: a Grave Bloom skeleton its
// half-second timer raised in the middle of the enemy turn (Go raises it at
// the kill, an approved change). Go can't follow that step, so the check
// takes the step's result from the recording.
func timerRaised(prev Snap, step Step) bool {
	had := map[int]bool{}
	for _, u := range prev.Units {
		had[u.ID] = true
	}
	for _, e := range step.Ev {
		if e.K == "hit" && e.B != nil && e.B.Team == state.Heroes && !had[e.B.ID] {
			return true
		}
	}
	return false
}
