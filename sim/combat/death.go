package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// died is everything a death sets off (old hurt line 1705-1707 and onDeath,
// line 1208): the unit leaves the board, a hero joins the fallen, clones and
// skeletons burst, Grave Bloom raises the dead, a Death Mark spreads poison,
// kills feed Blood Frenzy, Bloodlust and Oath of Ruin, a Black Knight
// terrifies, a Wanted Poster pays out, and the killer's side gains XP.
func died(c *content.Content, b *state.Battle, dice rng.Dice, u, v *state.Unit, emit action.Emit) {
	b.Remove(v.ID)
	emit(action.Event{Kind: action.Died, Unit: v.ID})
	if v.Team == state.Heroes && !v.Companion && !v.Summoned && !v.Decoy {
		b.Fallen = append(b.Fallen, v)
	}
	if v.Team == state.Enemies && c.Combat.MonstersRevive && !v.Summoned && !v.Decoy {
		b.FallenFoes = append(b.FallenFoes, v)
	}
	onDeath(c, b, dice, u, v, emit)
	if u.Team == state.Heroes {
		u.KilledThisAction = true
	}
	if v.Team == state.Enemies && isHero(u) && u.Ascension == "Necromancer" && v.Kind != "" && !v.Ally && !slices.Contains(u.Souls, v.Kind) {
		// Soul Harvest: each new kind of monster becomes a raise skill (old hurt, line 1706).
		u.Souls = append(u.Souls, v.Kind)
		b.Decisions = append(b.Decisions, state.Decision{Kind: state.DecideLearn, Unit: u.ID, Skill: content.RaisePrefix + v.Kind})
	}
	if u.Ascension == "Black Knight" {
		for _, z := range b.Units {
			if z.Team != u.Team && Distance(z, v) <= 1 {
				z.Status["stun"] = max(z.Status["stun"], 1)
			}
		}
	}
	if v.Team == state.Enemies && u.Team == state.Heroes && v.Wanted && v.WantedBy == u.ID {
		u.AbilityCooldown = 0 // the guaranteed item arrives with loot (step 6d)
	}
	if u.Team == state.Heroes {
		stats.Share(c, b, u, stats.KillXP(c, v), emit)
		if u.Ascension == "Berserker" {
			heal(u, float64(u.MaxHP)*c.Ascensions["Berserker"].N("killHeal"))
		}
	}
}

// onDeath is the old onDeath (index.html line 1208).
func onDeath(c *content.Content, b *state.Battle, dice rng.Dice, u, v *state.Unit, emit action.Emit) {
	if v.Undead || v.Decoy {
		explode(c, b, dice, v, emit)
	}
	if v.Team == state.Enemies {
		if nb := gravebloomer(b); nb != nil {
			b.Raises = append(b.Raises, state.Raise{By: nb.ID, X: v.X, Y: v.Y})
		} else {
			b.FloorKills = append(b.FloorKills, state.Tile{X: v.X, Y: v.Y})
		}
		if v.Status["dmark"] > 0 {
			for _, z := range b.Units {
				if z.Team == state.Enemies && z != v && Distance(z, v) <= 2 {
					z.Status["poison"] = max(z.Status["poison"], 3)
				}
			}
		}
	}
	if u != nil && u.Team == state.Heroes && v.Team == state.Enemies {
		if u.Status["frenzy"] > 0 {
			u.Status["frenzy"]++
		}
		if u.Status["blood"] > 0 {
			if u.P("blR", 0) != 0 {
				if i := slices.IndexFunc(u.Skills, func(s content.Skill) bool { return s.Cloak }); i >= 0 {
					u.Cooldowns[i] = 0
				}
			}
			if u.P("blX", 0) != 0 {
				u.Status["blood"]++
			}
		}
		if u.Status["ruin"] > 0 {
			u.ExtraTurn = true
			for _, z := range b.Units {
				if z.Team == state.Enemies && z != v && Distance(z, v) <= 2 {
					z.Status["stun"] = max(z.Status["stun"], 1)
				}
			}
		}
	}
}

// gravebloomer is a Necromancer whose Grave Bloom is raising every kill.
func gravebloomer(b *state.Battle) *state.Unit {
	for _, z := range b.Units {
		if z.Team == state.Heroes && z.Ascension == "Necromancer" && z.Status["gbloom"] > 0 {
			return z
		}
	}
	return nil
}

// Burst is a skeleton or clone destroyed by poison or bleeding bursting at
// the start of a turn.
func Burst(c *content.Content, b *state.Battle, dice rng.Dice, v *state.Unit, emit action.Emit) {
	explode(c, b, dice, v, emit)
}

// explode is a destroyed skeleton or shadow clone bursting (old explodeAt,
// line 1123): a skeleton poisons and entangles its neighbours; a clone
// deals 120% of its hero's attack to them.
func explode(c *content.Content, b *state.Battle, dice rng.Dice, v *state.Unit, emit action.Emit) {
	emit(action.Event{Kind: action.Exploded, Unit: v.ID})
	owner := b.Unit(v.Owner)
	for _, z := range slices.Clone(b.Units) {
		if z.Team == v.Team || Distance(z, v) > 1 || b.Unit(z.ID) == nil {
			continue
		}
		if v.Undead {
			z.Status["poison"] = max(z.Status["poison"], 3)
			z.Status["ivy"] = max(z.Status["ivy"], 2)
			setIvy(z, 0.3)
		} else if owner != nil {
			burst := content.Skill{Name: "Clone Burst", MaxRange: 9, Power: 1}
			Hurt(c, b, dice, Attack{By: owner, Skill: burst, Target: z, Sure: true,
				Fixed: max(1, int(Round(float64(owner.Atk)*1.2)))}, emit)
		}
	}
}
