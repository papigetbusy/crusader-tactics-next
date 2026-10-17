//go:build js && wasm

// Command wasm is the bridge between the browser page and the referee. It is
// the only code that knows about the browser. Everything crosses as plain
// JSON text, the same messages an online server will use later.
package main

import (
	"encoding/json"
	"errors"
	"runtime"
	"syscall/js"

	"github.com/papigetbusy/crusader-tactics-next/data"
	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
)

var build = "dev" // set by the build robot

var (
	rules    *content.Content
	practice []engine.Setup
	game     *engine.Game
)

func text(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"refused":{"code":"internal","reason":"` + err.Error() + `"}}`
	}
	return string(b)
}

func fn(f func(args []js.Value) any) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any { return f(args) })
}

func main() {
	var err error
	if rules, err = content.Default(); err != nil {
		panic(err)
	}
	var p struct {
		Practice []engine.Setup `json:"practice"`
	}
	raw, _ := data.Files.ReadFile("practice.json")
	if err := json.Unmarshal(raw, &p); err != nil {
		panic(err)
	}
	practice = p.Practice

	g := js.Global()
	g.Set("ctInfo", fn(func([]js.Value) any {
		return text(map[string]string{"go": runtime.Version(), "build": build})
	}))
	// ctGuide() -> the fixed rules the screen describes in words
	g.Set("ctGuide", fn(func([]js.Value) any { return text(engine.RulesGuide(rules)) }))
	// ctPractice() -> ["Floor 1", ...]
	g.Set("ctPractice", fn(func([]js.Value) any {
		var names []string
		for _, s := range practice {
			names = append(names, s.Name)
		}
		return text(names)
	}))
	// ctNew(index, seedHigh, seedLow) -> view
	g.Set("ctNew", fn(func(a []js.Value) any {
		i := a[0].Int()
		if i < 0 || i >= len(practice) {
			i = 0
		}
		seed := uint64(uint32(a[1].Float()))<<32 | uint64(uint32(a[2].Float()))
		game = engine.NewFight(rules, practice[i], seed)
		return text(game.View())
	}))
	// ctVersusRules() -> the versus picker's numbers and every class's ascensions
	g.Set("ctVersusRules", fn(func([]js.Value) any { return text(engine.VersusRules(rules)) }))
	// ctVersus(setupJSON, mapSeed, seedHigh, seedLow) -> view or {"error"}: a 1v1 match
	g.Set("ctVersus", fn(func(a []js.Value) any {
		var s engine.VersusSetup
		if err := json.Unmarshal([]byte(a[0].String()), &s); err != nil {
			return text(map[string]string{"error": err.Error()})
		}
		s.MapSeed = uint32(a[1].Float())
		seed := uint64(uint32(a[2].Float()))<<32 | uint64(uint32(a[3].Float()))
		v, err := engine.NewVersus(rules, s, seed)
		if err != nil {
			return text(map[string]string{"error": err.Error()})
		}
		game = v
		return text(game.ViewFor(game.B.Side)) // fog: the first player's view
	}))
	// ctFloor(floor, runSeed, seedHigh, seedLow) -> view: a built floor of a run
	g.Set("ctFloor", fn(func(a []js.Value) any {
		seed := uint64(uint32(a[2].Float()))<<32 | uint64(uint32(a[3].Float()))
		s := engine.FloorSetup(rules, uint32(a[1].Float()), max(1, a[0].Int()))
		game = engine.NewFight(rules, s, seed)
		return text(map[string]any{"name": s.Name, "view": game.View()})
	}))
	// ctClimb(runSeed, seedHigh, seedLow) -> view: a new climb from floor 1
	g.Set("ctClimb", fn(func(a []js.Value) any {
		seed := uint64(uint32(a[1].Float()))<<32 | uint64(uint32(a[2].Float()))
		game = engine.NewClimb(rules, uint32(a[0].Float()), seed)
		return text(game.View())
	}))
	// ctSave() -> the whole game as a versioned save (text)
	g.Set("ctSave", fn(func([]js.Value) any {
		b, err := game.Save()
		if err != nil {
			return text(map[string]string{"error": err.Error()})
		}
		return string(b)
	}))
	// ctLoad(saveText, seedHigh, seedLow) -> {view} or {error}
	g.Set("ctLoad", fn(func(a []js.Value) any {
		seed := uint64(uint32(a[1].Float()))<<32 | uint64(uint32(a[2].Float()))
		loaded, err := engine.Load(rules, []byte(a[0].String()), seed)
		if err != nil {
			return text(map[string]string{"error": err.Error()})
		}
		game = loaded
		return text(map[string]any{"view": game.View()})
	}))
	// ctApply(actionJSON) -> {events, view} or {refused}
	g.Set("ctApply", fn(func(a []js.Value) any {
		var act action.Action
		if err := json.Unmarshal([]byte(a[0].String()), &act); err != nil {
			return text(map[string]any{"refused": engine.Refused{Code: "bad-action", Reason: err.Error()}})
		}
		side := game.B.Side
		before := game.Visible(side)
		events, err := game.Apply(act)
		var r *engine.Refused
		if errors.As(err, &r) {
			return text(map[string]any{"refused": r})
		}
		// Versus fog on one phone: the events as the player who acted saw
		// them, then the board as the player to move now sees it.
		out := map[string]any{"events": game.EventsFor(side, events, before), "view": game.ViewFor(game.B.Side)}
		if err != nil {
			out["problem"] = err.Error()
		}
		return text(out)
	}))
	// ctReach(unit) -> [[x,y],...]
	g.Set("ctReach", fn(func(a []js.Value) any {
		var out [][2]int
		for _, t := range game.Reach(a[0].Int()) {
			out = append(out, [2]int{t.X, t.Y})
		}
		return text(out)
	}))
	// ctThreat(unit) -> [[x,y],...]: where a unit could walk next turn
	g.Set("ctThreat", fn(func(a []js.Value) any {
		var out [][2]int
		for _, t := range game.Threat(a[0].Int()) {
			out = append(out, [2]int{t.X, t.Y})
		}
		return text(out)
	}))
	tiles := func(ts []board.Tile) any {
		var out [][2]int
		for _, t := range ts {
			out = append(out, [2]int{t.X, t.Y})
		}
		return text(out)
	}
	// ctAttackTiles(unit) -> where a hero's ready attacks reach now (red tiles)
	g.Set("ctAttackTiles", fn(func(a []js.Value) any { return tiles(game.AttackTiles(a[0].Int())) }))
	// ctThreatHit(unit) -> where a monster could hit next turn (red zone)
	g.Set("ctThreatHit", fn(func(a []js.Value) any { return tiles(game.ThreatHit(a[0].Int())) }))
	// ctTargets(unit, slot[, set]) -> [unitID,...]
	g.Set("ctTargets", fn(func(a []js.Value) any { return text(game.TargetsWith(a[0].Int(), a[1].Int(), optInt(a, 2))) }))
	// ctPreview(unit, slot, target[, set]) -> forecast or null
	g.Set("ctPreview", fn(func(a []js.Value) any {
		f, ok := game.PreviewWith(a[0].Int(), a[1].Int(), a[2].Int(), optInt(a, 3))
		if !ok {
			return "null"
		}
		return text(f)
	}))
	// ctPreviewFrom(unit, slot, target, x, y) -> forecast after walking to x,y
	g.Set("ctPreviewFrom", fn(func(a []js.Value) any {
		f, ok := game.PreviewFrom(a[0].Int(), a[1].Int(), a[2].Int(), a[3].Int(), a[4].Int())
		if !ok {
			return "null"
		}
		return text(f)
	}))
	// ctAim(unit) -> {units:[ids]} or {tiles:[[x,y]...]}: where the class ability can be aimed
	g.Set("ctAim", fn(func(a []js.Value) any {
		units, tiles := game.AbilityAim(a[0].Int())
		return text(map[string]any{"units": units, "tiles": tiles})
	}))
	// ctSkillAim(unit, slot, picksJSON) -> {units:[ids]} or {tiles:[[x,y]...]}:
	// what to pick next for a skill, given {target, second, tiles} so far.
	g.Set("ctSkillAim", fn(func(a []js.Value) any {
		var so action.Action
		_ = json.Unmarshal([]byte(a[2].String()), &so)
		units, tiles := game.SkillAim(a[0].Int(), a[1].Int(), so)
		return text(map[string]any{"units": units, "tiles": tiles})
	}))
	// ctPlan(unit, target) -> {slot, walk} or null: which skill a tap on
	// that unit uses, and where to walk first if needed.
	g.Set("ctPlan", fn(func(a []js.Value) any {
		p, ok := game.AttackPlan(a[0].Int(), a[1].Int())
		if !ok {
			return "null"
		}
		return text(p)
	}))
	select {} // keep running so the page can keep asking
}

// optInt is argument i as a number, or 0 when it was left out.
func optInt(a []js.Value, i int) int {
	if i >= len(a) || a[i].IsUndefined() || a[i].IsNull() {
		return 0
	}
	return a[i].Int()
}
