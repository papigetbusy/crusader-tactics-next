// Package turn runs the rhythm of a battle: the start and end of each side's
// turn, the enemy phase, and win or loss (old startPhase, endPhase,
// enemyPhase and check, index.html lines 1735-1752).
package turn

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/ai"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func members(b *state.Battle, side int) []*state.Unit {
	var out []*state.Unit
	for _, u := range b.Units {
		if u.Team == side {
			out = append(out, u)
		}
	}
	return out
}

// Start begins a side's turn (old startPhase, line 1738): cooldowns and the
// class ability tick down, burn, bleed (doubled by Blood Pact, which heals the
// Blood Witch) and poison hurt, forts heal, flags reset, stunned units,
// frozen units and clones skip, a Sacred Vow heals. Units brought to 0 HP
// die; skeletons and clones burst.
func Start(c *content.Content, b *state.Battle, dice rng.Dice, side int, emit action.Emit) {
	t := c.Turns
	dot := func(u *state.Unit, status string) int {
		if u.Status[status] <= 0 {
			return 0
		}
		share := float64(float64(u.MaxHP) * t.DamageOverTime[status])
		if status == "bleed" && u.HeavyBleed {
			share *= 2
		}
		n := max(t.DamageOverTimeMin, int(combat.Round(share)))
		u.HP -= n
		u.Status[status]--
		emit(action.Event{Kind: action.Ticked, Unit: u.ID, Amount: n, Statuses: map[string]int{status: u.Status[status]}})
		return n
	}
	b.Side = side
	emit(action.Event{Kind: action.TurnStart, Side: side, Turn: b.Turn})
	combat.Kegs(c, b, dice, side, emit)
	for _, u := range members(b, side) {
		u.MomentumUsed, u.MoveLeft = false, 0
		if u.Quake > 0 { // Aftershock: the ground shakes again
			u.HP -= u.Quake
			emit(action.Event{Kind: action.Ticked, Unit: u.ID, Amount: u.Quake, Statuses: map[string]int{"quake": 0}})
			u.Quake = 0
		}
		for i, cd := range u.Cooldowns {
			if cd > 0 && cd < 99 {
				u.Cooldowns[i] = cd - 1
			}
		}
		if u.AbilityCooldown > 0 {
			u.AbilityCooldown--
		}
		u.QuickShots, u.QuickUsed = 0, nil
		dot(u, "burn")
		u.Defending, u.Done, u.Moved, u.Acted, u.Danced, u.TilesMoved, u.Undo = false, false, false, false, false, 0, nil
		if bd := dot(u, "bleed"); bd > 0 && u.HeavyBleed {
			if w := b.Unit(u.HeavyBleedBy); w != nil {
				w.HP = min(w.MaxHP, w.HP+bd)
			}
			if u.Status["bleed"] == 0 {
				u.HeavyBleed = false
			}
		}
		dot(u, "poison")
		if tr, ok := c.Terrain[board.GroundAt(b, u.X, u.Y)]; ok && tr.Heal > 0 && u.HP > 0 && u.HP < u.MaxHP {
			u.HP = min(u.MaxHP, u.HP+int(combat.Round(float64(u.MaxHP)*tr.Heal)))
		}
		if u.Status["regen"] > 0 && u.HP > 0 {
			u.HP = min(u.MaxHP, u.HP+int(combat.Round(float64(u.MaxHP)*t.RegenShare)))
		}
		for _, q := range t.Stunned {
			if u.Status[q] > 0 {
				u.Done = true
			}
		}
		if u.Decoy {
			u.Done = true
		}
		if u.Status["sky"] <= 0 {
			u.SkyReady = false
		}
		u.Overwatch = nil
		if u.Status["svow"] > 0 {
			u.HP = min(u.MaxHP, u.HP+int(combat.Round(float64(u.MaxHP)*0.1)))
		}
	}
	for _, u := range slices.Clone(b.Units) {
		if u.HP <= 0 && (u.Undead || u.Decoy) {
			combat.Burst(c, b, dice, u, emit)
		}
	}
	for _, u := range slices.Clone(b.Units) {
		if u.HP <= 0 {
			b.Remove(u.ID)
			if c.Combat.ReviveAnyDeath && u.Team == state.Heroes && !u.Summoned {
				b.Fallen = append(b.Fallen, u)
			}
			if c.Combat.MonstersRevive && u.Team == state.Enemies && !u.Summoned {
				b.FallenFoes = append(b.FallenFoes, u)
			}
			emit(action.Event{Kind: action.Died, Unit: u.ID})
		}
	}
	Check(b, emit)
}

// End finishes a side's turn (old endPhase, line 1739): timed statuses count
// down by one, a finished cloak starts its cooldown; after the heroes' turn,
// fields burn and tick down, and expired shadow clones fade.
func End(c *content.Content, b *state.Battle, dice rng.Dice, side int, emit action.Emit) {
	combat.WyrmClaws(c, b, dice, side, emit)
	for _, u := range members(b, side) {
		for _, q := range c.Turns.CountdownAtEnd {
			if u.Status[q] > 0 {
				u.Status[q]--
			}
		}
		if u.CloakUsed && u.Status["cloak"] <= 0 {
			combat.CloakCooldown(u)
		}
	}
	combat.Charged(b, side, emit)
	combat.FieldTick(b, side)
	if side == state.Heroes {
		for _, u := range slices.Clone(b.Units) {
			if u.Decoy && u.Status["clone"] <= 0 {
				b.Remove(u.ID)
				emit(action.Event{Kind: action.Faded, Unit: u.ID})
			}
		}
	}
	emit(action.Event{Kind: action.TurnEnd, Side: side})
}

// Check decides the battle once one side is gone (old check, line 1735):
// won when no monster is left, lost when no real hero is (companions,
// summons and clones do not count).
func Check(b *state.Battle, emit action.Emit) {
	if b.Outcome != state.Ongoing {
		return
	}
	heroes := 0
	for _, u := range members(b, state.Heroes) {
		if !u.Companion && !u.Summoned && !u.Decoy {
			heroes++
		}
	}
	if len(members(b, state.Enemies)) == 0 {
		b.Outcome = state.Won
		emit(action.Event{Kind: action.Won})
	} else if heroes == 0 {
		b.Outcome = state.Lost
		emit(action.Event{Kind: action.Lost})
	}
}

// EndTurn is the player ending their turn: the heroes' turn ends, every
// monster acts in board order, then a new turn starts for the heroes.
func EndTurn(c *content.Content, b *state.Battle, dice rng.Dice, emit action.Emit) error {
	End(c, b, dice, state.Heroes, emit)
	// Companions (bonded monsters and raised skeletons) act on their own first
	// (old enemyPhase, line 1747).
	for _, z := range members(b, state.Heroes) {
		if !z.Companion || b.Unit(z.ID) == nil || z.Status["stun"] > 0 {
			continue
		}
		if err := ai.Turn(c, b, dice, z, emit); err != nil {
			return err
		}
		combat.FlushRaises(c, b, emit)
		Check(b, emit)
		if len(members(b, state.Enemies)) == 0 {
			return nil
		}
	}
	Start(c, b, dice, state.Enemies, emit)
	for _, e := range members(b, state.Enemies) {
		if b.Unit(e.ID) == nil || e.Done {
			continue
		}
		if err := ai.Turn(c, b, dice, e, emit); err != nil {
			return err
		}
		combat.FlushRaises(c, b, emit)
		Check(b, emit)
		if b.Outcome == state.Lost {
			return nil
		}
	}
	End(c, b, dice, state.Enemies, emit)
	b.Turn++
	Start(c, b, dice, state.Heroes, emit)
	return nil
}
