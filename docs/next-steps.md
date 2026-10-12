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

## Next session, first thing: missing from the old build (owner, 2026-10-12)
Match the old game (screen only; check against papigetbusy/crusader-tactics):
1. **Done heroes dim:** when a hero can't do anything else this turn, it is drawn dimmed, showing it has made its turn.
2. **Health bars drain as damage lands:** the lost part of the bar should slide down when the hit lands (old shp/sh/sht, about 350 ms), not jump at the end of the turn. Today `st.shownHp` is just `u.hp` and the bar only changes when the new View is shown.
3. **Weakness indicator:** show that one of your moves does bonus damage (WEAK) to a monster, the way the old game marked it (before choosing the attack, not only in the forecast).

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

### 6. Phase 10: online-ready check
Prove the screen only talks to the referee through `web/game/js/referee.js`, so the same screen can talk to a server.

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
