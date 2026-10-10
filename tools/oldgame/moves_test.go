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

// problems collects the first few mismatches as plain sentences.
type problems []string

func (p *problems) add(format string, a ...any) {
	if len(*p) < 25 {
		*p = append(*p, fmt.Sprintf(format, a...))
	}
}

// TestOldGameMoves checks every move the old game made (hero moves,
// auto-walks before attacks, and every enemy move) against Go's movement
// rules, at the exact moment it happened.
func TestOldGameMoves(t *testing.T) {
	fights, err := Load(recordings)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var checked, heroes, enemies int
	var bad problems
	skipped, reinforced, dot := Walk(c, fights, func(m *Moment, e Event) {
		if e.K != "mv" {
			return
		}
		u := m.B.Unit(e.ID)
		if u == nil {
			bad.add("fight %d step %d: unit %d moved but is not on the board", m.Fight, m.Step, e.ID)
			return
		}
		if u.X != e.From[0] || u.Y != e.From[1] {
			bad.add("fight %d step %d: %s %d was at %d,%d in Go but the old game moved it from %d,%d",
				m.Fight, m.Step, u.Class, u.ID, u.X, u.Y, e.From[0], e.From[1])
		}
		if u.Team == state.Heroes && !board.CanMove(u) {
			bad.add("fight %d step %d: hero %s moved although Go says it cannot move", m.Fight, m.Step, u.Class)
		}
		if !board.CanReach(c, m.B, u, e.To[0], e.To[1]) {
			bad.add("fight %d step %d: the old game moved %s %d from %d,%d to %d,%d, but Go says that tile is out of reach",
				m.Fight, m.Step, u.Class, u.ID, u.X, u.Y, e.To[0], e.To[1])
		}
		checked++
		if u.Team == state.Heroes {
			heroes++
		} else {
			enemies++
		}
	})
	t.Logf("checked %d recorded moves (%d hero, %d enemy) across %d fights; %d reinforcements placed and %d poison/bleed deaths taken from the recording; skipped %d fights where the old game crashed",
		checked, heroes, enemies, len(fights)-skipped, reinforced, dot, skipped)
	if checked < 6000 {
		t.Errorf("expected about 6,600 recorded moves, checked only %d", checked)
	}
	for _, p := range bad {
		t.Error(p)
	}
}
