# Crusader Tactics (next): what's left

Written 2026-10-11 for the owner and Grok. The full plan is in `roadmap.md`, and every old-game feature is listed in `feature-checklist.md`.

## Where we are
- **Phases 0–7 are done.** The Go referee plays the whole game: combat, skills, gear, the purple tier, loot, floors, the climb, team select, shops, saves. Every old-game match check is green.
- **Phase 8 (the game screen) is done (2026-10-11).** The Play page is a new screen written from scratch. It uses only the old game's pictures, never its code. It only draws the referee's answers and sends actions, so online play stays possible. Done:
  - title screen
  - team picker and placing heroes
  - moving (tap or drag) and attacking with the full forecast
  - the command menu beside the hero, with skills that fly out
  - skills and class abilities aimed in several taps
  - walking, and the attack effects and poses ported from the old game (approved by the owner)
  - importing a save from a file into a chosen slot
  - the enemy turn with Skip, the danger zone and monster zones
  - the gear sheet with item pictures and comparisons
  - shop, floor list, saves with autosave, settings
  - flat and isometric views with rotate

## The new attack window: done 2026-10-12
The owner signed off design B with a skill-name box (no count) and ◀ ▶ to cycle attacks; built in `battle.js` (forecast, placeForecast) and `style.css`.

## Next session, first thing: missing from the old build — done 2026-10-12
The owner noticed these while playtesting. All four are on the screen only; the referee did not change.
1. **A hero who has finished the turn goes gray** (and a little see-through), so a spent hero no longer looks ready.
2. **Health bars drain as the hit lands**, over about a third of a second, instead of jumping at the end of the action. Heals fill the same way.
3. **WEAK** shows on a skill in the command menu when that move does bonus damage to someone in range, before you pick it. Once the attack window is open, WEAK or RESIST is also drawn on each target.
4. **Map controls are a narrow icon strip**, hidden until the clear eye in the top left is tapped (zoom, view, minimap, danger zone, speed, rotate, next hero). Tapping the map closes it. The ☰ menu no longer has the labeled map panel.
5. **Speed (owner, floor 41):** the ground is painted once and reused, so panning and pinching do not redraw every tile. Pinch zooms smoothly instead of snapping to the whole map. Water and trees hold still.
6. **Settings travel with the save** (owner, 2026-10-12). Isometric or flat, the danger zone, the minimap and the other options are stored in that save file. Loading it puts them back. Another save can look different.
7. **Formations** (owner, 2026-10-12). Before the battle, 1 2 3 remember where the heroes stand. Tap a gold number to place them. Hold a gold number to overwrite it, after a Yes. They are stored in that save, and they fit a different size floor.

## Play without internet (owner, 2026-10-12): for Grok, not done yet
The owner wants to play on the train with no signal. The old game did this with a helper file (`sw.js`, a service worker) that keeps a copy of the game on the phone. The new build has none yet, so it needs internet to open.
- Add a small `sw.js` to the published site that saves play.html, `game/js`, `referee.wasm`, `wasm_exec.js`, the data and `assets/` on first visit, and serves them when offline.
- Name its cache with the build number, so a new push replaces the old copy (the pages workflow already stamps `__BUILD__`).
- Saves already live on the phone, so they work offline as they are.
- Test: open the link once, turn on airplane mode, close and reopen the page, play a battle. Also add it to the home screen on the iPhone and open it from there.
- Screen-only change: the referee does not change.

## New attack animations: not in the game yet
The owner is rebuilding attack pictures in `new-animations/`. A bot (`python3 new-animations/bot/make.py`) can make the first review GIF for a pile of sheets dropped in `new-animations/inbox/`. It then stops. It must not write `new-animations/code/`, must not sign a row off, and must not load anything into the play page. A test GIF stays a test until the owner signs it off. Then one code file is written, and an in-game GIF is made of that move hitting an enemy. The effect plays with the sprite and must not cover the attacker. From 2026-10-12 the look is the signed Grand Hex: a small body motion, a game effect recolored to the skill, a black center fading out to that color, the color a little different on every frame, and the last frames back on the idle sprite with the glow gone. Nothing from that folder is loaded by the play page until every row of `new-animations/CHECKLIST.md` is signed off, and then they are added together. Blood Moon, Soul Nova and Grand Hex are signed off. Do not wire them in early.

## Left to do, in order

### 1. Finish Phase 8 (small polish): done 2026-10-11
- Zoom buttons (− and + beside the slider in the ☰ menu, with the menu's icons, as in v0.26)
- Title screen candle, lightning and mist effects
- Settings tabs: Options, How to Play, What's New, Saves, Report
- The damage-numbers on/off option
- First-time tips (off by default, each shown once)
- The level-cap notice

### 2. Attack animations, save-file picker and the Codex (done 2026-10-11)
- **The attack animations are ported and the owner approved them (2026-10-11).**
  - When we ran the old game, the FX2 code (`FX2assign`, `fxTheme`, `fx2Layers`, old lines 1386-1453) turned out never to be drawn. An earlier copy of the same step wins when the page runs. What really draws is each skill's own pixel effect (`drawP`, old lines 922-966) with a stack of recolored pictures on top (`makeRecipe`, old lines 1331-1380). The owner chose to port what really draws.
  - Ported, screen only (`web/game/js/fxbook.js`, `fxpixel.js`, `fxskill.js`, `fxsprite.js`, `playback.js`): the per-skill effects and their timings, the picture stacks, the older effects for skills without an entry, the poses and `anim` (guns and bows kick back, spells lift, charges and dives dash, potions and knives are thrown, blinks blink, breath rears, everything else lunges), the screen shake and the "fewer flashes" option. The referee did not change. `tools/skillbook` writes the list of skills the screen needs (a test keeps it in step with `data/`).
  - The 53 new skills and Last Light got their own entries (marked NEW in `fxbook.js`).
  - The owner checked Strike, Fireball, a gunshot, a heal and a charge on the published link and approved them. The new Codex can now be written.
- **The save-file picker is done (2026-10-11):** Save → Import picks a slot, then a file or pasted code, and asks before replacing a slot.

- **The Codex is done (2026-10-11), waiting for the owner to look it over.**
  - The old Codex (v0.21/0.22) is in `docs/codex/old-codex-v0.22.html`, unchanged.
  - The new Codex is published with the game as `codex.html` (links: the site's top bar, and Settings → Codex in the game). `tools/codex` builds it from the game's own data files, so it always matches the code; a test fails if it falls out of date (after changing `data/` or `changes-from-old-game.md`, run `go run ./tools/codex`).
  - It covers heroes with the new v0.26 sprites (every form and ascension, stats, weapons, class abilities, skills by level with the 53 new skills and Last Light marked New), monsters on the skill schedule, elites and bosses, every attack effect playing live with the game's own effect code, items with the purple tier and its bonuses, statuses, terrain and biomes, floors and rules (team slots 5 + one at floors 25, 50 and 75, no recruiting, shops before every 7th floor, saves, skill changes once per floor, no defeat penalty), mechanics, controls, and every change from `changes-from-old-game.md`.
  - `docs/codex-checklist.md` lists every entry of the old Codex and where it is now: nothing was lost (0 removed); changes such as new skill levels are listed. A test fails if an old entry ever disappears without being listed as removed.

### 3. The owner's new art
When Phase 8 is done and plays well, the owner swaps in his new sprites and attack animations. They replace files in `web/assets/`; no code changes needed for plain swaps. Claude wires in the hue and animation effects.

### 4. Balance autoplayer
Rebuild the old game's skilled autoplayer in Go. Tune the new skills, the purple tier and legendary power. Report where runs get stuck.

### 5. Phase 9: new systems (done 2026-10-11, as approved in `phase9-proposal.md`)
- Smarter enemies from floor 20 (bosses and elites from floor 10)
- Fire Emblem counterattacks with a full Strike, both sides, shown in the forecast
- Height: high ground and uphill, climbing costs
- Class-owned terrain: not now (owner)
- The owner moved the balance autoplayer and his new art to the end; these numbers get tuned by the autoplayer then.

### 6. Phase 10: online-ready check: done 2026-10-16
The screen only talks to the referee through `web/game/js/referee.js` (a house-rule test now enforces it), and a climb replays exactly from its seeds and actions. What online still needs, and what the owner must set up, is in `online-ready.md`.

### 6b. The fixing stage: before online (owner's playtest list, 2026-10-16)
The owner's order: Phase 9 is finished, then this whole list gets done, and online starts only once the game feels good. Bugs first. Anything that changes how the game plays needs the owner's OK (a proposal like `phase9-proposal.md`), and screen changes get a picture first.

**Fixes (bugs)**
- Done 2026-10-16: monsters raised for the heroes (Grave Bloom and the raise skill) were drawn as heroes because the referee left out their monster name. They now rise as a Skeleton with its moves, as in the old game (old raiseSkel).
- Done 2026-10-16: Settings and Gear (and Floor travel and Save, to match) have a ✕ pinned in the top corner that stays while scrolling.
- Done 2026-10-16: ground bonuses were invisible. A unit's badges now show what its tile does (from the referee, `combat.GroundEffects`), bonus tiles show their numbers, and a shield button in the eye strip hides the numbers.

**Quick polish (screen only, picture first)**
- A white glow on a cloaked hero (the old game only made it see-through).
- A small item picture with a + when loot is picked up.
- Bigger heroes and monsters, and bigger tap squares, for the iPhone.
- A skill window, and a Skills tab in the menu.

**Ideas (change how the game plays; owner OK needed)**
- Gear and gold:
  - each hero has their own bag, whose rarity sets the space (legendary holds the most)
  - more legendary weapons and armor
  - sets of every rarity, with new bonuses when a whole set is worn
  - exotic weapons: a 1–3% drop from certain bosses past a level, crafted into one exotic, with no other way to get one; the rarest thing in the game
  - a crafting system
  - a shop before every battle in team setup, plus a special shop every 7th floor that sells rare and above at a high price
  - more ways to earn gold
- Skills:
  - skills have their own level and get stronger as it rises
  - a passive tree for each class (Path of Exile style)
  - stamina and magic bars for every hero, to stop spamming one spell
  - run the numbers on every skill so each one is worth taking and fits a build or play style (the balance autoplayer; Grok reads its reports)
- Battlefield:
  - more terrains, and fixed terrain bonuses and penalties
  - each hero owns a terrain (the owner said "not now" in Phase 9; now on the list)
  - weather and fog of war
  - more random height on maps, so they aren't a flat perfect square (height itself is in, from Phase 9)
  - more movement for every class
- Content:
  - 5 new classes
  - a giant boss on floor 100
  - new weapon and gear sprites and new boss animations (Grok, art)
- The chess feel ("you think of moves and set your opponent up"). Claude's suggestions, for a proposal:
  - show what each enemy will do next turn
  - zones of control (a unit walking next to an enemy must stop)
  - facing, with flanking and backstabs
  - pushes into water, lava or off ledges
  - team-up attacks when allies are next to the target
  - goals other than "kill everything"
  - fog of war

  Claude's pick: the first three together.

Suggested order: bugs, then polish, then the chess-feel proposal, then gear, bags, shops and gold together, then skills, then battlefield, then new classes and the floor-100 boss.

### Online 1v1 versus (owner, 2026-10-11)
The owner's order: first iron out every flaw and match the original game as closely as possible, then go online. Online needs:
- versus rules in the referee (two player sides), tested first as "pass the phone" on one device
- a small server that runs the referee, with room codes. GitHub Pages can't run a server, so this needs a small host such as Fly.io or Render.
- a "play online" option on the screen, sending the same actions through `referee.js`

Versus rules (owner, 2026-10-11):
- Each player picks 8 heroes. All are set to Lv 30, and the player picks each hero's ascension.
- Everyone wears a full purple gear set, so every match is even.
- Chess.com-style clocks: a time limit per turn, plus an overall time for each player.
- A match lasts about 5 minutes at most.

Rules for the online work (owner, 2026-10-11). Do not build these now:
- A whole climb must replay from its seed plus the list of actions and come out identical.
- The turn clock and the match clock live in the referee and are shown in the View. The screen only displays them. Running out of time is the referee's decision.
- Still to settle with the owner: the seconds per turn and per player, the map (fixed or random, and how big), which purple set each class gets, and what happens when a clock runs out.

Estimate: about 3 sessions once the game matches the original.

### 7. After the full build (owner's reminders)
- More gear: named sets with play styles at every level
- Better gear sprites
