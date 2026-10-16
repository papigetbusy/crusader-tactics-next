package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

func call(t *testing.T, base, path string, body map[string]any) map[string]any {
	t.Helper()
	b, _ := json.Marshal(body)
	res, err := http.Post(base+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	if e, ok := out["error"]; ok {
		t.Fatalf("%s: %v", path, e)
	}
	return out
}

// listen reads a phone's stream into a channel of messages.
func listen(t *testing.T, ctx context.Context, base, tok string) chan map[string]any {
	req, _ := http.NewRequestWithContext(ctx, "GET", base+"/api/stream?token="+tok, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan map[string]any, 100)
	go func() {
		defer res.Body.Close()
		sc := bufio.NewScanner(res.Body)
		sc.Buffer(make([]byte, 1<<20), 1<<22)
		for sc.Scan() {
			if line := sc.Text(); strings.HasPrefix(line, "data: ") {
				var m map[string]any
				if json.Unmarshal([]byte(line[6:]), &m) == nil {
					ch <- m
				}
			}
		}
	}()
	return ch
}

// wait returns the first message for which ok is true.
func wait(t *testing.T, ch chan map[string]any, ok func(map[string]any) bool) map[string]any {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case m := <-ch:
			if ok(m) {
				return m
			}
		case <-deadline:
			t.Fatal("nothing came on the stream")
		}
	}
}

func TestOverTheInternet(t *testing.T) {
	t.Run("Two phones join one room by password, pick sides and teams, and both see the match and every turn", func(t *testing.T) {
		c := content.MustDefault()
		s := newServer(c)
		srv := httptest.NewServer(s.routes())
		defer srv.Close()
		ctx, stop := context.WithCancel(context.Background())
		defer stop()
		a := call(t, srv.URL, "/api/join", map[string]any{"room": "Dragons 7", "name": "Ann"})["token"].(string)
		b := call(t, srv.URL, "/api/join", map[string]any{"room": "dragons7", "name": "Bo"})["token"].(string)
		sa, sb := listen(t, ctx, srv.URL, a), listen(t, ctx, srv.URL, b)
		wait(t, sb, func(m map[string]any) bool { return m["type"] == "room" && len(m["people"].([]any)) == 2 })
		call(t, srv.URL, "/api/seat", map[string]any{"token": a, "side": 0})
		call(t, srv.URL, "/api/seat", map[string]any{"token": b, "side": 1})
		call(t, srv.URL, "/api/team", map[string]any{"token": a, "team": team(c, 0)})
		call(t, srv.URL, "/api/team", map[string]any{"token": b, "team": team(c, 2)})
		m := wait(t, sb, func(m map[string]any) bool { return m["type"] == "room" && m["playing"] == true })
		if m["match"].(map[string]any)["save"] == nil || m["yourSide"].(float64) != 1 {
			t.Fatalf("Bo's room: %v", m["yourSide"])
		}
		call(t, srv.URL, "/api/act", map[string]any{"token": a, "action": map[string]any{"kind": "endturn"}})
		r := wait(t, sb, func(m map[string]any) bool { return m["type"] == "result" })
		if r["match"].(map[string]any)["view"].(map[string]any)["side"].(float64) != 1 {
			t.Fatal("Bo did not see the turn pass to Red")
		}
		wait(t, sa, func(m map[string]any) bool { return m["type"] == "result" })
	})
	t.Run("A fifth person can't join a full room", func(t *testing.T) {
		s := newServer(content.MustDefault())
		srv := httptest.NewServer(s.routes())
		defer srv.Close()
		for i := 0; i < 4; i++ {
			call(t, srv.URL, "/api/join", map[string]any{"room": "full", "name": "p"})
		}
		b, _ := json.Marshal(map[string]any{"room": "full", "name": "late"})
		res, _ := http.Post(srv.URL+"/api/join", "application/json", bytes.NewReader(b))
		var out map[string]any
		_ = json.NewDecoder(res.Body).Decode(&out)
		if out["error"] == nil {
			t.Fatal("a fifth person got in")
		}
	})
}
