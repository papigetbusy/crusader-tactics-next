// Command practice writes data/practice.json: starting setups for practice
// fights, taken from real fights recorded in the old game. They stand in
// for the floor builder until it is ported (Phase 7).
//
//	go run ./tools/practice
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/tools/oldgame"
)

func main() {
	fights, err := oldgame.Load("testdata/oldgame/recordings/slice1-recordings.json.gz")
	if err != nil {
		log.Fatal(err)
	}
	var out []engine.Setup
	for _, want := range []int{1, 4, 5, 10, 15, 20} {
		for _, f := range fights {
			if !f.Usable() || f.Cfg.Floor != want || !hasEnemies(f) {
				continue
			}
			out = append(out, setupOf(f))
			break
		}
	}
	b, err := json.MarshalIndent(map[string]any{
		"_about":   "Practice fights: starting setups recorded from the old game, used until the floor builder is ported.",
		"practice": out}, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("data/practice.json", append(b, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}

func setupOf(f oldgame.Fight) engine.Setup {
	s := engine.Setup{Floor: f.Cfg.Floor, N: f.Init.N, Ground: f.Init.Ter}
	for _, row := range f.Init.Hgt {
		var r []int
		for _, v := range strings.Split(row, ",") {
			n, _ := strconv.Atoi(strings.TrimSpace(v))
			r = append(r, n)
		}
		s.Heights = append(s.Heights, r)
	}
	kind := ""
	for _, u := range f.Init.Snap.Units {
		s.Units = append(s.Units, engine.UnitSetup{Team: u.Team, Class: u.Class, Level: u.Lv, X: u.X, Y: u.Y,
			HP: u.MaxHP, Atk: u.Atk, Def: u.Def, Mov: u.Mov, Element: u.Elem, Boss: u.Boss != 0, Elite: u.Elite != 0})
		if u.Boss != 0 {
			kind = " (boss)"
		} else if u.Elite != 0 && kind == "" {
			kind = " (elite)"
		}
	}
	s.Name = fmt.Sprintf("Floor %d%s", f.Cfg.Floor, kind)
	return s
}

func hasEnemies(f oldgame.Fight) bool {
	for _, u := range f.Init.Snap.Units {
		if u.Team == 1 {
			return true
		}
	}
	return false
}
