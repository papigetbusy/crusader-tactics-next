package ai

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Scheduled reports a unit whose type learns on the even skill schedule
// (approved 2026-10-10); it plays with the rules in this file.
func Scheduled(c *content.Content, e *state.Unit) bool {
	return len(c.Classes[e.Class].Schedule) > 0
}

// usable reports a skill this monster could pick at all right now: ready,
// and one the monster knows how to aim (it never uses tile, roll, puppet or
// link aims, which no monster type has).
func usable(c *content.Content, b *state.Battle, e *state.Unit, i int) bool {
	s := e.Skills[i]
	if e.Cooldowns[i] != 0 || s.Free {
		return false
	}
	switch s.Aim {
	case content.AimTile, content.AimRoll, content.AimShove, content.AimTwo:
		return false
	case content.AimEcho:
		return e.LastSpell != ""
	}
	if s.Revive {
		return len(b.FallenFoes) > 0 && e.Monster()
	}
	return true
}

// chooseScheduled picks a scheduled monster's skill slot, or -1. A silenced
// monster can only use Strike. Healers bring back the fallen, then
// heal or shield the hurt, sometimes bless, else attack. Bosses and elites
// pick any skill at random; others use a special skill (one with a cooldown)
// 45% of the time, else their first plain attack.
func chooseScheduled(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit) int {
	if e.Status["silence"] > 0 {
		for i, s := range e.Skills {
			if s.Strike && e.Cooldowns[i] == 0 {
				return i
			}
		}
		return -1
	}
	var ready, special, attacks []int
	for i, s := range e.Skills {
		if !usable(c, b, e, i) {
			continue
		}
		ready = append(ready, i)
		if s.Cooldown > 0 || s.Kind == "u" {
			special = append(special, i)
		}
		if !s.Friendly() {
			attacks = append(attacks, i)
		}
	}
	if len(ready) == 0 {
		return -1
	}
	if e.Class == "Healer" {
		hurt, low := false, false
		for _, z := range b.Units {
			if z.Team == e.Team && z != e && float64(z.HP) < float64(z.MaxHP)*c.AI.HealWhenAllyBelow {
				hurt = true
			}
			if z.Team == e.Team && z.HP*2 < z.MaxHP {
				low = true
			}
		}
		for _, i := range ready {
			s := e.Skills[i]
			if s.Revive || (s.Trick == "lastLight" && low) {
				return i
			}
		}
		for j := len(ready) - 1; j >= 0; j-- {
			if s := e.Skills[ready[j]]; hurt && s.Heal && !s.AllAllies && s.MaxRange > 0 {
				return ready[j]
			}
		}
	}
	if e.Boss || e.Elite {
		return ready[rng.IntN(dice, RollPick, len(ready))]
	}
	if len(special) > 0 && dice.Roll(RollSpecial) < c.AI.SpecialSkillChance {
		i := special[rng.IntN(dice, RollPick, len(special))]
		if e.Class != "Healer" || !e.Skills[i].Heal {
			return i
		}
	}
	for _, i := range attacks {
		if e.Skills[i].Cooldown == 0 {
			return i
		}
	}
	if len(attacks) > 0 {
		return attacks[0]
	}
	return ready[0]
}

// scheduledTurn is one turn of a scheduled monster: Steady Aim first if a
// foe is already in reach, then the chosen skill. Line skills are used from
// where it stands if a line full of foes is open; otherwise, and whenever
// the chosen skill finds nobody after walking, it falls back to a plain
// attack. A kill with Momentum gives one more go.
func scheduledTurn(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit, emit action.Emit) error {
	steadyAim(c, b, dice, e, emit)
	slot := chooseScheduled(c, b, dice, e)
	if slot < 0 {
		return nil
	}
	s := e.Skills[slot]
	if s.Aim == content.AimLine {
		if t, ok := bestLine(c, b, e, s); ok {
			return combat.UseAimed(c, b, dice, e, slot, combat.Pick{Tiles: []state.Tile{t}}, emit)
		}
		if slot = plainAttack(e); slot < 0 {
			return nil
		}
		s = e.Skills[slot]
	}
	tn := taunter(b, e, s)
	aimAt := s
	if s.Aim == content.AimSelf || s.Support || s.Revive {
		if p := plainAttack(e); p >= 0 && !s.Around && s.Trick != "taunt" {
			aimAt = e.Skills[p] // buffs: walk as for a plain attack first
		} else {
			aimAt.MinRange, aimAt.MaxRange = 0, max(1, s.Area)
		}
	}
	var path []board.Tile
	if t, ok := ChooseTile(c, b, e, aimAt); ok && !e.Moved {
		path = board.MoveUnit(c, b, e, t.X, t.Y, emit)
	}
	if e.Monster() || e.PvP {
		combat.Trap(c, b, e, path, emit)
		combat.Overwatch(c, b, dice, e, emit)
		combat.LanceWall(c, b, dice, e, path, emit)
		if b.Unit(e.ID) == nil {
			return nil
		}
	}
	v := chooseTarget(c, b, e, s, tn)
	if v == nil && !s.Friendly() {
		if p := plainAttack(e); p >= 0 && p != slot {
			slot, s = p, e.Skills[p]
			v = chooseTarget(c, b, e, s, taunter(b, e, s))
		}
	}
	if v == nil {
		return nil
	}
	e.Moved = true
	if err := combat.UseSkill(c, b, dice, e, slot, v, emit); err != nil {
		return err
	}
	if b.Unit(e.ID) != nil && !e.Done && e.Monster() {
		return scheduledTurn(c, b, dice, e, emit) // Momentum
	}
	return nil
}

// plainAttack is the first ready damaging skill without a cooldown (or any
// ready damaging one), or -1.
func plainAttack(e *state.Unit) int {
	best := -1
	for i, s := range e.Skills {
		if e.Cooldowns[i] != 0 || s.Friendly() || s.Aim != content.AimUnit {
			continue
		}
		if s.Cooldown == 0 {
			return i
		}
		if best < 0 {
			best = i
		}
	}
	return best
}

// steadyAim uses a ready Steady Aim when a foe is already in shooting reach.
func steadyAim(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit, emit action.Emit) {
	for i, s := range e.Skills {
		if !s.Free || !s.Still || e.Cooldowns[i] != 0 || e.Moved || e.Status["silence"] > 0 {
			continue
		}
		if p := plainAttack(e); p >= 0 && len(combat.Targets(c, b, e, e.Skills[p])) > 0 {
			_ = combat.UseSkill(c, b, dice, e, i, e, emit)
		}
	}
}

// bestLine is where a line skill should go from here: the open straight
// line with the most foes on the way, or ok false if none has any.
func bestLine(c *content.Content, b *state.Battle, e *state.Unit, s content.Skill) (state.Tile, bool) {
	best, most := state.Tile{}, 0
	for _, d := range [4][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}} {
		for n := s.MaxRange; n >= 1; n-- {
			t := state.Tile{X: e.X + d[0]*n, Y: e.Y + d[1]*n}
			if !combat.PickOK(c, b, e, s, combat.Pick{Tiles: []state.Tile{t}}) {
				continue
			}
			foes := 0
			for k := 1; k < n; k++ {
				if z := b.UnitAt(e.X+d[0]*k, e.Y+d[1]*k); z != nil && z.Team != e.Team {
					foes++
				}
			}
			if foes > most {
				best, most = t, foes
			}
		}
	}
	return best, most > 0
}
