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
- ✅ Skill library: change a hero's skills (any number at once) once per floor (owner's rule, 2026-10-11)
- ✅ The purple (epic) tier, new (approved 2026-10-11): 8 weapons, 9 armor pieces, 3 charms, 14 purple bonuses

## 2. The run (Phase 7)

- ✅ 7a: endless floors; maps grow from 14x14 to 32x32; 7 biomes; rivers, bridges, mountain passes, forts, roads, forests, hills (1,500 old floors rebuilt exactly)
- ✅ 7a: monster count and level by floor (the v0.22.1 tuning), elites every 5 floors, bosses every 10, throne floors, reinforcement floors (now worked out by Go in every old fight), max 1 enemy healer, chests and companion spirits placed by the floor
- ✅ 7a: blue starting tiles on every floor (every open tile of the bottom 2 rows); placing heroes on them comes with team select (7c)
- ✅ 7b: persistent floors: going back gives the same map, monsters, opened chests and shop stock; replaying a cleared floor gives half gold and full XP; boss legendary memory
- ✅ 7b: floor travel list (cleared floors, chests left, boss loot); also from the middle of a battle (what was earned is kept)
- ✅ 7b: shop before every 7th floor: buy, sell, **combine** two items (mixes and amplifies their bonuses; two rares make a purple, two purples a legendary from Lv 30)
- ✅ 7c: recruiting is replaced (owner, 2026-10-11): all 10 heroes are yours from the start; every 25 floors (25, 50, 75) you get one more hero slot instead, from 5 up to 8
- ✅ 7b: a free legendary gift every 100 floors
- ✅ 7c: team select and deployment (approved 2026-10-10): look at the map, pick heroes for every slot, place them on blue starting tiles (more tiles than heroes); only heroes who fight earn XP; if you don't choose, your highest-level heroes go
- ✅ 7b: **change ascension once per floor**, between battles (the old "re-spec" and "Choose Ascension" buttons)
- ✅ 7b: floor-clear summary (XP, level-ups, loot, gold), and a defeat summary
- ✅ 7d: saves: 3 save slots, autosave every turn of a climb, resume mid-battle exactly (fresh dice), save export and import (copy, download, paste or file), version stamp; old saves not needed for now (owner, 2026-10-11), Claude makes a playtest save
- ✅ 7b: defeat re-runs the floor with no penalty (kept on purpose, owner 2026-10-11: getting stuck and farming earlier floors is the game)
- 🛠 Shops (owner, 2026-10-11): no floor number; every 7 floors the shop comes up between floors, then you fight that floor

## 3. Screen and controls (Phase 8)

The old screen is carried over and wired to the referee; each of these gets ticked off on an iPhone before Phase 8 is done.

**Camera**
- ✅ Drag anywhere to pan; pinch to zoom where your fingers are; mouse wheel zooms where you point (8a)
- ✅ Zoom buttons with fixed steps (1x, 1.6x, 2.4x, 3.4x) and a zoom slider; keyboard WASD / arrows to pan, + and - to zoom (8o: − and + beside the slider in the ☰ menu, as in v0.26)
- ✅ Never zoom out past a finger-sized tile on phones; "overview" zoom shows the whole map, then pinch or + to return (8a)
- ➕ Camera keeps a margin so units never hide under the bottom bar; it stays inside the map
- ✅ When aiming, the camera frames the attacker and the target above the forecast window (8h)
- ✏️ After a hero acts, the camera stays put; the "Next hero" button moves to the next one (owner's change 2026-10-11; was: the camera moved on by itself)
- ✅ Optional auto-zoom on the selected hero; the camera follows your selection when zoomed in (8i)
- ✅ Minimap you can tap to jump, with your view box drawn on it (8a)
- ✅ Units turn into simple markers only when zoomed far out (8a)
- ✅ Flat (Fire Emblem) view and isometric view with rotate (flat is the main view) (new screen, 8g)
- ✅ The camera waits for each animation to finish before moving on (your request) (8j)

**Moving and acting**
- ✅ (8i) **Drag a hero to a blue tile to move** (press, drag, release), with a preview line and target tile; the map scrolls when you drag near the screen edge; a short tap after a drag is ignored so it doesn't misfire; dropping on an unreachable tile says so
- ✅ Tap a hero, tap a blue tile to move, or tap the hero again to act in place (8a)
- ✅ Tap a monster to attack it (walks closer if needed; forecast first, tap again to confirm)
- ✅ Bottom command bar and the classic side menu (commands beside the hero, skills fly out) (8h)
- ✅ Aim strip: target list with ◀ ▶ to cycle targets, skill switcher, weapon-set buttons, Cancel and Attack (8h)
- ✅ Forecast window (at the bottom, as in v0.24): HP before → after with a KO tag, DEF, damage, hit %, WEAK/RESIST, skill and weapon, statuses; green window for heals and buffs (8h)
- ✅ Tap a monster to see its move and attack zones, weaknesses and skills (8h, 8i)
- ✅ Danger zone: shade every tile monsters can reach (Phase 9 builds on this) (8d)
- ✅ Attack-range tiles shown after moving (8i, from the referee)
- ✅ The system menu (End Turn, Gear, Save, map tools, floors) opens from ☰ (8d). In the final old game (v0.26.1) tapping an empty tile only clears the selection, so the new one does the same.
- ✅ End Turn glows when everyone has acted; optional auto end turn; asks before ending if heroes haven't acted (8a/8d)
- ✅ Battle speed (1x, 2x, Instant) and a Skip button during the enemy turn (8c/8d)

**Menus and info**
- ✅ Title screen (castle art, candle flicker, lightning and mist), New Game hero pick, Continue with save slots (8o)
- ✅ Gear sheet: CAPPED tag, bag sorting, stat comparison with worn gear, weapon ranks shown, Drop with confirmation, item pictures (8k)
- ✅ Settings: Options, How to Play (with a link to the Codex), What's New, Saves, a copyable bug Report (8o)
- ✅ Options: damage numbers, reduced flashing, low graphics, colorblind-friendly tiles, minimap, danger zone (8o)
- ✅ First-time tips, off by default (v0.25), each shown once; battle goal only at the start (your request) (8o)
- ✅ Level-cap notice; status badges only on the selected or targeted unit (8o)

**Look and feel**
- 🔜 Your new sprites with movement and attack animations; glows per evolution stage; companion blue glow; enraged boss red glow
- ✅ Gothic look: crimson and gold, night-graded battlefield, vignette, Grenze Gotisch and EB Garamond fonts (8a)
- ✅ Unique layered attack effects for every skill, monster attack and ability; caster poses; damage numbers; smooth tile-by-tile movement (8l, checked by the owner 2026-10-11)
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
