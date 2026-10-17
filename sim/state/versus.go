package state

// Versus is a 1v1 match (owner, 2026-10-11 and 2026-10-16): both sides are
// players' heroes, there are no monsters, and each player has clocks. The
// referee never reads a clock; the time used comes with each action.
type Versus struct {
	TurnMs   int    // the most one turn may last; 0 means no per-turn limit
	BankMs   [2]int // each player's time left for the whole match
	UsedMs   int    // time used so far in the current turn
	TimedOut int    // -1, or the side whose match clock ran out (that side lost)
}

// Player reports whether a player owns u: the heroes (and their summons) in
// single player, every unit in a versus match.
func (u *Unit) Player() bool { return u.Team == Heroes || u.PvP }

// Monster reports whether u is on the monsters' side (never in versus).
func (u *Unit) Monster() bool { return !u.Player() }

// Other is the side that is not s.
func Other(s int) int { return 1 - s }
