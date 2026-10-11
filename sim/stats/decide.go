package stats

import (
	"fmt"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Choices a player can give to a waiting decision.
const (
	ChooseOK      = "ok"      // accept (an evolution, or a skill that goes in by itself)
	ChooseKeep    = "keep"    // keep the current skills; the new one stays in the library
	ChooseReplace = "replace" // replace the skill in slot Slot (0-3), or the ultimate
	ChooseTake    = "take"    // take the new companion spirit instead (the old one leaves)
)

// Options lists what a player may answer to the first waiting decision
// (old showEv, index.html lines 714-726). For an ascension, the options
// are the class's three ascensions.
func Options(c *content.Content, b *state.Battle) []string {
	if len(b.Decisions) == 0 {
		return nil
	}
	d := b.Decisions[0]
	u := heroOf(b, d.Unit)
	if u == nil {
		return []string{ChooseOK}
	}
	switch d.Kind {
	case state.DecideAscend:
		return c.AscensionChoices[u.Class]
	case state.DecideSwap:
		return []string{ChooseKeep, ChooseTake}
	case state.DecideLearn:
		s, _ := c.SkillNamed(d.Skill)
		switch {
		case s.Kind == "u" && (u.Ultimate == "" || u.Ultimate == s.Name):
			return []string{ChooseOK}
		case s.Kind == "u":
			return []string{ChooseReplace, ChooseKeep}
		case slices.Contains(u.Equipped, s.Name) || len(u.Equipped) < 4:
			return []string{ChooseOK}
		default:
			return []string{ChooseReplace, ChooseKeep}
		}
	}
	return []string{ChooseOK}
}

// Decide answers the first waiting decision. slot is the skill slot to
// replace (0-3) when replacing a normal skill.
func Decide(c *content.Content, b *state.Battle, choice string, slot int, emit action.Emit) error {
	if !slices.Contains(Options(c, b), choice) {
		return fmt.Errorf("%q is not one of the choices %v", choice, Options(c, b))
	}
	d := b.Decisions[0]
	b.Decisions = b.Decisions[1:]
	u := heroOf(b, d.Unit)
	if u == nil {
		return nil
	}
	switch d.Kind {
	case state.DecideAscend:
		a := c.Ascensions[choice]
		u.Ascension = choice
		emit(action.Event{Kind: action.Ascended, Unit: u.ID, Skill: choice})
		// Its ultimate is learned next, before anything else waiting.
		b.Decisions = append([]state.Decision{{Kind: state.DecideLearn, Unit: u.ID, Skill: a.Ultimate}}, b.Decisions...)
	case state.DecideLearn:
		s, _ := c.SkillNamed(d.Skill)
		if !slices.Contains(u.Learned, s.Name) {
			u.Learned = append(u.Learned, s.Name)
		}
		switch {
		case s.Kind == "u" && choice != ChooseKeep:
			u.Ultimate = s.Name
		case s.Kind == "u":
		case choice == ChooseOK && !slices.Contains(u.Equipped, s.Name):
			u.Equipped = append(u.Equipped, s.Name)
		case choice == ChooseReplace:
			if slot < 0 || slot >= len(u.Equipped) {
				return fmt.Errorf("there is no skill slot %d", slot)
			}
			u.Equipped[slot] = s.Name
		}
		emit(action.Event{Kind: action.Learned, Unit: u.ID, Skill: s.Name})
	case state.DecideSwap:
		if choice == ChooseTake {
			u.Bond = &state.Bond{Kind: d.Skill, Element: d.Element}
			emit(action.Event{Kind: action.Bonded, Unit: u.ID, Skill: d.Skill})
		}
	}
	Refresh(c, b, u)
	return nil
}

// Refresh brings a hero on the board up to date after a choice: its skill
// bar (cooldowns follow the skills by name), its numbers (extra max HP is
// healed, never below 1 HP), its element and range from the ascension (old
// refreshAll, index.html line 444). A fallen hero is not refreshed.
func Refresh(c *content.Content, b *state.Battle, u *state.Unit) {
	if b.Unit(u.ID) == nil {
		return
	}
	cd := map[string]int{}
	for i, s := range u.Skills {
		if i < len(u.Cooldowns) {
			cd[s.Name] = u.Cooldowns[i]
		}
	}
	u.Skills = c.Bar(u.Class, u.Equipped, u.Ultimate)
	u.Cooldowns = make([]int, len(u.Skills))
	for i, s := range u.Skills {
		u.Cooldowns[i] = cd[s.Name]
	}
	n, base := Geared(c, u.Class, u.Level, u.Ascension, u.Gear)
	d := n.MaxHP - u.MaxHP
	u.MaxHP, u.Atk, u.Def, u.Mov, u.BaseAtk = n.MaxHP, n.Atk, n.Def, n.Mov, base
	u.HP = max(1, min(u.HP+max(0, d), u.MaxHP))
	u.Affix = AffixSum(c, u.Gear)
	a := c.Ascensions[u.Ascension]
	u.Element, u.RangeBonus = a.Element, a.Range
	InitSets(c, u)
}

// heroOf finds a hero on the board or among the fallen.
func heroOf(b *state.Battle, id int) *state.Unit {
	if u := b.Unit(id); u != nil {
		return u
	}
	for _, u := range b.Fallen {
		if u.ID == id {
			return u
		}
	}
	return nil
}
