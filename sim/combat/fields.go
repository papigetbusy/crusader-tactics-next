package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Trap springs the first hidden snare a monster walked over (old trapCheck,
// line 1126): rooted, poisoned and marked for the Trapper for 3 turns. The
// old game only saw the tile the monster stopped on (unless TrapsOnPath).
func Trap(c *content.Content, b *state.Battle, e *state.Unit, path []board.Tile, emit action.Emit) {
	if len(b.Traps) == 0 || (e.Player() && !e.PvP) {
		return
	}
	if len(path) == 0 || !c.Combat.TrapsOnPath {
		path = []board.Tile{{X: e.X, Y: e.Y}}
	}
	for _, t := range path {
		for i, tr := range b.Traps {
			if tr.X == t.X && tr.Y == t.Y && tr.Team != e.Team {
				b.Traps = append(b.Traps[:i:i], b.Traps[i+1:]...)
				e.Status["snare"] = max(e.Status["snare"], 3)
				e.Status["poison"] = max(e.Status["poison"], 3)
				e.Status["tmark"] = 3
				emit(action.Event{Kind: action.Trapped, Unit: e.ID})
				return
			}
		}
	}
}

// Overwatch is a Trickshot shooting a monster that ends its move inside the
// aimed 4-tile cone, once per monster, at 70% (old owCheck, line 1125).
func Overwatch(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit, emit action.Emit) {
	for _, h := range snapshot(b) {
		if h.Team == e.Team || h.Overwatch == nil || b.Unit(h.ID) == nil || b.Unit(e.ID) == nil {
			continue
		}
		ow := h.Overwatch
		if containsID(ow.Hit, e.ID) {
			continue
		}
		rx, ry := e.X-h.X, e.Y-h.Y
		along := rx*ow.DX + ry*ow.DY
		side := abs(rx*ow.DY - ry*ow.DX)
		if along < 1 || along > 4 || side > (along+1)/2 {
			continue
		}
		ow.Hit = append(ow.Hit, e.ID)
		shot, ok := c.Skills["Strike"]
		for _, s := range h.Skills {
			if !s.Heal && !s.Support && !s.Cloak && s.Summon == nil && !s.Strike {
				shot, ok = s, true
				break
			}
		}
		if !ok {
			continue
		}
		emit(action.Event{Kind: action.OverwatchShot, Unit: h.ID, Target: e.ID})
		Hurt(c, b, dice, Attack{By: h, Skill: shot, Target: e, Extra: 0.7}, emit)
	}
}

// FieldTick is the end of a side's turn for the fields that side made (old
// fieldTick, line 1127, for the heroes): every foe in a burning field or a
// war banner's reach burns for 2 turns, a swamp snares and poisons, and each
// field has one turn less.
func FieldTick(b *state.Battle, side int) {
	for i := range b.Fields {
		f := &b.Fields[i]
		if f.Team != side {
			continue
		}
		for _, z := range b.Units {
			if z.Team == side {
				continue
			}
			d := abs(z.X-f.X) + abs(z.Y-f.Y)
			if f.Square {
				d = max(abs(z.X-f.X), abs(z.Y-f.Y))
			}
			if d > f.Range {
				continue
			}
			switch f.Kind {
			case "static": // shocks at the end of the victim's own turn (Charged)
			case "swamp":
				putStatus(z, "snare", 1, func(action.Event) {})
				z.Status["poison"] = max(z.Status["poison"], 2)
			default:
				z.Status["burn"] = max(z.Status["burn"], 2)
			}
		}
		f.Turns--
	}
	kept := b.Fields[:0]
	for _, f := range b.Fields {
		if f.Turns > 0 || f.Team != side {
			kept = append(kept, f)
		}
	}
	b.Fields = kept
}

func containsID(ids []int, id int) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
