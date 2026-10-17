package engine

import (
	"encoding/json"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Fog of war (owner, 2026-10-16), versus only: each side sees the enemy units
// within Vision steps of one of its own units. The ground is always seen.
// Players get views, saves and events with what they can't see taken out, and
// can't target what they can't see. Watchers (side -1) see everything.

// Fogged reports whether this game has fog of war now. When the match is
// over, the fog lifts and everyone sees the whole board.
func (g *Game) Fogged() bool {
	return g.B.Versus != nil && g.C.Versus.Fog && g.B.Outcome == state.Ongoing
}

// Sight is every tile a side sees.
func (g *Game) Sight(side int) map[[2]int]bool {
	b, r, out := g.B, g.C.Versus.Vision, map[[2]int]bool{}
	for _, u := range b.Units {
		if u.Team != side || u.HP <= 0 {
			continue
		}
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				x, y := u.X+dx, u.Y+dy
				if abs(dx)+abs(dy) <= r && b.Inside(x, y) {
					out[[2]int{x, y}] = true
				}
			}
		}
	}
	return out
}

// Sees reports whether a side can see unit u: its own units always; enemies
// when there is no fog, or when they stand in its sight. Watchers (-1) see all.
func (g *Game) Sees(side int, u *state.Unit) bool {
	if u == nil || side < 0 || !g.Fogged() || u.Team == side {
		return true
	}
	return g.Sight(side)[[2]int{u.X, u.Y}]
}

// Visible is the IDs of every unit a side sees now.
func (g *Game) Visible(side int) map[int]bool {
	out := map[int]bool{}
	var sight map[[2]int]bool
	if side >= 0 && g.Fogged() {
		sight = g.Sight(side)
	}
	for _, u := range g.B.Units {
		if sight == nil || u.Team == side || sight[[2]int{u.X, u.Y}] {
			out[u.ID] = true
		}
	}
	return out
}

// ViewFor is the View as one side sees it: unseen enemies left out, and Seen
// marks the tiles in sight ('1') row by row, for the screen to shade the rest.
func (g *Game) ViewFor(side int) View {
	v := g.View()
	if side < 0 || !g.Fogged() {
		return v
	}
	vis, sight := g.Visible(side), g.Sight(side)
	units := v.Units[:0:0]
	for _, u := range v.Units {
		if vis[u.ID] {
			units = append(units, u)
		}
	}
	v.Units = units
	v.Seen = make([]string, g.B.N)
	for y := range v.Seen {
		row := make([]byte, g.B.N)
		for x := range row {
			row[x] = '0'
			if sight[[2]int{x, y}] {
				row[x] = '1'
			}
		}
		v.Seen[y] = string(row)
	}
	return v
}

// EventsFor keeps the events a side could see: those about its own units, or
// about enemies it saw before or after (before is Visible from just before the
// action). Events about nobody (turns, the win) are always kept.
func (g *Game) EventsFor(side int, events []action.Event, before map[int]bool) []action.Event {
	if side < 0 || !g.Fogged() {
		return events
	}
	after := g.Visible(side)
	ok := func(id int) bool { return id == 0 || before[id] || after[id] }
	var out []action.Event
	for _, e := range events {
		if ok(e.Unit) && ok(e.Target) {
			out = append(out, e)
		}
	}
	return out
}

// SaveFor is a save of the match without the enemies a side can't see, for
// that player's phone to answer its own questions with. It can't be used to
// peek. Watchers get the whole match.
func (g *Game) SaveFor(side int) ([]byte, error) {
	if side < 0 || !g.Fogged() {
		return g.Save()
	}
	raw, err := json.Marshal(g.B)
	if err != nil {
		return nil, err
	}
	var b state.Battle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	vis := g.Visible(side)
	units := b.Units[:0:0]
	for _, u := range b.Units {
		if vis[u.ID] {
			units = append(units, u)
		}
	}
	b.Units = units
	return (&Game{C: g.C, B: &b}).Save()
}

// unseen refuses an action aimed at a unit the acting side can't see.
func (g *Game) unseen(u *state.Unit, a action.Action) error {
	if !g.Fogged() {
		return nil
	}
	for _, id := range []int{a.Target, a.Second} {
		if t := g.B.Unit(id); id != 0 && t != nil && !g.Sees(u.Team, t) {
			return no("unseen", "You can't see that unit.")
		}
	}
	return nil
}
