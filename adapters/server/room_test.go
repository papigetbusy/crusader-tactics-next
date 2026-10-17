package main

import (
	"testing"
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

var t0 = time.Date(2026, 10, 16, 20, 0, 0, 0, time.UTC)

// team is 8 classes with their first ascension.
func team(c *content.Content, skip int) []engine.Pick {
	var out []engine.Pick
	for i, cl := range engine.VersusRules(c).Classes {
		if i == skip || i == skip+1 {
			continue
		}
		out = append(out, engine.Pick{Class: cl.Class, Ascension: cl.Ascensions[0].Name})
	}
	return out[:8]
}

// ready is a room where Ann (Blue) and Bo (Red) have started a match.
func ready(t *testing.T) (*Room, *Member, *Member) {
	c := content.MustDefault()
	r := NewRoom(c, "dragons")
	a, _ := r.Join("Ann")
	b, _ := r.Join("Bo")
	if r.Sit(a.ID, 0) != nil || r.Sit(b.ID, 1) != nil {
		t.Fatal("could not sit")
	}
	if err := r.Pick(a.ID, team(c, 0), t0); err != nil {
		t.Fatal(err)
	}
	if err := r.Pick(b.ID, team(c, 2), t0); err != nil {
		t.Fatal(err)
	}
	if !r.playing() {
		t.Fatal("the match did not start")
	}
	return r, a, b
}

// runOut lets the clocks run until the match ends.
func runOut(r *Room, from time.Time) time.Time {
	now := from
	for i := 0; i < 40 && r.playing(); i++ {
		now = now.Add(31 * time.Second)
		r.Tick(now)
	}
	return now
}

func TestRooms(t *testing.T) {
	c := content.MustDefault()
	t.Run("A room holds at most 4 people", func(t *testing.T) {
		r := NewRoom(c, "x")
		for i := 0; i < 4; i++ {
			if _, err := r.Join("p"); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := r.Join("fifth"); err != errFull {
			t.Fatalf("fifth person: %v", err)
		}
	})
	t.Run("Two people take Blue and Red, and a side that is taken can't be taken again", func(t *testing.T) {
		r := NewRoom(c, "x")
		a, _ := r.Join("Ann")
		b, _ := r.Join("Bo")
		if err := r.Sit(a.ID, 0); err != nil {
			t.Fatal(err)
		}
		if err := r.Sit(b.ID, 0); err != errSeat {
			t.Fatalf("second Blue: %v", err)
		}
		if err := r.Sit(b.ID, 1); err != nil || len(r.Line) != 0 {
			t.Fatalf("Red: %v, line %v", err, r.Line)
		}
	})
	t.Run("The match starts once both sides have picked their teams, on the same referee", func(t *testing.T) {
		r, _, _ := ready(t)
		if r.Game.B.Versus == nil || len(r.Game.B.Units) < 16 || r.Game.B.Side != 0 {
			t.Fatalf("game %+v", r.Game.B.Versus)
		}
	})
	t.Run("A bad team is refused with the referee's reason", func(t *testing.T) {
		r := NewRoom(c, "x")
		a, _ := r.Join("Ann")
		_ = r.Sit(a.ID, 0)
		if err := r.Pick(a.ID, team(c, 0)[:3], t0); err == nil {
			t.Fatal("3 heroes were accepted")
		}
	})
	t.Run("Only the player to move can act, and the server charges the time they took", func(t *testing.T) {
		r, a, b := ready(t)
		if _, err := r.Act(b.ID, action.Action{Kind: action.EndTurn}, t0); err != errNotTurn {
			t.Fatalf("Red acting on Blue's turn: %v", err)
		}
		if _, err := r.Act(a.ID, action.Action{Kind: action.EndTurn}, t0.Add(7*time.Second)); err != nil {
			t.Fatal(err)
		}
		v := r.Game.B.Versus
		charged := 7000 - r.c.Versus.GraceSeconds*1000
		if r.Game.B.Side != 1 || v.BankMs[0] != r.c.Versus.BankSeconds*1000-charged || r.Seq != 1 {
			t.Fatalf("side %d, banks %v, seq %d", r.Game.B.Side, v.BankMs, r.Seq)
		}
	})
	t.Run("The first two seconds after a move are free", func(t *testing.T) {
		r, a, _ := ready(t)
		if _, err := r.Act(a.ID, action.Action{Kind: action.EndTurn}, t0.Add(time.Duration(r.c.Versus.GraceSeconds)*time.Second)); err != nil {
			t.Fatal(err)
		}
		if r.Game.B.Versus.BankMs[0] != r.c.Versus.BankSeconds*1000 {
			t.Fatalf("bank %d", r.Game.B.Versus.BankMs[0])
		}
	})
	t.Run("A long think does not end the turn; running the 5 minutes out loses", func(t *testing.T) {
		r, _, _ := ready(t)
		if ev := r.Tick(t0.Add(30 * time.Second)); ev != nil || r.Game.B.Side != 0 {
			t.Fatal("a short think ended the turn")
		}
		r.Tick(t0.Add(time.Duration(r.c.Versus.BankSeconds+r.c.Versus.GraceSeconds)*time.Second + time.Second))
		if r.playing() || r.Game.B.Outcome != state.Lost || r.Winner != 1 {
			t.Fatalf("outcome %v winner %d playing %v", r.Game.B.Outcome, r.Winner, r.playing())
		}
	})
	t.Run("Phones can't send clock actions: the server keeps the clocks", func(t *testing.T) {
		r, a, _ := ready(t)
		if _, err := r.Act(a.ID, action.Action{Kind: action.Clock, Ms: 999999}, t0); err != errClock {
			t.Fatalf("%v", err)
		}
	})
	t.Run("The winner stays on: the loser goes to the back of the line and the next in line takes their side", func(t *testing.T) {
		r, a, b := ready(t)
		cy, _ := r.Join("Cy")
		di, _ := r.Join("Di")
		runOut(r, t0)
		if r.playing() || r.Winner < 0 {
			t.Fatal("the match did not end")
		}
		win, lose := a, b
		if r.Winner == 1 {
			win, lose = b, a
		}
		if r.Seats[r.Winner] != win.ID || r.Seats[1-r.Winner] != cy.ID {
			t.Fatalf("seats %v", r.Seats)
		}
		if len(r.Line) != 2 || r.Line[0] != di.ID || r.Line[1] != lose.ID {
			t.Fatalf("line %v", r.Line)
		}
	})
	t.Run("With nobody waiting, the same two can play again with the same teams", func(t *testing.T) {
		r, a, b := ready(t)
		now := runOut(r, t0)
		if err := r.Again(a.ID, now); err != nil {
			t.Fatal(err)
		}
		if r.playing() {
			t.Fatal("started before both were ready")
		}
		if err := r.Again(b.ID, now); err != nil || !r.playing() || r.Matches != 2 {
			t.Fatalf("again: %v, playing %v, matches %d", err, r.playing(), r.Matches)
		}
	})
	t.Run("A player who leaves mid-match keeps their seat, so their clock runs out and they lose", func(t *testing.T) {
		r, a, _ := ready(t)
		r.Leave(a.ID)
		if r.Seats[0] != a.ID {
			t.Fatal("seat was freed mid-match")
		}
		runOut(r, t0)
		if r.Game.B.Outcome != state.Lost || r.Winner != 1 {
			t.Fatalf("outcome %v, winner %d", r.Game.B.Outcome, r.Winner)
		}
	})
	t.Run("Giving up ends the match at once and the other player wins", func(t *testing.T) {
		r, a, _ := ready(t)
		ev, err := r.Forfeit(a.ID, t0)
		if err != nil || r.playing() || r.Winner != 1 || r.Game.B.Outcome != state.Lost {
			t.Fatalf("err %v outcome %v winner %d playing %v", err, r.Game.B.Outcome, r.Winner, r.playing())
		}
		if len(ev) == 0 || ev[0].Kind != action.GaveUp || ev[0].Side != 0 {
			t.Fatalf("events %v", ev)
		}
		r2, _, b := ready(t)
		if _, err := r2.Forfeit(b.ID, t0); err != nil || r2.Winner != 0 || r2.Game.B.Outcome != state.Won {
			t.Fatalf("red give up: %v winner %d outcome %v", err, r2.Winner, r2.Game.B.Outcome)
		}
		r3, _, _ := ready(t)
		w, _ := r3.Join("Cy")
		if _, err := r3.Forfeit(w.ID, t0); err == nil || !r3.playing() {
			t.Fatalf("a watcher ended the match: %v playing %v", err, r3.playing())
		}
	})
	t.Run("A phone gone for 20 seconds loses, and a watcher does not end the match", func(t *testing.T) {
		r, a, _ := ready(t)
		w, err := r.Join("Cy")
		if err != nil {
			t.Fatal(err)
		}
		w.Gone = t0
		if _, _, ok := r.Away(t0.Add(time.Minute)); ok {
			t.Fatal("a watcher ended the match")
		}
		a.Gone = t0
		if _, _, ok := r.Away(t0.Add(19 * time.Second)); ok {
			t.Fatal("gave up before 20 seconds")
		}
		id, ev, ok := r.Away(t0.Add(21 * time.Second))
		if !ok || id != a.ID || r.playing() || r.Winner != 1 || len(ev) == 0 || ev[0].Kind != action.GaveUp {
			t.Fatalf("ok %v id %s winner %d playing %v events %v", ok, id, r.Winner, r.playing(), ev)
		}
		r.Leave(a.ID)
		if r.member(a.ID) != nil || r.seatOf(a.ID) >= 0 {
			t.Fatal("the player who left is still in the room")
		}
		if r.Seats[1] == "" || r.member(r.Seats[1]) == nil {
			t.Fatal("the winner was removed")
		}
	})
	t.Run("Through the fog each player is sent only the heroes they can see, and a watcher is sent everything", func(t *testing.T) {
		r, a, b := ready(t)
		cy, _ := r.Join("Cy")
		count := func(id string) (mine, theirs int) {
			v := r.viewFor(id, t0).Match.View.(engine.View)
			for _, u := range v.Units {
				if u.Team == r.seatOf(id) || r.seatOf(id) < 0 && u.Team == 0 {
					mine++
				} else {
					theirs++
				}
			}
			return
		}
		if m, th := count(a.ID); m < 8 || th != 0 {
			t.Fatalf("Ann sees %d of hers and %d of Red's", m, th)
		}
		if m, th := count(b.ID); m < 8 || th != 0 {
			t.Fatalf("Bo sees %d of his and %d of Blue's", m, th)
		}
		if m, th := count(cy.ID); m < 8 || th < 8 {
			t.Fatalf("Cy sees %d and %d", m, th)
		}
	})
	t.Run("A whole match replays from its seeds and its actions, so the server's record is exact", func(t *testing.T) {
		r, a, b := ready(t)
		now := t0
		for i := 0; i < 6 && r.playing(); i++ {
			now = now.Add(5 * time.Second)
			id := a.ID
			if r.Game.B.Side == 1 {
				id = b.ID
			}
			if _, err := r.Act(id, action.Action{Kind: action.EndTurn}, now); err != nil {
				t.Fatal(err)
			}
		}
		want, _ := r.Game.Save()
		g, err := engine.NewVersus(c, r.Seed, r.Dice)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range r.Log {
			if _, err := g.Apply(a); err != nil {
				t.Fatal(err)
			}
		}
		got, _ := g.Save()
		if len(r.Log) != 6 || string(got) != string(want) {
			t.Fatalf("log %d, same %v", len(r.Log), string(got) == string(want))
		}
	})
}

func TestPasswords(t *testing.T) {
	t.Run("A room password ignores spaces and capital letters", func(t *testing.T) {
		if cleanPassword(" Dragons 7 ") != "dragons7" {
			t.Fatal(cleanPassword(" Dragons 7 "))
		}
	})
}
