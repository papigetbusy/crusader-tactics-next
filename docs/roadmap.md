# Crusader Tactics: Rebuild Roadmap (v1, updated 2026-10-11)

**Goal:** a clean single-player game where the rules live in one place (the *sim*, written in Go) and the screen (the *UI*) only asks it questions and shows the answers. Built so online play later is plug-and-play. Playable from a GitHub Pages link the whole way (Go compiled to WASM).

**Design aim (owner, 2026-10-10):** a strategy-heavy game; skill and game knowledge should always be rewarded.

**Mental model:** the sim is the referee, the UI is the screen. The screen never decides anything. It asks "can I move here?" or "what happens if I attack?", and shows the referee's answer. Online later = the same referee, with actions arriving over the network.

Related docs: `feature-checklist.md` (every old-game feature and the phase it lands in; ticked off before each phase is called done), `port-readiness-report.md` (audit of the current game), `design-lock.md` (rules every prompt must respect), `skeleton-design.md` (the map of the Go project), `art-checklist.md` (new art to finish before Phase 8), `skill-plan.md` (the approved even skill list), `purple-tier-proposal.md` (the approved purple tier). This roadmap also lives in the repo as `docs/roadmap.md`.

**Where the new game lives:** repo `papigetbusy/crusader-tactics-next`. Claude pushes; the build robot (GitHub Actions) tests, builds and publishes. Play: https://papigetbusy.github.io/crusader-tactics-next/ · Rule Book: `/rules.html` · Old-Game Match: `/match.html` · House Rules: `/house.html` · Changes: `/changes.html`. The old game stays untouched at https://papigetbusy.github.io/crusader-tactics/.

## Status
- **Phase 0: done.** Decisions confirmed 2026-10-09 (see below).
- **Phase 1: done.** Answer key built: 30 data tables dumped, 200 scripted fights recorded, repeat check 50 of 50 identical, plain-language report page.
- **Phase 2: done (2026-10-09).** The Go test page ran from GitHub Pages on the owner's iPhone.
- **Phase 3: done (2026-10-09/10).** Skeleton design approved, including the fair shuffle. New repo `crusader-tactics-next`; Claude pushes; the build robot builds and publishes.
- **Phase 4: done (2026-10-10).** Core fight in Go, checked against every recording.
- **Phase 5: done (2026-10-10).** Full combat, checked against 180 new recordings.
- **Phase 6: done (2026-10-11).** 6a, 6b, 6c and 6d all green (6d: loot d5aad27, purple tier 5d3b8d4, skill changes once per floor).
- **Phase 7: in progress.** 7a floors done (2026-10-11): 1,500 old floors rebuilt exactly; reinforcements by Go in every old fight; blue starting tiles; "Climb: floor" on the practice page. **Next: 7b between floors (shop, floor list, gifts, ascension change).**

## Phase 6 (heroes and gear), done
- **6a Levels and XP:** XP, level-ups and stat growth; evolutions at Lv 5 and 20; level cap tied to the best floor; only heroes who fought earn XP. All 24 recorded level-ups worked out by Go.
- **6b Skills, abilities and ascensions:** the old skills ported and proven (176 high-level fights, 14,289 of 14,289 steps, all 61 abilities used 1,054 times), then the new even skill schedule: 53 new skills plus Last Light, each with its own rule test. Monsters learn on the same schedule (4 newest normals, newest ultimate and Strike); monster healers can bring back fallen monsters; bosses keep their kits.
- **6c Gear:** weapons, armor, charms, rarities and all 27 bonuses; weapon ranks E–S; weapon sets and dual wielding; 3 armor sets and 19 legendaries; Equip, Drop and switch-set; unarmed heroes count as their class's usual weapon. 179 geared fights replay exactly (14,994 of 14,994 steps).
- **6d Loot:** 179 loot fights replay exactly (16,137 of 16,137 steps: 703 chests, 189 drops, 244 times gear taken off, 6 companion bonds). Legendaries only from Lv 30. Purple tier (P1–P42): 8 weapons (rank A), 9 armor pieces, 3 charms, 14 purple bonuses (two new: Executioner's and Warding); purple from floor 21; before Lv 30 a legendary roll gives purple; gold chests 65% rare, 30% purple, 5% legendary. Skills: a hero changes any number of skills once per floor.
- **Practice page:** multi-tap skill aiming; fields, swamps, kegs and chests on the board; four Lv 27 skill test fights, geared, with purple items in the bag; gear sheet with Take off; Skills button.

## Phases

| # | Phase | What happens | What you can check |
|---|---|---|---|
| 0 | Lock the plan | One-page design lock plus the three decisions. **Done.** | Read one page and say "yes, that's my game." |
| 1 | Record the old game | The old game becomes the answer key: real fights recorded with their dice, real numbers dumped. **Done.** | A green/red checklist page in plain words. |
| 2 | Upload pipeline | A tiny Go program runs in the browser from the GitHub link. **Done.** | Open a link on your phone and see it load. |
| 3 | Design the skeleton | Folder layout, the list of actions, where every number lives. **Done.** | Read the map of "who owns what." |
| 4 | Core fight in Go | Move, Attack, End Turn, compared against the recordings. **Done.** | Checklist goes green. Play one plain fight. |
| 5 | Full combat | Defend, Potion, Undo, tap-to-attack, revives, the throne, forecasts. **Done.** | Forecast numbers match real results. |
| 6 | Heroes and gear | Classes, levels, skills, ascensions, abilities, gear, loot, as data files. **Done.** | Change a number in a data file and see the effect. |
| 7 | The run (then the balance autoplayer) | Floors, the shop every 7 floors (buy, sell, combine), gifts, ascension change once per floor, team select and deployment (all 10 heroes, hero slots, blue starting tiles), and saves with a version stamp. | Look at the map, choose and place your team, climb floors. |
| 8 | Connect the real screen | Every screen feature on the checklist (camera, drag-to-move, command bar, aim strip, forecast window, danger zone, settings, saves screens), with the owner's **new sprites and attack animations**. | The new link plays like the old one, with the new art. |
| 9 | New systems | Counterattacks, move/attack/threat tiles, class-owned terrain, height, bushes, smarter enemies. | Each system gets its own checklist and a playtest. |
| 10 | Online-ready check | Record a whole run as a seed plus a list of actions, replay it in Go, get the identical result. | The replay matches the original run. |

## Phase 7 plan (the run)
- **7a Floors:** the floor builder (maps 14x14 to 32x32, 7 biomes, rivers, bridges, passes, forts, roads), monster count and level per floor, elites every 5, bosses every 10, throne and reinforcement floors, blue starting tiles (more than the heroes).
- **7b Between floors:** the shop every 7 floors, with no floor number of its own: it comes up before floors 7, 14, 21..., you shop, then fight that floor (owner, 2026-10-11). Buy, sell, combine (two rares make a purple, two purples a legendary from Lv 30). Legendary gift every 100 floors, ascension change once per floor, skill changes once per floor, floor-clear summary, going back to cleared floors (same map, chests and shop stock; half gold, full XP).
- **7c Team select and deployment:** each floor: look at the map, pick heroes for every slot from all 10, place them on blue tiles (swappable until the battle starts), fight. The monsters' level is set when the map is made, from your strongest heroes for that many slots, so it doesn't change while you choose (owner, 2026-10-11). Blue tiles: every open tile of the bottom 2 rows (owner: fine as is). The monsters' level is set when the map is made, from your strongest heroes for that many slots, so it doesn't change while you choose (owner, 2026-10-11). The blue tiles are every open tile of the bottom 2 rows (owner: fine as is).
- **7d Saves:** 3 slots, autosave every turn, mid-battle resume, export and import. Old saves are not needed for now (owner, 2026-10-11); Claude makes a save for playtesting.
- **Defeat:** re-runs the floor with no penalty (owner, 2026-10-11: the game is built around reaching a floor you can't beat yet and going back to farm; that must not be punished). The autoplayer reports how often runs get stuck, without making the game harder.

## New art (Phase 8), owner note 2026-10-10
- The owner is drawing **new character sprites that move** and **new attack animations**, and will hand over all the assets **before Phase 8 starts**, so Phase 8 is built on the new art from day one. See `art-checklist.md`.
- **Claude handles every hue effect and animation effect:** glows and tints (e.g. the enraged boss's red hue), hit flashes, skill and status effects, death and revive effects, and the timing that ties them together.
- The art never decides a rule. The referee sends events; the screen plays the matching animation.
- The camera waits for each animation to finish before jumping to the next hero.
- Monster attack animations are part of this work; the 53 new skills need effects too.

## Order of everything still to do
1. **Phase 7, the run:** 7a floors, 7b between floors, 7c team select and blue starting tiles, 7d saves.
2. **Balance autoplayer:** the old game's skilled bot rebuilt in Go; first tuning pass on the new skills, the purple tier and legendary power; report how often runs get stuck.
3. **Phase 8, the real screen**, with the owner's new sprites and animations.
4. **Phase 9, new systems**, each measured with the autoplayer.
5. **Phase 10, online-ready check.**
6. **After the full build:** more gear (owner's reminder): more named sets with play styles at every level, better gear sprites.

## Balance autoplayer (after Phase 7)
The old game was tuned in v0.22.1 with a skilled autoplayer. We rebuild it in Go so every balance change can be measured: win rate, floors reached, where runs get stuck, which skills and gear get used. Owner, 2026-10-11: get everything from the original game, and fix and tune what was overlooked.

## Confirmed decisions
1. Match the old game's combat numbers exactly (checked against the recordings); exact floor layouts do not have to match. (2026-10-09)
2. A fresh combat seed is rolled whenever a saved game loads. (2026-10-09)
3. Old saves: not needed for now; the owner starts fresh and Claude makes a playtest save (2026-10-11; was "existing saves keep working", 2026-10-09).
4. Fair shuffle replaces the old game's random-comparator shuffles (listed on the Changes page). (2026-10-09)
5. Mid-battle saves resume exactly where you left off; only future dice are fresh. (2026-10-10)
6. The old game is a proof of concept, not law: every idea carries over, exact matching proves it, improvements are deliberate and signed off by the owner. (2026-10-10)

## Owner sign-offs and reminders
Working agreement: Claude ports the original first, then flags improvements in plain words; the owner signs off or denies. Approved changes go on the Changes page.

**Approved, with the phase they land in:**
- **Phase 5 (done):** monsters use the same version of each skill as heroes; tap-to-attack shows the forecast first; enraged bosses (Phase 8: red glow on the real art).
- **Done (2026-10-10):** any fallen hero can be revived, however they died.
- **Phase 6 (done):** one skill schedule for every class (53 new skills plus Last Light); unarmed heroes count as their class weapon; gear swaps use the turn; level-ups keep the weapon in hand; legendaries from Lv 30; the purple tier.
- **Phase 6 (done 2026-10-11):** skill changes once per floor: any number of skills (and the ultimate) in one go, then fixed until the next floor.
- **Phase 6:** the "switching skills counts double for skill levels" rule stays, explained once in the tips.
- **Phase 7 (owner, 2026-10-11):** the shop comes up every 7 floors with no floor number of its own; a defeat re-runs the floor with no penalty; old saves not needed for now.
- **Phase 7 rules, Phase 8 screen: team select and deployment, Fire Emblem style (settled 2026-10-10; roster clarified 2026-10-11).**
  - **All 10 heroes (one of each class) are yours from the start. There is no recruiting** (replaces the old "a new hero every 25 floors").
  - Hero slots: 5 at the start; one more slot on reaching floor 25, 50 and 75 (8 is the most).
  - Each floor goes: **look at the map, pick your heroes, place them on starting tiles, fight.** Every slot must be filled.
  - If you don't choose, the game brings your highest-level heroes.
  - Only the heroes who fight on a floor earn its XP; benched heroes earn nothing.
  - You place your heroes yourself on blue starting tiles. Each map has more blue tiles than heroes, so where you start is a real choice.
  - Referee actions: choose team, place hero (swappable until the battle starts), start battle; each refusal gets a rule test.
- **Phase 8:** new sprites and attack animations from the owner; Claude makes all hue and animation effects.
- **Phase 9:** smarter enemies as floors rise: go after healers, and attack heroes who are weak to them.

**Reminders to raise at the right phase:**
- **After the full build (owner, 2026-10-11): more gear.** More gear, better gear sprites, more named sets with their own play styles at every level.
- **Phase 8:** explain each mechanic once in tips; show the battle goal only at the start, then keep it in the pause menu.
- **Phase 8:** the camera waits until a hero's animation finishes before jumping to the next hero.

## Known old-game issues (fixed once, in Go)
- Endless loop on auto-walk attack (screen code): the new tap-to-attack asks the referee for one plan, so it cannot loop.
- Forecast ignored crits, added-element damage and double hits: one math path in Go.
- Unseeded combat dice and random-comparator shuffles.
- Moving a hero, seizing the throne, dropping an item and auto end turn lived in screen code.

## How testing works for a non-technical owner
Every phase ends with a link and a green/red checklist in plain words. Claude runs the technical tests. The owner plays a few fights and says "this feels wrong."

## Rules that keep the new code from becoming another monolith
- No gameplay math in UI code.
- Every rule has a test.
- Numbers live in data files.
- At the end of each phase, check file sizes and the folder map before moving on.
