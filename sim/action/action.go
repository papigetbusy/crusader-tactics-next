// Package action is the vocabulary between the screen and the referee:
// actions are requests ("move this hero there"), events are facts the
// referee hands back ("the hero moved", "the Slime took 14 damage").
package action

// Kind names an action.
type Kind string

// The battle actions built so far.
const (
	Move    Kind = "move"    // Unit walks to X, Y
	Attack  Kind = "attack"  // Unit uses skill slot Skill on Target (any skill: damage, heal, buff)
	Wait    Kind = "wait"    // Unit ends its turn
	Defend  Kind = "defend"  // Unit braces: half damage until its next turn; ends its turn
	Potion  Kind = "potion"  // Unit drinks a potion; ends its turn
	Undo    Kind = "undo"    // Unit steps back to where it stood before moving this turn
	EndTurn Kind = "endturn" // the player's side is done; the enemies act
	Ability Kind = "ability" // Unit uses its class ability, aimed at Target (a unit) or Tiles, if it is aimed
	Decide  Kind = "decide"  // answers the first waiting decision with Choice (and Skill = slot to replace)
)

// Action is one request.
type Action struct {
	Kind   Kind     `json:"kind"`
	Unit   int      `json:"unit,omitempty"`
	X      int      `json:"x,omitempty"`
	Y      int      `json:"y,omitempty"`
	Skill  int      `json:"skill,omitempty"`  // slot on the unit's skill bar
	Target int      `json:"target,omitempty"` // unit ID
	Second int      `json:"second,omitempty"` // a second unit ID (Voodoo Link)
	Choice string   `json:"choice,omitempty"` // Decide: "ok", "keep", "replace", or an ascension's name
	Tiles  [][2]int `json:"tiles,omitempty"`  // the tiles picked, for tile-aimed abilities and skills
}

// Event is one fact, in the order it happened. The screen animates events;
// it never works anything out itself.
type Event struct {
	Kind     string         `json:"kind"`
	Unit     int            `json:"unit,omitempty"`   // who did it, or who it happened to
	Target   int            `json:"target,omitempty"` // who it was done to
	Skill    string         `json:"skill,omitempty"`
	Path     [][2]int       `json:"path,omitempty"`   // tiles walked, start first
	Amount   int            `json:"amount,omitempty"` // damage or healing
	Mult     float64        `json:"mult,omitempty"`   // above 1: weak spot; below 1: resisted
	Crit     bool           `json:"crit,omitempty"`
	Statuses map[string]int `json:"statuses,omitempty"`
	Side     int            `json:"side,omitempty"` // whose turn
	Turn     int            `json:"turn,omitempty"`
}

// Event kinds.
const (
	Moved         = "moved"      // Unit walked Path
	SkillUsed     = "skill"      // Unit used Skill on Target
	Hit           = "hit"        // Target took Amount damage
	Missed        = "missed"     // the attack on Target missed
	Healed        = "healed"     // Target healed Amount
	StatusOn      = "status"     // Statuses were put on Target
	Ticked        = "ticked"     // Unit lost Amount HP to a status at turn start
	Died          = "died"       // Unit died
	Finished      = "finished"   // Unit is done for this turn
	TurnEnd       = "turn-end"   // Side's turn ended
	TurnStart     = "turn-start" // Side's turn started (Turn is the turn number)
	Defended      = "defended"   // Unit is defending until its next turn
	Drank         = "potion"     // Unit drank a potion and healed Amount
	Undone        = "undone"     // Unit stepped back along Path
	Revived       = "revived"    // Target came back to life with Amount HP
	Seized        = "seized"     // Unit seized the throne: the enemies flee
	LevelUp       = "level-up"   // Unit reached level Amount
	Evolved       = "evolved"    // Unit evolved into the form named Skill (at levels 5 and 20)
	Guarded       = "guarded"    // Unit (a Guardian) takes the hit aimed at Target
	Immune        = "immune"     // Target's Glacial Scales stop Unit's ranged hit
	Shared        = "shared"     // Unit (a Templar's Sacred Vow) takes Amount of Target's damage
	Unbroken      = "unbroken"   // Unit survives a killing blow at 1 HP
	Riposte       = "riposte"    // Unit (a Duelist) strikes back at Target
	Exploded      = "exploded"   // Unit (a clone or skeleton) bursts as it is destroyed
	AbilityUsed   = "ability"    // Unit used its class ability Skill
	Summoned      = "summoned"   // Unit brought Target onto the board (clones, skeletons, raised monsters)
	Faded         = "faded"      // Unit (a shadow clone) fades away
	Trapped       = "trapped"    // Unit stepped on a hidden snare
	OverwatchShot = "overwatch"  // Unit shoots Target, who moved into its aimed cone
	ActAgain      = "act-again"  // Unit may act again this turn, thanks to Skill (Hit and Run, Quick Draw...)
	Learned       = "learned"    // Unit learned skill Skill
	Ascended      = "ascended"   // Unit ascended to Skill
	Won           = "won"        // every enemy is gone
	Lost          = "lost"       // every hero is gone
)

// Emit is how rules report events.
type Emit func(Event)
