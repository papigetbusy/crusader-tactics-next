package main

// A room is one password shared by up to 4 friends (owner, 2026-10-16). Two of
// them take Blue and Red and play a versus match on the same referee the phone
// uses; everyone in the room watches; the others wait in line, and the winner
// stays on. The room never reads the clock itself: every call passes now, so
// tests can move time by hand.

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"slices"
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// MaxPeople is how many can be in one room (owner, 2026-10-16).
const MaxPeople = 4

// Member is one person in a room.
type Member struct {
	ID    string
	Name  string
	Team  []engine.Pick // the team they will play with (seated players)
	Ready bool          // happy to start with that team
	Gone  time.Time     // when their stream closed (zero while connected)
	out   chan []byte   // messages for their stream
}

// Room is everything one password shares.
type Room struct {
	Password string
	Members  []*Member
	Seats    [2]string // member IDs on Blue (0) and Red (1); "" when free
	Line     []string  // member IDs waiting, first in line first
	Game     *engine.Game
	Last     time.Time // when the player to move last had an accepted action (or the turn began)
	Free     bool      // the time since Last starts with the free seconds (see charge)
	Winner   int       // -1, or the side that won the last match
	Matches  int
	Log      []action.Action // the current match's actions, with the times charged
	Seed     engine.VersusSetup
	Dice     uint64          // the match's battle seed: Seed, Dice and Log replay it exactly
	Before   [2]map[int]bool // fog: what each side saw just before the last action
	Seq      int             // how many actions this match has accepted
	Snaps    []*MatchView    // the whole board at the start of each turn, for watchers (one round behind)
	grace    time.Duration
	ping     time.Time // when the clocks were last told to the phones
	c        *content.Content
}

// NewRoom makes an empty room.
func NewRoom(c *content.Content, password string) *Room {
	return &Room{Password: password, Winner: -1, grace: time.Duration(c.Versus.GraceSeconds) * time.Second, c: c}
}

var (
	errFull    = errors.New("This room is full (4 people). Ask a friend for another password.")
	errNoOne   = errors.New("You are not in this room any more. Join again.")
	errPlaying = errors.New("A match is being played. Wait for it to end.")
	errSeat    = errors.New("That side is taken.")
	errNoSeat  = errors.New("Take Blue or Red first.")
	errNoMatch = errors.New("No match is being played.")
	errNotTurn = errors.New("It is not your turn.")
	errClock   = errors.New("The server keeps the clocks.")
)

func (r *Room) member(id string) *Member {
	for _, m := range r.Members {
		if m.ID == id {
			return m
		}
	}
	return nil
}

// seatOf is the side a member sits on, or -1.
func (r *Room) seatOf(id string) int {
	for s, x := range r.Seats {
		if x == id && id != "" {
			return s
		}
	}
	return -1
}

func (r *Room) playing() bool { return r.Game != nil && r.Game.B.Outcome == state.Ongoing }

// Join adds a person. They wait in line until they take a side.
func (r *Room) Join(name string) (*Member, error) {
	if len(r.Members) >= MaxPeople {
		return nil, errFull
	}
	if name == "" {
		name = "Player"
	}
	if len([]rune(name)) > 16 {
		name = string([]rune(name)[:16])
	}
	m := &Member{ID: token(), Name: name, out: make(chan []byte, 64)}
	r.Members = append(r.Members, m)
	r.Line = append(r.Line, m.ID)
	return m, nil
}

// Leave takes a person out of the room. A player who leaves mid-match keeps
// their seat: their clock runs out and they lose, as if they walked away.
func (r *Room) Leave(id string) {
	if s := r.seatOf(id); s >= 0 && r.playing() {
		return
	}
	if s := r.seatOf(id); s >= 0 {
		r.Seats[s] = ""
	}
	r.Line = slices.DeleteFunc(r.Line, func(x string) bool { return x == id })
	r.Members = slices.DeleteFunc(r.Members, func(m *Member) bool { return m.ID == id })
}

// Sit puts a member on Blue (0) or Red (1); -1 stands them up into the line.
func (r *Room) Sit(id string, side int) error {
	m := r.member(id)
	if m == nil {
		return errNoOne
	}
	if r.playing() {
		return errPlaying
	}
	cur := r.seatOf(id)
	if side == cur {
		return nil
	}
	if side >= 0 && r.Seats[side] != "" {
		return errSeat
	}
	if cur >= 0 {
		r.Seats[cur] = ""
		r.Line = append(r.Line, id)
	}
	if side >= 0 {
		r.Seats[side] = id
		r.Line = slices.DeleteFunc(r.Line, func(x string) bool { return x == id })
	}
	m.Ready = false
	return nil
}

// Pick sets a seated member's team and marks them ready. The match starts
// once both sides are ready.
func (r *Room) Pick(id string, team []engine.Pick, now time.Time) error {
	m := r.member(id)
	if m == nil {
		return errNoOne
	}
	if r.playing() {
		return errPlaying
	}
	if r.seatOf(id) < 0 {
		return errNoSeat
	}
	if err := engine.CheckTeam(r.c, team); err != nil {
		return err
	}
	m.Team, m.Ready = slices.Clone(team), true
	r.maybeStart(now)
	return nil
}

// Again is a seated player keeping the same team for the next match.
func (r *Room) Again(id string, now time.Time) error {
	m := r.member(id)
	if m == nil {
		return errNoOne
	}
	if r.seatOf(id) < 0 {
		return errNoSeat
	}
	if len(m.Team) == 0 {
		return errors.New("Choose your heroes first.")
	}
	m.Ready = true
	r.maybeStart(now)
	return nil
}

func (r *Room) maybeStart(now time.Time) {
	if r.playing() {
		return
	}
	for _, id := range r.Seats {
		if m := r.member(id); m == nil || !m.Ready {
			return
		}
	}
	setup := engine.VersusSetup{MapSeed: uint32(seed())}
	for s, id := range r.Seats {
		setup.Teams[s] = r.member(id).Team
	}
	dice := seed()
	g, err := engine.NewVersus(r.c, setup, dice)
	if err != nil {
		return
	}
	r.Game, r.Seed, r.Dice, r.Last, r.Winner, r.Log, r.Seq, r.Free = g, setup, dice, now, -1, nil, 0, true
	r.Snaps = []*MatchView{r.liveFor(-1)}
	r.Matches++
	for _, id := range r.Seats {
		r.member(id).Ready = false
	}
}

// Act is the player to move doing something. The server charges the time
// since their last accepted action.
func (r *Room) Act(id string, a action.Action, now time.Time) ([]action.Event, error) {
	if r.member(id) == nil {
		return nil, errNoOne
	}
	if !r.playing() {
		return nil, errNoMatch
	}
	if a.Kind == action.Clock {
		return nil, errClock
	}
	if a.Kind == action.Forfeit {
		return nil, errors.New("Use give up.")
	}
	if r.seatOf(id) != r.Game.B.Side {
		return nil, errNotTurn
	}
	a.Ms = r.charge(now)
	return r.apply(a, now)
}

func (r *Room) apply(a action.Action, now time.Time) ([]action.Event, error) {
	before := [2]map[int]bool{r.Game.Visible(0), r.Game.Visible(1)}
	side := r.Game.B.Side
	ev, err := r.Game.Apply(a)
	if err != nil {
		return nil, err
	}
	r.Last, r.Before, r.Free = now, before, earnsFree(a.Kind)
	r.Seq++
	r.Log = append(r.Log, a)
	if r.playing() && r.Game.B.Side != side { // a new turn: keep the board as it is now for watchers
		r.Snaps = append(r.Snaps, r.liveFor(-1))
		if len(r.Snaps) > 3 {
			r.Snaps = r.Snaps[len(r.Snaps)-3:]
		}
	}
	if !r.playing() {
		r.over()
	}
	return ev, nil
}

// Tick runs a few times a second: when the player to move runs out of turn
// time or match time, the referee is told how long they took.
func (r *Room) Tick(now time.Time) []action.Event {
	if !r.playing() {
		return nil
	}
	v, side := r.Game.B.Versus, r.Game.B.Side
	used := r.charge(now)
	turnLeft := v.TurnMs == 0 || v.UsedMs+used < v.TurnMs
	if turnLeft && v.BankMs[side]-used > 0 {
		return nil
	}
	ev, err := r.apply(action.Action{Kind: action.Clock, Ms: used}, now)
	if err != nil {
		return nil
	}
	return ev
}

// over is the end of a match: the winner stays on; the loser goes to the back
// of the line and the first in line takes their side (owner, 2026-10-16).
// With nobody waiting, the same two play again.
func (r *Room) over() {
	r.Winner = 0
	if r.Game.B.Outcome == state.Lost {
		r.Winner = 1
	}
	lose := 1 - r.Winner
	loser := r.Seats[lose]
	if len(r.Line) == 0 || loser == "" {
		return
	}
	next := r.Line[0]
	r.Line = append(r.Line[1:], loser)
	r.Seats[lose] = next
	if m := r.member(loser); m != nil {
		m.Ready = false
	}
}

// ---------- random numbers for tokens and seeds (the server, not the referee) ----------

func token() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	const abc = "abcdefghijkmnpqrstuvwxyz23456789"
	out := make([]byte, len(b))
	for i, x := range b {
		out[i] = abc[int(x)%len(abc)]
	}
	return string(out)
}

func seed() uint64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}
