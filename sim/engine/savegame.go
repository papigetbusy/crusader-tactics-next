package engine

import (
	"fmt"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/run"
	"github.com/papigetbusy/crusader-tactics-next/sim/save"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// saved is everything a game needs to resume exactly where it was: the
// climb, and the battle on the board (mid-fight included). The dice are not
// saved: a loaded game rolls with a fresh seed (decision of 2026-10-09).
type saved struct {
	Run         *run.Run          `json:"run,omitempty"`
	Phase       string            `json:"phase,omitempty"`
	Team        []string          `json:"team,omitempty"`
	Placed      []run.Placed      `json:"placed,omitempty"`
	StartChests []state.Chest     `json:"startChests,omitempty"`
	Before      map[string][2]int `json:"before,omitempty"`
	GoldBefore  int               `json:"goldBefore,omitempty"`
	BagBefore   int               `json:"bagBefore,omitempty"`
	Summary     *Summary          `json:"summary,omitempty"`
	Battle      *state.Battle     `json:"battle"`
}

// Save is the whole game as a versioned save.
func (g *Game) Save() ([]byte, error) {
	s := saved{Battle: g.B}
	kind := "fight"
	if cl := g.climb; cl != nil {
		kind = "climb"
		s.Run, s.Phase, s.Team, s.Placed, s.StartChests = cl.run, cl.phase, cl.team, cl.placed, cl.startChests
		s.Before, s.GoldBefore, s.BagBefore, s.Summary = cl.before, cl.goldBefore, cl.bagBefore, cl.summary
	}
	return save.Wrap(kind, s)
}

// Load resumes a saved game with fresh dice.
func Load(c *content.Content, data []byte, battleSeed uint64) (*Game, error) {
	var s saved
	kind, err := save.Open(data, &s)
	if err != nil {
		return nil, err
	}
	if s.Battle == nil {
		return nil, fmt.Errorf("this save has no battle in it")
	}
	g := &Game{C: c, B: s.Battle, Dice: rng.New(battleSeed, streamCombat)}
	if kind == "climb" {
		if s.Run == nil {
			return nil, fmt.Errorf("this climb save has no climb in it")
		}
		cl := &climb{run: s.Run, phase: s.Phase, team: s.Team, placed: s.Placed, startChests: s.StartChests,
			before: s.Before, goldBefore: s.GoldBefore, bagBefore: s.BagBefore, summary: s.Summary, battleSeed: battleSeed}
		cl.fl = cl.run.Build(c, cl.run.Floor)
		g.climb = cl
	}
	return g, nil
}
