// The online server (online step 3, owner 2026-10-16): rooms of up to 4
// friends who share a password, playing versus matches on the same referee
// the phones use. Phones send actions over plain HTTP and get everything that
// happens in their room over one Server-Sent Events stream. It also serves the
// game's pages, so friends need only one link.
//
//	go run ./adapters/server        (PORT, default 8080; SITE, default ./site)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
)

// Server holds every room. One lock keeps it simple: a few friends per room
// and a handful of rooms.
type Server struct {
	mu    sync.Mutex
	c     *content.Content
	rooms map[string]*Room
	who   map[string]*Room // member ID -> their room
	now   func() time.Time
}

func newServer(c *content.Content) *Server {
	return &Server{c: c, rooms: map[string]*Room{}, who: map[string]*Room{}, now: time.Now}
}

func main() {
	c, err := content.Default()
	if err != nil {
		log.Fatal(err)
	}
	s := newServer(c)
	go func() {
		for range time.Tick(250 * time.Millisecond) {
			s.tick()
		}
	}()
	site := os.Getenv("SITE")
	if site == "" {
		site = "site"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := s.routes()
	mux.Handle("/", http.FileServer(http.Dir(site)))
	log.Printf("Crusader Tactics server on :%s (pages from %s)", port, site)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/join", s.post(s.join))
	mux.HandleFunc("/api/seat", s.post(s.seat))
	mux.HandleFunc("/api/team", s.post(s.team))
	mux.HandleFunc("/api/again", s.post(s.again))
	mux.HandleFunc("/api/act", s.post(s.act))
	mux.HandleFunc("/api/leave", s.post(s.leave))
	mux.HandleFunc("/api/stream", s.stream)
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) { cors(w); fmt.Fprint(w, "ok") })
	return mux
}

// req is every POST body: who is asking, and what for.
type req struct {
	Token  string        `json:"token"`
	Room   string        `json:"room"`
	Name   string        `json:"name"`
	Side   int           `json:"side"`
	Team   []engine.Pick `json:"team"`
	Action action.Action `json:"action"`
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

// post wraps an API call: CORS, JSON in, {"ok":true,...} or {"error":...} out.
func (s *Server) post(f func(req) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			return
		}
		var q req
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&q); err != nil {
			reply(w, nil, fmt.Errorf("bad request"))
			return
		}
		out, err := f(q)
		reply(w, out, err)
	}
}

func reply(w http.ResponseWriter, out any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if out == nil {
		out = map[string]bool{"ok": true}
	}
	_ = json.NewEncoder(w).Encode(out)
}

// cleanPassword makes "Dragons 7 " and "dragons7" the same room.
func cleanPassword(p string) string {
	return strings.ToLower(strings.Join(strings.Fields(p), ""))
}

func (s *Server) join(q req) (any, error) {
	pw := cleanPassword(q.Room)
	if len(pw) < 3 || len(pw) > 32 {
		return nil, fmt.Errorf("Use a room password of 3 to 32 letters or numbers.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if q.Token != "" {
		if r := s.who[q.Token]; r != nil && r.Password == pw && r.member(q.Token) != nil {
			s.roomChanged(r)
			return map[string]string{"token": q.Token, "room": pw}, nil
		}
	}
	r := s.rooms[pw]
	if r == nil {
		r = NewRoom(s.c, pw)
		s.rooms[pw] = r
	}
	m, err := r.Join(strings.TrimSpace(q.Name))
	if err != nil {
		return nil, err
	}
	m.Gone = s.now() // until their stream connects
	s.who[m.ID] = r
	s.roomChanged(r)
	return map[string]string{"token": m.ID, "room": pw}, nil
}

// in finds the caller's room; the lock must be held.
func (s *Server) in(token string) (*Room, error) {
	r := s.who[token]
	if r == nil || r.member(token) == nil {
		return nil, errNoOne
	}
	return r, nil
}

func (s *Server) seat(q req) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.in(q.Token)
	if err != nil {
		return nil, err
	}
	if err := r.Sit(q.Token, q.Side); err != nil {
		return nil, err
	}
	s.roomChanged(r)
	return nil, nil
}

func (s *Server) team(q req) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.in(q.Token)
	if err != nil {
		return nil, err
	}
	if err := r.Pick(q.Token, q.Team, s.now()); err != nil {
		return nil, err
	}
	s.roomChanged(r)
	return nil, nil
}

func (s *Server) again(q req) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.in(q.Token)
	if err != nil {
		return nil, err
	}
	if err := r.Again(q.Token, s.now()); err != nil {
		return nil, err
	}
	s.roomChanged(r)
	return nil, nil
}

func (s *Server) act(q req) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.in(q.Token)
	if err != nil {
		return nil, err
	}
	if r.Game == nil {
		return nil, errNoMatch
	}
	side := r.Game.B.Side
	ev, err := r.Act(q.Token, q.Action, s.now())
	if err != nil {
		return nil, err
	}
	s.result(r, ev, side)
	return nil, nil
}

func (s *Server) leave(q req) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.in(q.Token)
	if err != nil {
		return nil, nil
	}
	s.drop(r, q.Token)
	return nil, nil
}

// drop takes a member out of their room for good.
func (s *Server) drop(r *Room, id string) {
	r.Leave(id)
	if r.member(id) == nil {
		delete(s.who, id)
	}
	if len(r.Members) == 0 {
		delete(s.rooms, r.Password)
		return
	}
	s.roomChanged(r)
}

// tick ends turns and matches whose time ran out, and lets go of people who
// closed the game more than a minute ago.
func (s *Server) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for _, r := range s.rooms {
		if r.playing() {
			side := r.Game.B.Side
			if ev := r.Tick(now); ev != nil {
				s.result(r, ev, side)
			} else if now.Sub(r.ping) >= time.Second {
				r.ping = now
				s.clocks(r, now)
			}
		}
		for _, m := range append([]*Member(nil), r.Members...) {
			if !m.Gone.IsZero() && now.Sub(m.Gone) > time.Minute {
				s.drop(r, m.ID)
			}
		}
	}
}

// ---------- sending to the phones ----------

func send(m *Member, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case m.out <- b:
	default: // a phone that stopped reading misses this; the next room message catches it up
	}
}

func (s *Server) roomChanged(r *Room) {
	for _, m := range r.Members {
		send(m, r.viewFor(m.ID, s.now()))
	}
}

func (s *Server) result(r *Room, ev []action.Event, sideBefore int) {
	now := s.now()
	r.ping = now
	for _, m := range r.Members {
		send(m, r.resultFor(m.ID, ev, sideBefore, now))
	}
	s.roomChanged(r)
}

func (s *Server) clocks(r *Room, now time.Time) {
	v := r.Game.B.Versus
	msg := ClockView{Type: "clock", Seq: r.Seq, ThinkMs: r.think(now), Side: r.Game.B.Side, BankMs: v.BankMs}
	for _, m := range r.Members {
		send(m, msg)
	}
}

// stream is one phone's live line to its room (Server-Sent Events).
func (s *Server) stream(w http.ResponseWriter, hr *http.Request) {
	cors(w)
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	id := hr.URL.Query().Get("token")
	s.mu.Lock()
	r, err := s.in(id)
	var m *Member
	if err == nil {
		m = r.member(id)
		m.Gone = time.Time{}
		send(m, r.viewFor(id, s.now()))
		s.roomChanged(r)
	}
	s.mu.Unlock()
	if m == nil {
		http.Error(w, "not in a room", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	keep := time.NewTicker(15 * time.Second)
	defer keep.Stop()
	for {
		select {
		case b := <-m.out:
			fmt.Fprintf(w, "data: %s\n\n", b)
			fl.Flush()
		case <-keep.C:
			fmt.Fprint(w, ": keep-alive\n\n")
			fl.Flush()
		case <-hr.Context().Done():
			s.mu.Lock()
			if r.member(id) == m {
				m.Gone = s.now()
				s.roomChanged(r)
			}
			s.mu.Unlock()
			return
		}
	}
}
