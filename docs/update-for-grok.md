# Update for Grok (2026-10-11)

Things that changed today that you may have missed reading the repo. Start with `next-steps.md` for the full order.

## Big decisions by the owner
- **The old game's code is never used in the new build** (design-lock 18). Its art may be used as files (now in `web/assets/`, named by `tools/art`); otherwise the old game is only the reference for look and feel.
  - The first Phase 8 attempt wired the referee under the old game's page. That is gone.
  - `web/game/` is a new screen written from scratch.
- **The page keeps no game state.** It stores only the referee's last answer (the View) to draw it, and sends actions. Everything goes through `web/game/js/referee.js`, so the same screen can talk to a server later.
- **Mindset** (design-lock 19): single-player is perfected first, matching the original in every rule, animation and movement. Online 1v1 versus comes only after that, and reuses all of it. Nothing may be built in a way that blocks multiplayer.
- **Online 1v1 rules (a skeleton for the vision, details later):**
  - each player picks 8 heroes, all at Lv 30, with chosen ascensions
  - full purple gear sets for everyone
  - chess.com-style clocks: per turn plus an overall time per player
  - matches of about 5 minutes
  - needs a small host such as Fly.io or Render, since GitHub Pages can't run a server
- **On-screen messages fade after 3 seconds.**

## What the new screen does now (Phase 8a–8k)
- **Title and setup:** title screen, team picker cards, placing heroes on blue tiles.
- **Moving and acting:**
  - tap or drag to move
  - tap a monster to attack; the hero walks closer first if needed, using the referee's plan
  - the full forecast window
  - the wooden command menu beside the hero, with a Skills fly-out
  - skills and class abilities aimed in several taps
- **Animations:** walking, strikes, effects, damage numbers. During the enemy turn the camera waits for each animation, and there's a Skip button.
- **Information on the map:**
  - red attack tiles
  - a monster's purple move zone and red attack zone
  - the danger zone
  - the monster info window and the terrain box
- **Menus:**
  - the ☰ menu
  - the gear sheet (item pictures, sorting, comparison with worn gear, CAPPED)
  - shop, floor list
  - three save slots plus an autosave, with Continue on the title screen
  - settings
- **Views:** flat and isometric, with rotate.
- Checked with scripted taps in a phone-sized browser after each step.

## New referee answers and fixes (each with a rule test)
- **New answers the screen asks for:**
  - threat zones and red attack tiles
  - targets and forecasts for a chosen weapon set
  - per-turn status amounts (one shared formula)
  - the hero's form stage, triangle weapon and "acted" flag
  - a rules guide (terrain, element wheel, weapon triangle)
  - which shop items can combine
  - how each bag item suits a hero, and whether a hero is capped
  - picture info for items
- **Fixes:**
  - gear changed before a battle no longer uses up the hero's turn
  - skill aiming now says when an aim is complete

## Known problems and next session
1. **The attack animations are wrong.** The screen uses a simplified effect list. Fix them against the old Codex, which lists every animation, and against the old game's effect recipes.
2. **Update the Codex:**
   - upload the old one (v0.21/0.22) to the repo
   - check every entry against the new game
   - write the new Codex from the data, using the new sprites
   - never delete an entry without listing it as changed or removed
3. **Phase 8 polish:** zoom buttons, the title screen's candle and lightning effects, How to Play / What's New / Report, the damage-numbers option, first-time tips, the level-cap notice.
4. **Then:** the owner's new art, the balance autoplayer, Phase 9, and online versus last.
