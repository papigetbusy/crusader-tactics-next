package floor

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Build makes floor number in a run for a team (old newFloor0). Each part
// draws from its own stream of the floor's seed, as in the old game, so
// the map stays the same whatever the team is.
func Build(c *content.Content, run uint32, number int, team []Hero, mem Memory, shuffle Shuffler) Floor {
	if shuffle == nil {
		shuffle = Fair
	}
	seed := rng.FloorSeed(run, number)
	n := Size(c, number)
	boss, shop, throne, rein := Is(c, number)
	fl := Floor{Number: number, N: n, Shop: shop, Boss: boss, Reinforce: rein, Level: Level(c, number, team)}
	if mem.Level > 0 {
		fl.Level = mem.Level
	}
	ground := terrain(n, throne || boss, rng.NewStream(seed+1))
	for _, row := range ground {
		fl.Ground = append(fl.Ground, string(row))
		h := make([]int, n)
		for x, ch := range row {
			h[x] = height(ch)
		}
		fl.Heights = append(fl.Heights, h)
	}
	if throne {
		fl.Throne = &state.Tile{X: n >> 1, Y: 1}
	}
	fl.Biome = c.Floors.Biomes[rng.NewStream(seed+2).N(len(c.Floors.Biomes))]
	if shop {
		return fl
	}
	fl.Enemies = monsters(c, &fl, number, len(team), shuffle, rng.NewStream(seed+3))
	extras(c, &fl, mem, rng.NewStream(seed+4))
	fl.Starts = starts(c, &fl, len(team))
	return fl
}

// monsters places the floor's monsters in the top rows (old newFloor0,
// lines 752-757): the boss in the top middle, an elite, then normal
// monsters, at most one of them a healer.
func monsters(c *content.Content, fl *Floor, number, team int, shuffle Shuffler, r *rng.Stream) []*state.Unit {
	f := c.Floors
	n := fl.N
	var kinds []byte
	if fl.Boss {
		kinds = append(kinds, 'B')
	}
	if (f.EliteBossFloors > 0 && number%f.EliteBossFloors == 0) || (number%f.EliteEvery == 0 && !fl.Boss) {
		kinds = append(kinds, 'E')
	}
	for len(kinds) < Count(c, number, team) {
		kinds = append(kinds, 'N')
	}
	var tiles [][2]int
	for y := 0; y < f.EnemyRows; y++ {
		for x := 0; x < n; x++ {
			if fl.Heights[y][x] >= 0 && !(x == n>>1 && y == 1) {
				tiles = append(tiles, [2]int{x, y})
			}
		}
	}
	shuffle(tiles, r)
	var out []*state.Unit
	healers := 0
	for _, k := range kinds {
		var pos [2]int
		if k == 'B' {
			pos = [2]int{n >> 1, 1}
			if i := slices.Index(tiles, pos); i >= 0 {
				tiles = slices.Delete(tiles, i, i+1)
			}
		} else {
			pos, tiles = tiles[len(tiles)-1], tiles[:len(tiles)-1]
		}
		var u *state.Unit
		switch k {
		case 'B':
			b := f.Bosses[(number/f.BossEvery-1)%len(f.Bosses)]
			u = Enemy(c, b.Name, fl.Level+1, b.Element, true, false)
		case 'E':
			e := f.Elites[r.N(len(f.Elites))]
			u = Enemy(c, e.Name, fl.Level, e.Element, false, true)
		default:
			var m content.Monster
			for {
				m = c.Monsters[r.N(len(c.Monsters))]
				if m.Class != "Healer" || healers < f.MaxHealers {
					break
				}
			}
			if m.Class == "Healer" {
				healers++
			}
			u = Monster(c, m, fl.Level, r)
		}
		u.X, u.Y = pos[0], pos[1]
		out = append(out, u)
	}
	return out
}

// Monster makes a normal monster of a kind; one with no element of its own
// rolls one (old newFloor0 line 757, reinforce line 1209).
func Monster(c *content.Content, m content.Monster, level int, d rng.Dice) *state.Unit {
	el := m.Element
	if el == "" {
		el = c.Floors.Elements[rng.IntN(d, RollKind, len(c.Floors.Elements))]
	}
	u := Enemy(c, m.Class, level, el, false, false)
	u.Kind = m.Name
	return u
}

// Enemy is a monster's numbers at a level (old calc with team 1, line 736):
// monsters grow faster than heroes, a little more at each level, up to a cap.
func Enemy(c *content.Content, class string, level int, element string, boss, elite bool) *state.Unit {
	f, p, b := c.Floors, c.Progress, c.Classes[class]
	lv := float64(level - 1)
	g := f.MonsterGrowth + min(f.MonsterGrowthCap, float64(f.MonsterGrowthLv*lv))
	hp := int(round(float64(b.HP) + float64(float64(p.GrowthHP*lv)*g)))
	atk := int(round(float64(b.Atk) + float64(lv*g) + p.AtkBonus))
	def := int(round(float64(b.Def) + float64(float64(lv*p.GrowthDef)*g)))
	u := &state.Unit{Team: state.Enemies, Class: class, Level: level, HP: hp, MaxHP: hp, Atk: atk, Def: def, Mov: b.Mov,
		Element: element, Weapon: b.Weapon, Boss: boss, Elite: elite, Status: map[string]int{}}
	u.Skills = c.MonsterSkills(class, level, boss || elite)
	u.Cooldowns = make([]int, len(u.Skills))
	return u
}

// extras puts chests on the floor and maybe a companion spirit (old
// spawnChests, line 1000, and newFloor0, line 758).
func extras(c *content.Content, fl *Floor, mem Memory, r *rng.Stream) {
	f, n := c.Floors, fl.N
	taken := func(x, y int) bool {
		for _, u := range fl.Enemies {
			if u.X == x && u.Y == y {
				return true
			}
		}
		for _, ch := range fl.Chests {
			if ch.X == x && ch.Y == y {
				return true
			}
		}
		return false
	}
	for _, o := range f.Chests {
		if r.Next() >= o.Chance {
			continue
		}
		for k := 0; k < 40; k++ {
			x := r.N(n)
			y := 2 + r.N(max(1, n-5))
			if fl.Heights[y][x] >= 0 && !taken(x, y) {
				fl.Chests = append(fl.Chests, state.Chest{X: x, Y: y, Kind: o.Kind})
				break
			}
		}
	}
	fl.Chests = slices.DeleteFunc(fl.Chests, func(ch state.Chest) bool {
		return slices.Contains(mem.Opened, state.Tile{X: ch.X, Y: ch.Y})
	})
	var normal []*state.Unit
	for _, u := range fl.Enemies {
		if u.Kind != "" {
			normal = append(normal, u)
		}
	}
	if !fl.Boss && len(normal) > 0 && r.Next() < f.AllyChance && !mem.Cleared {
		normal[r.N(len(normal))].Ally = true
	}
}

// starts are the blue starting tiles: every open tile of the bottom rows.
// With no starting rows (the old game) they are the old fixed spots, spread
// along the bottom row, one per hero.
func starts(c *content.Content, fl *Floor, team int) []state.Tile {
	n := fl.N
	var out []state.Tile
	if rows := c.Floors.StartRows; rows > 0 {
		for y := n - rows; y < n; y++ {
			for x := 0; x < n; x++ {
				if fl.Heights[y][x] >= 0 {
					out = append(out, state.Tile{X: x, Y: y})
				}
			}
		}
		return out
	}
	for i := 0; i < team; i++ {
		out = append(out, state.Tile{X: min(n-1, int(round((float64(i)+.5)*float64(n)/float64(team)-.5))), Y: n - 1})
	}
	return out
}
