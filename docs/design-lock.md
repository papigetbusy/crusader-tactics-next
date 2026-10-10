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
