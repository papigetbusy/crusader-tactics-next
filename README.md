# Crusader Tactics (next)

The clean rebuild of Crusader Tactics. The rules live in Go (the **referee**), the browser page only draws and asks questions (the **screen**).

- Live page: https://papigetbusy.github.io/crusader-tactics-next/
- The old game stays playable at https://papigetbusy.github.io/crusader-tactics/ until this one matches it.

## Where things are

| Folder | What it is |
|---|---|
| `docs/` | The design lock and the skeleton design (the map of this project) |
| `rehearsal/` | Rehearsal only: a tiny dice roller proving Go runs in the browser. Removed when real game code arrives. |
| `adapters/wasm/` | The bridge the browser page calls |
| `web/` | The page itself |
| `.github/workflows/` | The build robot: tests, builds and publishes on every change |

## Status

Rehearsal: proving the new route (Claude pushes, the robot tests, builds and publishes). Next: Phase 4, the core fight.
