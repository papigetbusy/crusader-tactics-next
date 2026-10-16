# Phase 10: the online-ready check (Claude, 2026-10-16)

The question: can the same game screen later play against a server instead of the referee inside the phone? The answer is yes, with the setup work listed below. Nothing here changes how single-player plays.

## Already ready (and now guarded by tests)
- **One doorway.** The screen talks to the referee only through `web/game/js/referee.js`. One side door was found (the lag tool built test floors directly) and moved into the doorway. A house-rule test now fails if any screen file calls the referee directly: "The game screen reaches the referee only through referee.js".
- **A game replays exactly.** New test: "A climb replayed from its seeds and its list of actions comes out identical". A bot plays three climbs to floor 3 (over 400 actions each), and replaying the same actions gives the same events and the same save, byte for byte. A server can rebuild or check any match from its seed and its actions.
- **The referee has no clock, network, files or unseeded dice** (house rules, already tested).
- **Every answer is plain data** (the View and the events), so it can travel over the internet as it is.
- **Saves don't carry the dice.** A loaded save gets fresh dice, so a copy of the game on a player's phone cannot be used to see the next roll.

## What has to be built for online (in order)
1. **Two player sides in the referee.** Today the second side is always the monsters' AI. Versus needs side 2 to wait for a player's actions. Tested first as "pass the phone" on one device. This is rules work, so it is Claude's.
2. **Clocks in the referee.** The referee can't read a clock, so the server writes the seconds used on each action it passes on, and the referee keeps the turn and match clocks and decides when time runs out. They are shown in the View.
3. **Waiting for answers.** Today every answer comes back instantly. Over the internet it takes a moment. The plan:
   - Real actions (move, attack, end turn) go to the server, which is the only judge and rolls the dice.
   - After each action the server sends back the events and the game (as save text). The phone loads it into its own copy of the referee, used only for "what if" questions (where can I walk, the attack window, the danger zone). Those stay instant and need no internet round trip.
   - Only `referee.js` changes, plus the few places that send real actions waiting for the reply.
4. **The server.** A small Go program that runs the same `sim/` referee, with room codes. GitHub Pages can't run it, so it needs a small host.
5. **"Play online" on the screen.** Make a room, join with a code, then the match.

## What the owner needs to set up (when we get to step 4)
- **An account on a small host** for the server, such as Render or Fly.io. We'll check current prices together then. A free plan that sleeps when nobody plays is fine for testing with friends.
- **Decisions still open** (from `next-steps.md`): seconds per turn and per player, the map (fixed or random, and its size), which purple set each class gets, and what happens when a clock runs out.

Nothing else is needed from the owner. No extra accounts or keys are needed for steps 1 to 3.
