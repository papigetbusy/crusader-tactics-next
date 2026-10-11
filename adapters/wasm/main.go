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
	// ctApply(actionJSON) -> {events, view} or {refused}
	g.Set("ctApply", fn(func(a []js.Value) any {
		var act action.Action
		if err := json.Unmarshal([]byte(a[0].String()), &act); err != nil {
			return text(map[string]any{"refused": engine.Refused{Code: "bad-action", Reason: err.Error()}})
		}
		events, err := game.Apply(act)
		var r *engine.Refused
		if errors.As(err, &r) {
			return text(map[string]any{"refused": r})
		}
		out := map[string]any{"events": events, "view": game.View()}
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
	// ctTargets(unit, slot) -> [unitID,...]
	g.Set("ctTargets", fn(func(a []js.Value) any { return text(game.Targets(a[0].Int(), a[1].Int())) }))
	// ctPreview(unit, slot, target) -> forecast or null
	g.Set("ctPreview", fn(func(a []js.Value) any {
		f, ok := game.Preview(a[0].Int(), a[1].Int(), a[2].Int())
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
