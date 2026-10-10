//go:build js && wasm

// Browser glue only: it exposes the roll package to the web page. No game rules live here.
package main

import (
	"encoding/json"
	"runtime"
	"syscall/js"

	"github.com/papigetbusy/crusader-tactics-next/rehearsal/roll"
)

// set at build time with -ldflags "-X main.build=..."
var build = "dev"

func main() {
	js.Global().Set("goInfo", js.FuncOf(func(this js.Value, args []js.Value) any {
		b, _ := json.Marshal(map[string]string{"go": runtime.Version(), "build": build, "arch": runtime.GOOS + "/" + runtime.GOARCH})
		return string(b)
	}))
	// goRoll(seed, count, sides) -> JSON text like [12,3,19]
	js.Global().Set("goRoll", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 3 {
			return "[]"
		}
		seed := uint64(args[0].Float())
		b, _ := json.Marshal(roll.Dice(seed, args[1].Int(), args[2].Int()))
		return string(b)
	}))
	select {} // keep the program alive so the page can keep calling it
}
