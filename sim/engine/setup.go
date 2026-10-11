package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/turn"
)

// Setup describes how a fight starts: the map and who stands where. Until
// the floor builder arrives (Phase 7), practice fights use starting setups
// recorded from the old game.
type Setup struct {
	Name    string      `json:"name"`
	Floor   int         `json:"floor"`
	N       int         `json:"n"`
	Ground  []string    `json:"ground"`
	Heights [][]int     `json:"heights"`
	Units   []UnitSetup `json:"units"`
	Throne  *state.Tile `json:"throne,omitempty"` // seize-the-throne floors only
}

// startingPotions is what every new hero carries (old recruit and new run,
// index.html lines 715 and 1960: bag {potion:3}).
const startingPotions = 3

// UnitSetup is one unit at the start of a fight.
type UnitSetup struct {
	Team    int    `json:"team"`
	Class   string `json:"class"`
	Level   int    `json:"level"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	HP      int    `json:"hp"`
	Atk     int    `json:"atk"`
	Def     int    `json:"def"`
	Mov     int    `json:"mov"`
	Element string `json:"element,omitempty"`
	Boss    bool   `json:"boss,omitempty"`
	Elite   bool   `json:"elite,omitempty"`
	// Equipped picks a hero's normal skills (skill test fights); empty means
	// the usual starting build.
	Equipped []string `json:"equipped,omitempty"`
}

// Combat dice come from their own stream of the battle seed, so later
// streams (loot, floors) can never shift them.
const streamCombat = 1

// NewFight starts a fight from a setup. The battle seed comes from outside
// (the browser's secure random, or later an online server); the referee
// never invents randomness itself.
func NewFight(c *content.Content, s Setup, battleSeed uint64) *Game {
	b := &state.Battle{N: s.N, Ground: s.Ground, Heights: s.Heights, Turn: 1, Side: state.Heroes}
	for i, us := range s.Units {
		u := &state.Unit{ID: i + 1, Team: us.Team, Class: us.Class, Level: us.Level, X: us.X, Y: us.Y,
			HP: us.HP, MaxHP: us.HP, Atk: us.Atk, Def: us.Def, Mov: us.Mov, Element: us.Element,
			Boss: us.Boss, Elite: us.Elite, Status: map[string]int{}}
		if u.Team == state.Heroes {
			u.Learned, u.Equipped, u.Ultimate = c.StartingBuild(u.Class, u.Level)
			if len(us.Equipped) > 0 {
				u.Equipped = us.Equipped
			}
			u.Skills = c.Bar(u.Class, u.Equipped, u.Ultimate)
			u.Potions = startingPotions
		} else {
			u.Skills = c.MonsterSkills(u.Class, u.Level, u.Boss || u.Elite)
			u.Weapon = c.Classes[u.Class].Weapon
		}
		u.Cooldowns = make([]int, len(u.Skills))
		b.Units = append(b.Units, u)
	}
	b.NextID = len(s.Units) + 1
	b.LevelCap = levelCap(c, b)
	b.Throne = s.Throne
	turn.Check(b, func(e action.Event) {})
	return &Game{C: c, B: b, Dice: rng.New(battleSeed, streamCombat)}
}

// levelCap is the level heroes stop at: the floor's monster level plus the
// gap (old capLv, index.html line 1462). Normal monsters and elites have the
// floor's level; a boss is one higher. The run will set this from the best
// floor reached once floors arrive (Phase 7).
func levelCap(c *content.Content, b *state.Battle) int {
	low, boss := 0, 0
	for _, u := range b.Units {
		switch {
		case u.Team != state.Enemies:
		case u.Boss:
			boss = u.Level - 1
		case low == 0 || u.Level < low:
			low = u.Level
		}
	}
	if low == 0 {
		low = boss
	}
	return low + c.Progress.CapAboveFloor
}
