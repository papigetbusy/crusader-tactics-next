package engine

import (
	"fmt"
	"hash/fnv"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/items"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
	"github.com/papigetbusy/crusader-tactics-next/sim/turn"
)

// Pick is one hero a player brings to a versus match.
type Pick struct {
	Class     string `json:"class"`
	Ascension string `json:"ascension"`
}

// VersusSetup is everything a 1v1 match starts from: both players' picks and
// the map seed. The battle seed comes separately, as for every fight.
type VersusSetup struct {
	Teams   [2][]Pick `json:"teams"` // side 0 moves first
	MapSeed uint32    `json:"mapSeed"`
}

// NewVersus starts a 1v1 match (owner, 2026-10-11 and 2026-10-16): every
// hero at the versus level with its chosen ascension and its class's purple
// set, on a mirrored map. Side 0 moves first.
func NewVersus(c *content.Content, s VersusSetup, battleSeed uint64) (*Game, error) {
	k := c.Versus
	for side, team := range s.Teams {
		if err := checkPicks(c, team, k.Team); err != nil {
			return nil, fmt.Errorf("player %d: %w", side+1, err)
		}
	}
	a := floor.NewArena(k.Size, k.Team, len(c.Floors.Biomes), s.MapSeed)
	b := &state.Battle{N: a.N, Ground: a.Ground, Heights: a.Heights, Turn: 1, Side: state.Heroes,
		Versus: &state.Versus{TurnMs: k.TurnSeconds * 1000, BankMs: [2]int{k.BankSeconds * 1000, k.BankSeconds * 1000}, TimedOut: -1}}
	if len(c.Floors.Biomes) > 0 {
		b.Biome = c.Floors.Biomes[a.Biome]
	}
	for side, team := range s.Teams {
		for i, p := range team {
			b.NextID++
			t := a.Starts[side][i]
			u := &state.Unit{ID: b.NextID, Team: side, PvP: true, Class: p.Class, Level: k.Level, X: t.X, Y: t.Y,
				Ascension: p.Ascension, SkillUses: map[string]int{}, Status: map[string]int{}, Potions: startingPotions}
			u.Learned, u.Equipped, u.Ultimate = c.StartingBuild(p.Class, k.Level)
			if asc, ok := c.Ascensions[p.Ascension]; ok && asc.Ultimate != "" {
				u.Ultimate = asc.Ultimate
				if !slices.Contains(u.Learned, asc.Ultimate) {
					u.Learned = append(u.Learned, asc.Ultimate)
				}
			}
			u.Gear = PurpleSet(c, p.Class)
			b.Units = append(b.Units, u)
			stats.Refresh(c, b, u)
			u.HP = u.MaxHP
			if side == 0 {
				b.Party = append(b.Party, u.ID)
			}
		}
	}
	b.NextID++
	turn.Check(b, func(action.Event) {})
	return &Game{C: c, B: b, Dice: rng.New(battleSeed, streamCombat)}, nil
}

// checkPicks is the rule for a versus team: the right number of different
// classes, each with one of its own ascensions.
func checkPicks(c *content.Content, team []Pick, want int) error {
	if len(team) != want {
		return fmt.Errorf("bring %d heroes, not %d", want, len(team))
	}
	seen := map[string]bool{}
	for _, p := range team {
		if _, ok := c.AscensionChoices[p.Class]; !ok {
			return fmt.Errorf("%q is not a hero class", p.Class)
		}
		if seen[p.Class] {
			return fmt.Errorf("%s is picked twice", p.Class)
		}
		seen[p.Class] = true
		if !slices.Contains(c.AscensionChoices[p.Class], p.Ascension) {
			return fmt.Errorf("%s can't take the ascension %q", p.Class, p.Ascension)
		}
	}
	return nil
}

// PurpleSet is a class's fixed purple gear for versus (owner, 2026-10-16):
// a purple weapon for each of its weapon sets, purple armor of its first
// armor weight, and a purple charm. The bonuses are rolled once from the
// class's name, so the same class always wears the same set.
func PurpleSet(c *content.Content, class string) map[string]content.Item {
	cg, ok := c.Gear.Classes[class]
	if !ok || c.Gear.Epic() < 0 {
		return map[string]content.Item{}
	}
	h := fnv.New64a()
	h.Write([]byte(class))
	dice := rng.New(h.Sum64(), 0)
	k, lv := c.Gear.Epic(), c.Versus.Level
	out := map[string]content.Item{}
	for i, w := range cg.Weapons {
		slot := "w"
		if i > 0 {
			slot = fmt.Sprintf("w%d", i+1)
		}
		out[slot] = items.Make(c, lv, dice, k, "w", w.Type, "", class) // a weapon keeps slot "w" in any set
	}
	weight := ""
	if len(cg.Armor) > 0 {
		weight = cg.Armor[0]
	}
	for _, sl := range []string{"h", "a", "b"} {
		out[sl] = items.Make(c, lv, dice, k, sl, "", weight, class)
	}
	out["c"] = items.Make(c, lv, dice, k, "c", "", "", class)
	return out
}

// clock charges the time used to the player to move (versus only). It
// reports true when that ended the turn or the match, and the action itself
// must not be done.
func (g *Game) clock(a action.Action, emit action.Emit) (bool, error) {
	b, v := g.B, g.B.Versus
	if v == nil || b.Outcome != state.Ongoing {
		return false, nil
	}
	if a.Ms < 0 {
		return false, no("bad-clock", "Time can't run backwards.")
	}
	side := b.Side
	v.UsedMs += a.Ms
	v.BankMs[side] -= a.Ms
	if v.TurnMs > 0 { // a turn limit, when there is one, stops the match clock at that limit
		if over := v.UsedMs - v.TurnMs; over > 0 {
			v.BankMs[side] += over
			v.UsedMs = v.TurnMs
		}
	}
	if v.BankMs[side] <= 0 {
		v.BankMs[side], v.TimedOut = 0, side
		emit(action.Event{Kind: action.TimeUp, Side: side})
		if side == state.Heroes {
			b.Outcome = state.Lost
			emit(action.Event{Kind: action.Lost, Side: state.Enemies})
		} else {
			b.Outcome = state.Won
			emit(action.Event{Kind: action.Won, Side: state.Heroes})
		}
		return true, nil
	}
	if v.TurnMs > 0 && v.UsedMs >= v.TurnMs {
		emit(action.Event{Kind: action.TurnTimeUp, Side: side})
		return true, turn.EndTurn(g.C, b, g.Dice, emit)
	}
	return a.Kind == action.Clock, nil
}

// VersusView is a 1v1 match as the screen shows it. The screen counts the
// seconds down between answers; the referee's numbers are the real ones.
type VersusView struct {
	TurnMs   int    `json:"turnMs"`   // the most one turn may last; 0 means there is none
	UsedMs   int    `json:"usedMs"`   // used so far in this turn
	BankMs   [2]int `json:"bankMs"`   // each player's time left
	TimedOut int    `json:"timedOut"` // -1, or the player whose match clock ran out
	Winner   int    `json:"winner"`   // -1 while playing, else 0 or 1
}

func versusView(b *state.Battle) *VersusView {
	vs := b.Versus
	if vs == nil {
		return nil
	}
	v := &VersusView{TurnMs: vs.TurnMs, UsedMs: vs.UsedMs, BankMs: vs.BankMs, TimedOut: vs.TimedOut, Winner: -1}
	switch b.Outcome {
	case state.Won:
		v.Winner = 0
	case state.Lost:
		v.Winner = 1
	}
	return v
}

// CheckTeam is the versus team rule for a server or a picker: the right
// number of different classes, each with one of its own ascensions.
func CheckTeam(c *content.Content, team []Pick) error { return checkPicks(c, team, c.Versus.Team) }
