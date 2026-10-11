package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
	"github.com/papigetbusy/crusader-tactics-next/sim/turn"
)

// Setup describes how a fight starts: the map and who stands where. Until
// the floor builder arrives (Phase 7), practice fights use starting setups
// recorded from the old game.
type Setup struct {
	Name    string         `json:"name"`
	Floor   int            `json:"floor"`
	N       int            `json:"n"`
	Ground  []string       `json:"ground"`
	Heights [][]int        `json:"heights"`
	Units   []UnitSetup    `json:"units"`
	Throne  *state.Tile    `json:"throne,omitempty"` // seize-the-throne floors only
	Bag     []content.Item `json:"bag,omitempty"`    // items carried but not worn
	Chests  []state.Chest  `json:"chests,omitempty"` // unopened chests on the map
	// Built floors (Phase 7): its biome, reinforcement floor, blue starting tiles.
	Biome     string       `json:"biome,omitempty"`
	Reinforce bool         `json:"reinforce,omitempty"`
	Starts    []state.Tile `json:"starts,omitempty"`
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
	// Gear is what a hero wears, by slot (w, w2, w3, h, a, b, c). With gear,
	// the hero's numbers are worked out from its class, level and gear.
	Gear map[string]content.Item `json:"gear,omitempty"`
	// Bond is a hero's companion spirit (a monster kind and element); it
	// starts the fight beside the hero.
	Bond *state.Bond `json:"bond,omitempty"`
	Kind string      `json:"kind,omitempty"` // monsters: which kind
	Ally bool        `json:"ally,omitempty"` // a companion spirit to bond with
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
			Boss: us.Boss, Elite: us.Elite, Kind: us.Kind, Ally: us.Ally, Status: map[string]int{}}
		if u.Team == state.Heroes {
			u.Learned, u.Equipped, u.Ultimate = c.StartingBuild(u.Class, u.Level)
			if len(us.Equipped) > 0 {
				u.Equipped = us.Equipped
			}
			u.Skills = c.Bar(u.Class, u.Equipped, u.Ultimate)
			u.Potions = startingPotions
			u.Gear, u.BaseAtk = us.Gear, u.Atk
			if us.Gear != nil { // worked out from class, level and gear (even with no gear)
				n, base := stats.Geared(c, u.Class, u.Level, "", us.Gear)
				u.HP, u.MaxHP, u.Atk, u.Def, u.Mov, u.BaseAtk = n.MaxHP, n.MaxHP, n.Atk, n.Def, n.Mov, base
			}
			stats.InitSets(c, u)
			u.Bond = us.Bond
		} else {
			u.Skills = c.MonsterSkills(u.Class, u.Level, u.Boss || u.Elite)
			u.Weapon = c.Classes[u.Class].Weapon
		}
		u.Cooldowns = make([]int, len(u.Skills))
		b.Units = append(b.Units, u)
	}
	b.NextID = len(s.Units) + 1
	for _, u := range b.Units {
		if u.Team == state.Heroes {
			b.Party = append(b.Party, u.ID)
		}
	}
	b.Floor, b.Reinforce, b.Biome, b.Starts = s.Floor, s.Reinforce, s.Biome, s.Starts
	for _, id := range b.Party {
		combat.SpawnCompanion(c, b, b.Unit(id))
	}
	b.LevelCap = levelCap(c, b)
	b.Throne = s.Throne
	b.Bag = append([]content.Item(nil), s.Bag...)
	b.Chests = append([]state.Chest(nil), s.Chests...)
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
