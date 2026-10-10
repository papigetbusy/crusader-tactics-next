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
)

// Action is one request.
type Action struct {
	Kind   Kind `json:"kind"`
	Unit   int  `json:"unit,omitempty"`
	X      int  `json:"x,omitempty"`
	Y      int  `json:"y,omitempty"`
	Skill  int  `json:"skill,omitempty"`  // slot on the unit's skill bar
	Target int  `json:"target,omitempty"` // unit ID
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
	Moved     = "moved"      // Unit walked Path
	SkillUsed = "skill"      // Unit used Skill on Target
	Hit       = "hit"        // Target took Amount damage
	Missed    = "missed"     // the attack on Target missed
	Healed    = "healed"     // Target healed Amount
	StatusOn  = "status"     // Statuses were put on Target
	Ticked    = "ticked"     // Unit lost Amount HP to a status at turn start
	Died      = "died"       // Unit died
	Finished  = "finished"   // Unit is done for this turn
	TurnEnd   = "turn-end"   // Side's turn ended
	TurnStart = "turn-start" // Side's turn started (Turn is the turn number)
	Defended  = "defended"   // Unit is defending until its next turn
	Drank     = "potion"     // Unit drank a potion and healed Amount
	Undone    = "undone"     // Unit stepped back along Path
	Revived   = "revived"    // Target came back to life with Amount HP
	Seized    = "seized"     // Unit seized the throne: the enemies flee
	Won       = "won"        // every enemy is gone
	Lost      = "lost"       // every hero is gone
)

// Emit is how rules report events.
type Emit func(Event)
