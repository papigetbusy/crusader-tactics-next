# Crusader Tactics (next)

The clean rebuild of Crusader Tactics. The rules live in Go (the **referee**), the browser page only draws and asks questions (the **screen**).

- Play the game: https://papigetbusy.github.io/crusader-tactics-next/play.html
- Practice fights: https://papigetbusy.github.io/crusader-tactics-next/
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
| `web/` | The practice page |
| `web/game/` | The game screen (Play): draws the referee's answers and sends actions; it keeps no rules or game state |
| `web/assets/` | Pictures (from the old game for now; the owner's new art replaces them), with `manifest.json` naming each one |
| `tools/art/` | One-time tool that saved the old game's pictures as named files |
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
- Phase 8, the real screen (in progress): a new game screen written from scratch, using only the old game's pictures (none of its code; owner's rule 2026-10-11). It only draws the referee's answers and sends actions, so online play stays possible. Done so far:
  - 8a: the title, the map with the old textures, trees, water and buildings, units with health bars, the camera (drag, pinch, zoom), the minimap, moving, attacking with a forecast, Defend, Wait, Potion, Undo, End Turn, the enemy turn and the floor summary.
  - 8b: the team picker (cards for all ten heroes, like the old "Choose 5 heroes" screen) and placing heroes on the blue tiles.
  - 8c: animations: heroes and monsters walk tile by tile, lunge when they strike, the old effect pictures play for each skill (fire, lightning, slashes, heals...), damage numbers float up, hit units shake, the fallen fade out, the camera follows the monsters, and a Skip button speeds up the enemy turn.
  - 8d: the ☰ menu: End Turn, Gear, Save, Settings, zoom, minimap, danger zone, battle speed, rotate, next hero and the floor list. The gear sheet (weapon sets, worn gear, the bag, changing skills once per floor, ascension), three save slots plus an autosave after every turn, export and import, Continue on the title screen, and the Settings window (auto end turn, ask before ending, low graphics, fewer flashes, colorblind tiles).
  - Checked in a phone-sized browser: importing a save, equipping, switching weapon sets, taking gear off, dropping, changing skills once per floor, changing ascension, the danger zone, and moving after rotating the map 90°, 180° and 270°. Fixed: gear changed before a battle no longer uses up the hero's turn. Messages now fade after 3 seconds.
  - 8e: the shop before every 7th floor: Buy, Sell and Combine tabs (the referee lists which bag items can be combined), then Leave shop goes on to fight the floor.
  - 8f: aiming in several taps: tile and line skills (Blink, Bog of Despair, Powder Keg, Trample...), two-target and shove skills, and the aimed class abilities (Snare Field, Absolute Freeze, War Banner...), with gold tiles to tap and a Cancel button. The referee now says when an aim is complete.
  - 8g: the isometric view (☰ → View, or Settings → Map view): the same Tiny Swords ground on diamond tiles, raised hills and mountains with cliff sides, trees, houses and the castle, rotate included. Checked: moving by tapping works in all four turns.
  - 8h (side-by-side with the old game): the wooden command menu beside the hero (Attack, Skills ▸ flying out, ✦ ability, Item, Move, Undo Move, Defend, Wait, Cancel); the full forecast window (HP before → after with KO, damage, hit %, WEAK/RESIST, skill switcher, ◀ ▶ targets, ⇄ weapon sets); tapping a monster out of reach walks the hero closer first (the referee's plan); the monster info window and the terrain box; the battle goal when a floor starts; the camera moves to the next ready hero; End Turn hides while a hero is chosen. The Phase 8 lines of the feature checklist are being ticked off.
  - 8i: drag a hero to a blue tile (dashed line and target tile, the map scrolls at the screen edge), red attack-range tiles after choosing or moving a hero, a monster's red attack zone around its purple move zone (both from the referee), auto-zoom on the selected hero (Settings), and panning never counts as a tap.
  - 8j: during the enemy turn the camera waits for each strike and effect to finish before moving to the next monster.
  - 8k: the gear sheet extras: item pictures (gear sheet and shop), bag sorting (Newest, Rarity, Slot), each bag item compared with what is worn (▲ ▼) or why it can't be worn (from the referee), and the CAPPED tag.
  - Still to do in Phase 8 (from the checklist): zoom buttons, the title screen's candle and lightning effects, How to Play / What's New / Report in Settings, the damage-numbers option, first-time tips, and the level-cap notice.
- Phase 7d, saves: done. Three save slots and an autosave on every turn of a climb, kept in the browser; a saved battle resumes exactly where it was, mid-fight (only the dice are fresh); saves can be exported (copied and downloaded) and imported (pasted or from a file); every save carries a version stamp so later versions can read it.
- Phase 7b and 7c, the climb: done. "Start a climb" on the practice page plays the real run from floor 1 with all ten heroes: look at the map, choose a hero for each slot (5, then one more at floors 25, 50 and 75), place them on the blue tiles, fight. Only the heroes who fight gain XP. A floor-clear summary follows each win; a defeat keeps everything earned and the floor is fought again. The shop comes up before every 7th floor (buy, sell, combine two copies into one a rarity higher). The floor list travels back to any floor reached; floors keep their map, monster level, opened chests and shop stock, and replays pay half gold. Every 100th floor gives a free legendary; a Lv 30 hero can change ascension once per floor.
- Phase 7a, floors: done. The floor builder makes every floor of the climb from the run's seed: map size, ground (ridges with passes, rivers with bridges, forests, hills, lakes, forts, a road along a guaranteed route), biome, monsters and their level (from the floor and the team's strength), bosses, elites, throne and reinforcement floors, chests and companion spirits, plus the new blue starting tiles. 1,500 floors recorded from the old game are built again exactly in Go, and Go now brings in reinforcements in every old fight (732 arrivals, all matching). Shops no longer take a floor (owner's rule). The practice page has "Climb: floor" to build and play any floor.
- Phase 6d, loot: the old loot is done. Monsters drop items (elites always, bosses sometimes a legendary), wood, silver and gold chests open when a hero stops on them, kills pay gold (more for elites and bosses, double for a Bounty Hunter, triple for a Wanted Poster target), loot is made for a hero who can use it, rare monsters can bond with the hero who defeats them as a companion spirit, and heroes can take gear off into the bag. 179 new old-game loot fights replay exactly (16,137 of 16,137 steps), with 703 chests, 189 drops, 244 times gear was taken off and 6 companion bonds. Legendaries now wait until a hero reaches Lv 30 (approved). The purple (epic) tier is in too (approved 2026-10-11): 8 weapons, 9 armor pieces, 3 charms and 14 purple bonuses, each with a rule test; the skill test fights carry a few purple items to try. Each hero's Skills button changes its learned skills, any number at once, once per floor (the old skill library, with the owner's once-per-floor rule). 6d is done; next is Phase 7, the run.
- Phase 5b–5d: Defend, Potion, Undo and tap-to-attack (the hero walks closer if needed, with a forecast first) on the practice page; forecasts for heals, revives and area attacks; enraged bosses glow red.
