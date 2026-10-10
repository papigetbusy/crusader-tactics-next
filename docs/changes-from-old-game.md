# Changes from the old game

Every place where the new game intentionally behaves differently from the old one (v0.26.1, commit 7289405). Nothing changes silently: if it is not on this list, the new game matches the old one.

## Decided and in place
- The damage forecast runs the same math as the real hit, so the numbers you see before attacking are the numbers you get. (Agreed fix: the old forecast ignored crits and bonus damage, up to 54% off.)
- Every fight gets a new battle seed, and loading a save rolls a new one, so reloading never replays the same dice. (Agreed.)
- Shuffles are fair: when a multi-target skill picks its extra targets, every enemy in range has an equal chance. The old game used a shortcut that favoured some picks and could differ between browsers. (Approved 2026-10-09.)
- Formulas that were written in several places in the old code (XP needed, potion healing, monster level, "is this skill usable?") exist exactly once. This changes no numbers.

## Decided, arriving with later phases
- Traps trigger along the path a unit walks, decided by the rules, not by animation frames. (Agreed fix; traps arrive in Phase 6.)

## Not changed yet (conscious design calls for later)
- Shop floors count as cleared without a fight.
- A defeat re-runs the same floor with no penalty.
