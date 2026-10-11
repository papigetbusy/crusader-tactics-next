# Old-game feature checklist

Every feature of the old game (v0.26.1), and where it lands in the rebuild. Nothing gets dropped: before a phase is called done, every line for that phase is ticked off here. Made 2026-10-11 from the old game's code, its built-in "What's new" history (v0.1 to v0.22) and the later update blocks (v0.23 to v0.26).

✅ in the new game and proven against the old one · 🔜 planned (phase named) · ➕ was missing from the roadmap, now added · 🛠 old behaviour we will improve (needs your sign-off unless already approved)

## 1. Battle rules (the referee)

- ✅ Moving, terrain costs, flying and mounted units, height and water
- ✅ Hit chance, damage, crits, weapon triangle, element wheel, terrain cover
- ✅ Every status (bleed, burn, poison, snare, stun, shock, weak, armor break, ivy, lock and the rest)
- ✅ Defend, Potion, Undo move, Wait, End Turn
- ✅ Every skill, the skill bar (4 skills + ultimate + Strike), skill learning and replacing
- ✅ Skills level up with use (+5% per level; switching skills counts double)
- ✅ Class abilities at Lv 1, 5 and 20, the 30 ascension abilities and all 30 ascension passives
- ✅ Revives, seize-the-throne, reinforcement floors (where reinforcements appear comes in Phase 7)
- ✅ Monster AI (copied exactly; smarter AI comes in Phase 9)
- ✅ Levels, XP, XP sharing, level cap, evolutions at Lv 5 and 20, ascension at Lv 30
- ✅ Gear: weapons, armor, charms, rarities, all 27 bonuses, weapon ranks E to S, weapon sets, dual wielding, armor sets, legendaries
- ✅ Necromancer raises, shadow clones, traps, banners, fire fields, Overwatch, Sky Lord
- ✅ Keep the hero selected for extra shots (Quick Draw, Oath of Ruin): the referee simply keeps that hero able to act
- ✅ Loot drops, chests (wood, silver, gold), gold, loot that targets gear your party can use
- ✅ Companion spirits (rare blue-glowing monsters that bond with the hero who defeats them; the blue glow comes with the new sprites in Phase 8)
- ✅ **Unequip** (take an item off into the bag; the old gear sheet allowed it by tapping a slot)
- ✅ The purple (epic) tier, new (approved 2026-10-11): 8 weapons, 9 armor pieces, 3 charms, 14 purple bonuses

## 2. The run (Phase 7)

- 🔜 Endless floors; maps grow from 14x14 to 32x32; 7 biomes; rivers, bridges, mountain passes, forts, roads, forests, hills
- 🔜 Monster count and level by floor (the v0.22.1 tuning), elites every 5 floors, bosses every 10, shops every 7, throne floors, reinforcement floors, max 1 enemy healer
- 🔜 Persistent floors: going back gives the same map, monsters, opened chests and shop stock; replaying a cleared floor gives half gold and full XP; boss legendary memory
- 🔜 Floor travel list (cleared floors, chests left, boss loot)
- 🔜 Shop: buy, sell, **combine** two items (mixes and amplifies their bonuses)
- 🔜 Recruit a new hero every 25 floors (up to 8); a free legendary gift every 100 floors
- 🔜 Team select and deployment on blue tiles (approved 2026-10-10)
- ➕ **Change ascension once per floor** from the gear sheet (the old "re-spec" and "Choose Ascension" buttons)
- 🔜 Floor-clear summary (XP, level-ups, loot, gold)
- 🔜 Saves: 3 save slots, autosave at the start of every turn, resume mid-battle, save export and import (copy, download, paste or file), existing saves keep loading
- 🔜 Defeat re-runs the floor; shop floors count as cleared (both flagged for a design call)

## 3. Screen and controls (Phase 8)

The old screen is carried over and wired to the referee; each of these gets ticked off on an iPhone before Phase 8 is done.

**Camera**
- ➕ Drag anywhere to pan; pinch to zoom where your fingers are; mouse wheel zooms where you point
- ➕ Zoom buttons with fixed steps (1x, 1.6x, 2.4x, 3.4x) and a zoom slider; keyboard WASD / arrows to pan, + and - to zoom
- ➕ Never zoom out past a finger-sized tile on phones; "overview" zoom shows the whole map, then pinch or + to return
- ➕ Camera keeps a margin so units never hide under the bottom bar; it stays inside the map
- ➕ When aiming, the camera frames the attacker and the target above the forecast window
- ➕ After a hero acts, the camera moves to the next ready hero ("Next hero" button too)
- ➕ Optional auto-zoom on the selected hero; the camera follows your selection when zoomed in
- ➕ Minimap you can tap to jump, with your view box drawn on it
- ➕ Units turn into simple markers only when zoomed far out
- ➕ Flat (Fire Emblem) view and isometric view with rotate (flat is the main view)
- 🔜 The camera waits for each animation to finish before moving on (your request)

**Moving and acting**
- ➕ **Drag a hero to a blue tile to move** (press, drag, release), with a preview line and target tile; the map scrolls when you drag near the screen edge; a short tap after a drag is ignored so it doesn't misfire; dropping on an unreachable tile says so
- ➕ Tap a hero, tap a blue tile to move, or tap the hero again to act in place
- ✅ Tap a monster to attack it (walks closer if needed; forecast first, tap again to confirm)
- ➕ Bottom command bar and the classic side menu (commands beside the hero, skills fly out)
- ➕ Aim strip: target list with ◀ ▶ to cycle targets, skill switcher, weapon-set buttons, Cancel and Attack
- ➕ Floating forecast window beside the target: HP before → after with a KO tag, DEF, damage, hit %, WEAK/RESIST, skill and weapon, statuses; green window for heals and buffs
- ➕ Tap a monster to see its move and attack zones, weaknesses and skills
- ➕ Danger zone: shade every tile monsters can reach (Phase 9 builds on this)
- ➕ Attack-range tiles shown after moving
- ➕ Tap an empty tile for the system menu (End Turn, Gear, Save, map tools, floors)
- ➕ End Turn glows when everyone has acted; optional auto end turn; asks before ending if heroes haven't acted
- ➕ Battle speed (1x, 2x, Instant) and a Skip button during the enemy turn

**Menus and info**
- ➕ Title screen (castle art, candle flicker, lightning and mist), New Game hero pick, Continue with save slots
- ➕ Gear sheet: CAPPED tag, bag sorting, stat comparison with worn gear, weapon ranks shown, Drop with confirmation
- ➕ Settings: Options, How to Play (with a link to the Codex), What's New, a copyable bug Report
- ➕ Options: damage numbers, reduced flashing, low graphics, colorblind-friendly tiles, minimap, danger zone
- 🔜 First-time tips, off by default (v0.25), each shown once; battle goal only at the start (your request)
- ➕ Level-cap notice; status badges only on the selected or targeted unit

**Look and feel**
- 🔜 Your new sprites with movement and attack animations; glows per evolution stage; companion blue glow; enraged boss red glow
- ➕ Gothic look: crimson and gold, night-graded battlefield, vignette, Grenze Gotisch and EB Garamond fonts
- ➕ Unique layered attack effects for every skill, monster attack and ability; caster poses; damage numbers; smooth tile-by-tile movement
- ➕ Map art (Tiny Swords tiles, coasts, bridges, roads, cliffs, trees, houses, castle), item icons
- (There was no sound in the old game.)

## 4. Tools

- ➕ **Balance autoplayer.** The old game was tuned in v0.22.1 with a skilled autoplayer that played thousands of fights. We rebuild it in Go after Phase 7 so every balance change (new skills, gear, AI) can be measured: win rate, floors reached, which skills get used. This is how we tune things that were overlooked, with numbers instead of guesses.

## 5. Things to iron out

Already fixed with your sign-off: forecast now matches real damage; traps work along the whole walk; Grave Bloom no longer depends on a timer; a fallen hero can be revived however they died; one skill version for monsters and heroes; gear swaps cost the turn; level-ups keep the weapon in hand; the auto-walk endless loop cannot happen.

Signed off by the owner on 2026-10-11 (nothing built yet; each lands in the step named):
1. ✅ done: the battle ends the moment it is won (the old game let heroes keep acting until the next floor loaded).
2. ✅ done: an undo that can't happen (a shadow clone stands on the old tile) says why instead of silently doing nothing.
3. ✅ approved, **6d design, tuned after Phase 7:** (the Lv 30 rule is in place) legendaries only start to drop once any one hero reaches Lv 30, and only ones that hero can use. (purple built 2026-10-11) A new **purple tier (epic)** sits between rare (blue) and legendary (gold): weapons, armor and charms with their own bonuses (prefixes), built for the middle levels. Claude designs the purple items and bonuses and sends them for approval (like the skill list) before anything goes in; their numbers are then tuned with the autoplayer.
4. ✅ settled by point 3: legendary weapons needing rank S fits, since legendaries now arrive at Lv 30, when heroes reach rank S.
5. ✅ approved, **Phase 7:** a design call on what a defeat costs and whether shop floors need a fight.
6. ✅ approved, **Phase 9:** smarter monsters as floors rise.
