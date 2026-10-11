package engine

import (
	"encoding/json"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/data"
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// TestPracticeFights plays every practice fight to the end with real dice:
// each hero walks toward the nearest enemy and hits whatever is in range.
func practiceSetups(t *testing.T) []Setup {
	var p struct {
		Practice []Setup `json:"practice"`
	}
	raw, err := data.Files.ReadFile("practice.json")
	if err != nil || json.Unmarshal(raw, &p) != nil || len(p.Practice) == 0 {
		t.Fatal("practice.json is missing or empty")
	}
	return p.Practice
}

func TestPracticeFights(t *testing.T) {
	p := struct{ Practice []Setup }{practiceSetups(t)}
	c := content.MustDefault()
	t.Run("Every practice fight can be played to the end without a hitch", func(t *testing.T) {
		for i, s := range p.Practice {
			for seed := uint64(1); seed <= 5; seed++ {
				g := NewFight(c, s, seed)
				for turns := 0; turns < 80 && g.B.Outcome == state.Ongoing; turns++ {
					playHeroes(t, g)
					if g.B.Outcome != state.Ongoing {
						break
					}
					if _, err := g.Apply(action.Action{Kind: action.EndTurn}); err != nil {
						t.Fatalf("fight %d seed %d: end turn: %v", i+1, seed, err)
					}
				}
			}
		}
	})
}

// answerAll answers every waiting choice with its first option, replacing
// the first skill slot when asked to replace.
func answerAll(t *testing.T, g *Game) {
	for len(g.B.Decisions) > 0 {
		opts := g.Options()
		if _, err := g.Apply(action.Action{Kind: action.Decide, Choice: opts[0]}); err != nil {
			t.Fatalf("decide %v: %v", opts, err)
		}
	}
}

func playHeroes(t *testing.T, g *Game) {
	answerAll(t, g)
	for _, u := range append([]*state.Unit(nil), g.B.Units...) {
		answerAll(t, g)
		if u.Team != state.Heroes || u.Companion || g.B.Unit(u.ID) == nil || g.B.Outcome != state.Ongoing {
			continue
		}
		if !tryAttack(t, g, u) {
			if r := g.Reach(u.ID); len(r) > 0 {
				best, bestD := r[0], 1<<30
				for _, tile := range r {
					for _, e := range g.B.Units {
						if e.Team == state.Enemies {
							if d := abs(e.X-tile.X) + abs(e.Y-tile.Y); d < bestD {
								best, bestD = tile, d
							}
						}
					}
				}
				if _, err := g.Apply(action.Action{Kind: action.Move, Unit: u.ID, X: best.X, Y: best.Y}); err != nil {
					t.Fatalf("move: %v", err)
				}
				answerAll(t, g)
				tryAttack(t, g, u)
			}
		}
	}
}

func tryAttack(t *testing.T, g *Game, u *state.Unit) bool {
	if u.Done {
		return false
	}
	for slot, s := range u.Skills {
		if s.Heal || s.Support || s.Revive || s.Summon != nil || s.Friendly() {
			continue
		}
		switch s.Aim { // skills aimed in more than one tap are tried in TestNewSkillsInFullFights
		case content.AimTile, content.AimLine, content.AimRoll, content.AimShove, content.AimTwo, content.AimSelf:
			continue
		}
		if ts := g.Targets(u.ID, slot); len(ts) > 0 {
			if _, err := g.Apply(action.Action{Kind: action.Attack, Unit: u.ID, Skill: slot, Target: ts[0]}); err != nil {
				t.Fatalf("attack: %v", err)
			}
			return true
		}
	}
	return false
}
