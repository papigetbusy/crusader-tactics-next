# Crusader Tactics: Skeleton Design (Phase 3, draft for review)

**What this is:** a map of the new Go project. No game code is written in this phase, and the old game is untouched. Every choice has one "Why" line. Section 7 lists the decisions I need from you.

**How to read it:** three words matter. The **referee** (the "sim", written in Go) knows every rule. The **screen** (the "UI") only draws and asks questions. An **action** is a request ("move this hero here"). An **event** is a fact the referee hands back ("the hero moved", "the Slime took 14 damage").

```
  YOU TAP  ->  SCREEN (web/)  ->  BRIDGE (adapters/wasm/)  ->  REFEREE (sim/)
                                                                checks the rule,
                                                                rolls the dice,
                                                                updates the game
  YOU SEE  <-  SCREEN plays   <-  BRIDGE hands back        <-  new game state
               the events         the events                    + list of events
```

## 1. Folder layout and the repository decision

### 1.1 Recommendation: a new repository, not a subfolder

Working name: `crusader-tactics-next` (you can pick another). The old repository stays exactly as it is.

- **Why:** the old repo is one 5 MB file plus its own offline helper, and mixing a folder tree into it would blur which file belongs to which game.
- **Why:** the design lock says the old game stays playable at its link until the new one matches it, and a separate repo makes that automatic.
- **Why:** future developers, the online server and the Steam and iOS wrappers can join the new repo without ever touching the old code.
- **Why:** when the new game matches, we can swap the two by renaming the repos, so the link you share stays the same (I will double-check this when we get there).
- **Saves:** browsers keep saves per website address, and every Pages site under your account shares one address (`papigetbusy.github.io`). So the new game should be able to read your old saves wherever it lives. The new game will use its own save names and only read the old ones, never write to them. I am confident of the web rule, and Phase 7 will prove it with a real save.

### 1.2 The catch: getting a folder tree into GitHub from your iPhone

- Phase 2 worked because it was three flat files. As far as I know, the iPhone file picker uploads single files, not folders, and the real project will have around a hundred source files in folders.
- My workspace is temporary. Today, source code that only lives there disappears when the session ends, so the repo has to be its permanent home.
- Today my GitHub link is **not connected**: when I tried to list your repositories, it answered "no GitHub account linked".
- **Recommendation:** connect GitHub in your Claude settings so I can push straight to the new repo, but only when you ask me to. I have not tested pushing yet, so the first step after you approve is a tiny rehearsal (section 7).
- **Fallback if you would rather not connect it:** I hand you the finished site as a few flat files each phase, but the source then has no permanent home, so I do not recommend it.

### 1.3 The build robot

On every change, GitHub's built-in robot (called Actions) runs all the tests, builds the Go program into WASM, publishes the site, and puts a green tick or a red cross next to the change.

- **Why:** you never upload built files by hand, and you can see on your phone whether everything passes.
- It needs one switch in the new repo's Pages settings (publish from "GitHub Actions"). I have not tested this path yet, which is why it gets its own rehearsal.

### 1.4 The folder map

```
crusader-tactics-next/
  README.md           start here: this map in plain words
  go.mod              Go's project file (the project's name tag)
  docs/               design lock, this map, decision log, "differences from old game"
  data/               EVERY game number, as files you can edit (section 4)
  sim/                THE REFEREE: all rules. No screen, no clock, no internet.
    rng/                our own small dice generator
    content/            reads and checks the data/ files
    state/              the shape of a battle and of a whole run
    action/             the list of actions and events (the vocabulary, section 2)
    board/              grid, terrain, distance, who can walk where
    stats/              numbers for a hero or monster from class, level, gear, ascension
    combat/             hit chance, damage, crits, statuses; preview and apply
    items/              gear, affixes, loot, shop prices
    floor/              floor builder, spawns, chests, monster level formula
    save/               versioned saves, plus the loader for old saves
    ai/                 enemy and companion decisions (they send actions like you do)
    turn/               phases, start and end of turn effects, win and lose
    run/                the whole run: party, bag, gold, waiting decisions, progress
    engine/             the front door: Apply, Preview, LegalActions, Reach, Targets
  adapters/
    wasm/               the bridge the browser page calls (the only code that knows the browser)
  web/                  THE SCREEN: page, drawing, menus, camera, animations
    art/                sprites and tiles (pictures, not rules)
  testdata/
    oldgame/            Phase 1 answer key: recordings, tables, checksum
    scenarios/          hand-written "scenario cards" (section 6)
  tools/                helper programs: old-game match checker, data checker, rule-book maker
  .github/workflows/    the build robot
```

Later, `adapters/server/` joins for online play. It reuses `sim/` untouched.

**The layering rule:** a folder may only use folders listed *above* it inside `sim/`. For example, `combat/` can use `board/` and `state/`, but `board/` can never use `combat/`.

- **Why:** rules cannot quietly depend on each other in circles, which is how the old game became tangled.
- **Three guard tests enforce it:** `sim/` may not touch the browser, the clock, the network or Go's global random; no file may be longer than 400 lines (a number you can change); and the layering above. A rule broken by accident turns the robot red.
- `state/` exists so that "what a unit is" is defined in exactly one place, and rules receive the state as an argument instead of reading hidden globals (fixes the audit's risk R3).

## 2. Actions and events

**How it works:** you call `Apply(state, action)`. The referee returns the new state and a list of events, or a plain "No, because..." and nothing changes. Players, enemy AI and companions all use this same door.

- **Why:** there are no special cases for enemies, and online play later is just the same actions sent over the internet.

**Four questions the screen may ask without changing anything:**

| Question | What it answers |
|---|---|
| `LegalActions` | Everything the player may do right now (the screen greys out the rest). |
| `Reach(unit)` | The tiles this unit can walk to. |
| `Targets(unit, skill)` | Who or what this skill can hit. |
| `Preview(action)` | What would happen, including hit chance and damage range. It runs the same code as `Apply` and just does not commit. |

- **Why:** this is the "one math path" rule, and it ends the forecast drift (+54% off in the audit).

**Not actions:** selecting a unit, inspecting, zooming, camera, menus, settings, animation speed. **Why:** they change nothing in the game, so the referee never hears about them.

**How the bridge talks:** plain JSON text in, plain JSON text out, exactly as the Phase 2 page did with the dice. **Why:** it proved it works on your iPhone, and a future server can speak the identical messages.

### 2.1 Battle actions

| Action | Inputs | Allowed when | Events it can produce |
|---|---|---|---|
| **Move** | unit, destination tile | its side is up; the unit has not moved this turn; not snared, locked or shocked; tile within walking reach | `UnitMoved` (with the path), `ChestOpened`, `ItemFound`, `GoldFound`, `TrapTriggered`, `OverwatchFired`, `ObjectiveReached` (throne seized, then `BattleWon`) |
| **UndoMove** | unit | it moved but has not acted; the old tile is free | `MoveUndone` |
| **Attack** (any skill) | unit, skill slot, target (unit or tile) | skill off cooldown; target in range; unit not finished | `SkillUsed`, `Hit` (damage, crit, element), `Missed`, `Healed`, `ShieldGained`, `StatusApplied`, `UnitDied`, `UnitSummoned`, `UnitRevived`, `XpGained`, `LevelUp`, `GoldFound`, `ItemFound`, `ExtraTurnGranted`, `BattleWon`, `BattleLost` |
| **UseAbility** | unit, and a target unit or tiles for the targeted ones | ability off cooldown; unit not finished (a few are free) | `AbilityUsed`, `StatusApplied`, `HpPaid`, `FieldPlaced`, `TrapPlaced`, `UnitSummoned`, `UnitRevived`, `TurnRefreshed` |
| **Potion** | unit | HP below max; potions left | `PotionUsed`, `Healed` |
| **Defend** | unit | unit not finished | `DefendStarted`, `UnitFinished` |
| **Wait** | unit | unit not finished | `UnitFinished` |
| **SwapWeaponSet** | unit, set | the skill needs the other set | `WeaponSwapped` |
| **EndTurn** | none | your side is up; no decision is waiting | `PhaseEnded`, then the **whole enemy phase** as events (every enemy and companion action, `StatusTicked`, `ReinforcementsArrived`), `TurnAdvanced`, `PhaseStarted`, then `BattleWon`, `BattleLost` or your turn again |

Notes, each with its reason:

- **"Attack" is one verb for every skill**, including heals, buffs, summons and revives, as the old game's `act` does. **Why:** one door means one math path. (If you prefer the name `UseSkill`, it is a one-word change.)
- **The enemy phase comes back all at once**, and the screen plays it with delays. **Why:** no timers live inside the rules, and the old `busy` flag splits cleanly into the referee's phase and the screen's "still playing animations".
- **Auto end-turn is a screen setting.** The screen sends `EndTurn` itself when everyone is done. **Why:** the old game let a drawing function end your turn; the referee never will.
- **"Keep the hero selected for extra shots"** (Quick Draw and similar) becomes `LegalActions` still listing that unit. **Why:** the old 12-second polling hack disappears.
- **Traps trigger along the walked path, decided by the rules, not by animation.** **Why:** this is the animation-frame bug we agreed to fix.
- Whether Swap uses up the turn, and every other detailed rule, is copied from the old game's code when that action is built in Phase 4 or 5.

### 2.2 Run-level actions (outside a fight)

| Action | Inputs | Allowed when | Events it can produce |
|---|---|---|---|
| **StartRun** | starting party, run seed | no run in progress | `RunStarted` |
| **EnterFloor** | floor number, **battle seed** (supplied from outside) | no decision waiting; floor is next or already cleared; also used to retry after a defeat | `FloorStarted` (map and units), `ShopOpened` (shop floors), `DecisionQueued` |
| **Equip** | hero, item, slot | item in bag; hero's weapon rank and armor class allow it | `GearChanged` |
| **Unequip** | hero, slot | slot holds an item | `GearChanged` |
| **Drop** | item | item in bag | `ItemDiscarded` |
| **Buy** | shop item | shop open; enough gold | `ItemBought`, `GoldChanged` |
| **Sell** | bag item | shop open | `ItemSold`, `GoldChanged` |
| **Combine** | two matching items | shop open; two copies | `ItemsCombined` (rolls dice), `GoldChanged` if a cost applies |
| **LeaveShop** | none | shop open | `ShopClosed`; the floor-70 "shop first, then boss" flow lives here |
| **Recruit** | which offered hero | a recruit decision is waiting | `HeroRecruited`, `DecisionResolved` |
| **ChooseAscension** | hero, ascension | level 30 reached | `AscensionChosen`, `StatsChanged` |
| **Respec** | hero, ascension | once per floor | `AscensionChosen` |
| **LearnSkill** | hero, skill, slot to replace | a learn decision is waiting | `SkillLearned` or `SkillReplaced` |
| **ReplaceUltimate** | hero, ultimate | a decision is waiting | `UltimateSet` |
| **BondCompanion** | hero, keep or swap | a companion decision is waiting | `CompanionBonded` |
| **AcceptGift** | which offered gift | a gift decision is waiting (every 100 floors) | `GiftReceived` |

- **Waiting decisions are saved state.** The referee keeps a queue of them (the old `EVQ`) and only accepts the answer to the first one. **Why:** the old game lost a recruit or a skill if you closed it before answering.
- **Shop-floor clearing and defeat-retry** are the "conscious design calls for later". They live in `LeaveShop` and `EnterFloor`, so changing them later touches one place.
- **Not actions:** `Save(state)` gives bytes, and `Load(bytes)` gives a state with a **fresh combat seed**. Export, import and new-game screens call these.
- A refused action returns a reason code such as `TooFar`, `OnCooldown`, `NotYourTurn` or `DecisionWaiting`, and the screen shows it in words.

## 3. State: who owns what

**The rule of thumb (and the Why):** if a thing can change who wins or what you own, the referee owns it; if it only changes how things look, the screen owns it.

| Thing | Owner | Notes |
|---|---|---|
| Heroes (class, level, XP, gear, learned skills, ascension), bag, gold, best floor | Referee (run state) | changed only through events |
| Floor memory (opened chests, cleared, boss-legendary given, monster level, biome, shop stock) | Referee (run state) | monster level is computed once, then frozen |
| Run seed, milestone counters, decisions waiting | Referee (run state) | saved, unlike today |
| Map, heights, terrain, biome, units with **stable IDs**, HP, position, statuses, cooldowns, flags | Referee (battle state) | every unit, including summons and reinforcements, gets an ID |
| Turn number, objective, reinforcement schedule, chests, fields, traps, fallen list, **phase** (explicit) | Referee (battle state) | replaces the `busy` flag |
| Last-move record (for Undo) | Referee | it is a rule, so it lives with the unit |
| Dice state | Referee | part of the saved state |
| Selection, inspecting, current mode, open menus and dialogs | Screen | never saved |
| Camera, zoom, rotation, minimap | Screen | |
| Animations: units mid-walk, shake, flashes, floating numbers, death ghosts, effects | Screen | a table keyed by unit ID, replacing the animation fields stuck on units today |
| Settings: speed, confirm end-turn, auto end-turn, flat view, tips seen | Screen | saved under its own storage name, apart from the game save |
| Which labels, icons and colors a status shows | Screen | presentation only |

- **Why one bundle:** the referee's state is a single plain bundle of data that can be saved whole at any moment, including mid-fight, so a saved fight resumes exactly (fixes the lossy battle snapshot).
- **What the screen receives:** a read-only copy of the state plus the events, never the referee's internals. **Why:** the screen cannot change a rule by accident.

## 4. Where every game number lives

### 4.1 Three kinds of "number", three homes

1. **Plain numbers and tables** (class stats, skill power, prices, terrain costs) live in `data/` files you can edit.
2. **Formulas** keep their *shape* in Go (for example "damage is power times attack, once rounded"), but every constant inside them lives in `data/`.
3. **Behaviours** (what a skill, ability, status, affix or ascension passive *does*) are written in data as a short recipe from a fixed menu of effects. Each effect is written once in Go and tested.

- **Why:** you can change a number or reshape a recipe without touching code, and the few truly unusual cases stay few and visible.

### 4.2 The 30 dumped tables become these files

| New file | Built from (dumped table, entries) | What it holds |
|---|---|---|
| `classes.json` | `CL` 26, `FORMS` 10, `ROLE` 10 | 10 hero classes plus the elite and boss types: base stats, weapon, starting skills, evolution names |
| `skills.json` | `SK` 109, `MSK` | every skill, plus monster skills unlocked at set levels |
| `ascensions.json` | `SPEC` 30, `SPECS` 10 | the 30 ascensions: stat changes and passive recipes |
| `abilities.json` | `AB2` (10 classes) | about 60 class abilities as recipes. `ABIL` is dead data and is dropped; `SU` is a builder function that the data itself replaces |
| `statuses.json` | `STD` 46 | the rules half of each status (duration, ticks, what it modifies) |
| `elements.json` | `EB` 5 | the element wheel |
| `monsters.json` | `MON` 30, `MNAMES` 30, `ELT` 9, `BOSS` 6 | monster names, elements, and which are elites or bosses |
| `terrain.json`, `biomes.json` | `TINFO` 9, `BIOMES` 7 | move cost, defense, healing per terrain; biome looks and mixes |
| `affixes.json` | `AFX` 27, `AFW` 26 | gear affixes and how often each rolls |
| `weapons.json` | `WPN` 32, `WTN` 8, `RANKS` 6, `RREQ` 4 | weapons, types, rank requirements |
| `uniques.json`, `sets.json`, `gear_tiers.json` | `UNQ` 19, `SETS` 3, `TIER` 4 | unique items, set bonuses, rarity multipliers |
| `tuning.json` | `TUNE` 18 | the global dials (renamed in plain words as we port each one) |
| `version.json` | `VER`, `VNAME` | the game's label (this is not the save-format version) |

- **Pictures leave the rules.** `MON` is 555 KB because almost all of it is sprite images. Art moves to `web/art/`. **Why:** the rules data stays small enough to read on a phone.
- **Display text leaves the rules too:** a status's label formula, icon and color are presentation, so they move to the screen. Skill descriptions stay with their skill.

### 4.3 Numbers hiding in the old code move into data

Found in the audit, to move into `data/rules/` files with plain names:

- the XP needed per level (`lv x 6`, written three times) and the potion heal (`10 + 3 x level`, twice)
- the damage spread (90% to 110%), skill power growth per level, enemy AI choices (heal under 75%, special skill 45% of the time)
- floor rhythm: boss every 10, elite every 5 and 50, shop every 7, seize-the-throne and reinforcement floors and their turns
- map size and enemy count formulas, the monster-level formula, the gold halving on cleared floors
- recruit every 25 best floors, legendary gift every 100, hero cap 8, evolution at levels 5, 20 and 30, potions refilling to 3

**Why:** this is what "numbers live in data, not buried in code" means in practice, and it also removes the audit's copies of the same number.

### 4.4 Abilities: the closures become recipes

Today each ability is a mini program. Looking at all ten classes, nearly every one does the same two things: give a unit a timed status, and set some numbers for it. The status's effect is then applied elsewhere through if-chains. In the new design an ability is a **card** like this (the real file is JSON in this same shape):

```
Blood Frenzy   (Berserker ascension ability, cooldown 5)
  text:   "Spend 20% of your HP: for 3 turns, +50% damage and 25% life steal.
           Each kill adds a turn."
  steps:  1. pay_hp        20% of max HP, never below 1 HP
          2. gain_status   frenzy, 3 turns
```

and the status itself is a card in `statuses.json`: `frenzy` = +50% damage dealt, 25% life steal, each kill adds a turn.

About a dozen step types cover the roughly 60 abilities: `gain_status`, `pay_hp`, `place_field`, `place_traps`, `mark_target`, `revive_ally`, `summon_clones`, `raise_skeletons`, `refresh_ally`, `set_overwatch`, `take_off`, and `end_turn`.

- **Why:** a status's effect is written once and every ability that grants it shares it, instead of the same logic copied across `dcalc`, `hurt` and `act`.

### 4.5 One vocabulary for statuses, ascension passives, affixes and set bonuses

All four are the same idea: "when X happens, if Y is true, change Z by this much". So they share one **modifier** format, and gear affixes (which are already shaped that way) fit in as they are.

| Part | What it can say |
|---|---|
| When | stats are built, hit chance, damage dealt, damage taken, on hit, on kill, turn start, turn end, an action rule |
| If | HP below a percent, the unit has not moved, the target has status X, the target is an elite or boss, the skill's element, the skill is single-target, an ally is adjacent or within N tiles, a charge of 3+ tiles |
| Then | add or multiply damage, crit chance or damage, ignore a share of defense, dodge, immunity to a status, heal or shield, apply a status, hit one more target for a share, gain an extra move or action |
| Who | the unit itself, or allies within N tiles (an aura) |

### 4.6 The 30 ascension passives

I read all 30 passive texts. My first read:

- **About 24 fit the menu directly.** Examples: Berserker (+40% damage while below half HP, kills heal 15%), Sniper (+35% if you have not moved, ignores 30% defense), Phantom (30% dodge), Viper (every hit bleeds, +30% damage to bleeding enemies), Warlord (allies within 2 tiles deal 15% more).
- **About 6 need a small named Go routine**, with their numbers still in data: Necromancer's Soul Harvest, Saint's overheal-to-shield, Guardian's bodyguard redirect, Templar's charge heal, Duelist's riposte (until the general counter system arrives in Phase 9), and Oracle's Foresight.
- **The risk, in one sentence:** matching the old numbers exactly depends on applying each modifier at the same step of the damage calculation with the same rounding, so every modifier carries an explicit step number and the old-game match checks will catch any slip.
- **The honest gap:** the Phase 1 recordings do not cover ascensions or abilities, so each gets its own recordings *before* it is ported.

### 4.7 Skills, exactness, and the safety net

- **Skills:** the 109 skills are already almost pure data (power, range, cooldown, effects, flags like `aoe`, `line`, `multi`). They stay data, with a schema file documenting every flag. A flag's behaviour is written once in Go.
- **Exactness:** one rounding helper copies JavaScript's rounding (it differs from Go's for negative halves), all math is 64-bit decimals, and rounding happens at the same places as the old game. **Why:** this is how "match the old numbers exactly" is kept.
- **The data checker:** a test reads every data file and checks it (every skill a class names exists, no negative cooldown, every status has a duration rule). **Why:** a typo in a data file turns the robot red with a plain message, instead of crashing the game.
- **Editing a number from your phone:** you change a value in a data file on GitHub, the robot re-runs the tests, and the pages in section 6 update. The old-game match checks always read the *saved old tables* in `testdata/oldgame/`, never the live `data/`, so your edits never break them. A generated **"Changes from the old game"** page lists every difference between your live data and the old tables.

## 5. How dice work

### 5.1 One generator, our own

- A small 64-bit generator (about 20 lines) written inside `sim/rng/`.
- **Why:** Go's built-in generator is allowed to change between Go versions, which could silently break replays and online play; ours produces the same numbers on every machine forever. (Phase 2's page used Go's built-in one for the rehearsal only.)

### 5.2 Two seeds, and who supplies them

- The **run seed** is fixed for the whole run and builds the floors, so a floor looks the same every time you visit it.
- The **battle seed** is new for every fight, and a new one is rolled when a saved game loads.
- **The referee never asks the clock or the system for randomness.** The outside world hands the battle seed in with `EnterFloor` (the browser uses the same secure random the Phase 2 page used, and an online server will choose it itself so nobody can shop for lucky seeds). It is recorded with the action. **Why:** this keeps the referee fully replayable, which is what Phase 10 needs (a whole run as the run seed plus the list of actions).

### 5.3 Separate streams for separate jobs

- Floor building, combat, enemy AI and loot each draw from their own stream derived from the seed.
- **Why:** adding a loot roll can never shift the combat rolls, so replays and balance edits stay stable.

### 5.4 What a roll looks like

- The dice give a fraction from 0 up to just under 1 (like the old `Math.random`), with whole-number helpers on top, and every draw is labelled with its job ("hit roll", "damage spread", "crit roll", "AI skill pick").
- **Why:** Phase 1 stored the old dice as fractions with the code location that drew them, so the labels line up one for one, and a failed check can say exactly which roll was out of order.

### 5.5 The recorded-dice feed (for the old-game match checks)

- In tests only, the dice are replaced by a feed that serves the Phase 1 recorded values in the old order, and the referee cannot tell the difference.
- After every step the test compares positions, HP, statuses **and how many dice were used**, and reports the first difference in plain words.
- **Why:** this is how the numbers are proven identical without needing the old game's random generator.

### 5.6 The shuffles

The old game has six places that shuffle a list with `sort(() => Math.random() - .5)`. By the audit's own note, the result of that trick depends on the browser, and it is known to favour some orders over others. Checking where each one lives:

| Where | Used for | Status |
|---|---|---|
| `newFloor0` (line 753) | where enemies start on a floor | live; floors do not need to match |
| `combineAf` (1017) | combining two items | live; loot is not compared number for number |
| `rollAffixes` (1203) | which affixes an item gets | live |
| `rollAffixes` (1013 and 1069) | the same job | **dead duplicate copies**, they vanish on their own |
| `act` (1721) | picking the extra targets of a multi-target skill | live, and the only one that affects combat |

- **Decision:** replace all four live ones with one proper fair shuffle (a standard method) drawing from the seeded streams. **Why:** it behaves the same on every machine, and Go cannot copy a result that depends on the browser.
- **What changes in play:** the extra targets of multi-target skills are now chosen evenly at random instead of with the old quirk. **Approved by the owner on 2026-10-09** and listed as an intentional difference.
- **A gap to close first:** the Phase 1 fights never drew from that combat shuffle, so multi-target skills need recordings of their own before they are ported. For that spot the check is that the extras are valid enemies in range and the count is right.
- **Cosmetic randomness** (sparkles, flicker) stays in the screen, unseeded, and never reaches the referee. **Why:** frame timing can never leak into outcomes.

## 6. How you read a test without knowing code

### 6.1 Four pages, all in the gothic look, published beside the game

| Page | What you see |
|---|---|
| **Rule Book** | One plain sentence per rule, grouped (Movement, Hitting, Damage, Statuses, Floors, Loot, Saves), each with a tick or a cross. The sentence is the test's name. |
| **Old-Game Match** | One row per recorded fight. Green means Go matched the old game at every step; a red row names the first difference (for example "Fight 57, step 14: the old game's Mage took 31 damage, Go says 29"). This grows out of the Phase 1 report page. |
| **House Rules** | The guard tests: the referee never touches the screen, clock or network; no file is too long; every number lives in `data/`; every data file is valid. |
| **Changes from the old game** | Every intentional difference, so nothing changes silently. It starts with: forecast now equals the real hit (agreed), traps follow the walked path (agreed), duplicated formulas unified (no number changes), a new battle seed each fight and each load (agreed), and the fair shuffle (approved 2026-10-09). |

- **Why:** this is the Phase 1 green/red checklist idea, kept up to date by the robot after every change.

### 6.2 Scenario cards

Small plain-text files in `testdata/scenarios/`: a tiny map, who stands where, the dice fixed, one thing done, and what must happen. Two cards, with **invented numbers just to show the shape**:

```
CARD: Water blocks walking                     (map: . plains  ~ water  K Knight)
  Map:      K . ~ . .
  Do:       Move Knight to the far right tile
  Expect:   No, because the path is blocked

CARD: The lowest damage roll
  Dice:     hit roll 0.00 (always hits), damage roll 0.00 (lowest, 90%)
  Do:       Warrior attacks Slime with Slash
  Expect:   Slime loses 11 HP          (real cards use numbers from the old recordings)
```

- **Why:** you can read a card like a board-game rule, and you or Claude can write a new one in a minute without code.

### 6.3 How it works on your iPhone

1. After each change, GitHub shows a green tick or a red cross.
2. You open the Rule Book and Old-Game Match links.
3. If something is red, tap the row, copy the sentence, and send it to me.

You never need to read Go. A phase is **done** when all four pages are green and you have played it by hand.

## 7. Decisions I need from you

**Needed before Phase 4:**

1. **New repository** called `crusader-tactics-next` (or your name), with the old one untouched. OK?
2. **Connect GitHub** in your Claude settings so I can push to the new repo when you ask (I never push on my own). OK?
3. **Build robot:** may I run a small rehearsal first? It would publish the Phase 2 test page from the new repo through the robot, to prove the new pipeline before real code depends on it.
4. ~~Fair shuffle~~ **Decided 2026-10-09: yes.** All four live shuffles use one fair shuffle.

**I will assume yes unless you object:** the single verb "Attack" for every skill; Undo and weapon-swap stay as actions because the old game has them; our own dice generator instead of Go's; a 400-line file limit.

**What I could not verify:** pushing from my side to a new repo, the Actions-to-Pages publishing, and the iPhone folder-upload limit (stated "as far as I know"). The exact split of the 30 passives is a first read, settled as each is ported. Folder names may shift slightly when the first slice is built, but the layering rule stays.

**After you approve:** the rehearsal, then Phase 4 (Move, Attack, End Turn checked against the Phase 1 recordings).
