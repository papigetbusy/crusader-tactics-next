package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// View is the board as the screen sees it: a read-only copy, so the screen
// can never change a rule by accident.
type View struct {
	N       int        `json:"n"`
	Ground  []string   `json:"ground"`
	Turn    int        `json:"turn"`
	Side    int        `json:"side"`
	Outcome string     `json:"outcome"`
	Units   []UnitView `json:"units"`
}

// UnitView is one unit as the screen sees it.
type UnitView struct {
	ID       int            `json:"id"`
	Team     int            `json:"team"`
	Class    string         `json:"class"`
	X        int            `json:"x"`
	Y        int            `json:"y"`
	HP       int            `json:"hp"`
	MaxHP    int            `json:"maxHp"`
	Atk      int            `json:"atk"`
	Def      int            `json:"def"`
	Mov      int            `json:"mov"`
	Level    int            `json:"level"`
	Element  string         `json:"element,omitempty"`
	Boss     bool           `json:"boss,omitempty"`
	Elite    bool           `json:"elite,omitempty"`
	Statuses map[string]int `json:"statuses,omitempty"`
	Done     bool           `json:"done"`
	CanMove  bool           `json:"canMove"`
	Skills   []SkillView    `json:"skills"`
}

// SkillView is one skill on a unit's bar.
type SkillView struct {
	Name     string  `json:"name"`
	Text     string  `json:"text,omitempty"`
	MinRange int     `json:"minRange"`
	MaxRange int     `json:"maxRange"`
	Power    float64 `json:"power"`
	Element  string  `json:"element,omitempty"`
	Area     int     `json:"area,omitempty"`
	Heal     bool    `json:"heal,omitempty"`
	Support  bool    `json:"support,omitempty"`
	Wait     int     `json:"wait"` // turns until ready (0 = ready)
}

// View returns the current board.
func (g *Game) View() View {
	b := g.B
	v := View{N: b.N, Ground: make([]string, b.N), Turn: b.Turn, Side: b.Side, Outcome: b.Outcome}
	for y := 0; y < b.N; y++ {
		row := make([]byte, b.N)
		for x := range row {
			row[x] = board.GroundAt(b, x, y)
		}
		v.Ground[y] = string(row)
	}
	for _, u := range b.Units {
		uv := UnitView{ID: u.ID, Team: u.Team, Class: u.Class, X: u.X, Y: u.Y, HP: u.HP, MaxHP: u.MaxHP,
			Atk: u.Atk, Def: u.Def, Mov: u.Mov, Level: u.Level, Element: u.Element, Boss: u.Boss, Elite: u.Elite,
			Done: u.Done, CanMove: !u.Done && board.CanMove(u), Statuses: map[string]int{}}
		for k, n := range u.Status {
			if n > 0 {
				uv.Statuses[k] = n
			}
		}
		for i, s := range u.Skills {
			uv.Skills = append(uv.Skills, SkillView{Name: s.Name, Text: s.Text, MinRange: s.MinRange, MaxRange: s.MaxRange,
				Power: s.Power, Element: s.Element, Area: s.Area, Heal: s.Heal, Support: s.Support, Wait: u.Cooldowns[i]})
		}
		v.Units = append(v.Units, uv)
	}
	return v
}

// Reach is where a hero may walk now (empty if it cannot move).
func (g *Game) Reach(unit int) []board.Tile {
	u := g.B.Unit(unit)
	if u == nil || u.Team != state.Heroes || u.Done || !board.CanMove(u) || g.B.Outcome != state.Ongoing {
		return nil
	}
	return board.Reach(g.C, g.B, u)
}

// Targets are the unit IDs a hero's skill can be used on now.
func (g *Game) Targets(unit, slot int) []int {
	u := g.B.Unit(unit)
	if u == nil || u.Done || slot < 0 || slot >= len(u.Skills) || u.Cooldowns[slot] > 0 {
		return nil
	}
	var out []int
	for _, t := range combat.Targets(g.C, g.B, u, u.Skills[slot]) {
		out = append(out, t.ID)
	}
	return out
}

// Forecast is the preview of a skill use. For attacks it runs the same
// damage math as the real hit; for heals it gives the amount restored.
type Forecast struct {
	combat.Forecast
	Heal int `json:"heal,omitempty"`
}

// Preview forecasts a skill use without changing anything.
func (g *Game) Preview(unit, slot, target int) (Forecast, bool) {
	u, v := g.B.Unit(unit), g.B.Unit(target)
	if u == nil || v == nil || slot < 0 || slot >= len(u.Skills) {
		return Forecast{}, false
	}
	s := u.Skills[slot]
	if s.Heal {
		h := int(combat.Round(combat.HealAmount(g.C, u, s)))
		return Forecast{Forecast: combat.Forecast{HitChance: 100}, Heal: min(h, v.MaxHP-v.HP)}, true
	}
	if s.Support {
		return Forecast{Forecast: combat.Forecast{HitChance: 100}}, true
	}
	_, classBase, _ := g.C.SkillFor(u.Team, u.Class, s.Name)
	return Forecast{Forecast: combat.Preview(g.C, g.B, combat.Attack{By: u, Skill: s, ClassBase: classBase, Target: v})}, true
}
