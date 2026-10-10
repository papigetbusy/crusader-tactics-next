# Crusader Tactics (next)

The clean rebuild of Crusader Tactics. The rules live in Go (the **referee**), the browser page only draws and asks questions (the **screen**).

- Play: https://papigetbusy.github.io/crusader-tactics-next/ (practice fight)
- Rule Book: https://papigetbusy.github.io/crusader-tactics-next/rules.html
- Old-Game Match: https://papigetbusy.github.io/crusader-tactics-next/match.html
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
| `tools/convert/` | One-time converter that made the data files from the old tables |
| `tools/practice/` | Makes the practice fights from recorded starting setups |
| `tools/report/` | Turns test results into the Rule Book, Old-Game Match, House Rules and Changes pages |
| `adapters/wasm/` | The bridge the browser page calls |
| `web/` | The page itself |
| `.github/workflows/` | The build robot: tests, builds and publishes on every change |

## Status

- Rehearsal: done (Claude pushes, the robot tests, builds and publishes).
- Phase 4a, foundation and movement: done. Every one of the 6,624 moves recorded in the old game is a legal move in Go.
- Phase 4b, attack math: done. All 2,508 hits and misses recorded in the old game give the same result in Go (same damage, same weak-spot multiplier, same HP after, same dice), using the exact dice the old game rolled.
- Phase 4c, turns and enemy AI: done. All 197 recorded fights replay in Go from the first turn, 9,217 of 9,217 steps identical to the old game, with Go making all 1,746 enemy decisions itself. Taken from the recordings for now: where reinforcements appear (floors, Phase 7), level-ups (XP, Phase 6), and loot rolls (Phase 6).
- Phase 4d: done. A practice fight you can play from the link (six real starting setups from the old game, floors 1 to 20), plus the Rule Book, Old-Game Match, House Rules and Changes pages, rebuilt by the robot after every change.
- **Phase 4 is complete.** Next: Phase 5, full combat (the rest of the skills, statuses and the forecast in every case).
