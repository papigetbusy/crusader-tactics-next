package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// UseAimed is a unit using the skill in a slot with everything it picked.
// The old game's skills go the old way; the even-schedule skills come here.
func UseAimed(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, slot int, p Pick, emit action.Emit) error {
	s := u.Skills[slot]
	if !IsNew(s) {
		return UseSkill(c, b, dice, u, slot, p.Target, emit)
	}
	return useNew(c, b, dice, u, slot, s, p, emit)
}

// useNew is one even-schedule skill: what it does for its kind of aim, the
// usual after-skill rules, then Free skills, Momentum and Relentless Charge.
func useNew(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, slot int, s content.Skill, p Pick, emit action.Emit) error {
	_, classBase, _ := c.SkillFor(u.Team, u.Class, s.Name)
	who := p.Target
	if who == nil {
		who = u
	}
	emit(action.Event{Kind: action.SkillUsed, Unit: u.ID, Target: who.ID, Skill: s.Name})
	foes := map[int]bool{}
	for _, z := range b.Units {
		if z.Team != u.Team {
			foes[z.ID] = true
		}
	}
	switch s.Aim {
	case content.AimSelf:
		if s.Support { // Last Light: an aura like the old support skills
			support(c, b, u, s, emit)
			lastLight(b, u, s, emit)
		} else {
			selfSkill(c, b, dice, u, s, classBase, emit)
		}
	case content.AimAlly:
		allySkill(c, b, u, s, p.Target, emit)
	case content.AimAny:
		swap(u, p.Target, emit)
	case content.AimTile:
		tileSkill(c, b, u, s, p.Tiles[0], emit)
	case content.AimLine:
		lineSkill(c, b, dice, u, s, p.Tiles[0], classBase, emit)
	case content.AimRoll:
		jump(u, p.Tiles[0], emit)
		strike(c, b, dice, u, s, p.Target, classBase, emit)
	case content.AimShove:
		jump(p.Target, p.Tiles[0], emit)
	case content.AimTwo:
		for _, pair := range [2][2]*state.Unit{{p.Target, p.Second}, {p.Second, p.Target}} {
			pair[0].Status["link"], pair[0].LinkedTo = 3, pair[1].ID
			emit(action.Event{Kind: action.StatusOn, Target: pair[0].ID, Statuses: map[string]int{"link": 3}})
		}
	case content.AimEcho:
		last, cb, _ := c.SkillFor(u.Team, u.Class, u.LastSpell)
		last.Power = float64(last.Power * 0.6)
		strike(c, b, dice, u, last, p.Target, cb, emit)
	default:
		if s.Support {
			support(c, b, u, s, emit)
			lastLight(b, u, s, emit)
		} else {
			strike(c, b, dice, u, s, p.Target, classBase, emit)
		}
	}
	if s.Friendly() && !s.Support && isHero(u) {
		stats.Gain(c, b, u, c.Progress.SupportXP, emit)
	}
	killed := false
	for id := range foes {
		killed = killed || b.Unit(id) == nil
	}
	finishSkill(c, b, u, slot, s, emit)
	afterNew(b, u, s, killed, emit)
	return nil
}

// afterNew is who may go on after a new skill: Free skills leave the turn
// unused (Blink and Steady Aim end the walking), Momentum's kill earns one
// more action per turn, Relentless Charge lets the knight walk on.
func afterNew(b *state.Battle, u *state.Unit, s content.Skill, killed bool, emit action.Emit) {
	if s.Free {
		u.Done = false
		if s.Still || s.Trick == "blink" {
			u.Moved, u.Undo = true, nil
		}
	}
	if s.Trick == "momentum" && killed && !u.MomentumUsed && b.Unit(u.ID) != nil {
		u.MomentumUsed = true
		u.Done, u.Moved, u.Acted, u.Undo = false, false, false, nil
		emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: s.Name})
	}
	if s.Trick == "relentless" && b.Unit(u.ID) != nil {
		if left := board.Budget(b, u) - 2*u.TilesMoved; left > 0 {
			u.MoveLeft = left
			u.Done, u.Moved, u.Acted = false, false, true
			emit(action.Event{Kind: action.ActAgain, Unit: u.ID, Skill: s.Name})
		}
	}
}

// strike is a new skill aimed at one unit: a leap or teleport first, the
// hits (the old attack rules, with +50% for a Slam), then what follows a
// hit that landed (shove, pull, stun, stripping buffs, Mana Burn, Open
// Veins), Aftershock's second quake, Gravity Well, ground left behind, and
// a Disengage leap.
func strike(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, v *state.Unit, classBase bool, emit action.Emit) {
	if s.Leap > 0 {
		if t, ok := landing(c, b, u, v, s.Leap); ok && (t.X != u.X || t.Y != u.Y) {
			jump(u, t, emit)
		}
	}
	if s.Behind {
		jump(u, behindOf(u, v), emit)
	}
	cx, cy := v.X, v.Y
	if s.NoDamage {
		curse(c, b, dice, u, s, v, emit)
		if s.Trick == "contagion" {
			contagion(b, u, v, emit)
		}
	} else {
		dealt := map[int]int{}
		rec := func(e action.Event) {
			if e.Kind == action.Hit && e.Unit == u.ID {
				dealt[e.Target] += e.Amount
			}
			emit(e)
		}
		plain := s
		if s.Around {
			plain.Area = 0
		}
		slam := s.Slam && shoveBlocked(c, b, u.X, u.Y, v, s.Push)
		a := Attack{By: u, Skill: plain, ClassBase: classBase, Target: v}
		if slam {
			a.Extra = 1.5
		}
		attack(c, b, dice, a, rec)
		if s.Around {
			for _, z := range around(b, u, s.Area) {
				if z != v {
					Hurt(c, b, dice, Attack{By: u, Skill: plain, ClassBase: classBase, Target: z}, rec)
				}
			}
		}
		if dealt[v.ID] > 0 && b.Unit(v.ID) != nil {
			afterLanded(c, b, dice, u, s, v, slam, emit)
		}
		if s.Trick == "aftershock" {
			for id, d := range dealt {
				if z := b.Unit(id); z != nil && z.Team != u.Team {
					z.Quake += int(Round(float64(d) * 0.5))
				}
			}
		}
	}
	if s.Gather > 0 {
		for _, z := range snapshot(b) {
			if z.Team != u.Team && b.Unit(z.ID) != nil && abs(z.X-cx)+abs(z.Y-cy) <= s.Gather {
				drag(c, b, z, cx, cy, 1, emit)
			}
		}
	}
	if s.Field != "" {
		b.Fields = append(b.Fields, state.Field{X: cx, Y: cy, Range: s.FieldRange, Turns: s.FieldTurns, Kind: s.Field, Owner: u.ID, Team: u.Team})
	}
	if s.Away > 0 && b.Unit(u.ID) != nil {
		shove(c, b, cx, cy, u, s.Away, emit)
	}
}

// afterLanded is what a new skill does to a target it hit that still stands.
func afterLanded(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, v *state.Unit, slam bool, emit action.Emit) {
	if slam {
		putStatus(v, "stun", 1, emit)
	}
	if s.Push > 0 {
		shove(c, b, u.X, u.Y, v, s.Push, emit)
	}
	if s.Pull {
		drag(c, b, v, u.X, u.Y, 0, emit)
	}
	if s.Dispel {
		dispel(v, emit)
	}
	switch s.Trick {
	case "manaBurn":
		for i, sk := range v.Skills {
			if sk.Cooldown > 0 && i < len(v.Cooldowns) && v.Cooldowns[i] < 98 {
				v.Cooldowns[i]++
			}
		}
		if v.AbilityCooldown > 0 {
			v.AbilityCooldown++
		}
	case "openVeins":
		openVeins(c, b, dice, u, v, emit)
	}
}

// around are the units on the other side within r tiles of u.
func around(b *state.Battle, u *state.Unit, r int) []*state.Unit {
	var out []*state.Unit
	for _, z := range snapshot(b) {
		if z.Team != u.Team && Distance(z, u) <= r && b.Unit(z.ID) != nil {
			out = append(out, z)
		}
	}
	return out
}

// selfSkill is a new skill aimed at nobody: blows or roars around the user
// (Tail Sweep, Wyrm Roar), Taunting Roar, then the user's own buffs and
// shield.
func selfSkill(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, classBase bool, emit action.Emit) {
	if s.Around {
		for _, z := range around(b, u, s.Area) {
			if b.Unit(z.ID) == nil {
				continue
			}
			landed := true
			if s.NoDamage {
				curse(c, b, dice, u, s, z, emit)
			} else {
				landed = !Hurt(c, b, dice, Attack{By: u, Skill: s, ClassBase: classBase, Target: z}, emit).Missed
			}
			if landed && s.Push > 0 && b.Unit(z.ID) != nil {
				shove(c, b, u.X, u.Y, z, s.Push, emit)
			}
		}
	}
	if s.Trick == "taunt" {
		for _, z := range around(b, u, s.Area) {
			z.Status["taunted"], z.TauntedBy = 1, u.ID
			emit(action.Event{Kind: action.StatusOn, Target: z.ID, Statuses: map[string]int{"taunted": 1}})
		}
	}
	if s.Around {
		return
	}
	giveBuffs(u, s.Buffs, emit)
	giveShield(u, u, s.Shield)
}

// allySkill is a new skill aimed at an ally: Bodyguard, Rescue, a shield.
func allySkill(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, v *state.Unit, emit action.Emit) {
	switch s.Trick {
	case "bodyguard":
		v.GuardedBy = u.ID
		giveBuffs(v, map[string]int{"guarded": 2}, emit)
	case "rescue":
		if t, ok := landing(c, b, v, u, 99); ok {
			jump(v, t, emit)
		}
		giveBuffs(v, map[string]int{"cover": 2}, emit)
	}
	giveShield(u, v, s.Shield)
}

// tileSkill is a new skill aimed at a tile: Blink, Powder Keg, Bog of Despair.
func tileSkill(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, t state.Tile, emit action.Emit) {
	switch s.Trick {
	case "blink":
		jump(u, t, emit)
	case "keg":
		b.Kegs = append(b.Kegs, state.Keg{X: t.X, Y: t.Y, Owner: u.ID, Team: u.Team})
	}
	if s.Field != "" {
		b.Fields = append(b.Fields, state.Field{X: t.X, Y: t.Y, Range: s.FieldRange, Turns: s.FieldTurns, Kind: s.Field, Owner: u.ID, Team: u.Team})
	}
}

// lineSkill is a new skill aimed along a straight line: Ember Trail flies
// and leaves fire, Trample rides through and hits every enemy passed,
// Unstoppable hits each enemy in the way and shoves it aside (stopping in
// front of one that cannot be shoved).
func lineSkill(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, t state.Tile, classBase bool, emit action.Emit) {
	path := lineTiles(u, t.X, t.Y, s.MaxRange)
	dx, dy := sign(t.X-u.X), sign(t.Y-u.Y)
	walk := [][2]int{{u.X, u.Y}}
	for _, p := range path {
		walk = append(walk, [2]int{p.X, p.Y})
	}
	hit := func(z *state.Unit) bool {
		return !Hurt(c, b, dice, Attack{By: u, Skill: s, ClassBase: classBase, Target: z}, emit).Missed
	}
	switch s.Trick {
	case "emberTrail":
		place(u, walk, emit)
		for _, p := range path {
			b.Fields = append(b.Fields, state.Field{X: p.X, Y: p.Y, Turns: 2, Kind: "fire", Owner: u.ID, Team: u.Team})
		}
	case "trample":
		place(u, walk, emit)
		for _, p := range path {
			if z := b.UnitAt(p.X, p.Y); z != nil && z.Team != u.Team && z != u {
				hit(z)
			}
		}
	case "unstoppable":
		for i, p := range path {
			z := b.UnitAt(p.X, p.Y)
			if z == nil || z.Team == u.Team {
				continue
			}
			hit(z)
			if b.Unit(z.ID) != nil && !sideStep(c, b, z, dx, dy, emit) {
				walk = walk[:i+1]
				break
			}
		}
		place(u, walk, emit)
	}
}
