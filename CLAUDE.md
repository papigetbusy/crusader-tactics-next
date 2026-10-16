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
- Claude and Grok both write code here: follow `docs/working-together.md` (pull first, claim the task, tests and green tick before and after pushing).

## If your turn was restarted (owner, 2026-10-16)
Claude's turns sometimes fail, or the owner edits and resends a prompt. The new turn then finds the earlier turn's half-done work in the folder. That work is yours, not someone else's.
- **Start of every turn:** run `git status` and read `.claude-wip.md` if it exists, before changing anything.
- **`.claude-wip.md` exists:** an earlier run of you started this task. Its uncommitted changes are yours. Read them, run the tests, and carry on from the steps not ticked. Don't redo them and don't throw them away, unless the owner's new prompt changed the task. Then adjust them to fit.
- **Uncommitted changes but no `.claude-wip.md`:** they may be Grok's or the owner's. Look before touching them, and ask if unsure.
- **Before your first edit on a task,** write `.claude-wip.md`: the owner's request in his words, the time, the plan as `- [ ]` steps, and the files you expect to touch. Tick steps as you finish them. Delete it once the push is green.
- It is listed in `.gitignore`, so it never reaches GitHub. It only lives in this folder, for the next run of you.
- **Never commit with `git commit -a` or `-am`.** Run `git status`, then `git add` only the files of the task you are finishing. That way a half-done task never slips into an unrelated commit.

