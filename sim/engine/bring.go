package engine

import (
	"slices"
	"strconv"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Choosing which learned skills a hero brings into a battle (the old game's
// skill library on the gear sheet, index.html lines 722-728). The new game
// allows it while preparing, before any hero has moved or acted.

// Preparing reports a battle that has not started yet: the first turn,
// before any hero has moved or acted.
func (g *Game) Preparing() bool {
	b := g.B
	if b.Turn > 1 || b.Side != state.Heroes || b.Outcome != state.Ongoing {
		return false
	}
	for _, u := range b.Units {
		if u.Team == state.Heroes && (u.Moved || u.Acted || u.Done) {
			return false
		}
	}
	return true
}

func (g *Game) checkBring(a action.Action) error {
	u := g.B.Unit(a.Unit)
	if u == nil || u.Team != state.Heroes || u.Companion || u.Decoy || u.Summoned {
		return no("no-such-hero", "That is not one of your heroes.")
	}
	if !g.Preparing() {
		return no("battle-started", "Skills can only be changed before the battle starts.")
	}
	s, ok := g.C.SkillNamed(a.Choice)
	if !ok || !slices.Contains(u.Learned, a.Choice) {
		return no("not-learned", "%s has not learned %s.", u.Class, a.Choice)
	}
	if s.Kind == "u" {
		if u.Ultimate == a.Choice {
			return no("already-on-bar", "%s is already the ultimate.", a.Choice)
		}
		return nil
	}
	if slices.Contains(u.Equipped, a.Choice) {
		return no("already-on-bar", "%s is already on the skill bar.", a.Choice)
	}
	if a.Slot == "" {
		if len(u.Equipped) >= 4 {
			return no("bar-full", "All 4 skill slots are full. Pick one to replace.")
		}
		return nil
	}
	if i, err := strconv.Atoi(a.Slot); err != nil || i < 1 || i > len(u.Equipped) {
		return no("no-such-slot", "There is no skill slot %s.", a.Slot)
	}
	return nil
}

func (g *Game) bring(a action.Action, emit action.Emit) {
	u := g.B.Unit(a.Unit)
	s, _ := g.C.SkillNamed(a.Choice)
	switch {
	case s.Kind == "u":
		u.Ultimate = a.Choice
	case a.Slot == "":
		u.Equipped = append(u.Equipped, a.Choice)
	default:
		i, _ := strconv.Atoi(a.Slot)
		u.Equipped[i-1] = a.Choice
	}
	u.Skills = g.C.Bar(u.Class, u.Equipped, u.Ultimate)
	u.Cooldowns = make([]int, len(u.Skills))
	emit(action.Event{Kind: action.SkillBrought, Unit: u.ID, Skill: a.Choice})
}

// library is a hero's learned skills that are not on its bar.
func (g *Game) library(u *state.Unit) []LibraryView {
	var out []LibraryView
	for _, n := range u.Learned {
		if slices.Contains(u.Equipped, n) || n == u.Ultimate {
			continue
		}
		if s, ok := g.C.SkillNamed(n); ok {
			out = append(out, LibraryView{Name: n, Text: s.Text, Ultimate: s.Kind == "u"})
		}
	}
	return out
}
