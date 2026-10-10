# Crusader Tactics (next)

The clean rebuild of Crusader Tactics. The rules live in Go (the **referee**), the browser page only draws and asks questions (the **screen**).

- Live page: https://papigetbusy.github.io/crusader-tactics-next/
- The old game stays playable at https://papigetbusy.github.io/crusader-tactics/ until this one matches it.

## Where things are

| Folder | What it is |
|---|---|
| `docs/` | The design lock and the skeleton design (the map of this project) |
| `data/` | Every game number, as files you can edit |
| `sim/` | The referee: all game rules (no screen, no clock, no internet) |
| `testdata/oldgame/` | The answer key: the old game's tables and 200 recorded fights |
| `tools/oldgame/` | Checks Go against the recorded fights |
| `tools/guard/` | House rules: keeps the code clean |
| `rehearsal/` | Rehearsal only: a tiny dice roller proving Go runs in the browser. Removed when real game code arrives. |
| `adapters/wasm/` | The bridge the browser page calls |
| `web/` | The page itself |
| `.github/workflows/` | The build robot: tests, builds and publishes on every change |

## Status

- Rehearsal: done (Claude pushes, the robot tests, builds and publishes).
- Phase 4a, foundation and movement: done. Every one of the 6,624 moves recorded in the old game is a legal move in Go.
- Phase 4b, attack math: done. All 2,508 hits and misses recorded in the old game give the same result in Go (same damage, same weak-spot multiplier, same HP after, same dice), using the exact dice the old game rolled.
- Next: 4c, turns and enemy AI.
