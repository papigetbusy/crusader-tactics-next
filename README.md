# Crusader Tactics (next)

The clean rebuild of Crusader Tactics. The rules live in Go (the **referee**), the browser page only draws and asks questions (the **screen**).

- Play the game: https://papigetbusy.github.io/crusader-tactics-next/play.html
- Practice fights: https://papigetbusy.github.io/crusader-tactics-next/
- Rule Book: https://papigetbusy.github.io/crusader-tactics-next/rules.html
- Old-Game Match: https://papigetbusy.github.io/crusader-tactics-next/match.html
- What's left, in order: `docs/next-steps.md`. Latest update for Grok: `docs/update-for-grok.md`.
- The old game stays playable at https://papigetbusy.github.io/crusader-tactics/ until this one matches it.

## Two AIs code here: read this first (Claude and Grok)

Claude and Grok both write code in this repository for the owner. To stay consistent and not cause bugs:

1. **Follow `docs/working-together.md` exactly.** Pull first, claim your task in its "Now working on" list, run `gofmt`, `go vet` and `go test ./...` before pushing, pull again, never force-push, and wait for the green tick.
2. **The rules of the project:** `CLAUDE.md`, `docs/design-lock.md` and `docs/skeleton-design.md`.
   - Game rules live only in `sim/` (the referee); the screen in `web/game/` only draws the View and sends actions.
   - Numbers live in `data/`.
   - Every rule has a plain-sentence test.
   - The old game's code is never used, only its art.
   - Nothing changes how the game plays without the owner's approval, and approved changes are listed in `docs/changes-from-old-game.md`.
3. **Log every change in this README.** Add a short plain-words line under "Change log" below for each piece of work you finish: date, who, what changed. This README is the shared log of everything, so the other AI and the owner can see what happened.
4. **What to work on:** `docs/next-steps.md` (the next-session list at the top, then "Left to do, in order").
5. **When Claude comes back,** it audits everything that changed since its last commit (diff, tests, project rules) before doing anything else.
6. **Who does what (owner, 2026-10-12).** Claude does the heavy coding and new systems. Grok does not start those. Grok works on look and feel, art direction, lag, and fine-tuning mechanics, and checks that the architecture stays clean (rules in the referee, the screen only draws). New attack sheets are paused until the owner has OpenArt credits again. The prompt rule is already locked in `new-animations/PROMPT.md`. Signed off, still not in the game: Blood Moon, Soul Nova, Grand Hex, Blood Frenzy.

### Change log
(newest first; one line per finished piece of work: date · who · what changed)
- 2026-10-17 · Grok · New pictures replace the old monsters. Four bosses and six elites are in the game: Bone Hydra, Fallen Seraph, Brood Matron, Ash Colossus, Crowned Slime, Goblin Captain, Grave Knight, Dire Wolf, Iron Minotaur, Blight Reaper. The huge demon is not in the game yet. Codex rebuilt from the data.
- 2026-10-17 · Claude · Fix: online, the hero picker no longer "closes" when a friend joins, sits down or gets ready. The room screen was being redrawn over it; now it waits until the picker closes, then shows the latest.
- 2026-10-17 · Claude · Versus leftovers fixed. An enemy coming out of the fog shows only the seen part of its walk. Watchers see the whole board one round behind (live again when the match ends). Grave Bloom raises only the other player's dead. Raised monsters are the same for both sides (test). The room's view code moved to `adapters/server/roomview.go`.
- 2026-10-16 · Claude · Versus exploit fixes. Move-and-undo can't stop the clock any more (the free seconds only follow real turn actions), a move that uncovers an enemy can't be undone, the other player's traps stay hidden (their fire and kegs only in sight), and each player has their own bag. A test for each.
- 2026-10-16 · Grok · Continue and loading a save leave an online room first, so the match cannot freeze the climb. Versus matches cannot overwrite a climb save. A phone gone for 20 seconds loses the match; a refresh inside that keeps the seat. A dead room says the match is over instead of joining you as someone new.
- 2026-10-16 · Grok · Title screen is in the ☰ menu in every mode. Online Give up ends the match now (the other player wins) so you can start again. Opening single player or pass-the-phone while an online match is going ends that match for both, instead of freezing on "It is not your turn."
- 2026-10-16 · Grok · Versus clocks are 5 minutes each with no per-turn limit. Online: the server's clock is the one on screen (2 free seconds after a move), a refresh keeps your seat, and a missed move snaps to the server's board.
- 2026-10-16 · Claude · Versus and online: fog of war (see enemies within 6 steps of your units, can't target what you can't see, darkened tiles, each player's phone only gets what they can see, watchers see all, fog lifts at the end) and 32×32 maps, the climb's biggest. Each player's camera starts at their own edge of the map.
- 2026-10-16 · Claude · Online step 3: online rooms. Versus → Online, a room password for up to 4 friends, Blue and Red play while the others watch and wait in line, and the winner stays on. A small Go server (`adapters/server`) runs the same referee and keeps the clocks. It's ready for Render (`render.yaml`, `Dockerfile`), with the owner's setup steps in `docs/online-setup.md`.
- 2026-10-16 · Claude · Online step 2, pass the phone: Versus on the title screen (Pass the phone / Online soon), pick 8 heroes and ascensions per player, hand-over screens between turns, both clocks on screen (turn and match time, enforced by the referee), and a winner screen with Rematch. Also a restart note for Claude (CLAUDE.md, .claude-wip.md).
- 2026-10-16 · Claude · Online step 1: versus rules in the referee. Two players, 8 heroes each at Lv 30 with ascensions and a fixed purple set per class, a mirrored 12×12 map, turn and match clocks, win by wipe-out or time. Player 2's heroes keep every hero power (about 100 team checks now ask "is this a player's unit?"). Single player is unchanged. New tests include whole bot matches, an exact replay, and every skill used by both sides. Nothing on screen yet: that is step 2.
- 2026-10-16 · Claude · Plan: the playtest bugs and polish are done. The owner moved the ideas (chess feel first) to after online 1v1, so online is next.
- 2026-10-16 · Claude · The red attack window says what the chosen attack does besides its numbers (its words, plus element, area, cooldown, free action). It now never covers the target or any enemy it could hit: it tries beside the hero, then the screen's corners and edges, and may cover heroes.
- 2026-10-16 · Claude · New Skills window (☰ → Skills, and Skills ▸ in Gear): each hero's class ability, brought skills, ultimate and learned skills, with range, power, area, element, cooldown, the level learned and turns until ready. Changing skills (once per floor) moved here from Gear: tap cards to pick up to 4 and an ultimate. The referee now sends cooldowns and levels.
- 2026-10-16 · Claude · Owner's pick B: on a phone the map starts 30% closer, so heroes, monsters and squares are bigger to tap. When the battle starts the camera slides so the map fills the screen (no empty band under the heroes). Pinch and zoom work as before.
- 2026-10-16 · Claude · Owner's picks: a cloaked hero gets a soft white glow that pulses (still see-through), and loot shows a small card (item picture, + and name in its rarity color) floating up where the monster fell.
- 2026-10-16 · Claude · Ground bonuses are shown: tapping a unit shows a green badge for what its tile does (forest, hill, fort, high ground; numbers from the referee), and bonus tiles show 🛡 and ✚ numbers on the map. A shield button in the eye strip turns the tile numbers on and off (saved per save).
- 2026-10-16 · Claude · Fix: the Settings, Gear, Floor travel and Save windows have a round ✕ pinned in the top corner, so they close without scrolling to the bottom. The Close button at the bottom stays.
- 2026-10-16 · Claude · Fix: monsters raised for the heroes (Grave Bloom and the raise skill) now keep their monster name, so they show as a Skeleton with a monster's moves instead of a hero picture. Test added.
- 2026-10-16 · Claude · Wrote the owner's playtest list into docs/next-steps.md as the fixing stage (bugs, polish, ideas). It all gets done before online.
- 2026-10-16 · Claude · Online-ready check done (docs/online-ready.md). New tests: a climb replays exactly from its seeds and actions, and the screen may only reach the referee through referee.js. The lag tool now goes through that doorway too.
- 2026-10-16 · Claude · A Duelist no longer does the new counterattack; his Riposte replaces it (owner's choice). The forecast says "ripostes instead".
- 2026-10-16 · Claude · Audited Grok's 39 commits: tests, format and checks all pass; rules stay in the referee. Wrote down the play-without-internet task for Grok in docs/next-steps.md.
- 2026-10-12 · Grok · The lag watch can play a move, an attack, or the enemy turn, so those can be timed the same way as opening a floor. A normal game does not run it.
- 2026-10-12 · Grok · Lag watch. Open the game with ?lag=1 and it records slow frames and which part of the picture was slow. A normal game does not run it. The probe is tools/lag/probe.mjs.
- 2026-10-12 · Grok · Job split: Claude builds systems. Grok handles feel, lag, art direction, and keeping the rules in the referee. New attack sheets wait until OpenArt credits are back.
- 2026-10-12 · Grok · Blood Frenzy signed off. The prompt rule is locked to that sheet. Not in the game.
- 2026-10-12 · Grok · Idle sprites are also in new-animations/Idle, one file per hero, so they can be opened from the repo. Not in the game.
- 2026-10-12 · Grok · Blood Frenzy review GIF. The ring stays on the Berserker. Not signed off. Not in the game.
- 2026-10-12 · Grok · Named idle sprites for the art AI are in new-animations/idle-sprites.zip. First prompt batch is prompts/01-warrior-ascension.md. Not in the game.
- 2026-10-12 · Grok · Idle cards for all 60 hero forms are in new-animations/idle-cards. Skill names to copy are in new-animations/HERO-NAMES.md. Not in the game.
- 2026-10-12 · Grok · Tested the animation bot on six warrior sheets. Review GIFs only. Not signed off. Not in the game.
- 2026-10-12 · Grok · The animation bot makes a first review GIF for every sheet dropped in new-animations/inbox, then stops. It does not write code and it does not sign off.
- 2026-10-12 · Grok · Grand Hex is signed off. Code is in new-animations/code/grand-hex.js. It is not in the game. Queen's Garden is next.
- 2026-10-12 · Grok · Grand Hex aura is black in the middle, fading to green, and the green changes from frame to frame. Not signed off.
- 2026-10-12 · Grok · Grand Hex aura is now the game's warp ring, recolored green, around her. Not signed off.
- 2026-10-12 · Grok · Grand Hex review now gives her a green glow and the game's absorb ring, recolored green, around her while she powers up. Not signed off.
- 2026-10-12 · Grok · Grand Hex review uses the v3 Hex Queen sheet. The seal grows onto the enemy and she returns to idle. Not signed off.
- 2026-10-12 · Grok · Sprite prompts now call the saved hero with @. Grand Hex uses @hex queen.
- 2026-10-12 · Grok · Sprite prompts now use one locked template in new-animations/PROMPT.md. Grand Hex is being sent again. The last sheet mixed in the other witches.
- 2026-10-12 · Grok · Queen's Garden has a review GIF. Ivy lands on the enemy, the ring is centered, and she returns to idle. Not signed off. Not in the game.
- 2026-10-12 · Grok · Grand Hex has a review GIF. The seal hits the enemy, the circle is centered on it, and she returns to idle. Not signed off. Not in the game.
- 2026-10-12 · Grok · Soul Nova is signed off. Code is in new-animations/code/soul-nova.js. It is not in the game. Grand Hex is next.
- 2026-10-12 · Grok · Soul Nova's opening circle is centered on the black ball and drawn behind it. Not signed off. Not in the game.
- 2026-10-12 · Grok · Soul Nova's hit now uses the black ball from the first GIF, on the glow version. She still returns to idle. Not signed off. Not in the game.
- 2026-10-12 · Grok · Animation rule: small body motion, recolored effects the game already has, the hero glow, and the last frames back on idle. Soul Nova's glow GIF now ends on idle. Not in the game.
- 2026-10-12 · Grok · Soul Nova glow test. She keeps the stiff cast. A purple power-up uses effects already in the game, recolored. Not signed off. Not in the game.
- 2026-10-12 · Grok · Soul Nova has an in-game review GIF. It is not signed off and it is not in the game. Blood Moon is still the only approved attack.
- 2026-10-12 · Grok · Blood Moon is approved. Its code stays in new-animations/code for later. It is checked off the animation list. It is not in the game.
- 2026-10-12 · Grok · Blood Moon's in-game GIF now also plays the effect the codex already uses: a red beam on the enemy, the red burst inside it, and a ring that grows under his feet and fades. The signed swing and the red moon stay.
- 2026-10-12 · Grok · Blood Moon's in-game GIF keeps the red moon circle from the signed test. It hangs in the air, then the pillar and the blast hit the enemy. An effect is only removed if the owner says so.
- 2026-10-12 · Grok · Blood Moon is the first signed attack. Its code is in new-animations/code and is not loaded by the game. The in-game GIF shows it hitting an enemy, with the blast kept off her sprite.
- 2026-10-12 · Grok · Blood Moon v2 is a test GIF only, at assets/preview. It is not in the game and not in new-animations yet.
- 2026-10-12 · Grok · New attack tests live in `new-animations/` (not in the game). `gifs/` holds every try. `code/` gets one file per attack only after that GIF is signed off. They go into the game together, once the checklist is done. The witch v1 GIFs are the first three, still not signed off.
- 2026-10-12 · Grok · A separate attack preview (not in the game) is at assets/preview, so the witch GIFs can be opened and saved.
- 2026-10-12 · Grok · On the place-your-heroes screen, 1 2 3 save a starting formation. An empty number saves where they stand and turns gold. Tap a gold number to put them back. Hold a gold number to overwrite it; Cancel does nothing, Yes replaces it. Formations are kept in that save.
- 2026-10-12 · Grok · Each save file remembers its own settings (isometric or flat, danger zone, minimap and the rest). Loading a save puts that look back. An older save with no settings keeps whatever is on screen.
- 2026-10-12 · Grok · Floor 41 was redrawing the whole map every moment, and pinch-zoom on a phone snapped instead of zooming. The ground is now painted once and moved, pinch zooms smoothly from the whole map to up close, and the picture is drawn a bit softer on a sharp phone so it keeps up.
- 2026-10-12 · Grok · The map-icon strip stays hidden. A clear eye in the top left opens it, and tapping the map closes it again.
- 2026-10-12 · Grok · Finished heroes turn gray, health bars drain as the hit lands (and fill on a heal), a skill says WEAK before you pick it when it does bonus damage and the target is marked WEAK or RESIST once the attack window is open, and map controls are a narrow icon strip on the left edge instead of a labeled panel in the menu.
- 2026-10-12 · Claude · The attack window is menu-sized in a crimson wooden frame, floating beside the hero, with a skill-name box and arrows to cycle attacks (owner's design B). Added `docs/working-together.md` for Claude and Grok.
- 2026-10-12 · Grok · A floor 40 purple-team save ships with the test build and loads into an empty slot; the shield icon became the site icon.
- 2026-10-11 · Claude · Phase 9: smarter enemies from floor 20, full-power counterattacks, height. Phase 8 polish: zoom buttons, title effects, Settings tabs, damage-numbers switch, tips, level-cap notice. The new Codex. Ported attack effects. Save-file import.

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
| `tools/codex/` | Builds the Codex (`web/codex/`) and `docs/codex-checklist.md` from the game's data |
| `tools/skillbook/` | Writes `web/assets/skillbook.json`, the list of skills the screen needs to pick each attack's look |
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
  - 8c: animations: heroes and monsters walk tile by tile, the old effect pictures play for each skill (fire, lightning, slashes, heals...), damage numbers float up, hit units shake, the fallen fade out, the camera follows the monsters, and a Skip button speeds up the enemy turn.
  - 8d: the ☰ menu: End Turn, Gear, Save, Settings, zoom, minimap, danger zone, battle speed, rotate, next hero and the floor list. The gear sheet (weapon sets, worn gear, the bag, changing skills once per floor, ascension), three save slots plus an autosave after every turn, export and import, Continue on the title screen, and the Settings window (auto end turn, ask before ending, low graphics, fewer flashes, colorblind tiles).
  - Checked in a phone-sized browser: importing a save, equipping, switching weapon sets, taking gear off, dropping, changing skills once per floor, changing ascension, the danger zone, and moving after rotating the map 90°, 180° and 270°. Fixed: gear changed before a battle no longer uses up the hero's turn. Messages now fade after 3 seconds.
  - 8e: the shop before every 7th floor: Buy, Sell and Combine tabs (the referee lists which bag items can be combined), then Leave shop goes on to fight the floor.
  - 8f: aiming in several taps: tile and line skills (Blink, Bog of Despair, Powder Keg, Trample...), two-target and shove skills, and the aimed class abilities (Snare Field, Absolute Freeze, War Banner...), with gold tiles to tap and a Cancel button. The referee now says when an aim is complete.
  - 8g: the isometric view (☰ → View, or Settings → Map view): the same Tiny Swords ground on diamond tiles, raised hills and mountains with cliff sides, trees, houses and the castle, rotate included. Checked: moving by tapping works in all four turns.
  - 8h (side-by-side with the old game): the wooden command menu beside the hero (Attack, Skills ▸ flying out, ✦ ability, Item, Move, Undo Move, Defend, Wait, Cancel); the full forecast window (HP before → after with KO, damage, hit %, WEAK/RESIST, skill switcher, ◀ ▶ targets, ⇄ weapon sets); tapping a monster out of reach walks the hero closer first (the referee's plan); the monster info window and the terrain box; the battle goal when a floor starts; End Turn hides while a hero is chosen. The Phase 8 lines of the feature checklist are being ticked off.
  - 8i: drag a hero to a blue tile (dashed line and target tile, the map scrolls at the screen edge), red attack-range tiles after choosing or moving a hero, a monster's red attack zone around its purple move zone (both from the referee), auto-zoom on the selected hero (Settings), and panning never counts as a tap.
  - 8j: during the enemy turn the camera waits for each strike and effect to finish before moving to the next monster.
  - 8k: the gear sheet extras: item pictures (gear sheet and shop), bag sorting (Newest, Rarity, Slot), each bag item compared with what is worn (▲ ▼) or why it can't be worn (from the referee), and the CAPPED tag.
  - 8l: the attack effects, ported from what the old game really draws: each skill's own drawn effect (the fireball flies, bullets trace, arrows fly, meteors fall, pillars of light), its own stack of recolored effect pictures on top, and the old poses (guns and bows kick back, spells lift, charges and dives dash, potions and knives are thrown, blinks vanish, dragons rear, swords lunge), with the old timings and screen shake. The 53 new skills got effects of their own. Checked and approved by the owner.
  - 8m: importing a save from a file: Save → Import picks a slot, then a save file or pasted code; it asks before replacing a slot that has a game, and a file that isn't a save is refused without touching the slot.
  - 8n: the Codex (`codex.html`, also Settings → Codex): every hero with the new sprites, skills by level (new ones marked), monsters, elites, bosses, every attack effect playing live, items with the purple tier, statuses, terrain, floors and rules, controls and every change from the old game, built from the game's data so it always matches. The old Codex is kept unchanged in `docs/codex/`, and `docs/codex-checklist.md` shows where each of its entries is now.
  - 8o, the last polish: − and + zoom buttons and icons in the ☰ menu; the title screen's flickering candles, lightning and mist; Settings with Options, How to Play (with the Codex), What's New, Saves and a copyable Report; the damage-numbers switch; first-time tips (off until Settings → Turn tips on, each shown once); and the level-cap notice. Phase 8 is done: next is the owner's new art.
  - 8p (owner's design, signed off 2026-10-12): the attack window is now menu-sized in a crimson wooden frame, floating beside the hero on whichever side covers the fewest units. It shows the target's health bar, HP before → after, damage, hit and the counter; a box with the skill's name and ◀ ▶ to cycle attacks; ◀ ▶ for enemies in range, ⇄ weapon set, ✕ cancel; and a slim ⚔ Attack button.
  - 8q (owner, 2026-10-12): a hero who has finished the turn is drawn gray; health bars slide as the hit or heal lands; WEAK is written on a skill before you choose it, and WEAK or RESIST is drawn on the target in the attack window. Map controls are a narrow icon strip, hidden until the clear eye in the top left is tapped.
  - Owner's changes after testing (2026-10-11): choosing a hero no longer zooms in, and the camera no longer jumps to the next hero after one acts.
- Phase 9, new systems (in progress; rules in `docs/phase9-proposal.md`, approved 2026-10-11):
  - 9a: smarter enemies from floor 20 (bosses and elites from floor 10): they go for a hero they can defeat, then healers, then heroes weak to their attack; from floor 50 they avoid ending close to two heroes when they can't attack. Floors 1-19 still play like the old game.
  - 9b: Fire Emblem counterattacks: a hero or monster that survives a single-target hit strikes back once with a full Strike if the attacker is in its Strike range; not while stunned, frozen, rooted or cloaked, nor against area skills, ultimates, heals and buffs. The forecast shows the counter (damage, hit %, your HP after) or why there is none, and the counter plays as its own strike.
  - 9c: height: from a hill onto lower ground +10% damage and +10% hit, uphill −10% hit, climbing a hill costs 1 more move (not for fliers, Rangers and Knights); the forecast says HIGH GROUND or UPHILL.
  - Phase 9 is done as approved; class-owned terrain was left out for now (owner).
- Phase 7d, saves: done. Three save slots and an autosave on every turn of a climb, kept in the browser; a saved battle resumes exactly where it was, mid-fight (only the dice are fresh); saves can be exported (copied and downloaded) and imported (pasted or from a file); every save carries a version stamp so later versions can read it.
- Phase 7b and 7c, the climb: done. "Start a climb" on the practice page plays the real run from floor 1 with all ten heroes: look at the map, choose a hero for each slot (5, then one more at floors 25, 50 and 75), place them on the blue tiles, fight. Only the heroes who fight gain XP. A floor-clear summary follows each win; a defeat keeps everything earned and the floor is fought again. The shop comes up before every 7th floor (buy, sell, combine two copies into one a rarity higher). The floor list travels back to any floor reached; floors keep their map, monster level, opened chests and shop stock, and replays pay half gold. Every 100th floor gives a free legendary; a Lv 30 hero can change ascension once per floor.
- Phase 7a, floors: done. The floor builder makes every floor of the climb from the run's seed: map size, ground (ridges with passes, rivers with bridges, forests, hills, lakes, forts, a road along a guaranteed route), biome, monsters and their level (from the floor and the team's strength), bosses, elites, throne and reinforcement floors, chests and companion spirits, plus the new blue starting tiles. 1,500 floors recorded from the old game are built again exactly in Go, and Go now brings in reinforcements in every old fight (732 arrivals, all matching). Shops no longer take a floor (owner's rule). The practice page has "Climb: floor" to build and play any floor.
- Phase 6d, loot: the old loot is done. Monsters drop items (elites always, bosses sometimes a legendary), wood, silver and gold chests open when a hero stops on them, kills pay gold (more for elites and bosses, double for a Bounty Hunter, triple for a Wanted Poster target), loot is made for a hero who can use it, rare monsters can bond with the hero who defeats them as a companion spirit, and heroes can take gear off into the bag. 179 new old-game loot fights replay exactly (16,137 of 16,137 steps), with 703 chests, 189 drops, 244 times gear was taken off and 6 companion bonds. Legendaries now wait until a hero reaches Lv 30 (approved). The purple (epic) tier is in too (approved 2026-10-11): 8 weapons, 9 armor pieces, 3 charms and 14 purple bonuses, each with a rule test; the skill test fights carry a few purple items to try. Each hero's Skills button changes its learned skills, any number at once, once per floor (the old skill library, with the owner's once-per-floor rule). 6d is done; next is Phase 7, the run.
- Phase 5b–5d: Defend, Potion, Undo and tap-to-attack (the hero walks closer if needed, with a forecast first) on the practice page; forecasts for heals, revives and area attacks; enraged bosses glow red.
