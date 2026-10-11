package engine

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// View is the board as the screen sees it: a read-only copy, so the screen
// can never change a rule by accident.
type View struct {
	N        int           `json:"n"`
	Ground   []string      `json:"ground"`
	Turn     int           `json:"turn"`
	Side     int           `json:"side"`
	Outcome  string        `json:"outcome"`
	Units    []UnitView    `json:"units"`
	Throne   *state.Tile   `json:"throne,omitempty"`
	LevelCap int           `json:"levelCap,omitempty"`
	Decision *DecisionView `json:"decision,omitempty"` // the first waiting choice
	Fields   []state.Field `json:"fields,omitempty"`
	Kegs     []state.Keg   `json:"kegs,omitempty"`   // Powder Kegs waiting to blow up
	Traps    int           `json:"traps,omitempty"`  // how many hidden traps (where is a secret only to monsters)
	Fallen   []string      `json:"fallen,omitempty"` // classes of fallen heroes, last to fall at the end
	Bag      []ItemView    `json:"bag,omitempty"`    // items carried but not worn
	Gold     int           `json:"gold"`
	Chests   []state.Chest `json:"chests,omitempty"`
}

// UnitView is one unit as the screen sees it.
type UnitView struct {
	ID        int            `json:"id"`
	Team      int            `json:"team"`
	Class     string         `json:"class"`
	X         int            `json:"x"`
	Y         int            `json:"y"`
	HP        int            `json:"hp"`
	MaxHP     int            `json:"maxHp"`
	Atk       int            `json:"atk"`
	Def       int            `json:"def"`
	Mov       int            `json:"mov"`
	Level     int            `json:"level"`
	XP        int            `json:"xp,omitempty"`     // heroes: experience toward the next level
	XPNext    int            `json:"xpNext,omitempty"` // heroes: experience the next level needs
	Form      string         `json:"form,omitempty"`   // heroes: the form's name ("Vanguard")
	Ascension string         `json:"ascension,omitempty"`
	Ability   *AbilityView   `json:"ability,omitempty"` // heroes: the class ability
	Decoy     bool           `json:"decoy,omitempty"`
	Companion bool           `json:"companion,omitempty"`
	Pet       bool           `json:"pet,omitempty"`
	Element   string         `json:"element,omitempty"`
	Boss      bool           `json:"boss,omitempty"`
	Elite     bool           `json:"elite,omitempty"`
	Statuses  map[string]int `json:"statuses,omitempty"`
	Enraged   bool           `json:"enraged,omitempty"`
	Potions   int            `json:"potions"`
	CanUndo   bool           `json:"canUndo"`
	Done      bool           `json:"done"`
	CanMove   bool           `json:"canMove"`
	Skills    []SkillView    `json:"skills"`
	Gear      *GearView      `json:"gear,omitempty"` // heroes: what they wear
	Bond      *state.Bond    `json:"bond,omitempty"` // heroes: their companion spirit
	Kind      string         `json:"kind,omitempty"` // monsters: what kind (Goblin, Slime...)
}

// AbilityView is a hero's class ability as the screen shows it.
type AbilityView struct {
	Name   string                 `json:"name"`
	Text   string                 `json:"text"`
	Wait   int                    `json:"wait"` // turns until ready
	Free   bool                   `json:"free,omitempty"`
	Target *content.AbilityTarget `json:"target,omitempty"`
}

// DecisionView is a waiting choice, with every option spelled out.
type DecisionView struct {
	Kind    string       `json:"kind"`
	Unit    int          `json:"unit"`
	Skill   string       `json:"skill,omitempty"`
	Text    string       `json:"text,omitempty"` // what the skill does
	Options []OptionView `json:"options"`
}

// OptionView is one button of a choice.
type OptionView struct {
	Choice string `json:"choice"`
	Slot   int    `json:"slot"`
	Label  string `json:"label"`
	Text   string `json:"text,omitempty"`
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
	Revive   bool    `json:"revive,omitempty"`
	Support  bool    `json:"support,omitempty"`
	Aim      string  `json:"aim,omitempty"`  // how it is aimed (content.Aim...); "" = tap a unit
	Free     bool    `json:"free,omitempty"` // does not use up the turn
	Wait     int     `json:"wait"`           // turns until ready (0 = ready)
}

// View returns the current board.
func (g *Game) View() View {
	b := g.B
	v := View{N: b.N, Ground: make([]string, b.N), Turn: b.Turn, Side: b.Side, Outcome: b.Outcome, Throne: b.Throne, LevelCap: b.LevelCap}
	for _, f := range b.Fallen {
		v.Fallen = append(v.Fallen, f.Class)
	}
	v.Fields, v.Kegs, v.Traps = b.Fields, b.Kegs, len(b.Traps)
	v.Decision = g.decisionView()
	v.Bag, v.Gold, v.Chests = g.bagView(), b.Gold, b.Chests
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
			Enraged: combat.Enraged(g.C, u), Potions: u.Potions, CanUndo: u.Undo != nil && !u.Done,
			Done: u.Done, CanMove: !u.Done && board.CanMove(u), Statuses: map[string]int{}}
		if u.Team == state.Heroes && !u.Summoned && !u.Companion && !u.Decoy {
			uv.XP, uv.XPNext, uv.Form = u.XP, stats.XPToNext(g.C, u.Level), stats.FormName(g.C, u.Class, u.Level)
			if u.Ascension != "" {
				uv.Form = u.Ascension
			}
			uv.Gear, uv.Bond = g.gearView(u), u.Bond
			if ab, ok := combat.AbilityOf(g.C, u); ok {
				wait := u.AbilityCooldown
				if ab.Free {
					wait = 0
				}
				uv.Ability = &AbilityView{Name: ab.Name, Text: ab.Text, Wait: wait, Free: ab.Free, Target: ab.Target}
			}
		}
		uv.Ascension, uv.Decoy, uv.Companion, uv.Pet, uv.Kind = u.Ascension, u.Decoy, u.Companion, u.Summoned, u.Kind
		for k, n := range u.Status {
			if n > 0 {
				uv.Statuses[k] = n
			}
		}
		for i, s := range u.Skills {
			uv.Skills = append(uv.Skills, SkillView{Name: s.Name, Text: s.Text, MinRange: s.MinRange, MaxRange: s.MaxRange,
				Power: s.Power, Element: s.Element, Area: s.Area, Heal: s.Heal, Revive: s.Revive, Support: s.Support, Aim: s.Aim, Free: s.Free, Wait: u.Cooldowns[i]})
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
	set, err := g.setFor(u, u.Skills[slot], 0)
	if err != nil {
		return nil
	}
	var out []int
	g.inSet(u, set, func() {
		for _, t := range combat.Targets(g.C, g.B, u, u.Skills[slot]) {
			out = append(out, t.ID)
		}
	})
	return out
}

// Forecast is the preview of a skill use. For attacks it runs the same
// damage math as the real hit; for heals it gives the amount restored.
type Forecast struct {
	combat.Forecast
	Heal    int    `json:"heal,omitempty"`
	Revives string `json:"revives,omitempty"` // the class a revive brings back
	Also    []int  `json:"also,omitempty"`    // other units an area or piercing skill also hits
}

// PreviewFrom forecasts a skill use as if the unit first walked to (x, y),
// for tap-to-attack. Nothing changes: the unit is put back afterwards.
func (g *Game) PreviewFrom(unit, slot, target, x, y int) (Forecast, bool) {
	u := g.B.Unit(unit)
	if u == nil {
		return Forecast{}, false
	}
	ox, oy := u.X, u.Y
	defer func() { u.X, u.Y = ox, oy }()
	u.X, u.Y = x, y
	return g.Preview(unit, slot, target)
}

// Preview forecasts a skill use without changing anything.
func (g *Game) Preview(unit, slot, target int) (f Forecast, ok bool) {
	u, v := g.B.Unit(unit), g.B.Unit(target)
	if u == nil || v == nil || slot < 0 || slot >= len(u.Skills) {
		return Forecast{}, false
	}
	set, err := g.setFor(u, u.Skills[slot], 0)
	if err != nil {
		return Forecast{}, false
	}
	g.inSet(u, set, func() { f, ok = g.preview(u, v, slot) })
	return f, ok
}

func (g *Game) preview(u, v *state.Unit, slot int) (Forecast, bool) {
	s := u.Skills[slot]
	if s.Heal {
		h := int(combat.Round(combat.HealAmount(g.C, u, s)))
		return Forecast{Forecast: combat.Forecast{HitChance: 100}, Heal: min(h, v.MaxHP-v.HP)}, true
	}
	if s.Revive && len(g.B.Fallen) > 0 {
		f := g.B.Fallen[len(g.B.Fallen)-1]
		return Forecast{Forecast: combat.Forecast{HitChance: 100}, Heal: int(combat.Round(float64(f.MaxHP) * g.C.Combat.ReviveShare)), Revives: f.Class}, true
	}
	if s.Support {
		return Forecast{Forecast: combat.Forecast{HitChance: 100}}, true
	}
	_, classBase, _ := g.C.SkillFor(u.Team, u.Class, s.Name)
	f := Forecast{Forecast: combat.Preview(g.C, g.B, combat.Attack{By: u, Skill: s, ClassBase: classBase, Target: v})}
	f.Also = combat.AlsoHit(g.B, u, s, v)
	return f, true
}

// Options lists the answers to the first waiting choice (empty if none).
func (g *Game) Options() []string { return stats.Options(g.C, g.B) }

// decisionView spells out the first waiting choice for the screen.
func (g *Game) decisionView() *DecisionView {
	if len(g.B.Decisions) == 0 {
		return nil
	}
	d := g.B.Decisions[0]
	dv := &DecisionView{Kind: d.Kind, Unit: d.Unit, Skill: d.Skill}
	if s, ok := g.C.SkillNamed(d.Skill); ok {
		dv.Text = s.Text
	}
	switch d.Kind {
	case state.DecideBond:
		dv.Text = "It joins every battle next to its hero and acts on its own. If it dies, it is gone for good."
	case state.DecideSwap:
		dv.Text = "A hero can only keep one companion spirit."
	}
	u := g.B.Unit(d.Unit)
	for _, o := range g.Options() {
		switch {
		case d.Kind == state.DecideAscend:
			a := g.C.Ascensions[o]
			dv.Options = append(dv.Options, OptionView{Choice: o, Label: o, Text: a.Bonus + ". " + a.Passive + " Ultimate: " + a.Ultimate + "."})
		case o == stats.ChooseReplace && u != nil:
			if s, _ := g.C.SkillNamed(d.Skill); s.Kind == "u" {
				dv.Options = append(dv.Options, OptionView{Choice: o, Label: "Replace " + u.Ultimate})
				continue
			}
			for i, n := range u.Equipped {
				dv.Options = append(dv.Options, OptionView{Choice: o, Slot: i, Label: "Replace " + n})
			}
		case d.Kind == state.DecideSwap && o == stats.ChooseKeep && u != nil && u.Bond != nil:
			dv.Options = append(dv.Options, OptionView{Choice: o, Label: "Keep the " + u.Bond.Kind})
		case d.Kind == state.DecideSwap && o == stats.ChooseTake:
			dv.Options = append(dv.Options, OptionView{Choice: o, Label: "Take the " + d.Skill})
		case o == stats.ChooseKeep:
			dv.Options = append(dv.Options, OptionView{Choice: o, Label: "Keep current skills"})
		default:
			dv.Options = append(dv.Options, OptionView{Choice: o, Label: "Continue"})
		}
	}
	return dv
}

// AbilityAim lists where a hero's ability can be aimed now: unit IDs for a
// unit-aimed ability, tiles for a tile-aimed one.
func (g *Game) AbilityAim(unit int) (units []int, tiles [][2]int) {
	u := g.B.Unit(unit)
	if u == nil {
		return nil, nil
	}
	ab, ok := combat.AbilityOf(g.C, u)
	if !ok || ab.Target == nil {
		return nil, nil
	}
	if ab.Target.Kind == "unit" {
		for _, z := range combat.AbilityUnits(g.B, u, ab.Target) {
			units = append(units, z.ID)
		}
		return units, nil
	}
	for y := 0; y < g.B.N; y++ {
		for x := 0; x < g.B.N; x++ {
			if combat.AbilityTileOK(g.B, u, ab.Target, x, y) {
				tiles = append(tiles, [2]int{x, y})
			}
		}
	}
	return nil, tiles
}
