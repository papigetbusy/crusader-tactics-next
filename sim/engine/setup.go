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
}

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
			u.Skills = c.HeroSkills(u.Class, u.Level)
		} else {
			u.Skills = c.MonsterSkills(u.Class, u.Level, u.Boss || u.Elite)
			u.Weapon = c.Classes[u.Class].Weapon
		}
		u.Cooldowns = make([]int, len(u.Skills))
		b.Units = append(b.Units, u)
	}
	b.NextID = len(s.Units) + 1
	turn.Check(b, func(e action.Event) {})
	return &Game{C: c, B: b, Dice: rng.New(battleSeed, streamCombat)}
}
