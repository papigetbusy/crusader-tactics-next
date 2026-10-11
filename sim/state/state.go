// Package state defines the shape of a battle: the map and the units on it.
// Rules receive a *Battle instead of reading hidden globals.
package state

import "github.com/papigetbusy/crusader-tactics-next/sim/content"

// Team numbers, as in the old game.
const (
	Heroes  = 0
	Enemies = 1
)

// Unit is anything standing on the board.
type Unit struct {
	ID     int            // stable for the whole battle, including summons
	Team   int            // Heroes or Enemies
	Class  string         // "Knight", "Bat Lord"...
	X, Y   int            // position
	HP     int            // current hit points
	MaxHP  int            // maximum hit points
	Atk    int            // attack
	Def    int            // defense
	Mov    int            // move points
	Level  int            // level
	XP     int            // heroes: experience toward the next level
	Status map[string]int // status name -> turns left ("snare": 2)

	Weapon     string // sword, axe, lance... ("" for heroes without a triangle weapon)
	Element    string // fire, ice, elec, light, dark, or ""
	Boss       bool
	Elite      bool
	Defending  bool           // chose Defend this turn: takes half damage
	TilesMoved int            // tiles walked this turn (charge skills)
	RangeBonus int            // extra range (ascensions)
	Shield     int            // absorbs damage before HP
	CritChance int            // % (comes from gear; 0 without gear)
	SkillUses  map[string]int // heroes: uses per skill (skills level up with use)
	LastSkill  string         // heroes: the skill used last
	Summoned   bool           // summons and companions do not learn or level

	Skills    []content.Skill // the skill bar, in order
	Cooldowns []int           // turns until each skill is ready again (0 = ready)

	Moved bool // has moved this turn
	Acted bool // has used its action this turn
	Done  bool // finished for this turn

	CloakUsed bool // Cloak was used; its 2-turn cooldown starts when the cloak ends

	Potions int   // potions left this floor (heroes)
	Undo    *Tile // where the unit stood before its move this turn, while Undo is still allowed
}

// Tile is a board position.
type Tile struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Outcomes of a battle.
const (
	Ongoing = ""
	Won     = "won"
	Lost    = "lost"
)

// Battle is everything the referee knows about one fight.
type Battle struct {
	N        int      // the map is N x N tiles
	Ground   []string // one row of terrain letters per y ("ggfr..."); may be short
	Heights  [][]int  // -2 mountain, -1 water, 0 flat, 1 hill
	Units    []*Unit  // in turn order; dead units are removed
	Turn     int
	Side     int     // whose turn it is: Heroes or Enemies
	Outcome  string  // Ongoing, Won or Lost
	NextID   int     // the ID the next new unit gets
	Throne   *Tile   // seize-the-throne floors: a hero standing here wins the fight
	LevelCap int     // heroes stop gaining XP at this level (0: no cap); set by the run from the best floor
	Fallen   []*Unit // heroes who died in combat this floor, last one first to be revived
}

// UnitAt returns the unit standing on (x, y), or nil.
func (b *Battle) UnitAt(x, y int) *Unit {
	for _, u := range b.Units {
		if u.X == x && u.Y == y {
			return u
		}
	}
	return nil
}

// Unit returns the unit with this ID, or nil.
func (b *Battle) Unit(id int) *Unit {
	for _, u := range b.Units {
		if u.ID == id {
			return u
		}
	}
	return nil
}

// Remove takes a unit off the board (it died).
func (b *Battle) Remove(id int) {
	out := b.Units[:0]
	for _, u := range b.Units {
		if u.ID != id {
			out = append(out, u)
		}
	}
	b.Units = out
}

// Inside reports whether (x, y) is on the map.
func (b *Battle) Inside(x, y int) bool { return x >= 0 && y >= 0 && x < b.N && y < b.N }
