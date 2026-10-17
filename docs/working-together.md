# Working together: Claude and Grok on one game

Paste this to either AI at the start of a session. Both of us write code in this repository, so we follow the same rules and never break each other's work.

## Who does what (owner, 2026-10-12)

Claude builds the systems. That is the heavy coding.

Grok does not start a new system. Grok works on how the game looks and feels, the art direction, the lag, and fine-tuning mechanics that already exist. Grok also checks that the split stays clean: rules in the referee, the screen only draws.

New attack sprite sheets are paused. The owner is out of OpenArt credits. Do not ask for new sheets and do not invent new attack art until he says credits are back. The prompt rule is locked in `new-animations/PROMPT.md`. Already signed off, and still not in the game: Blood Moon, Soul Nova, Grand Hex, Blood Frenzy.

## Showing the owner a picture (owner, 2026-10-17)

Pictures pasted in the chat do not reach his phone. When he asks to see art, save real PNG files, put them in a folder under `new-animations/` with a plain `index.html` that shows every picture, push that folder, and give him the GitHub Pages link (`https://papigetbusy.github.io/crusader-tactics-next/new-animations/...`). Label each picture on the page. Do not send only names or numbers. Once the owner has picked, add the folder to `new-animations/unpublished.txt`: it stays in the repo but is no longer published.

## Item art style (owner, 2026-10-17)

Item pictures are painted pixel art, the same kind as the class legendary weapons and the sheets the owner sent (swords, axes, lances, bows, daggers, staves, guns). Dark navy background. Each icon sits in its own rounded square frame. The object is fully drawn: metal shading, a dark outline, gems, glow, and small ornaments. Three different designs, not three recolors of one shape. No letters on the picture. Never draw flat shapes, stick figures, or simple blobs. The stick-figure icon page was rejected and removed.

## The prompt

You are working on Crusader Tactics (next) in the repository papigetbusy/crusader-tactics-next. Another AI (Claude or Grok) also writes code here. Follow these rules exactly.

**Before you start**
1. Run `git pull --rebase`. Never start from an old copy.
2. Read `CLAUDE.md`, `docs/design-lock.md`, `docs/next-steps.md` and the "Now working on" list at the bottom of this file.
3. Add one line to "Now working on": who you are, the task, the files you expect to touch, the date. Commit and push that line first. If the other AI is listed on the same files, pick a different task or ask the owner.

**While you work**
4. Rules live only in `sim/` (the referee). The screen (`web/game/`) only draws the referee's View and sends actions. Never put game rules or game state in the page. Never use the old game's code; its pictures only.
5. Numbers go in `data/`, never written into code.
6. Every new rule gets a test named as a plain sentence (`t.Run("A unit cannot walk through water", ...)`).
7. Don't change a rule, a number or how something plays unless the owner approved it. Approved changes are listed in `docs/changes-from-old-game.md` (and in `sim/content/changes.go` if data changed).
8. Keep each change small and finished. One task, one or a few commits, each with a plain-words message.
9. Don't edit a file the other AI listed under "Now working on" unless the owner asks.

**Before you push**
10. Run `gofmt -l .` (must print nothing), `go vet ./...` and `go test ./...` (all must pass, including the old-game match checks).
11. If you changed `data/` or `docs/changes-from-old-game.md`, run `go run ./tools/skillbook` and `go run ./tools/codex`, and commit what they write.
12. If you changed the screen, open the game in a phone-sized browser and play the part you changed. Fix any console errors.
13. Run `git pull --rebase` again, re-run the tests if anything came in, then push. Never force-push. Never rewrite history someone else pushed.

**After you push**
14. Wait for the green tick on GitHub. If it is red, fixing it comes before anything else, whoever broke it.
15. Remove your line from "Now working on". Add a short line at the top of `docs/changelog.md` saying what changed, in plain words, and update `docs/next-steps.md` if the work was on it.
16. Tell the owner in plain language what changed and give the test link.

**If you're unsure**
17. If two docs disagree, `docs/design-lock.md` wins, then the owner's latest words. Ask the owner rather than guess on anything that changes how the game plays.

## Now working on
(one line each: who · task · files · date; remove it when done)











