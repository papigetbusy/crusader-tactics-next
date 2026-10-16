package combat

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Pick is everything a skill was aimed at: a unit, a second unit (Voodoo
// Link) and tiles (Blink, Powder Keg, a line, a roll, a puppet's new spot).
type Pick struct {
	Target *state.Unit
	Second *state.Unit
	Tiles  []state.Tile
}

// IsNew reports a skill of the even schedule that needs the new rules.
func IsNew(s content.Skill) bool {
	return s.Aim != content.AimUnit || s.Push > 0 || s.Pull || s.Gather > 0 || s.Leap > 0 || s.Behind ||
		s.Away > 0 || s.NoDamage || s.Dispel || s.Shield > 0 || s.Field != "" || s.Trick != ""
}

// newTargets is Targets for the new skills' first pick: units for unit
// aims, the user itself for skills aimed at nobody or at tiles.
func newTargets(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill) []*state.Unit {
	inRange := func(z *state.Unit) bool { d := Distance(u, z); return d >= s.MinRange && d <= s.MaxRange }
	var out []*state.Unit
	switch s.Aim {
	case content.AimSelf, content.AimTile, content.AimLine:
		return []*state.Unit{u}
	case content.AimAlly:
		for _, z := range b.Units {
			if z.Team == u.Team && z != u && !z.Decoy && inRange(z) {
				out = append(out, z)
			}
		}
		return out
	case content.AimAny:
		for _, z := range b.Units {
			if z != u && inRange(z) {
				out = append(out, z)
			}
		}
		return out
	case content.AimEcho:
		last, _, ok := c.SkillFor(u.Team, u.Class, u.LastSpell)
		if !ok || u.LastSpell == "" {
			return nil
		}
		return Targets(c, b, u, last)
	case content.AimRoll:
		// Who can be shot after the roll depends on where it lands; the
		// first pick is the tile (RollTargets lists the rest).
		return []*state.Unit{u}
	}
	hi := s.MaxRange + u.RangeBonus
	if u.Status["aim"] > 0 {
		hi += 2
	}
	for _, z := range b.Units {
		d := Distance(u, z)
		if z.Team == u.Team || d < s.MinRange || d > hi || ((u.Monster() || u.PvP) && hidden(z)) {
			continue
		}
		if s.Leap > 0 {
			if _, ok := landing(c, b, u, z, s.Leap); !ok {
				continue
			}
		}
		if s.Behind {
			if t := behindOf(u, z); !Open(c, b, u, t.X, t.Y) {
				continue
			}
		}
		out = append(out, z)
	}
	return out
}

// AimTiles are the tiles a new skill can pick next, given what is picked so
// far: a tile in range (Blink, Powder Keg, Bog), a straight line (Trample,
// Unstoppable, Ember Trail), a roll spot, or a puppet's new spot.
func AimTiles(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, p Pick) []state.Tile {
	var out []state.Tile
	for y := 0; y < b.N; y++ {
		for x := 0; x < b.N; x++ {
			if tileOK(c, b, u, s, p, x, y) {
				out = append(out, state.Tile{X: x, Y: y})
			}
		}
	}
	return out
}

func tileOK(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, p Pick, x, y int) bool {
	d := abs(x-u.X) + abs(y-u.Y)
	switch s.Aim {
	case content.AimTile:
		if d < s.MinRange || d > s.MaxRange || !b.Inside(x, y) {
			return false
		}
		return s.Trick != "blink" || Open(c, b, u, x, y)
	case content.AimRoll:
		return d >= 1 && d <= s.Dash && Open(c, b, u, x, y)
	case content.AimShove:
		v := p.Target
		return v != nil && abs(x-v.X)+abs(y-v.Y) >= 1 && abs(x-v.X)+abs(y-v.Y) <= 2 && Open(c, b, v, x, y)
	case content.AimLine:
		path := lineTiles(u, x, y, s.MaxRange)
		if path == nil || !Open(c, b, u, x, y) {
			return false
		}
		for _, t := range path[:len(path)-1] {
			z := b.UnitAt(t.X, t.Y)
			if s.Trick == "emberTrail" {
				continue // flying: nothing on the way matters
			}
			if (z != nil && z.Team == u.Team) || wayBlocked(c, b, u, t) {
				return false
			}
		}
		return true
	}
	return false
}

// wayBlocked is ground u cannot cross on its way (enemies do not block).
func wayBlocked(c *content.Content, b *state.Battle, u *state.Unit, t state.Tile) bool {
	o := b.UnitAt(t.X, t.Y)
	if o != nil && o.Team != u.Team {
		return false // enemies on the way are what these skills are for
	}
	return !Open(c, b, u, t.X, t.Y)
}

// RollTargets are who a Roll Out can shoot after rolling to t.
func RollTargets(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, t state.Tile) []*state.Unit {
	ox, oy := u.X, u.Y
	u.X, u.Y = t.X, t.Y
	defer func() { u.X, u.Y = ox, oy }()
	plain := s
	plain.Aim = content.AimUnit
	return newTargets(c, b, u, plain)
}

// PickOK reports whether a new skill may be used with this pick.
func PickOK(c *content.Content, b *state.Battle, u *state.Unit, s content.Skill, p Pick) bool {
	if s.Still && u.Moved {
		return false
	}
	tileIn := func(i int, q Pick) bool {
		return i < len(p.Tiles) && tileOK(c, b, u, s, q, p.Tiles[i].X, p.Tiles[i].Y)
	}
	switch s.Aim {
	case content.AimTile, content.AimLine:
		return len(p.Tiles) == 1 && tileIn(0, p)
	case content.AimRoll:
		return len(p.Tiles) == 1 && tileIn(0, p) && p.Target != nil &&
			slices.Contains(RollTargets(c, b, u, s, p.Tiles[0]), p.Target)
	case content.AimShove:
		return p.Target != nil && slices.Contains(Targets(c, b, u, s), p.Target) && len(p.Tiles) == 1 && tileIn(0, p)
	case content.AimTwo:
		ts := Targets(c, b, u, s)
		return p.Target != nil && p.Second != nil && p.Target != p.Second &&
			slices.Contains(ts, p.Target) && slices.Contains(ts, p.Second)
	}
	return p.Target != nil && slices.Contains(Targets(c, b, u, s), p.Target)
}
