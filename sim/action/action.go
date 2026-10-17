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
	Equip   Kind = "equip"   // Unit puts on bag item Item (a weapon goes in Slot: w, w2 or w3; empty = a free one)
	Drop    Kind = "drop"    // throws bag item Item away for good
	Unequip Kind = "unequip" // Unit takes off the item in Slot, into the bag
	UseSet  Kind = "set"     // Unit takes weapon set Set (1-3) in hand; it does not use up the turn
	// Bring changes Unit's skills, once per floor: Bar is the new bar of
	// 1-4 learned normal skills, Choice the ultimate (empty: keep it).
	Bring Kind = "bring"

	// The climb between battles (Phase 7).
	Team     Kind = "team"     // choose the heroes for this floor: Heroes (classes), one per slot
	Place    Kind = "place"    // put hero Hero on starting tile X, Y (swaps with a hero already there)
	Start    Kind = "start"    // start the battle (with the strongest heroes if none were chosen)
	Continue Kind = "continue" // after a battle or the shop: go on to the next step
	Travel   Kind = "travel"   // go to floor Floor (one already reached)
	Buy      Kind = "buy"      // buy shop item Item
	Sell     Kind = "sell"     // sell bag item Item
	Combine  Kind = "combine"  // combine bag items Item and Second (two copies) into one a rarity higher
	Ascend   Kind = "ascend"   // change hero Hero's ascension to Choice (once per floor)

	// Versus (online 1v1).
	Clock   Kind = "clock"   // only reports Ms, the time the player to move has used (their clocks may run out)
	Forfeit Kind = "forfeit" // the player on Side gives up; the other player wins at once
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
	Bar    []string `json:"bar,omitempty"`    // Bring: the 1-4 normal skills to carry, in order
	Heroes []string `json:"heroes,omitempty"` // Team: hero classes
	Hero   string   `json:"hero,omitempty"`   // Place, Ascend: a hero class
	Floor  int      `json:"floor,omitempty"`  // Travel
	Set    int      `json:"set,omitempty"`    // Attack: weapon set 1-3 to use (0 = the one in hand, or the first that fits the skill)
	Item   int      `json:"item,omitempty"`   // Equip, Drop: index in the bag
	Slot   string   `json:"slot,omitempty"`   // Equip: weapon slot for a weapon
	Ms     int      `json:"ms,omitempty"`     // versus: milliseconds the player to move used since their last action (the server or screen measures it)
	Side   int      `json:"side,omitempty"`   // Forfeit: the player who gives up (0 or 1)
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
	GearOn        = "equipped"   // Unit put on item Skill (the old one went to the bag)
	Looted        = "loot"       // item Skill (rarity Amount) went into the bag
	GoldGained    = "gold"       // Amount gold
	GearOff       = "unequipped" // Unit took item Skill off into the bag
	Bonded        = "bonded"     // Unit bonded with a companion spirit of kind Skill
	GearDropped   = "dropped"    // item Skill was thrown away
	SkillBrought  = "brought"    // Unit changed its skills (once per floor)
	Reinforced    = "reinforced" // Amount monsters arrived (reinforcement floors)
	Combined      = "combined"   // the shop combined two items into Skill (rarity Amount)
	SetSwitched   = "set"        // Unit took weapon set Amount (1-3) in hand
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
	Countered     = "counter"    // Unit strikes back at Target, who just hit it (Phase 9)
	ActAgain      = "act-again"  // Unit may act again this turn, thanks to Skill (Hit and Run, Quick Draw...)
	Learned       = "learned"    // Unit learned skill Skill
	Ascended      = "ascended"   // Unit ascended to Skill
	TimeUp        = "time-up"    // versus: player Side's match clock ran out; they lose
	TurnTimeUp    = "turn-time"  // versus: player Side's turn clock ran out; their turn ends
	GaveUp        = "gave-up"    // versus: player Side gave up; they lose
	Won           = "won"        // every enemy is gone
	Lost          = "lost"       // every hero is gone
)

// Emit is how rules report events.
type Emit func(Event)
