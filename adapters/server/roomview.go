package main

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
)

// ---------- what each person is sent ----------

// PersonView is one person as the room shows them.
type PersonView struct {
	Name    string `json:"name"`
	Side    int    `json:"side"` // -1 waiting or watching
	Line    int    `json:"line"` // place in line, from 1; 0 when seated
	Ready   bool   `json:"ready"`
	HasTeam bool   `json:"hasTeam"`
	Online  bool   `json:"online"`
	You     bool   `json:"you"`
}

// RoomView is the room as one person sees it.
type RoomView struct {
	Type     string        `json:"type"` // "room"
	Room     string        `json:"room"`
	People   []PersonView  `json:"people"`
	Seats    [2]string     `json:"seats"` // names on Blue and Red
	YourSide int           `json:"yourSide"`
	YourTeam []engine.Pick `json:"yourTeam,omitempty"`
	Playing  bool          `json:"playing"`
	Winner   int           `json:"winner"`
	Matches  int           `json:"matches"`
	Max      int           `json:"max"`
	Seq      int           `json:"seq"`
	ThinkMs  int           `json:"thinkMs"` // time the server would charge the player to move right now
	Match    *MatchView    `json:"match,omitempty"`
}

// MatchView is the match so far: the referee's View and its save, for the
// phone's own copy that answers "where can I walk" and the attack window.
type MatchView struct {
	View any             `json:"view"`
	Save json.RawMessage `json:"save"`
}

func (r *Room) viewFor(id string, now time.Time) RoomView {
	rv := RoomView{Type: "room", Room: r.Password, YourSide: r.seatOf(id), Playing: r.playing(), Winner: r.Winner, Matches: r.Matches, Max: MaxPeople, Seq: r.Seq, ThinkMs: r.think(now)}
	for s, x := range r.Seats {
		if m := r.member(x); m != nil {
			rv.Seats[s] = m.Name
		}
	}
	for _, m := range r.Members {
		p := PersonView{Name: m.Name, Side: r.seatOf(m.ID), Ready: m.Ready, HasTeam: len(m.Team) > 0, Online: m.Gone.IsZero(), You: m.ID == id}
		if i := slices.Index(r.Line, m.ID); i >= 0 {
			p.Line = i + 1
		}
		rv.People = append(rv.People, p)
	}
	if m := r.member(id); m != nil {
		rv.YourTeam = m.Team
	}
	if r.Game != nil {
		rv.Match = r.matchFor(rv.YourSide)
	}
	return rv
}

// matchFor is the match as one side sees it through the fog. Watchers (-1)
// see the whole board, but one round behind (owner, 2026-10-16): the board
// as it was when the turn before last began, so anything they tell a player
// is old news. When the match is over they see it as it is.
func (r *Room) matchFor(side int) *MatchView {
	if side < 0 && r.playing() && len(r.Snaps) > 0 {
		return r.Snaps[max(0, len(r.Snaps)-3)]
	}
	return r.liveFor(side)
}

// liveFor is the match as it is now, as one side sees it.
func (r *Room) liveFor(side int) *MatchView {
	save, err := r.Game.SaveFor(side)
	if err != nil {
		return nil
	}
	return &MatchView{View: r.Game.ViewFor(side), Save: save}
}

// resultFor is one action's outcome as one person sees it.
func (r *Room) resultFor(id string, ev []action.Event, sideBefore int, now time.Time) ResultView {
	side := r.seatOf(id)
	if side >= 0 && r.Before[side] != nil {
		ev = r.Game.EventsFor(side, ev, r.Before[side])
	}
	if side < 0 && r.playing() { // watchers don't see moves live: their board jumps a round behind
		ev = nil
	}
	return ResultView{Type: "result", Events: ev, SideBefore: sideBefore, Seq: r.Seq, ThinkMs: r.think(now), Match: r.matchFor(side)}
}

// ResultView is one action's outcome, sent to everyone in the room.
type ResultView struct {
	Type       string         `json:"type"` // "result"
	Events     []action.Event `json:"events"`
	SideBefore int            `json:"sideBefore"`
	Seq        int            `json:"seq"`
	ThinkMs    int            `json:"thinkMs"`
	Match      *MatchView     `json:"match"`
}
