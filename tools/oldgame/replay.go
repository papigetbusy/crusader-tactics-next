package oldgame

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Moment is where a replay stands when it hands an event to a check.
type Moment struct {
	Fight, Step int // 1-based, as shown in reports
	F           *Fight
	S           *Step
	B           *state.Battle          // the board just before the event
	TilesMoved  map[int]int            // per unit, this turn (for charge skills)
	Uses        map[int]map[string]int // heroes: skill uses before the current action
}

// Walk replays every usable fight. For each event it calls check with the
// board as it was just before the event, then applies what the old game
// recorded (moves, hit results, deaths), so the next event starts from the
// old game's truth. It returns how many fights were skipped.
//
// Some rules are not ported yet and are taken from the recording instead:
// where reinforcements arrive, and poison or bleed deaths at turn start.
func Walk(c *content.Content, fights []Fight, check func(m *Moment, e Event)) (skipped, reinforced, dotDeaths int) {
	for fi := range fights {
		f := &fights[fi]
		if !f.Usable() {
			skipped++
			continue
		}
		moved := map[int]int{}
		uses := map[int]map[string]int{}
		last := map[int]string{}
		actUses := map[int]map[string]int{}
		prev := f.Init.Snap
		for si := range f.Steps {
			step := &f.Steps[si]
			b := f.Battle(prev)
			reinforced += addArrivals(f, b, *step)
			m := &Moment{Fight: fi + 1, Step: si + 1, F: f, S: step, B: b, TilesMoved: moved, Uses: actUses}
			for ei, e := range step.Ev {
				for _, u := range b.Units {
					u.TilesMoved = moved[u.ID]
					u.SkillUses = actUses[u.ID]
				}
				check(m, e)
				switch e.K {
				case "startPhase":
					dotDeaths += removeTurnStartDeaths(b, *step, ei)
					for _, u := range b.Units {
						if u.Team == e.Side {
							moved[u.ID] = 0
						}
					}
				case "mv":
					if u := b.Unit(e.ID); u != nil {
						n := abs(u.X-e.To[0]) + abs(u.Y-e.To[1])
						if p := board.Path(c, b, u, e.To[0], e.To[1]); len(p) >= 2 {
							n = max(n, len(p)-1)
						} else if n > 0 {
							n = max(n, 1)
						}
						moved[u.ID] += n
						u.X, u.Y, u.Moved = e.To[0], e.To[1], true
					}
				case "act":
					if u := b.Unit(e.ID); u != nil && u.Team == state.Heroes {
						if uses[u.ID] == nil {
							uses[u.ID] = map[string]int{}
						}
						actUses[u.ID] = copyUses(uses[u.ID])
						add := 1
						if last[u.ID] != "" && last[u.ID] != e.S {
							add = 2
						}
						uses[u.ID][e.S] += add
						last[u.ID] = e.S
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
			prev = step.Snap
		}
	}
	return
}

func copyUses(m map[string]int) map[string]int {
	out := make(map[string]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
