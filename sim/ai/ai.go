// Package ai is how monsters (and later, companions) decide what to do. It
// copies the old game's aiTurn (index.html lines 1740-1746). A monster
// decides exactly like a player: it picks a skill, walks, and attacks, using
// the same rules and the same dice as everyone else.
package ai

import (
	"math"
	"slices"
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Roll labels for the AI's own dice.
const (
	RollPick    = "AI skill pick"
	RollSpecial = "AI special chance"
)

// ChooseSkill picks which skill slot a monster uses this turn, or -1.
// Bosses and elites pick any ready skill at random. Healers heal when an
// ally is hurt. Others use a special skill 45% of the time, else their basic
// attack.
func ChooseSkill(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit) int {
	base := max(1, len(c.Classes[e.Class].Skills))
	var ready, special, heals, attacks []int
	for i, s := range e.Skills {
		if e.Cooldowns[i] != 0 {
			continue
		}
		ready = append(ready, i)
		if i >= base {
			special = append(special, i)
		}
		if s.Heal && !s.AllAllies && s.MaxRange > 0 {
			heals = append(heals, i)
		}
		if !s.Heal && !s.Support && s.Summon == nil {
			attacks = append(attacks, i)
		}
	}
	allyHurt := false
	for _, z := range b.Units {
		if z.Team == e.Team && z != e && float64(z.HP) < float64(z.MaxHP)*c.AI.HealWhenAllyBelow {
			allyHurt = true
		}
	}
	at := func(list []int, i int) int {
		if i >= 0 && i < len(list) {
			return list[i]
		}
		return -1
	}
	slot := func(i int) int {
		if i < len(e.Skills) {
			return i
		}
		return -1
	}
	switch {
	case e.Boss || e.Elite:
		return at(ready, rng.IntN(dice, RollPick, len(ready)))
	case e.Class == "Healer":
		if allyHurt && len(heals) > 0 {
			return heals[len(heals)-1]
		}
		if len(attacks) > 0 {
			return attacks[len(attacks)-1]
		}
		return slot(1)
	case len(special) > 0 && dice.Roll(RollSpecial) < c.AI.SpecialSkillChance && !e.Skills[special[0]].Heal:
		return special[rng.IntN(dice, RollPick, len(special))]
	case len(attacks) > 0:
		return attacks[0]
	}
	return slot(0)
}

// foes are who the chosen skill is aimed at: hurt allies for a heal,
// otherwise every visible opponent. A monster attacking heroes must go for
// a taunting Paladin within 4 tiles, or else for shadow clones within 8
// (old aiTurn, line 1743).
func foes(b *state.Battle, e *state.Unit, s content.Skill) []*state.Unit {
	if tn := taunter(b, e, s); tn != nil {
		return []*state.Unit{tn}
	}
	if dc := decoys(b, e, s); len(dc) > 0 {
		return dc
	}
	var out []*state.Unit
	for _, v := range b.Units {
		if s.Heal {
			if v.Team == e.Team && v != e && v.HP < v.MaxHP {
				out = append(out, v)
			}
		} else if v.Team != e.Team && !(v.Team == state.Heroes && (v.Status["cloak"] > 0 || v.Status["blood"] > 0 || v.Status["sky"] > 0)) {
			out = append(out, v)
		}
	}
	return out
}

// ChooseTile scores every tile the monster can reach and returns the best:
// inside the skill's range if possible, ranged units as far back as their
// range allows, melee units as close as possible. Ties go to the first tile
// in board order. ok is false when the monster should not move.
func ChooseTile(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill) (board.Tile, bool) {
	fs := foes(b, e, s)
	if len(fs) == 0 || e.Status["snare"] > 0 || e.Status["lock"] > 0 || e.Status["shock"] > 0 {
		return board.Tile{}, false
	}
	goals := make([]board.Tile, len(fs))
	for i, f := range fs {
		goals[i] = board.Tile{X: f.X, Y: f.Y}
	}
	walk := board.Distances(c, b, e, goals)
	best, bestScore, found := board.Tile{}, 1e9, false
	for _, t := range board.Reach(c, b, e) {
		md := math.MaxInt
		for _, f := range fs {
			md = min(md, abs(f.X-t.X)+abs(f.Y-t.Y))
		}
		var pen float64
		switch {
		case md < s.MinRange:
			pen = float64(s.MinRange-md) + c.AI.TooClosePenalty
		case md > s.MaxRange:
			pen = math.Max(float64(md), walk(t.X, t.Y)) - float64(s.MaxRange)
		}
		spread := float64(md)
		if s.MaxRange > 1 {
			spread = -spread
		}
		pen += float64(spread * c.AI.SpreadWeight) // float64() keeps Go from fusing the multiply and add
		if pen < bestScore {
			best, bestScore, found = t, pen, true
		}
	}
	return best, found
}

// taunter is a hero within 4 tiles whose Oath of Protection forces attacks.
func taunter(b *state.Battle, e *state.Unit, s content.Skill) *state.Unit {
	if t := b.Unit(e.TauntedBy); e.Status["taunted"] > 0 && t != nil && !s.Friendly() {
		return t // Taunting Roar
	}
	if s.Heal || e.Team == state.Heroes {
		return nil
	}
	for _, h := range b.Units {
		if h.Team == state.Heroes && h.Status["taunt"] > 0 && abs(h.X-e.X)+abs(h.Y-e.Y) <= 4 {
			return h
		}
	}
	return nil
}

// decoys are shadow clones within 8 tiles.
func decoys(b *state.Battle, e *state.Unit, s content.Skill) []*state.Unit {
	if s.Heal || e.Team == state.Heroes {
		return nil
	}
	var out []*state.Unit
	for _, h := range b.Units {
		if h.Decoy && abs(h.X-e.X)+abs(h.Y-e.Y) <= 8 {
			out = append(out, h)
		}
	}
	return out
}

// ChooseTarget picks who to hit among those in range: a taunting hero, then
// a shadow clone, then the weakest (lowest HP; ties go to the first in board
// order). nil if nobody is in range.
func ChooseTarget(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill) *state.Unit {
	return chooseTarget(c, b, e, s, taunter(b, e, s))
}

// chooseTarget is ChooseTarget with the taunting hero found before walking.
func chooseTarget(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill, tn *state.Unit) *state.Unit {
	cands := combat.Targets(c, b, e, s)
	if tn != nil && slices.Contains(cands, tn) {
		return tn
	}
	for _, z := range cands {
		if z.Decoy {
			return z
		}
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].HP < cands[j].HP })
	if len(cands) == 0 {
		return nil
	}
	return cands[0]
}

// Turn plays one monster's turn: choose a skill, walk, use it.
func Turn(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit, emit action.Emit) error {
	if Scheduled(c, e) {
		return scheduledTurn(c, b, dice, e, emit)
	}
	slot := ChooseSkill(c, b, dice, e)
	if slot < 0 {
		return nil
	}
	s := e.Skills[slot]
	tn := taunter(b, e, s) // decided before walking, as in the old game
	var path []board.Tile
	if t, ok := ChooseTile(c, b, e, s); ok {
		path = board.MoveUnit(c, b, e, t.X, t.Y, emit)
	}
	if e.Team == state.Enemies {
		combat.Trap(c, b, e, path, emit)
		combat.Overwatch(c, b, dice, e, emit)
		combat.LanceWall(c, b, dice, e, path, emit)
		if b.Unit(e.ID) == nil {
			return nil
		}
	}
	if v := chooseTarget(c, b, e, s, tn); v != nil {
		return combat.UseSkill(c, b, dice, e, slot, v, emit)
	}
	return nil
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
