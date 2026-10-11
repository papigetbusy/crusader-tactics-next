// Package floor builds the floors of a run (old newFloor0 and its helpers,
// index.html lines 736-760, 1000 and 1209-1228): the map, its biome, the
// monsters and their level, chests, a companion spirit, throne and
// reinforcement floors, and the blue starting tiles. A run seed and a floor
// number always give the same floor.
package floor

import (
	"math"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Hero is what the builder needs to know about a team member: monsters are
// as strong as the team that comes (old power).
type Hero struct {
	Level int
	Gear  map[string]content.Item
}

// Memory is what a floor remembers between visits (old fmem).
type Memory struct {
	Opened  []state.Tile // chests already opened
	Cleared bool         // beaten before: no companion spirit
	Level   int          // the monster level set on the first visit (0: work it out)
}

// Floor is a built floor, ready for the team to be placed.
type Floor struct {
	Number    int
	N         int
	Ground    []string
	Heights   [][]int
	Biome     string
	Shop      bool // the old game's shop floors (no monsters)
	Boss      bool
	Throne    *state.Tile // seize this tile to win (throne floors)
	Reinforce bool        // monsters arrive on the reinforcement turns
	Level     int         // the monsters' level
	Enemies   []*state.Unit
	Chests    []state.Chest
	Starts    []state.Tile // blue starting tiles
}

// Shuffler orders the tiles monsters start on. The new game uses a fair
// shuffle (approved 2026-10-09); the old-game check plays back the old one.
type Shuffler func(tiles [][2]int, r *rng.Stream)

// Fair is the fair shuffle.
func Fair(tiles [][2]int, r *rng.Stream) { rng.Shuffle(r, "start tiles", tiles) }

// Size is a floor's map size (old mapN).
func Size(c *content.Content, number int) int {
	f := c.Floors
	return min(f.MapMax, f.MapStart+(number-1)/f.MapEvery)
}

// Count is how many monsters a floor has for a team of this size (old mobN).
func Count(c *content.Content, number, team int) int {
	f := c.Floors
	extra := int(round(float64(max(0, team-f.PartyBase)) * f.PerExtraHero))
	return min(f.MonstersMax, f.MonstersStart+(number-1)/f.MonstersEvery+extra)
}

// Power is a team's strength: its average level plus gear (old power and
// gearPts, line 742).
func Power(c *content.Content, number int, team []Hero) float64 {
	if len(team) == 0 {
		return 1
	}
	f := c.Floors
	sum := 0.0
	for _, h := range team {
		pts := 0.0
		for _, sl := range []string{"h", "a", "b", "c", "w", "w2", "w3"} {
			it, ok := h.Gear[sl]
			if !ok {
				continue
			}
			af := 0.0
			for _, a := range it.Affixes {
				w, ok := f.GearWeights[a.Key]
				if !ok {
					w = .2
				}
				af = af + float64(float64(float64(a.Value)*w)*(1+float64(number)/f.GearFloorShare))
			}
			pts = pts + float64(it.Atk) + float64(it.Def) + float64(it.HP)/3 + float64(it.Mov*2) + af
		}
		sum = sum + float64(h.Level) + pts/f.GearPerLevel
	}
	return sum / float64(len(team))
}

// Level is the monsters' level on a floor (old elv and newFloor0, line 744):
// it rises with the floor and with the team's power, and is lower when
// there are more monsters.
func Level(c *content.Content, number int, team []Hero) int {
	f := c.Floors
	p := Power(c, number, team)
	late := float64(max(0, number-f.LateFrom))
	elv := max(1, int(round(1+float64(float64(number-1)*f.LevelPerFloor)+float64(late*f.LatePerFloor)+float64((p-1)*f.LevelPerPower))))
	cnt := Count(c, number, len(team))
	return max(1, elv-int(round(f.PerExtraFoe*float64(cnt-3))))
}

// round is JavaScript's Math.round (halves go up).
func round(x float64) float64 { return math.Floor(x + .5) }

// Is reports what kind of floor a number is.
func Is(c *content.Content, number int) (boss, shop, throne, reinforce bool) {
	f := c.Floors
	boss = f.BossEvery > 0 && number%f.BossEvery == 0
	shop = f.ShopEvery > 0 && number%f.ShopEvery == 0
	throne = !boss && !shop && number > 2 && f.ThroneEvery > 0 && number%f.ThroneEvery == f.ThroneAt
	reinforce = !boss && !shop && number > 3 && f.ReinforceEvery > 0 && number%f.ReinforceEvery == f.ReinforceAt
	return
}
