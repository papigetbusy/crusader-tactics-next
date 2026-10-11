# Working on Crusader Tactics (next)

Read `docs/design-lock.md` and `docs/skeleton-design.md` before changing anything. They are the rules of this project.

- The owner is not a programmer. Explain in plain language, one sentence per choice. He checks work on his iPhone: the green tick on GitHub, the published pages, and playing the link.
- The old game (repo `papigetbusy/crusader-tactics`) is the answer key. Do not change it unless the owner asks.
- The old game's code is never used in the new build, including the game screen. Its art (images, sprites, textures) may be used as files. Otherwise it is only a reference for how the new game must look, feel and play. The screen is a new, thin client: it only draws the referee's View and sends actions, so online play stays possible.
- `sim/` is the referee. It may not use the browser, clock, network, files or unseeded random (`tools/guard` enforces this, plus the layer order and a 400-line file limit).
- Every rule gets a test named as a plain sentence (`t.Run("A unit cannot walk through water", ...)`). Those names become the Rule Book page.
- Old-game match checks use `testdata/oldgame/` only, never the live `data/`, so editing data never breaks them. Port the old code exactly; note the old line numbers.
- Intentional differences from the old game must be listed in docs (fair shuffle is approved).
- Push only when the owner has asked for the work. Commit messages say what changed in plain words.
