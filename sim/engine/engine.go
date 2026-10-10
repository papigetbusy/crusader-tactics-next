// Package engine is the referee's front door. The screen (and later, an
// online server) only ever calls Apply with an action, and gets back either
// the events that happened or a plain "no, because...". A refused action
// changes nothing.
package engine

import (
	"fmt"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/turn"
)

// Game is one battle in progress.
type Game struct {
	C    *content.Content
	B    *state.Battle
	Dice rng.Dice
}

// Refused explains why an action was not allowed.
type Refused struct {
	Code   string `json:"code"`   // short code for the screen: "too-far", "on-cooldown"...
	Reason string `json:"reason"` // plain words
}

func (r *Refused) Error() string { return r.Reason }

func no(code, format string, a ...any) error {
	return &Refused{Code: code, Reason: fmt.Sprintf(format, a...)}
}

// Apply performs one action and returns what happened.
func (g *Game) Apply(a action.Action) ([]action.Event, error) {
	var events []action.Event
	emit := func(e action.Event) { events = append(events, e) }
	if err := g.check(a); err != nil {
		return nil, err
	}
	b := g.B
	switch a.Kind {
	case action.Move:
		u := b.Unit(a.Unit)
		from := state.Tile{X: u.X, Y: u.Y}
		board.MoveUnit(g.C, b, u, a.X, a.Y, emit)
		u.Moved, u.Undo = true, &from
		if u.Acted { // a hero that already acted (Hit and Run) is finished after moving
			u.Done, u.Undo = true, nil
		}
		seizeThrone(b, u, emit)
	case action.Defend:
		u := b.Unit(a.Unit)
		u.Defending, u.Done = true, true
		emit(action.Event{Kind: action.Defended, Unit: u.ID})
	case action.Potion:
		u := b.Unit(a.Unit)
		n := min(u.MaxHP-u.HP, g.C.Combat.PotionBase+g.C.Combat.PotionPerLevel*u.Level)
		u.HP += n
		u.Potions--
		u.Done = true
		emit(action.Event{Kind: action.Drank, Unit: u.ID, Amount: n})
	case action.Undo:
		u := b.Unit(a.Unit)
		to := *u.Undo
		u.Undo = nil
		board.MoveUnit(g.C, b, u, to.X, to.Y, func(e action.Event) { e.Kind = action.Undone; emit(e) })
		u.Moved, u.TilesMoved = false, 0
	case action.Attack:
		u := b.Unit(a.Unit)
		if err := combat.UseSkill(g.C, b, g.Dice, u, a.Skill, b.Unit(a.Target), emit); err != nil {
			return events, err
		}
		turn.Check(b, emit)
	case action.Wait:
		b.Unit(a.Unit).Done = true
		emit(action.Event{Kind: action.Finished, Unit: a.Unit})
	case action.EndTurn:
		if err := turn.EndTurn(g.C, b, g.Dice, emit); err != nil {
			return events, err
		}
	}
	return events, nil
}

// seizeThrone is the seize-the-throne win: a hero who reaches the throne
// makes every enemy flee (old click handler, index.html line 1761).
func seizeThrone(b *state.Battle, u *state.Unit, emit action.Emit) {
	if b.Throne == nil || u.X != b.Throne.X || u.Y != b.Throne.Y || u.Summoned {
		return
	}
	emit(action.Event{Kind: action.Seized, Unit: u.ID})
	for _, e := range append([]*state.Unit(nil), b.Units...) {
		if e.Team != state.Heroes {
			b.Remove(e.ID)
		}
	}
	turn.Check(b, emit)
}

// check is every "is this allowed?" rule in one place.
func (g *Game) check(a action.Action) error {
	b := g.B
	if b.Outcome != state.Ongoing {
		return no("battle-over", "The battle is over.")
	}
	if b.Side != state.Heroes {
		return no("not-your-turn", "It is not your turn.")
	}
	if a.Kind == action.EndTurn {
		return nil
	}
	u := b.Unit(a.Unit)
	if u == nil || u.Team != state.Heroes {
		return no("no-such-hero", "That is not one of your heroes.")
	}
	if u.Done {
		return no("finished", "%s has finished for this turn.", u.Class)
	}
	switch a.Kind {
	case action.Move:
		if !board.CanMove(u) {
			return no("cannot-move", "%s cannot move now.", u.Class)
		}
		if !board.CanReach(g.C, b, u, a.X, a.Y) {
			return no("too-far", "%s cannot reach that tile.", u.Class)
		}
	case action.Attack:
		if a.Skill < 0 || a.Skill >= len(u.Skills) {
			return no("no-such-skill", "%s has no skill in that slot.", u.Class)
		}
		if u.Cooldowns[a.Skill] > 0 {
			return no("on-cooldown", "%s is not ready yet (%d turns).", u.Skills[a.Skill].Name, u.Cooldowns[a.Skill])
		}
		v := b.Unit(a.Target)
		ok := false
		for _, t := range combat.Targets(g.C, b, u, u.Skills[a.Skill]) {
			ok = ok || t == v
		}
		if !ok {
			return no("bad-target", "%s cannot reach that target.", u.Skills[a.Skill].Name)
		}
	case action.Wait, action.Defend:
	case action.Potion:
		if u.Potions <= 0 {
			return no("no-potions", "%s has no potions left.", u.Class)
		}
		if u.HP >= u.MaxHP {
			return no("full-hp", "%s is already at full health.", u.Class)
		}
	case action.Undo:
		if u.Undo == nil {
			return no("nothing-to-undo", "%s has no move to undo.", u.Class)
		}
		if o := b.UnitAt(u.Undo.X, u.Undo.Y); o != nil && o != u {
			return no("tile-taken", "Your old tile is occupied now.")
		}
	default:
		return no("unknown", "Unknown action %q.", a.Kind)
	}
	return nil
}
