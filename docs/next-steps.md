# Crusader Tactics (next): what's left

For the owner, Claude and Grok. Open work is at the top, finished work is a short list at the bottom (the details of each piece are in `changelog.md`, the docs it names, and the git history). The full plan is in `roadmap.md`, and every old-game feature is listed in `feature-checklist.md`. Trimmed 2026-10-17 (owner).

## Where we are (2026-10-17)
- The Go referee plays the whole game and every old-game match check is green. The Play page is a new thin screen that only draws the referee's View.
- Single player, pass-the-phone versus and online rooms (on Render, tested by the owner with a friend) all work.
- Latest: the legendary rework (one legendary per class, `legendary-proposal.md`) and Grok's monster redraw (4 new bosses, 6 new elites).

## Open: waiting on the owner
- **The pit-lord demon:** only a picture (`new-animations/monster-picks/pit-lord.png`), not a unit. It has to be fitted as something bigger than one tile.

## Open: to do
- **Online extras (owner's order):** chat and emojis, then the ideas list below (chess feel first).
- **Play without internet (for Grok).** The owner wants to play on the train with no signal. The old game did this with a service worker (`sw.js`) that keeps a copy of the game on the phone; the new build has none.
  - Add a small `sw.js` that saves play.html, `game/js`, the wasm, `wasm_exec.js`, the data and `assets/` on first visit, and serves them offline.
  - Name its cache with the build number, so a new push replaces the old copy (the pages workflow already stamps `__BUILD__`).
  - Saves already live on the phone. Test: open the link once, airplane mode, close and reopen, play a battle; also from the iPhone home screen.
  - Screen-only change: the referee does not change.
- **Parked attack animations (signed off, not in the game).** Blood Moon, Soul Nova, Grand Hex and Blood Frenzy have their code in `new-animations/code/`. They go into the game together, once every row of `new-animations/CHECKLIST.md` is signed off, not before. New sheets wait until the owner has OpenArt credits again.
  - A bot (`python3 new-animations/bot/make.py`) makes the first review GIF for sheets dropped in `new-animations/inbox/`, then stops. It never writes `code/`, never signs a row off and never loads anything into the play page.
  - The look, from 2026-10-12, is the signed Grand Hex: small body motion, a game effect recolored to the skill, a black center fading out to that color, and the last frames back on the idle sprite with the glow gone. The effect must not cover the attacker.
- **The owner's new art:** his sprites and attack animations replace files in `web/assets/`; plain swaps need no code. Claude wires in the hue and animation effects.
- **Balance autoplayer:** rebuild the old game's skilled autoplayer in Go. Tune the new skills, the purple tier, the legendary powers (`sim/content/rework_legendary.go`) and the new monsters. Report where runs get stuck.

## Ideas (they change how the game plays; owner OK needed)
Suggested order: the chess-feel proposal, then gear, bags, shops and gold together, then skills, then battlefield, then new classes and the floor-100 boss.
- **The chess feel** ("you think of moves and set your opponent up"). Claude's suggestions, for a proposal: show what each enemy will do next turn; zones of control (a unit walking next to an enemy must stop); facing, with flanking and backstabs; pushes into water, lava or off ledges; team-up attacks when allies are next to the target; goals other than "kill everything". Claude's pick: the first three together.
- **Gear and gold:**
  - each hero has their own bag, whose rarity sets the space (legendary holds the most)
  - more legendary armor; named sets with play styles at every level
  - exotic weapons: a 1–3% drop from certain bosses past a level, crafted into one exotic, with no other way to get one; the rarest thing in the game
  - a crafting system
  - a shop before every battle in team setup, plus a special shop every 7th floor that sells rare and above at a high price
  - more ways to earn gold
  - better gear sprites
- **Skills:**
  - skills have their own level and get stronger as it rises
  - a passive tree for each class (Path of Exile style)
  - stamina and magic bars for every hero, to stop spamming one spell
  - run the numbers on every skill so each one is worth taking (the balance autoplayer; Grok reads its reports)
- **Battlefield:**
  - more terrains, and fixed terrain bonuses and penalties
  - each hero owns a terrain (the owner said "not now" in Phase 9; now on the list)
  - weather (fog of war is in for versus)
  - more random height on maps (height itself is in, from Phase 9)
  - more movement for every class
- **Content:**
  - 5 new classes
  - a giant boss on floor 100
  - new boss animations (Grok, art)

## Done
- 2026-10-17 · Menu frame (Grok): the owner picked corner vines. It is on the floor bar, the ☰ menu, gear, skills and settings. Move and attack menus unchanged.
- 2026-10-17 · Cleanup the owner signed off: one shared status-immunity check, Lifeline covers every death, tests for the new monsters, finished pick pages not published (`new-animations/unpublished.txt`), docs trimmed, change log moved to `changelog.md`.
- 2026-10-17 · Legendary rework: one legendary weapon per class with 2 unique powers and a 4-piece class bonus, 5 bonuses on every legendary, exact bonus counts by rarity, epic set bonuses (`legendary-proposal.md`).
- 2026-10-17 · Monster redraw (Grok): new pictures, 4 new bosses after the old six, 6 new elites; pictures recut so bodies are whole.
- 2026-10-17 · Sheets: the ✕ stays at the top right of Gear, Skills and the other sheets; the floor bar hides while one is open; the sheet sits below the phone's clock.
- 2026-10-16 · Online 1v1 versus, steps 1–3 (owner's design): versus rules in the referee, pass the phone, online rooms with a Go server (`adapters/server`, on Render: `online-setup.md`), 5-minute clocks with no per-turn limit, 32×32 mirrored maps with fog of war, exploit fixes, leaving a mode cleanly. Design and rules: `online-ready.md`, `design-lock.md`.
- 2026-10-16 · The fixing stage (owner's playtest list): raised monsters drawn as monsters, pinned ✕, ground bonus badges, cloak glow, loot cards, a closer camera, the Skills window, an attack window that says what each attack does.
- 2026-10-16 · Phase 10, online-ready check: the screen only talks to the referee through `referee.js` (a test enforces it), and a climb replays exactly from its seeds and actions.
- 2026-10-12 · The new attack window (design B, ◀ ▶ to cycle attacks); spent heroes go gray, health bars drain, WEAK and RESIST shown, map controls in an icon strip, faster panning, settings and formations saved with the game.
- 2026-10-11 · Phase 9 (`phase9-proposal.md`): smarter enemies from floor 20, Fire Emblem counterattacks, height.
- 2026-10-11 · Phase 8, the game screen, written from scratch; the old game's attack animations ported (the owner approved them); the save-file picker; the new Codex (`codex.html`, built by `tools/codex`, with `codex-checklist.md`).
- 2026-10-09 to 10-11 · Phases 0–7: the Go referee plays the whole game (combat, skills, gear, the purple tier, loot, floors, the climb, team select, shops, saves), checked against the old game.
