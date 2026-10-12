# Phase 9 proposal: new systems (for the owner to approve, item by item)

Written 2026-10-11. The ideas are already signed in `design-lock.md` (rule 9 and "Planned after the port matches"); this turns them into exact rules with numbers. Nothing here is built until the owner says yes. Every rule lives in the referee and in data files, with a plain-sentence test, so online play keeps working.

The owner moved the balance autoplayer to the end, so these numbers are a careful first guess, checked by playtesting; the autoplayer tunes them later.

## Already done (no work left)
- **Move, attack and enemy-threat tiles:** blue move tiles, red attack tiles, a monster's purple move zone and red attack zone, and the danger zone (Phase 8).
- **The map influences who you bring:** you see the map before choosing your team (Phase 7).

## P9-1 Smarter enemies as you climb (approved in principle; numbers to approve)
- **Floors 1-19:** monsters think exactly like the old game (no change).
- **From floor 20:** when a monster picks whom to attack, it also weighs:
  - a hero it can **defeat this turn** (strongest pull),
  - a **healer** (Healer class, or any hero carrying a heal),
  - a hero **weak to its element** (the WEAK hit).
- **From floor 50:** the same, plus monsters avoid ending their move next to two or more heroes when they can't attack anyone.
- Bosses and elites use the smarter thinking from floor 10.

## P9-2 Counterattacks (Fire Emblem style)
- When a unit is hit by a **single-target attack** and survives, it **strikes back once with its Strike**, if the attacker is inside its Strike range (melee hits back at melee, bows and guns hit back at range 2-6, staffs and wands at 2-6).
- **Power:** 50% of a normal Strike. Can miss and crit like any hit.
- **No counter when:** the defender is stunned, frozen, rooted by a net or cloaked; the attack was an area skill, an ultimate, a heal or buff; the attacker is out of the defender's range; or the attack killed.
- **Both sides counter** (heroes and monsters), so positioning matters.
- The **forecast shows both sides**: "Counter: 8 dmg · 72%" or "No counter: out of range".
- The Duelist's Riposte stays and is stronger than a plain counter (60%, twice with Flurry); a Duelist uses Riposte instead of the plain counter.

## P9-3 Class-owned terrain
- **Bush (new tile, in grass and forest edges, a few per map):** anyone in a bush takes 10% less damage. An **Assassin** in a bush is **cloaked** (can't be targeted) and gets **+25% crit**; attacking breaks the cloak until the Assassin leaves the bush and enters one again.
- **Tree (forest):** a **Ranger** standing in a forest gets **+1 range** with bows, and forest costs a Ranger only 1 move.
- **Mountain:** **Knight and Dragon Knight** can stand on mountains (the Knight pays 3 move; the Dragon Knight already flies) and take **20% less damage** there. Mountains still block everyone else.
- **Fort:** a **Tank** in a fort takes **40% less damage** (others 25%) and heals **15%** a turn (others 10%).

## P9-4 Height
- Hills are high ground (the maps already have them).
- **Attacking from higher ground:** +10% damage and +10% hit.
- **Attacking uphill:** −10% hit.
- Capped: only one step counts (hill vs flat); mountains count as hills for the Knight and Dragon Knight.
- **Climbing a hill** costs 1 extra move for everyone except fliers, Rangers and Knights.
- The forecast shows "High ground +10%" or "Uphill −10% hit".

## Order of work, once approved
1. P9-1 smarter enemies (smallest change, already approved in principle).
2. P9-2 counterattacks (the biggest change to how fights feel).
3. P9-4 height.
4. P9-3 class-owned terrain (needs the new bush tile on the maps and in the art).

Each one: rules and tests in the referee, the forecast and Codex updated, listed in `changes-from-old-game.md`, pushed, then a playtest link.
