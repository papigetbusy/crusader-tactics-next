// Package run is the climb between battles: the roster of all ten heroes,
// the floors reached and what each floor remembers, hero slots, gold, the
// bag, the shop and the gifts. A battle is built from it, and its results go
// back into it.
package run

import (
	"slices"
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Classes are the ten heroes, one of each class, in roster order.
var Classes = []string{"Warrior", "Tank", "Mage", "Healer", "Ranger", "Assassin", "Gunslinger", "Dragon Knight", "Witch", "Knight"}

// Hero is a hero kept between battles.
type Hero struct {
	Class     string                  `json:"class"`
	Level     int                     `json:"level"`
	XP        int                     `json:"xp"`
	Learned   []string                `json:"learned"`
	Equipped  []string                `json:"equipped"`
	Ultimate  string                  `json:"ultimate"`
	Ascension string                  `json:"ascension,omitempty"`
	Souls     []string                `json:"souls,omitempty"`
	Bond      *state.Bond             `json:"bond,omitempty"`
	SkillUses map[string]int          `json:"skillUses,omitempty"`
	Gear      map[string]content.Item `json:"gear"`
	WeaponSet int                     `json:"weaponSet,omitempty"`
	// SkillsVisit and AscensionVisit: the floor visit on which the hero last
	// changed skills or ascension (each once per floor).
	SkillsVisit    int `json:"skillsVisit,omitempty"`
	AscensionVisit int `json:"ascensionVisit,omitempty"`
}

// Memory is what one floor remembers (old fmem): its monster level, the
// chests it had and those opened, whether it was cleared, whether its boss
// gave its legendary, and its shop stock.
type Memory struct {
	Level         int            `json:"level"`
	Chests        int            `json:"chests"`
	Opened        []state.Tile   `json:"opened,omitempty"`
	Cleared       bool           `json:"cleared,omitempty"`
	BossLegendary bool           `json:"bossLegendary,omitempty"`
	Biome         string         `json:"biome,omitempty"`
	Shop          []content.Item `json:"shop,omitempty"`
	ShopMade      bool           `json:"shopMade,omitempty"`
}

// Run is one climb.
type Run struct {
	Seed      uint32          `json:"seed"`
	Floor     int             `json:"floor"` // the floor being played or prepared
	Best      int             `json:"best"`  // the highest floor reached
	Gold      int             `json:"gold"`
	Bag       []content.Item  `json:"bag"`
	Heroes    []*Hero         `json:"heroes"`
	Floors    map[int]*Memory `json:"floors"`
	Visit     int             `json:"visit"`     // counts floor visits (once-per-floor choices)
	ShopVisit int             `json:"shopVisit"` // the floor whose shop was last visited
	Gifts     int             `json:"gifts"`     // legendary gifts given (one per 100 floors)
	Combined  int             `json:"combined"`  // items combined so far (each combination rolls anew)
}

// New starts a climb on floor 1 with all ten heroes at level 1.
func New(c *content.Content, seed uint32) *Run {
	r := &Run{Seed: seed, Floor: 1, Best: 1, Floors: map[int]*Memory{}, Visit: 1}
	for _, class := range Classes {
		h := &Hero{Class: class, Level: 1, Gear: map[string]content.Item{}}
		h.Learned, h.Equipped, h.Ultimate = c.StartingBuild(class, 1)
		r.Heroes = append(r.Heroes, h)
	}
	return r
}

// Hero finds a hero by class.
func (r *Run) Hero(class string) *Hero {
	for _, h := range r.Heroes {
		if h.Class == class {
			return h
		}
	}
	return nil
}

// Slots is how many heroes go into a battle: 5, plus one on reaching floors
// 25, 50 and 75 (owner, 2026-10-10).
func (r *Run) Slots() int {
	n := 5
	for _, f := range []int{25, 50, 75} {
		if r.Best >= f {
			n++
		}
	}
	return n
}

// Strongest is the team the game brings when the player doesn't choose:
// the highest-level heroes (roster order breaks ties).
func (r *Run) Strongest() []string {
	hs := slices.Clone(r.Heroes)
	sort.SliceStable(hs, func(i, j int) bool { return hs[i].Level > hs[j].Level })
	var out []string
	for _, h := range hs[:min(r.Slots(), len(hs))] {
		out = append(out, h.Class)
	}
	return out
}

// Memory is a floor's memory, made on the first visit.
func (r *Run) Memory(f int) *Memory {
	m, ok := r.Floors[f]
	if !ok {
		m = &Memory{}
		r.Floors[f] = m
	}
	return m
}

// team is the floor builder's view of some heroes.
func (r *Run) team(classes []string) []floor.Hero {
	var t []floor.Hero
	for _, cl := range classes {
		if h := r.Hero(cl); h != nil {
			t = append(t, floor.Hero{Level: h.Level, Gear: h.Gear})
		}
	}
	return t
}

// Build builds a floor of this run. Its monster level is set on the first
// visit, from the strongest heroes for the slots, and kept after that
// (owner, 2026-10-11).
func (r *Run) Build(c *content.Content, f int) floor.Floor {
	m := r.Memory(f)
	fl := floor.Build(c, r.Seed, f, r.team(r.Strongest()), floor.Memory{Opened: m.Opened, Cleared: m.Cleared, Level: m.Level}, nil)
	if m.Level == 0 {
		m.Level, m.Chests, m.Biome = fl.Level, len(fl.Chests), fl.Biome
	}
	return fl
}

// LevelCap is the level heroes stop at: the best floor's monster level plus
// the gap (old capLv).
func (r *Run) LevelCap(c *content.Content) int {
	level := r.Memory(r.Best).Level
	if level == 0 {
		level = floor.Level(c, r.Best, r.team(r.Strongest()))
	}
	return level + c.Progress.CapAboveFloor
}
