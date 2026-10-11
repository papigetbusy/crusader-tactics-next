package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// buffs are the good statuses Sunder strips.
var buffs = []string{"grd", "rally", "haste", "prism", "regen", "riposte", "trance", "dscale", "lwall", "wyrm",
	"lastlight", "aim", "coin", "cover", "taunt", "bulwark", "svow", "vow", "frenzy", "blood", "beast", "inner",
	"imh", "ironm", "enl", "heye", "dbl", "qd", "deadeye", "unb", "scales"}

// curses are the bad statuses Contagion copies.
var curses = append(slices.Clone(cleansable), "hmark", "silence", "slow", "wither")

// putStatus puts a status on v unless v cannot have it (a Juggernaut cannot
// be stunned, snared or broken; a Frost Wyrm cannot be snared).
func putStatus(v *state.Unit, q string, n int, emit action.Emit) {
	if (v.Ascension == "Juggernaut" && slices.Contains(juggernautImmune, q)) || (v.Ascension == "Frost Wyrm" && q == "snare") {
		return
	}
	v.Status[q] = max(v.Status[q], n)
	emit(action.Event{Kind: action.StatusOn, Target: v.ID, Statuses: map[string]int{q: v.Status[q]}})
}

// curse is a skill that only puts statuses on (it never misses).
func curse(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, s content.Skill, v *state.Unit, emit action.Emit) {
	if put := applyStatuses(c, b, dice, Attack{By: u, Skill: s, Target: v, Fixed: 1}, v); len(put) > 0 {
		emit(action.Event{Kind: action.StatusOn, Target: v.ID, Statuses: put})
	}
}

// contagion copies v's curses and poisons onto its allies within 2 tiles.
func contagion(b *state.Battle, u, v *state.Unit, emit action.Emit) {
	for _, z := range b.Units {
		if z == v || z.Team != v.Team || Distance(z, v) > 2 {
			continue
		}
		for _, q := range curses {
			if v.Status[q] > 0 {
				putStatus(z, q, v.Status[q], emit)
			}
		}
	}
}

// dispel strips v's buffs and shield.
func dispel(v *state.Unit, emit action.Emit) {
	gone := map[string]int{}
	for _, q := range buffs {
		if v.Status[q] > 0 {
			v.Status[q], gone[q] = 0, 0
		}
	}
	v.Shield = 0
	emit(action.Event{Kind: action.StatusOn, Target: v.ID, Statuses: gone})
}

// giveBuffs puts good statuses on z, unless Wither stops it.
func giveBuffs(z *state.Unit, bs map[string]int, emit action.Emit) {
	if len(bs) == 0 || z.Status["wither"] > 0 {
		return
	}
	for q, n := range bs {
		z.Status[q] = max(z.Status[q], n)
	}
	emit(action.Event{Kind: action.StatusOn, Target: z.ID, Statuses: bs})
}

// giveShield gives z a shield worth share of u's max HP (Wither stops it).
func giveShield(u, z *state.Unit, share float64) {
	if share > 0 && z.Status["wither"] <= 0 {
		z.Shield = max(z.Shield, int(Round(float64(u.MaxHP)*share)))
	}
}

// lastLight heals everyone a Last Light covered by 30% of their max HP.
func lastLight(b *state.Battle, u *state.Unit, s content.Skill, emit action.Emit) {
	if s.Trick != "lastLight" {
		return
	}
	for _, z := range b.Units {
		if z.Team == u.Team && Distance(z, u) <= s.Area {
			if n := heal(z, float64(z.MaxHP)*0.3); n > 0 {
				emit(action.Event{Kind: action.Healed, Unit: u.ID, Target: z.ID, Amount: n})
			}
		}
	}
}

// openVeins makes v's bleed, poison and burn hurt right now, then doubles
// the turns they have left.
func openVeins(c *content.Content, b *state.Battle, dice rng.Dice, u, v *state.Unit, emit action.Emit) {
	for _, q := range []string{"bleed", "poison", "burn"} {
		if v.Status[q] <= 0 {
			continue
		}
		share := float64(float64(v.MaxHP) * c.Turns.DamageOverTime[q])
		if q == "bleed" && v.HeavyBleed {
			share *= 2
		}
		n := max(c.Turns.DamageOverTimeMin, int(Round(share)))
		v.HP -= n
		v.Status[q] *= 2
		emit(action.Event{Kind: action.Ticked, Unit: v.ID, Amount: n, Statuses: map[string]int{q: v.Status[q]}})
		if v.HP <= 0 {
			died(c, b, dice, u, v, emit)
			return
		}
	}
}

// Kegs blows up every Powder Keg the other side left, at the start of this
// side's turn: the keg's skill hits everyone of this side within 1 tile.
func Kegs(c *content.Content, b *state.Battle, dice rng.Dice, side int, emit action.Emit) {
	var kept []state.Keg
	for _, k := range b.Kegs {
		if k.Team == side {
			kept = append(kept, k)
			continue
		}
		owner := b.Unit(k.Owner)
		s, _, ok := c.SkillFor(k.Team, classOf(owner), "Powder Keg")
		if owner == nil || !ok {
			continue
		}
		emit(action.Event{Kind: action.Exploded, Unit: owner.ID, Path: [][2]int{{k.X, k.Y}}})
		for _, z := range snapshot(b) {
			if z.Team == side && abs(z.X-k.X)+abs(z.Y-k.Y) <= s.Area && b.Unit(z.ID) != nil {
				Hurt(c, b, dice, Attack{By: owner, Skill: s, Target: z}, emit)
			}
		}
	}
	b.Kegs = kept
}

func classOf(u *state.Unit) string {
	if u == nil {
		return ""
	}
	return u.Class
}

// WyrmClaws are the dragons of Wyrmbond: at the end of their rider's side's
// turn each claws the weakest enemy next to it.
func WyrmClaws(c *content.Content, b *state.Battle, dice rng.Dice, side int, emit action.Emit) {
	for _, u := range snapshot(b) {
		if u.Team != side || u.Status["wyrm"] <= 0 || b.Unit(u.ID) == nil {
			continue
		}
		s, _, _ := c.SkillFor(u.Team, u.Class, "Wyrmbond")
		claw := content.Skill{Name: "Wyrmbond", Power: s.Power, MinRange: 1, MaxRange: 1}
		var best *state.Unit
		for _, z := range around(b, u, 1) {
			if best == nil || z.HP < best.HP {
				best = z
			}
		}
		if best != nil {
			Hurt(c, b, dice, Attack{By: u, Skill: claw, Target: best}, emit)
		}
	}
}

// LanceWall is a knight's Lance Wall striking an enemy that just walked up
// next to it (once per walk).
func LanceWall(c *content.Content, b *state.Battle, dice rng.Dice, e *state.Unit, path []board.Tile, emit action.Emit) {
	if len(path) < 2 {
		return
	}
	for _, h := range snapshot(b) {
		if h.Team == e.Team || h.Status["lwall"] <= 0 || Distance(h, e) != 1 || b.Unit(h.ID) == nil || b.Unit(e.ID) == nil {
			continue
		}
		s, _, _ := c.SkillFor(h.Team, h.Class, "Lance Wall")
		Hurt(c, b, dice, Attack{By: h, Skill: content.Skill{Name: "Lance Wall", Power: s.Power, MinRange: 1, MaxRange: 1}, Target: e, Counter: true}, emit)
	}
}

// Charged is the end of a side's turn on charged ground: anyone of that
// side standing on the other side's Static Field is shocked.
func Charged(b *state.Battle, side int, emit action.Emit) {
	for _, f := range b.Fields {
		if f.Kind != "static" || f.Team == side {
			continue
		}
		for _, z := range b.Units {
			if z.Team == side && abs(z.X-f.X)+abs(z.Y-f.Y) <= f.Range {
				putStatus(z, "shock", 1, emit)
			}
		}
	}
}
