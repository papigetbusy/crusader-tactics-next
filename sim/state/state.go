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
	Kind   string         // monsters: which kind ("Goblin"), when known
	Ally   bool           // a monster glowing blue: defeating it bonds it as a companion (Phase 7)
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
	Defending  bool // chose Defend this turn: takes half damage
	TilesMoved int  // tiles walked this turn (charge skills)
	RangeBonus int  // extra range (ascensions)
	Shield     int  // absorbs damage before HP

	// Gear (heroes). Gear holds the items by slot (w, w2, w3, h, a, b, c).
	// WeaponSet is the set in hand; Affix is every bonus active right now
	// (armor, charm, set bonuses and the weapons in hand, old u.af); BaseAtk
	// is attack without weapons (old atk0). Armed is a hero that carries
	// weapon sets (old u.sets); clones, pets and monsters do not.
	Gear      map[string]content.Item
	WeaponSet int
	Affix     map[string]int
	BaseAtk   int
	Armed     bool
	SkillUses map[string]int // heroes: uses per skill (skills level up with use)
	LastSkill string         // heroes: the skill used last
	Summoned  bool           // summons and companions do not learn or level

	// Heroes: what the hero knows (the old hero record's learned, eq, ult, spec).
	Learned   []string // every skill learned
	Equipped  []string // up to four skills on the bar
	Swapped   bool     // has changed its skills on this floor (once per floor)
	Ultimate  string   // the skill in the ultimate slot
	Ascension string   // chosen at Lv 30 ("" before)
	Souls     []string // Necromancer: monster kinds whose raise skill it learned
	Bond      *Bond    // heroes: a bonded companion spirit (nil: none)

	Skills    []content.Skill // the skill bar, in order
	Cooldowns []int           // turns until each skill is ready again (0 = ready)

	Moved bool // has moved this turn
	Acted bool // has used its action this turn
	Done  bool // finished for this turn

	CloakUsed bool // Cloak was used; its 2-turn cooldown starts when the cloak ends

	// Class abilities and the effects they leave behind (old u.abcd, P2 numbers and flags).
	AbilityCooldown  int                // turns until the class ability is ready (old abcd)
	Params           map[string]float64 // numbers the active ability set (old P2: dblN, imhM, inP...); also ivyP on targets
	Decoy            bool               // a Phantom's shadow clone: no turn of its own, explodes when destroyed
	Undead           bool               // a raised skeleton: poisons neighbours when destroyed
	Companion        bool               // acts on its own in the enemy phase (bonded monsters, skeletons)
	Owner            int                // a clone's hero, or a summon's master (old owner/master)
	WantedBy         int                // Bounty Hunter's Wanted Poster: the hunter's ID (only if Wanted)
	Wanted           bool
	Overwatch        *Overwatch // Trickshot aiming its cone during the enemy turn
	SkyReady         bool       // Storm Dragoon in the sky: Crash Down replaces the ability
	UnbrokenUsed     bool       // Juggernaut's one saved killing blow is spent
	Danced           bool       // Blade Dancer already acted again this turn
	KilledThisAction bool       // made a kill during the current action (old _kill)
	ExtraTurn        bool       // Oath of Ruin: a kill earned another full turn (old _ruin)
	QuickShots       int        // Quick Draw shots fired this turn (old qdShot)
	QuickUsed        []string   // skills fired this turn with Quick Draw (old qdUsed)
	HeavyBleedBy     int        // Blood Pact: bleed damage heals this unit (only if HeavyBleed)
	HeavyBleed       bool

	// What the even-schedule skills leave on a unit.
	TauntedBy    int    // Taunting Roar: must attack this unit (while "taunted")
	GuardedBy    int    // Bodyguard: this unit takes the hits (while "guarded")
	LinkedTo     int    // Voodoo Link: shares damage with this unit (while "link")
	Quake        int    // Aftershock: damage waiting for the start of this unit's next turn
	MomentumUsed bool   // Momentum already gave an extra action this turn
	MoveLeft     int    // Relentless Charge: move budget left (half-steps) for one more walk
	LastSpell    string // the last spell cast (Echo repeats it)

	Potions int   // potions left this floor (heroes)
	Undo    *Tile // where the unit stood before its move this turn, while Undo is still allowed
}

// P is one of the numbers the unit's active ability set, or def when it is
// missing or zero (the old code wrote u.dblN||2).
func (u *Unit) P(key string, def float64) float64 {
	if v := u.Params[key]; v != 0 {
		return v
	}
	return def
}

// Overwatch is a Trickshot's aimed cone: direction and who it already shot.
type Overwatch struct {
	DX, DY int
	Hit    []int
}

// Field is a burning or bannered area that lasts a few turns (old FIELDS).
type Field struct {
	X, Y   int
	Range  int
	Square bool // range counts like a king's move (3x3), not in straight steps
	Turns  int
	Kind   string // "fire", "banner", "static" or "swamp"
	Owner  int
	Team   int // the side that made it: it works on the other side
}

// Keg is a Powder Keg waiting to blow up at the start of the other side's turn.
type Keg struct {
	X, Y  int
	Owner int
	Team  int
}

// Trap is a hidden Trapper snare (old TRAPS).
type Trap struct {
	X, Y  int
	Owner int
}

// Raise is a skeleton a Necromancer raises where a monster died.
type Raise struct {
	By   int
	X, Y int
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
	N          int      // the map is N x N tiles
	Ground     []string // one row of terrain letters per y ("ggfr..."); may be short
	Heights    [][]int  // -2 mountain, -1 water, 0 flat, 1 hill
	Units      []*Unit  // in turn order; dead units are removed
	Turn       int
	Side       int            // whose turn it is: Heroes or Enemies
	Outcome    string         // Ongoing, Won or Lost
	NextID     int            // the ID the next new unit gets
	Throne     *Tile          // seize-the-throne floors: a hero standing here wins the fight
	LevelCap   int            // heroes stop gaining XP at this level (0: no cap); set by the run from the best floor
	Fields     []Field        // burning ground and war banners
	Traps      []Trap         // hidden snares
	FloorKills []Tile         // where monsters died this floor (Grave Bloom raises them)
	Raises     []Raise        // skeletons waiting to rise once the current action is over
	Decisions  []Decision     // choices waiting for the player, answered first to last (old EVQ)
	Fallen     []*Unit        // heroes who died in combat this floor, last one first to be revived
	Lost       []*Unit        // party heroes who died where no revive reaches them (they still share the loot)
	FallenFoes []*Unit        // monsters who died, for monster healers' Resurrection (only with MonstersRevive)
	Kegs       []Keg          // Powder Kegs waiting to blow up
	Bag        []content.Item // items carried but not worn (the run will own this from Phase 7)

	// Loot and the run's numbers this fight needs (the run sets them from Phase 7).
	Floor         int     // which floor: loot gets stronger deeper down
	Gold          int     // gold carried
	Party         []int   // the heroes' IDs in party order, alive or fallen (loot goes to the least geared)
	Chests        []Chest // unopened chests on the map
	BossLegendary bool    // this floor's boss already gave a legendary
	Replayed      bool    // this floor was cleared before (half gold)
}

// Chest is a chest on the map: wood, silver or gold.
type Chest struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Kind string `json:"kind"`
}

// Bond is a hero's companion spirit: the monster kind and its element. It
// joins the hero in battle and acts on its own; if it dies it is gone.
type Bond struct {
	Kind    string `json:"kind"`
	Element string `json:"element,omitempty"`
}

// Decision is a choice the player must make before playing on: a skill
// learned at a level-up, an evolution to acknowledge, or the Lv 30 ascension.
type Decision struct {
	Kind    string `json:"kind"`              // DecideLearn, DecideEvolve or DecideAscend
	Unit    int    `json:"unit"`              // the hero it is about
	Skill   string `json:"skill,omitempty"`   // the skill learned (DecideLearn)
	Element string `json:"element,omitempty"` // DecideSwap: the new companion's element
}

// Kinds of decisions.
const (
	DecideLearn  = "learn"
	DecideEvolve = "evolve"
	DecideAscend = "ascend"
	DecideBond   = "bond"     // a hero bonded with a companion spirit (acknowledge)
	DecideSwap   = "bondSwap" // a hero with a companion met another: keep or take the new one
)

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

// Af is one active gear bonus (0 when the unit has none).
func (u *Unit) Af(key string) int { return u.Affix[key] }
