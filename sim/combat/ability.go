package combat

import (
	"fmt"
	"maps"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// AbilityOf is the class ability a hero can use now (old abOf, line 1191).
func AbilityOf(c *content.Content, u *state.Unit) (content.Ability, bool) {
	if !u.Player() || u.Summoned || u.Companion || u.Decoy {
		return content.Ability{}, false
	}
	return c.AbilityFor(u.Class, u.Level, u.Ascension, u.SkyReady && u.Status["sky"] > 0)
}

// AbilityUnits lists who a unit-aimed ability can be used on (old
// useAbility, line 1195): enemies, or allies other than the user, clones and
// companions, within range.
func AbilityUnits(b *state.Battle, u *state.Unit, t *content.AbilityTarget) []*state.Unit {
	var out []*state.Unit
	for _, z := range b.Units {
		ok := z.Team != u.Team
		if !t.Foe {
			ok = z.Team == u.Team && z != u && !z.Decoy && !z.Companion
		}
		if ok && Distance(z, u) <= t.Range {
			out = append(out, z)
		}
	}
	return out
}

// AbilityTileOK reports whether a tile-aimed ability may pick (x, y) (old
// click handler, line 1759): in range, walkable ground, empty if required.
func AbilityTileOK(b *state.Battle, u *state.Unit, t *content.AbilityTarget, x, y int) bool {
	d := abs(x-u.X) + abs(y-u.Y)
	if !b.Inside(x, y) || d > t.Range || d < t.Min || heightAt(b, x, y) < 0 {
		return false
	}
	return !t.Empty || b.UnitAt(x, y) == nil
}

// UseAbility is a hero using its class ability (old useAbility, lines
// 1193-1195): the statuses and their numbers, the ability's special effect,
// then its cooldown. It assumes the caller already checked it is allowed.
func UseAbility(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, ab content.Ability, target *state.Unit, tiles []state.Tile, emit action.Emit) error {
	emit(action.Event{Kind: action.AbilityUsed, Unit: u.ID, Skill: ab.Name})
	switch ab.Effect {
	case "payHP":
		u.HP = max(1, u.HP-int(Round(float64(u.MaxHP)*ab.Amount)))
	case "resetUnbroken":
		u.UnbrokenUsed = false
	}
	for q, n := range ab.Statuses {
		u.Status[q] = n
	}
	if len(ab.Params) > 0 {
		if u.Params == nil {
			u.Params = map[string]float64{}
		}
		maps.Copy(u.Params, ab.Params)
	}
	if _, ok := ab.Statuses["qd"]; ok {
		u.QuickShots, u.QuickUsed = 0, nil
	}
	switch ab.Effect {
	case "", "payHP", "resetUnbroken":
	case "reviveOne":
		reviveOne(c, b, u, emit)
	case "timeShift":
		target.Done, target.Moved, target.Acted, target.Undo = false, false, false, nil
		target.Status["haste"] = max(target.Status["haste"], 2)
	case "freeze":
		if !shrugs(target, "frozen") { // Wyrm's Hide (approved 2026-10-17)
			target.Status["frozen"] = 2
		}
	case "deathMark":
		target.Status["dmark"] = 3
	case "wanted":
		target.Wanted, target.WantedBy = true, u.ID
	case "standStill":
		u.Moved = true
	case "traps":
		for _, t := range tiles {
			b.Traps = append(b.Traps, state.Trap{X: t.X, Y: t.Y, Owner: u.ID, Team: u.Team})
		}
	case "banner":
		b.Fields = append(b.Fields, state.Field{X: tiles[0].X, Y: tiles[0].Y, Range: 3, Turns: 3, Kind: "banner", Owner: u.ID, Team: u.Team})
	case "fireField":
		b.Fields = append(b.Fields, state.Field{X: tiles[0].X, Y: tiles[0].Y, Range: 1, Square: true, Turns: 3, Kind: "fire", Owner: u.ID, Team: u.Team})
	case "overwatch":
		dx, dy := tiles[0].X-u.X, tiles[0].Y-u.Y
		if abs(dx) >= abs(dy) {
			u.Overwatch = &state.Overwatch{DX: sign(dx)}
		} else {
			u.Overwatch = &state.Overwatch{DY: sign(dy)}
		}
		u.Done = true
	case "clones":
		makeClones(b, u, emit)
	case "skyLord":
		u.SkyReady, u.Done = true, true
	case "crashDown":
		crashDown(c, b, dice, u, ab, tiles[0], emit)
	case "graveBloom":
		if b.Versus != nil { // versus: the other player's fallen heroes rise, never your own
			foe := 1 - u.Team
			for _, k := range b.SideKills[foe] {
				RaiseSkeleton(c, b, u, k.X, k.Y, true, emit)
			}
			b.SideKills[foe] = nil
			break
		}
		for _, k := range b.FloorKills {
			RaiseSkeleton(c, b, u, k.X, k.Y, true, emit)
		}
		b.FloorKills = nil
	default:
		return fmt.Errorf("ability %s: unknown effect %q", ab.Name, ab.Effect)
	}
	if !ab.Free {
		u.AbilityCooldown = ab.Cooldown
	}
	if u.Done {
		PairDone(b, u)
	}
	return nil
}

// crashDown lands a Storm Dragoon on the chosen tile and hits every enemy
// within 2 tiles for 2.5x, shocking them; it never misses (old SKYC).
func crashDown(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, ab content.Ability, t state.Tile, emit action.Emit) {
	u.X, u.Y = t.X, t.Y
	u.Status["sky"], u.SkyReady = 0, false
	emit(action.Event{Kind: action.Moved, Unit: u.ID, Path: [][2]int{{t.X, t.Y}}})
	slam := content.Skill{Name: "Crash Down", MaxRange: 9, Power: ab.Amount, Element: "elec", Physical: true}
	for _, z := range snapshot(b) {
		if z.Team != u.Team && Distance(z, u) <= 2 && b.Unit(z.ID) != nil {
			Hurt(c, b, dice, Attack{By: u, Skill: slam, Target: z, Sure: true}, emit)
		}
	}
	u.Done = true
}

// FreeNear is the nearest open tile around (x, y), ring by ring up to r
// steps, in the old game's scan order (old freeNear, line 1119).
func FreeNear(b *state.Battle, x, y, r int) (state.Tile, bool) {
	for k := 1; k <= r; k++ {
		for dy := -k; dy <= k; dy++ {
			for dx := -k; dx <= k; dx++ {
				if abs(dx)+abs(dy) != k {
					continue
				}
				a, bb := x+dx, y+dy
				if b.Inside(a, bb) && heightAt(b, a, bb) >= 0 && b.UnitAt(a, bb) == nil {
					return state.Tile{X: a, Y: bb}, true
				}
			}
		}
	}
	return state.Tile{}, false
}

// reviveOne is Divine Choir's revive: the last hero to fall returns near the
// Saint with 40% HP (old reviveOne, line 1120).
func reviveOne(c *content.Content, b *state.Battle, u *state.Unit, emit action.Emit) {
	if len(b.Fallen) == 0 {
		return
	}
	t, ok := FreeNear(b, u.X, u.Y, 2)
	if !ok {
		return
	}
	f := b.Fallen[len(b.Fallen)-1]
	b.Fallen = b.Fallen[:len(b.Fallen)-1]
	f.X, f.Y = t.X, t.Y
	f.HP = int(Round(float64(f.MaxHP) * c.Combat.ReviveShare))
	f.Status, f.Done = map[string]int{}, true
	b.Units = append(b.Units, f)
	emit(action.Event{Kind: action.Revived, Unit: u.ID, Target: f.ID, Amount: f.HP})
}

// makeClones summons two shadow clones near the Phantom: 30% of its max HP
// (at least 5), no skills, gone after 3 turns (old mkClones, line 1122).
func makeClones(b *state.Battle, u *state.Unit, emit action.Emit) {
	for i := 0; i < 2; i++ {
		t, ok := FreeNear(b, u.X, u.Y, 2)
		if !ok {
			return
		}
		hp := max(5, int(Round(float64(u.MaxHP)*0.3)))
		cl := *u
		cl.ID, b.NextID = b.NextID, b.NextID+1
		cl.X, cl.Y, cl.HP, cl.MaxHP, cl.Shield = t.X, t.Y, hp, hp, 0
		cl.Decoy, cl.Owner = true, u.ID
		cl.Status = map[string]int{"clone": 3}
		cl.Skills, cl.Cooldowns, cl.Params = nil, nil, maps.Clone(u.Params)
		cl.Done, cl.Moved, cl.AbilityCooldown = true, true, 0
		cl.Affix, cl.Gear, cl.Armed = nil, nil, false
		cl.Learned, cl.Equipped, cl.SkillUses = nil, nil, nil
		b.Units = append(b.Units, &cl)
		emit(action.Event{Kind: action.Summoned, Unit: u.ID, Target: cl.ID})
	}
}

// RaiseSkeleton raises a skeleton for the Necromancer: on the given tile if
// it is open (else nearby), at one level below the Necromancer; it fights on
// its own and poisons when destroyed. At most 10 at a time (old raiseSkel,
// line 1121).
func RaiseSkeleton(c *content.Content, b *state.Battle, u *state.Unit, x, y int, at bool, emit action.Emit) {
	m, _, ok := c.MonsterNamed("Skeleton")
	if !ok {
		return
	}
	n := 0
	for _, z := range b.Units {
		if z.Undead && z.Companion {
			n++
		}
	}
	if n >= 10 {
		return
	}
	t := state.Tile{X: x, Y: y}
	if !at || !b.Inside(x, y) || heightAt(b, x, y) < 0 || b.UnitAt(x, y) != nil {
		var found bool
		if t, found = FreeNear(b, x, y, 3); !found {
			if t, found = FreeNear(b, u.X, u.Y, 3); !found {
				return
			}
		}
	}
	z := monsterUnit(c, b, m, max(1, u.Level-1), "dark", u.Team, t)
	z.Companion, z.Undead, z.Done = true, true, true
	b.Units = append(b.Units, z)
	emit(action.Event{Kind: action.Summoned, Unit: u.ID, Target: z.ID})
}

// monsterUnit makes a monster of this kind on a team at a tile. On the
// heroes' side it gets a hero's numbers for its class and level and no
// weapon, as in the old game (mk with team 0 used calc without a form, and
// the weapon-set setup left it unarmed).
func monsterUnit(c *content.Content, b *state.Battle, m content.Monster, level int, element string, team int, t state.Tile) *state.Unit {
	if m.Element != "" {
		element = m.Element
	}
	n := stats.Plain(c, m.Class, level)
	z := &state.Unit{ID: b.NextID, Team: team, Class: m.Class, Kind: m.Name, X: t.X, Y: t.Y, Level: level,
		HP: n.MaxHP, MaxHP: n.MaxHP, Atk: n.Atk, Def: n.Def, Mov: n.Mov, Element: element,
		Status: map[string]int{}}
	z.PvP = b.Versus != nil
	if team != state.Heroes && !z.PvP {
		z.Weapon = c.Classes[m.Class].Weapon
	}
	b.NextID++
	z.Skills = c.MonsterSkills(m.Class, level, false)
	z.Cooldowns = make([]int, len(z.Skills))
	return z
}

// PairDone ends the turn of a hero's raised pet too, and the hero's when the
// pet acts (old pairDone, line 644): they share one action.
func PairDone(b *state.Battle, u *state.Unit) {
	for _, z := range b.Units {
		if z != u && ((u.Summoned && z.ID == u.Owner) || (z.Summoned && z.Owner == u.ID)) {
			z.Done, z.Moved = true, true
		}
	}
}

// SpawnCompanion puts a hero's bonded companion spirit next to it at the
// start of a fight (old newFloor0, index.html line 747): two levels below
// the hero, on the first open tile above or beside it.
func SpawnCompanion(c *content.Content, b *state.Battle, h *state.Unit) {
	if h.Bond == nil {
		return
	}
	m, _, ok := c.MonsterNamed(h.Bond.Kind)
	if !ok {
		return
	}
	for _, d := range [][2]int{{0, -1}, {1, -1}, {-1, -1}, {1, 0}, {-1, 0}} {
		x, y := h.X+d[0], h.Y+d[1]
		if b.Inside(x, y) && heightAt(b, x, y) >= 0 && b.UnitAt(x, y) == nil {
			z := monsterUnit(c, b, m, max(1, h.Level-2), h.Bond.Element, h.Team, state.Tile{X: x, Y: y})
			z.Element, z.Kind = h.Bond.Element, m.Name
			z.Companion, z.Owner = true, h.ID
			b.Units = append(b.Units, z)
			return
		}
	}
}
