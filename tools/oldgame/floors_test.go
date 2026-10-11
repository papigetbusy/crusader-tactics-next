package oldgame

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

const floorRecordings = "../../testdata/oldgame/recordings/floors7-recordings.json.gz"

// recFloor is one floor the old game built (Phase 7a recorder).
type recFloor struct {
	Cfg struct {
		Seed    uint32 `json:"seed"`
		Floor   int    `json:"floor"`
		Cleared bool   `json:"cleared"`
	} `json:"cfg"`
	Party []struct {
		Lv   int                `json:"lv"`
		Gear map[string]RecItem `json:"gear"`
	} `json:"party"`
	N   int      `json:"n"`
	Bio string   `json:"bio"`
	Ter []string `json:"ter"`
	Hgt []string `json:"hgt"`
	Obj *struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"obj"`
	Rein  bool `json:"rein"`
	Shop  bool `json:"shop"`
	L     int  `json:"L"`
	Units []struct {
		T   int      `json:"t"`
		N   string   `json:"n"`
		M   *int     `json:"m"`
		E   string   `json:"e"`
		Lv  int      `json:"lv"`
		Mx  int      `json:"mx"`
		Atk int      `json:"atk"`
		Def int      `json:"def"`
		Mov int      `json:"mov"`
		X   int      `json:"x"`
		Y   int      `json:"y"`
		Bo  int      `json:"bo"`
		El  int      `json:"el"`
		Al  int      `json:"al"`
		W   *string  `json:"w"`
		Sk  []string `json:"sk"`
	} `json:"units"`
	Chests  [][]any `json:"chests"`
	Shuffle *struct {
		Draws int      `json:"draws"`
		Order [][2]int `json:"order"`
	} `json:"shuffle"`
}

// playedBack is the old game's own enemy-start order: the old shuffle sorted
// with a random comparison, which no other sort reproduces, so the check
// takes its result from the recording and uses up as many draws as it did.
func playedBack(f *recFloor, t *testing.T) floor.Shuffler {
	return func(tiles [][2]int, r *rng.Stream) {
		if f.Shuffle == nil {
			return
		}
		a, b := slices.Clone(tiles), slices.Clone(f.Shuffle.Order)
		less := func(p, q [2]int) int { return (p[1]*99 + p[0]) - (q[1]*99 + q[0]) }
		slices.SortFunc(a, less)
		slices.SortFunc(b, less)
		if !slices.Equal(a, b) {
			t.Errorf("seed %d floor %d: Go's monster start tiles differ from the old game's", f.Cfg.Seed, f.Cfg.Floor)
		}
		copy(tiles, f.Shuffle.Order)
		for range f.Shuffle.Draws {
			r.Next()
		}
	}
}

// TestOldGameFloors builds every recorded floor in Go from the same run seed,
// floor number and team, and checks it is the old game's floor: size, ground,
// biome, throne and reinforcement floors, monster level, every monster (kind,
// numbers, element, skills, place), chests and the companion spirit.
func TestOldGameFloors(t *testing.T) {
	fh, err := os.Open(floorRecordings)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	gz, err := gzip.NewReader(fh)
	if err != nil {
		t.Fatal(err)
	}
	var floors []recFloor
	if err := json.NewDecoder(gz).Decode(&floors); err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var bad problems
	same, monsters, chests := 0, 0, 0
	for i := range floors {
		f := &floors[i]
		var team []floor.Hero
		for _, p := range f.Party {
			h := floor.Hero{Level: p.Lv, Gear: map[string]content.Item{}}
			for sl, it := range p.Gear {
				h.Gear[sl] = it.Item()
			}
			team = append(team, h)
		}
		got := floor.Build(c, f.Cfg.Seed, f.Cfg.Floor, team, floor.Memory{Cleared: f.Cfg.Cleared}, playedBack(f, t))
		d := floorDifferences(got, f)
		if len(d) == 0 {
			same++
			monsters += len(got.Enemies)
			chests += len(got.Chests)
			continue
		}
		bad.add("seed %d floor %d: %s (and %d more differences)", f.Cfg.Seed, f.Cfg.Floor, d[0], len(d)-1)
	}
	t.Logf("%d of %d floors built exactly as the old game built them, with %d monsters and %d chests", same, len(floors), monsters, chests)
	for _, p := range bad {
		t.Error(p)
	}
}

func floorDifferences(g floor.Floor, f *recFloor) []string {
	var out []string
	if g.N != f.N || !slices.Equal(g.Ground, f.Ter) {
		return []string{fmt.Sprintf("map %dx%d %v, old game %dx%d %v", g.N, g.N, g.Ground, f.N, f.N, f.Ter)}
	}
	for y, row := range g.Heights {
		if s := fmt.Sprint(row); s != fmt.Sprint(parseInts(f.Hgt[y])) {
			out = append(out, fmt.Sprintf("heights of row %d: %s, old game %s", y, s, f.Hgt[y]))
		}
	}
	if g.Biome != f.Bio || g.Shop != f.Shop || g.Reinforce != f.Rein || (g.Throne != nil) != (f.Obj != nil) {
		out = append(out, fmt.Sprintf("biome %s, shop %v, reinforcements %v, throne %v; old game %s, %v, %v, %v",
			g.Biome, g.Shop, g.Reinforce, g.Throne != nil, f.Bio, f.Shop, f.Rein, f.Obj != nil))
	}
	if f.Shop {
		return out
	}
	if g.Level != f.L {
		out = append(out, fmt.Sprintf("monster level %d, old game %d", g.Level, f.L))
	}
	var old []string
	for _, u := range f.Units {
		if u.T == state.Enemies {
			old = append(old, fmt.Sprintf("%s Lv %d %d/%d/%d/%d at %d,%d %s boss %d elite %d ally %d weapon %s skills %v",
				u.N, u.Lv, u.Mx, u.Atk, u.Def, u.Mov, u.X, u.Y, u.E, u.Bo, u.El, u.Al, deref(u.W), u.Sk))
		}
	}
	var mine []string
	for _, u := range g.Enemies {
		var sk []string
		for _, s := range u.Skills {
			sk = append(sk, s.Name)
		}
		b := func(v bool) int {
			if v {
				return 1
			}
			return 0
		}
		mine = append(mine, fmt.Sprintf("%s Lv %d %d/%d/%d/%d at %d,%d %s boss %d elite %d ally %d weapon %s skills %v",
			u.Class, u.Level, u.MaxHP, u.Atk, u.Def, u.Mov, u.X, u.Y, u.Element, b(u.Boss), b(u.Elite), b(u.Ally), u.Weapon, sk))
	}
	for i := range max(len(old), len(mine)) {
		if i >= len(old) || i >= len(mine) || old[i] != mine[i] {
			a, b := "nothing", "nothing"
			if i < len(mine) {
				a = mine[i]
			}
			if i < len(old) {
				b = old[i]
			}
			out = append(out, fmt.Sprintf("monster %d: %s; old game %s", i+1, a, b))
			break
		}
	}
	var oc []string
	for _, ch := range f.Chests {
		oc = append(oc, fmt.Sprintf("%v,%v %v", ch[0], ch[1], ch[2]))
	}
	var gc []string
	for _, ch := range g.Chests {
		gc = append(gc, fmt.Sprintf("%d,%d %s", ch.X, ch.Y, ch.Kind))
	}
	if fmt.Sprint(oc) != fmt.Sprint(gc) {
		out = append(out, fmt.Sprintf("chests %v, old game %v", gc, oc))
	}
	return out
}

func parseInts(s string) []int {
	var out []int
	n, neg, any := 0, false, false
	for _, ch := range s + "," {
		switch {
		case ch == '-':
			neg = true
		case ch >= '0' && ch <= '9':
			n, any = n*10+int(ch-'0'), true
		case ch == ',' && any:
			if neg {
				n = -n
			}
			out = append(out, n)
			n, neg, any = 0, false, false
		}
	}
	return out
}
