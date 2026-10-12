# Crusader Tactics (next): what's left

Written 2026-10-11 for the owner and Grok. The full plan is in `roadmap.md`, and every old-game feature is listed in `feature-checklist.md`.

## Where we are
- **Phases 0–7 are done.** The Go referee plays the whole game: combat, skills, gear, the purple tier, loot, floors, the climb, team select, shops, saves. Every old-game match check is green.
- **Phase 8 (the game screen) is nearly done.** The Play page is a new screen written from scratch. It uses only the old game's pictures, never its code. It only draws the referee's answers and sends actions, so online play stays possible. Done:
  - title screen
  - team picker and placing heroes
  - moving (tap or drag) and attacking with the full forecast
  - the command menu beside the hero, with skills that fly out
  - skills and class abilities aimed in several taps
  - walking (attack animations are NOT done; see section 2)
  - the enemy turn with Skip, the danger zone and monster zones
  - the gear sheet with item pictures and comparisons
  - shop, floor list, saves with autosave, settings
  - flat and isometric views with rotate

## Left to do, in order

### 1. Finish Phase 8 (small polish)
- Zoom buttons
- Title screen candle, lightning and mist effects
- Settings tabs: How to Play, What's New, Report
- The damage-numbers on/off option
- First-time tips (off by default, each shown once)
- The level-cap notice

### 2. Fix the attack animations, then update the Codex (next session, first thing)
- **The attack animations are not done (owner, 2026-10-11).** The screen plays a simplified effect list and one lunge (`web/game/js/playback.js`).
  - Port the version that actually runs in the old game (papigetbusy/crusader-tactics, index.html): `FX2assign`, `fxTheme` and `fx2Layers` (each skill gets its own stack of pictures from its name and element, including the 53 new skills), and the ACT pose table and `anim` (guns and bows kick back, spells lift, charges and dives dash, potions and knives are thrown, blinks blink). Screen only; the referee does not change. Use the effect files already in `web/assets`.
  - Do not port the earlier `drawSPXfor` that only uses the E1–E18 list; it was replaced.
  - The old Codex (v0.21/0.22) is older than these animations. It is an archive, not the authority for how attacks look.
  - Done when Strike, Fireball, a gunshot and a heal look different from each other, and a charge dashes instead of lunging. The new Codex is written only after the owner says those look right.

- The Codex is the old game's reference page (v0.21/0.22). Upload the old Codex to this repository as-is, so nothing is lost.
- Make a checklist comparing every entry in it with the new game: in the code, in the roadmap, changed, or missing.
- Write the new Codex from the game's data files, so it always matches the code. It should cover:
  - the 53 new skills plus Last Light
  - the purple tier
  - the team picker and hero slots (5, plus one at floors 25, 50 and 75)
  - shops before every 7th floor
  - saves
  - skill changes once per floor
  - no defeat penalty
  - every other change in `changes-from-old-game.md`
- **Use the new sprites** (the v0.26 hero and ascension pictures now in `web/assets/heroes`). The old Codex shows outdated sprites.
- Never delete an entry without listing it as changed or removed.

### 3. The owner's new art
When Phase 8 is done and plays well, the owner swaps in his new sprites and attack animations. They replace files in `web/assets/`; no code changes needed for plain swaps. Claude wires in the hue and animation effects.

### 4. Balance autoplayer
Rebuild the old game's skilled autoplayer in Go. Tune the new skills, the purple tier and legendary power. Report where runs get stuck.

### 5. Phase 9: new systems
Each one is measured with the autoplayer:
- smarter enemy AI as you climb
- class-owned terrain
- height advantage
- Fire Emblem counterattacks
- and the rest in `roadmap.md`

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
