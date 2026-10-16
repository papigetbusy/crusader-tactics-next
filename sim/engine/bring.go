package engine

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Changing which learned skills a hero carries (the old game's skill library
// on the gear sheet, index.html lines 722-728). Owner's rule (2026-10-11):
// once per floor. One change can swap any number of skills; after it the
// hero's skills are fixed until the next floor. It does not use up the turn.

func (g *Game) checkBring(a action.Action) error {
	u := g.B.Unit(a.Unit)
	if u == nil || u.Team != state.Heroes || u.Companion || u.Decoy || u.Summoned {
		return no("no-such-hero", "That is not one of your heroes.")
	}
	if u.Swapped {
		return no("swapped", "%s has already changed skills on this floor.", u.Class)
	}
	if u.Done {
		return no("finished", "%s has finished for this turn.", u.Class)
	}
	if len(a.Bar) == 0 || len(a.Bar) > 4 {
		return no("bar-size", "Bring 1 to 4 skills.")
	}
	for i, n := range a.Bar {
		s, ok := g.C.SkillNamed(n)
		if !ok || !slices.Contains(u.Learned, n) {
			return no("not-learned", "%s has not learned %s.", u.Class, n)
		}
		if s.Kind == "u" {
			return no("ultimate", "%s is an ultimate; it goes in the ultimate slot.", n)
		}
		if slices.Contains(a.Bar[:i], n) {
			return no("twice", "%s is on the bar twice.", n)
		}
	}
	if a.Choice != "" {
		s, ok := g.C.SkillNamed(a.Choice)
		if !ok || !slices.Contains(u.Learned, a.Choice) || s.Kind != "u" {
			return no("not-learned", "%s has not learned the ultimate %s.", u.Class, a.Choice)
		}
	}
	return nil
}

// bring sets the new bar. A skill kept from the old bar keeps its cooldown;
// a skill brought in starts ready.
func (g *Game) bring(a action.Action, emit action.Emit) {
	u := g.B.Unit(a.Unit)
	wait := map[string]int{}
	for i, s := range u.Skills {
		if i < len(u.Cooldowns) {
			wait[s.Name] = u.Cooldowns[i]
		}
	}
	u.Equipped = slices.Clone(a.Bar)
	if a.Choice != "" {
		u.Ultimate = a.Choice
	}
	u.Skills = g.C.Bar(u.Class, u.Equipped, u.Ultimate)
	u.Cooldowns = make([]int, len(u.Skills))
	for i, s := range u.Skills {
		u.Cooldowns[i] = wait[s.Name]
	}
	u.Swapped = true
	emit(action.Event{Kind: action.SkillBrought, Unit: u.ID})
}

// library is a hero's learned skills that are not on its bar.
func (g *Game) library(u *state.Unit) []LibraryView {
	var out []LibraryView
	for _, n := range u.Learned {
		if slices.Contains(u.Equipped, n) || n == u.Ultimate {
			continue
		}
		if s, ok := g.C.SkillNamed(n); ok {
			out = append(out, LibraryView{Name: n, Text: s.Text, Ultimate: s.Kind == "u", MinRange: s.MinRange, MaxRange: s.MaxRange,
				Power: s.Power, Element: s.Element, Area: s.Area, Heal: s.Heal, Support: s.Support, Revive: s.Revive,
				Cooldown: s.Cooldown, Level: s.Level})
		}
	}
	return out
}
