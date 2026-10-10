# Changes from the old game

Every place where the new game intentionally behaves differently from the old one (v0.26.1, commit 7289405). Nothing changes silently: if it is not on this list, the new game matches the old one.

## Decided and in place
- The damage forecast runs the same math as the real hit, so the numbers you see before attacking are the numbers you get. (Agreed fix: the old forecast ignored crits and bonus damage, up to 54% off.)
- Every fight gets a new battle seed, and loading a save rolls a new one, so reloading never replays the same dice. (Agreed.)
- Shuffles are fair: when a multi-target skill picks its extra targets, every enemy in range has an equal chance. The old game used a shortcut that favoured some picks and could differ between browsers. (Approved 2026-10-09.)
- Monsters use the same version of a skill as heroes. In the old game six differed: monster Fireball, Thunder and Frost Bolt reached 1–3 tiles instead of 2–6, and monster Shot, Snaring Shot and Power Shot reached 7 tiles instead of 6. One version for everyone now (approved 2026-10-10). Their look will differ, not their rules.
- A fallen hero can be revived however they died. In the old game, heroes killed by poison or bleeding could not be revived. (Approved 2026-10-10.)
- Formulas that were written in several places in the old code (XP needed, potion healing, monster level, "is this skill usable?") exist exactly once. This changes no numbers.

## Decided, arriving with later phases
- Team select and deployment: start with 5 hero slots; one more slot at floors 25, 50 and 75 (8 is the most). Before each map you choose who to bring from your whole roster and must fill every slot. Only the heroes who fight on a floor earn its XP. You place them yourself on blue starting tiles, and each map has more blue tiles than heroes, so where you start is a real choice. If you don't choose, the game brings your highest-level heroes. The old game took the whole party into every fight. (Approved 2026-10-10; arrives in Phase 7, screen in Phase 8.)
- A hero with no weapon counts as their class's usual weapon in the weapon triangle, both ways. The old game skipped the triangle for unarmed heroes. (Approved 2026-10-10; arrives in Phase 6.)
- Traps trigger along the path a unit walks, decided by the rules, not by animation frames. (Agreed fix; traps arrive in Phase 6.)

## Not changed yet (conscious design calls for later)
- Shop floors count as cleared without a fight.
- A defeat re-runs the same floor with no penalty.
