# Changes from the old game

Every place where the new game intentionally behaves differently from the old one (v0.26.1, commit 7289405). Nothing changes silently: if it is not on this list, the new game matches the old one.

## Decided and in place
- The damage forecast runs the same math as the real hit, so the numbers you see before attacking are the numbers you get. (Agreed fix: the old forecast ignored crits and bonus damage, up to 54% off.)
- Every fight gets a new battle seed, and loading a save rolls a new one, so reloading never replays the same dice. (Agreed.)
- Shuffles are fair: when a multi-target skill picks its extra targets, every enemy in range has an equal chance. The old game used a shortcut that favoured some picks and could differ between browsers. (Approved 2026-10-09.)
- Monsters use the same version of a skill as heroes. In the old game six differed: monster Fireball, Thunder and Frost Bolt reached 1–3 tiles instead of 2–6, and monster Shot, Snaring Shot and Power Shot reached 7 tiles instead of 6. One version for everyone now (approved 2026-10-10). Their look will differ, not their rules.
- A fallen hero can be revived however they died. In the old game, heroes killed by poison or bleeding could not be revived. (Approved 2026-10-10.)
- Formulas that were written in several places in the old code (XP needed, potion healing, monster level, "is this skill usable?") exist exactly once. This changes no numbers.

- A Trapper's traps spring on any tile a monster walks over. The old game only noticed a trap on the tile a monster stopped on, because it checked after the walking animation. (Agreed fix; in place since Phase 6.)
- Every class learns the same number of skills on one schedule: two normal skills and an ultimate at Lv 1, a normal skill at Lv 3, 6, 9, 12, 15, 18, 21, 24 and 27, a second ultimate at Lv 20 and the ascension ultimate at Lv 30. All skills are learned by Lv 30. In the old game the Healer had 12 skills and most classes 7, with nothing new between Lv 20 and 30. 53 new skills and the Healer's new ultimate Last Light were added; Execution, Fortress Slam and Resurrection became normal skills (Resurrection is still once per battle); the Mage's Thunder moved to Lv 3; most old skills now arrive a little later. The full list is in skill-plan.md. (Approved skill by skill 2026-10-10; in place since Phase 6.)
- Monster types learn on the same schedule, from the hero skills plus their old monster skills, and fight with their 4 newest normal skills, their newest ultimate and Strike. Elites follow the schedule too; bosses keep their own kits. Monster healers can bring back a fallen monster with Resurrection. In the old game a monster had a fixed list plus 2-3 skills at set levels, and elites got none of the level skills. (Approved 2026-10-10; in place since Phase 6.)
- A hero with no weapon in hand counts as their class's usual weapon in the weapon triangle (Warrior sword, Tank axe, Knight and Dragon Knight lance), both when attacking and when attacked. The old game skipped the triangle for unarmed heroes. (Approved 2026-10-10; in place since Phase 6.)
- Putting on gear in the middle of a battle uses up the hero's turn. The old game let you swap gear any number of times for free. Switching weapon sets is still free. (Approved 2026-10-11.)
- After a level-up, a hero's attack counts the weapons in hand. The old game quietly went back to weapon set 1's weapons until the next set switch. (Approved 2026-10-11.)
- A Necromancer's Grave Bloom raises the skeleton right after the kill. The old game raised it on a half-second timer, so it could appear in the middle of the enemy turn depending on how fast you played. (The referee has no clock; in place since Phase 6.)

## Decided, arriving with later phases
- Team select and deployment: start with 5 hero slots; one more slot at floors 25, 50 and 75 (8 is the most). Before each map you choose who to bring from your whole roster and must fill every slot. Only the heroes who fight on a floor earn its XP. You place them yourself on blue starting tiles, and each map has more blue tiles than heroes, so where you start is a real choice. If you don't choose, the game brings your highest-level heroes. The old game took the whole party into every fight. (Approved 2026-10-10; arrives in Phase 7, screen in Phase 8.)

## Not changed yet (conscious design calls for later)
- Shop floors count as cleared without a fight.
- A defeat re-runs the same floor with no penalty.
