// Package oldgame reads the Phase 1 answer key (the recorded fights from the
// old game) and turns recorded moments into referee state, so tests can check
// Go against what the old game actually did.
package oldgame

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Fight is one recorded fight.
type Fight struct {
	Cfg struct {
		Seed  int `json:"seed"`
		Floor int `json:"floor"`
	} `json:"cfg"`
	Init struct {
		N    int      `json:"N"`
		Ter  []string `json:"ter"`
		Hgt  []string `json:"hgt"`
		Snap Snap     `json:"snap"`
	} `json:"init"`
	Steps      []Step   `json:"steps"`
	Sites      []string `json:"sites"`
	PageErrors []any    `json:"pageErrors"`
}

// Step is one recorded player input and everything that followed it.
type Step struct {
	Kind string  `json:"kind"` // select, move, attack, wait, endturn
	ID   int     `json:"id"`
	Got  []int   `json:"got"`
	Ev   []Event `json:"ev"`
	Dice struct {
		G []uint32 `json:"g"` // gameplay rolls, as fraction * 2^32
		S []int    `json:"s"` // which code site rolled each one (index into Fight.Sites)
	} `json:"dice"`
	Snap Snap `json:"snap"`
}

// Event is one thing the old game reported happening.
type Event struct {
	K    string   `json:"k"`
	ID   int      `json:"id"`
	Side int      `json:"side"` // startPhase/endPhase: whose turn
	From []int    `json:"from"`
	To   []int    `json:"to"`
	S    string   `json:"s"` // act/hit: skill name
	V    int      `json:"v"` // act: target unit
	A    *HitUnit `json:"a"` // hit: attacker just before the hit
	B    *HitUnit `json:"b"` // hit: target just before the hit
	G0   int      `json:"g0"`
	G1   int      `json:"g1"` // hit: dice used are Step.Dice[G0:G1]
	R    *struct {
		Dm   int     `json:"dm"`
		M    float64 `json:"m"`
		Miss int     `json:"miss"`
	} `json:"r"`
	BHP *int `json:"bhp"` // hit: target HP right after
}

// HitUnit is a unit as recorded at the moment of a hit.
type HitUnit struct {
	ID      int            `json:"id"`
	Team    int            `json:"t"`
	Class   string         `json:"n"`
	X       int            `json:"x"`
	Y       int            `json:"y"`
	HP      int            `json:"hp"`
	MaxHP   int            `json:"mx"`
	Atk     int            `json:"atk"`
	Def     int            `json:"def"`
	Mov     int            `json:"mov"`
	Lv      int            `json:"lv"`
	Shield  int            `json:"sh"`
	Element *string        `json:"el"`
	Weapon  *string        `json:"w"`
	RB      int            `json:"rb"`
	Spec    *string        `json:"sp"`
	Defend  any            `json:"df"`
	Boss    any            `json:"bo"`
	Elite   any            `json:"elt"`
	St      map[string]int `json:"st"`
	Af      map[string]any `json:"af"`
	Ter     string         `json:"ter"`
}

// Snap is the state of every unit after a step.
type Snap struct {
	Turn   int        `json:"turn"`
	Units  []SnapUnit `json:"u"`
	Fallen []int      `json:"fallen"` // heroes who can be revived, oldest first (Phase 5 recordings)
	Seized []int      `json:"seized"` // the throne tile on seize-the-throne floors (Phase 5 recordings)
}

// SnapUnit is one unit in a Snap.
type SnapUnit struct {
	ID    int            `json:"id"`
	Team  int            `json:"t"`
	Class string         `json:"n"`
	X     int            `json:"x"`
	Y     int            `json:"y"`
	HP    int            `json:"hp"`
	MaxHP int            `json:"mx"`
	Atk   int            `json:"atk"`
	Def   int            `json:"def"`
	Mov   int            `json:"mov"`
	Lv    int            `json:"lv"`
	Moved int            `json:"mv"`
	Acted int            `json:"ac"`
	Done  int            `json:"dn"`
	St    map[string]int `json:"st"`
	Elem  string         `json:"e"`
	Boss  int            `json:"bo"`
	Elite int            `json:"el"`
	Cd    []*int         `json:"cd"` // one per skill slot; null = never used
	Def2  int            `json:"df"`
	Sh    int            `json:"sh"`
	Pot   *int           `json:"pot"` // potions left (heroes; Phase 5 recordings only)
}

// Load reads the recorded fights file.
func Load(path string) ([]Fight, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	var fights []Fight
	return fights, json.NewDecoder(z).Decode(&fights)
}

// Usable is false for the 3 fights where the old game itself crashed
// (the line-2067 bug); tests skip those.
func (f *Fight) Usable() bool { return len(f.PageErrors) == 0 }

// Battle turns a recorded moment into referee state.
func (f *Fight) Battle(s Snap) *state.Battle {
	b := &state.Battle{N: f.Init.N, Ground: f.Init.Ter, Turn: s.Turn}
	for _, row := range f.Init.Hgt {
		var r []int
		for _, v := range strings.Split(row, ",") {
			n, _ := strconv.Atoi(strings.TrimSpace(v))
			r = append(r, n)
		}
		b.Heights = append(b.Heights, r)
	}
	for _, u := range s.Units {
		st := map[string]int{}
		for k, v := range u.St {
			st[k] = v
		}
		b.Units = append(b.Units, &state.Unit{
			ID: u.ID, Team: u.Team, Class: u.Class, X: u.X, Y: u.Y,
			HP: u.HP, MaxHP: u.MaxHP, Atk: u.Atk, Def: u.Def, Mov: u.Mov, Level: u.Lv,
			Status: st, Moved: u.Moved != 0, Acted: u.Acted != 0, Done: u.Done != 0,
		})
	}
	return b
}
