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
| `testdata/oldgame/` | The answer key: the old game's tables and 380 recorded fights |
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
- **Phase 4 is complete.**
- Phase 5a, full-combat recordings: done. 180 more old-game fights recorded with every menu skill, Defend, Potion, Undo, revives and seize-the-throne floors; all 178 usable ones replay in Go with 13,555 of 13,555 steps identical (including 22 revives, 5 throne wins, 512 defends, 515 undos and 86 potions).
- Phase 6a, levels and experience: done. Heroes earn XP for hits, misses, kills, heals and buffs, level up, grow and evolve at levels 5 and 20, exactly as in the old game: all 24 recorded level-ups and every hero's XP after every step now come from Go, not from the recordings.
- Phase 6b, skills, abilities and ascensions: the old rules are done. Heroes carry their learned skills and bar; level-ups offer skills (fill a slot, replace one, or keep); Lv 30 offers three ascensions; all 30 class abilities and 30 ascension abilities work, with every ascension passive. 176 new high-level old-game fights replay exactly (14,289 of 14,289 steps), using all 61 abilities 1,054 times. The new even skill list (approved 2026-10-10) is in too: 53 new skills plus Last Light, every class on one schedule, monsters on the same schedule, each new skill with its own rule test, and four Lv 27 skill test fights on the practice page. The old-game match checks still use the old lists and stay fully green.
- Phase 6c, gear: done. Weapons, armor, charms, rarities and all 27 item bonuses, weapon ranks E-S, weapon sets (switching sets, dual wielding), the 3 armor sets, the 19 legendaries, and equipping and dropping items. A hero with no weapon counts as its class's usual weapon in the weapon triangle (approved change). 179 new old-game fights with geared parties replay exactly (14,994 of 14,994 steps), including 342 times a hero changed gear mid-battle. The practice page has a gear sheet: what each hero wears, its weapon sets, and a bag to equip from. Two old-game oversights are fixed (approved): putting on gear in battle uses up the hero's turn, and a level-up no longer drops the weapons in hand from the hero's attack.
- Phase 5b–5d: Defend, Potion, Undo and tap-to-attack (the hero walks closer if needed, with a forecast first) on the practice page; forecasts for heals, revives and area attacks; enraged bosses glow red.
