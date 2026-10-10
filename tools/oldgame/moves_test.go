package oldgame

import (
	"fmt"
	"os"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

const recordings = "../../testdata/oldgame/recordings/slice1-recordings.json.gz"

func oldContent(t *testing.T) *content.Content {
	c, err := content.FromOld(os.DirFS("../../testdata/oldgame/tables"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestOldGameMoves replays every recorded fight and checks each move the old
// game made (hero moves, auto-walks before attacks, and every enemy move)
// against Go's movement rules, at the exact moment it happened.
func TestOldGameMoves(t *testing.T) {
	fights, err := Load(recordings)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var checked, heroMoves, enemyMoves, skipped, reinforced, dotDeaths int
	var problems []string
	report := func(format string, a ...any) {
		if len(problems) < 20 {
			problems = append(problems, fmt.Sprintf(format, a...))
		}
	}

	for fi := range fights {
		f := &fights[fi]
		if !f.Usable() {
			skipped++
			continue
		}
		prev := f.Init.Snap
		for si, step := range f.Steps {
			b := f.Battle(prev)
			reinforced += addArrivals(f, b, step)
			for ei, e := range step.Ev {
				switch e.K {
				case "startPhase":
					dotDeaths += removeTurnStartDeaths(b, step, ei)
				case "mv":
					u := b.Unit(e.ID)
					if u == nil {
						report("fight %d step %d: unit %d moved but is not on the board", fi+1, si+1, e.ID)
						continue
					}
					if u.X != e.From[0] || u.Y != e.From[1] {
						report("fight %d step %d: %s %d was at %d,%d in Go but the old game moved it from %d,%d",
							fi+1, si+1, u.Class, u.ID, u.X, u.Y, e.From[0], e.From[1])
					}
					if u.Team == state.Heroes && !board.CanMove(u) {
						report("fight %d step %d: hero %s moved although Go says it cannot move", fi+1, si+1, u.Class)
					}
					if !board.CanReach(c, b, u, e.To[0], e.To[1]) {
						report("fight %d step %d: the old game moved %s %d from %d,%d to %d,%d, but Go says that tile is out of reach",
							fi+1, si+1, u.Class, u.ID, u.X, u.Y, e.To[0], e.To[1])
					}
					u.X, u.Y, u.Moved = e.To[0], e.To[1], true
					checked++
					if u.Team == state.Heroes {
						heroMoves++
					} else {
						enemyMoves++
					}
				case "hit":
					if e.B != nil && e.BHP != nil {
						if v := b.Unit(e.B.ID); v != nil {
							v.HP = *e.BHP
							if v.HP <= 0 {
								b.Remove(v.ID)
							}
						}
					}
				}
			}
			// After the step, every unit Go tracked must stand where the old game says.
			for _, su := range step.Snap.Units {
				if u := b.Unit(su.ID); u != nil && (u.X != su.X || u.Y != su.Y) {
					report("fight %d step %d: after the step, %s %d is at %d,%d in Go but %d,%d in the old game",
						fi+1, si+1, su.Class, su.ID, u.X, u.Y, su.X, su.Y)
				}
			}
			prev = step.Snap
		}
	}

	t.Logf("checked %d recorded moves (%d hero, %d enemy) across %d fights; %d reinforcements placed and %d poison/bleed deaths taken from the recording; skipped %d fights where the old game crashed",
		checked, heroMoves, enemyMoves, len(fights)-skipped, reinforced, dotDeaths, skipped)
	if checked < 6000 {
		t.Errorf("expected about 6,700 recorded moves, checked only %d", checked)
	}
	for _, p := range problems {
		t.Error(p)
	}
}

// addArrivals puts reinforcements on the board. Where they arrive is a floor
// rule ported later, so for now they are placed where the old game put them:
// the start of their first recorded move, or where they ended the step.
func addArrivals(f *Fight, b *state.Battle, step Step) int {
	added := 0
	for _, su := range step.Snap.Units {
		if b.Unit(su.ID) != nil {
			continue
		}
		x, y := su.X, su.Y
		for _, e := range step.Ev {
			if e.K == "mv" && e.ID == su.ID {
				x, y = e.From[0], e.From[1]
				break
			}
		}
		u := f.Battle(Snap{Units: []SnapUnit{su}}).Units[0]
		u.X, u.Y, u.Moved = x, y, false
		b.Units = append(b.Units, u)
		added++
	}
	return added
}

// removeTurnStartDeaths handles units killed by poison or bleed at the start
// of a turn. That rule is ported in step 4c; until then the deaths are taken
// from the recording: a unit that is gone after the step and plays no further
// part after this point died here.
func removeTurnStartDeaths(b *state.Battle, step Step, at int) int {
	alive := map[int]bool{}
	for _, su := range step.Snap.Units {
		alive[su.ID] = true
	}
	later := map[int]bool{}
	for _, e := range step.Ev[at+1:] {
		if e.K == "mv" || e.K == "ai0" || e.K == "ai1" || e.K == "act" {
			later[e.ID] = true
		}
		if e.B != nil {
			later[e.B.ID] = true
		}
	}
	removed := 0
	for _, u := range append([]*state.Unit(nil), b.Units...) {
		if !alive[u.ID] && !later[u.ID] {
			b.Remove(u.ID)
			removed++
		}
	}
	return removed
}
