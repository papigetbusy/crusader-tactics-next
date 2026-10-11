# Crusader Tactics: Design Lock (v1, single-player first)

Paste this at the top of every big prompt. Decisions confirmed by the owner on 2026-10-09.

## What the game is
- Turn-based tactics RPG with Fire Emblem-style controls. Party of up to 5 heroes from 10 classes, gear with affixes, ascensions, endless floors.
- Gothic look: crimson and gold, Castlevania mood, no bright colors.
- Made as the free game the owner wishes existed. Quality over speed.

## Scope right now
- Single-player only. No online play, Steam or iOS in any task until the owner says so.
- Do not rewrite everything in one prompt. One phase at a time.

## Architecture (non-negotiable)
1. The **sim (Go)** owns every rule and number: movement, hit, crit, damage, statuses, enemy AI, loot, XP, floors, shop, progression choices.
2. The **UI** never decides anything. It asks the sim for legal moves, previews and results, and draws them.
3. Everything a player or the AI does is an **action** (data): Move, Attack, UseAbility, Potion, Defend, Wait, EndTurn, and run-level actions such as Equip, Drop, Buy, Sell, Recruit, ChooseAscension, LearnSkill.
4. The sim takes state + action and returns new state + a list of events. The UI animates the events. An illegal action changes nothing.
5. **One math path:** the forecast and the real hit use the same code.
6. **Randomness:** every gameplay roll goes through the sim's seeded random. Floors are fixed per run. Every fight gets a fresh combat seed, and a fresh seed is rolled when a saved game loads, so reloading never replays the same dice. Only cosmetics may use unseeded random.
7. Game numbers live in data files, not buried in code.
8. Go compiles to WASM and runs on GitHub Pages. The old game stays playable at its link until the new one matches it.

## Confirmed decisions (2026-10-09)
1. **Match level:** the Go version matches the old game's combat numbers exactly (checked against recorded fights). Exact floor layouts are not required to match, because the old floor generator is not reproducible across browsers.
2. **Fight seeds and saves:** a fresh combat seed is rolled whenever a saved game loads.
3. **Old saves:** existing saves keep working, through a loader that reads today's save format.
4. **Fair shuffle:** the old game's four live random-comparator shuffles (enemy start slots, item combining, affix picks, multi-target extra targets) are replaced by one fair shuffle on the sim's seeded random. Listed as an intentional difference from the old game.

## Confirmed decisions (2026-10-10)
5. **The old game is a proof of concept, not law.** Every idea and mechanic carries over; matching it exactly is how we prove an idea came across correctly. Improvements are welcome, but each one is explained to the owner and signed off (or turned down) before it goes in, and then listed in `changes-from-old-game.md`.
6. **Mid-battle saves resume exactly where you left off** (same board, health, statuses and turn); only future dice are fresh (decision 2). Exact resume of old-game mid-battle saves is attempted in Phase 7.
7. **Monsters use the same skills as heroes,** the same version of each skill; only their animations differ (monster animations come in a later phase). The old game had a few monster-only versions; this is listed as a change.
8. **The weapon triangle (sword, lance, axe) shows in every forecast,** both when heroes attack and when monsters attack heroes.
9. **Enemy AI gets smarter as you climb** (Phase 9): it learns to go after healers and heroes weak to its attacks.
10. **Look and feel notes for Phase 8:** an enraged boss glows red; the camera waits for an attack animation to finish before jumping to the next hero; the battle goal shows once at the start and then lives in the pause menu; the skill-switching bonus is explained once in tips.
11. **Team select and deployment (rules in Phase 7, screen in Phase 8), like Fire Emblem's preparation screen:** 5 hero slots, plus one at floors 25, 50 and 75 (up to 8). Before each map you pick heroes from your whole roster and must fill every slot; only the heroes who fight earn XP. Heroes go on blue starting tiles you choose, with more blue tiles than heroes. If you don't choose, the game brings your highest-level heroes. Design aim: a strategy-heavy game where skill and game knowledge are always rewarded.
12. **Any fallen hero can be revived,** however they died (poison and bleeding included). In the game since 2026-10-10.
13. **An unarmed hero counts as their class's usual weapon** in the weapon triangle (for example Warrior sword, Tank axe, Knight lance), both when attacking and when being attacked. (Phase 6)
14. **New art (Phase 8):** the owner is drawing new character sprites with movement and new attack animations, and will hand over all the assets before Phase 8 starts, so Phase 8 is built on the new art from day one (no backtracking; the old sprites are not wired up). Claude wires them in and makes every hue effect and animation effect (glows, tints such as the enraged boss's red, hit flashes, skill effects, timing), with the camera waiting for each animation to finish. The art never decides a rule: it only plays the events the referee sends.
15. **Every class learns skills on one schedule (Phase 6):** 3 skills at Lv 1 (2 normal + 1 ultimate), then a new skill every 3 levels (Lv 3, 6, 9, 12, 15, 18, 21, 24, 27), a second ultimate at Lv 20, and the ascension ultimate at Lv 30. Every class ends with the same number of skills. By Lv 30 every skill is learned; after that, growth comes from levelling skills up. About 45 new skills are needed; each must be new and fresh (not a renamed copy), fit its class, and be approved by the owner before going in. Resurrection becomes a normal skill instead of an ultimate (so the Healer needs a new Lv 20 ultimate). Monsters follow the same skill schedule too, for more variety in fights. (Done: 53 new skills plus Last Light, approved by the owner on 2026-10-10; see skill-plan.md.)

16. **Loot tiers (decided 2026-10-11, built in 6d):** legendaries only start to drop once any one hero reaches Lv 30, and a legendary only drops for a hero who can use it (class, weapon rank and armor weight). A new purple tier (epic) sits between rare and legendary: weapons, armor and charms with their own bonuses, made for the middle levels, designed by Claude and approved by the owner item by item before going in, then balanced with the autoplayer. (Purple approved in full on 2026-10-11, items P1-P42 in purple-tier-proposal.md, with: power 6 kept, Executioner's and Warding in from the start, and gold chests 65% rare / 30% purple / 5% legendary from Lv 30. Built 2026-10-11.)
17. **Approved iron-outs (2026-10-11):** a battle ends the moment it is won; an undo that can't happen says why. Everything on the old-game feature checklist (feature-checklist.md) must come across; each phase ticks off its lines before it is called done.

## Existing rules to preserve
- Floors stay the same within a run; revisiting a cleared floor gives less gold and the same XP.
- Level cap tied to the monster level of the best floor reached.
- Boss first-clear legendary bonus, then back to a low base chance. This is a flag in run/save state, never UI.

## Planned after the port matches (Phase 9)
- Fire Emblem-style counterattacks; the forecast shows both sides, or "No counter" with the reason.
- Move, attack and enemy-threat tiles.
- Class-owned terrain: bush = Assassin (free cloak and crit, breaks on attack), tree = Ranger (climb for range), mountain = Knight / Dragon Knight, fort = Tank.
- Height: higher ground is harder to hit and hits a bit harder, with capped bonuses; climbing is slower for non-owner classes.
- The map influences who and what to bring.

## Clean-code rules
- Small files, clear folders. No gameplay math in UI code.
- Every rule has a test that runs without a screen.
- Changes must be explainable in one sentence.
- No change to unrelated systems when fixing a bug; bug fixes come with a regression test.
