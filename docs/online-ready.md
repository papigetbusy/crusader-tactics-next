# Phase 10: the online-ready check (Claude, 2026-10-16)

The question: can the same game screen later play against a server instead of the referee inside the phone? The answer is yes, with the setup work listed below. Nothing here changes how single-player plays.

## Already ready (and now guarded by tests)
- **One doorway.** The screen talks to the referee only through `web/game/js/referee.js`. One side door was found (the lag tool built test floors directly) and moved into the doorway. A house-rule test now fails if any screen file calls the referee directly: "The game screen reaches the referee only through referee.js".
- **A game replays exactly.** New test: "A climb replayed from its seeds and its list of actions comes out identical". A bot plays three climbs to floor 3 (over 400 actions each), and replaying the same actions gives the same events and the same save, byte for byte. A server can rebuild or check any match from its seed and its actions.
- **The referee has no clock, network, files or unseeded dice** (house rules, already tested).
- **Every answer is plain data** (the View and the events), so it can travel over the internet as it is.
- **Saves don't carry the dice.** A loaded save gets fresh dice, so a copy of the game on a player's phone cannot be used to see the next roll.

## What has to be built for online (in order)
1. **Two player sides in the referee.** Done 2026-10-16. Pass the phone is on the title screen.
2. **Clocks in the referee.** Done. The server writes the time on each action. Each player has 5 minutes and there is no per-turn limit (owner, 2026-10-16). The first 2 seconds after a move are not charged. Running out of the 5 minutes loses.
3. **Waiting for answers.** Done. Real moves go to the server. The phone's own copy only answers "where can I walk" and the attack window.
4. **The server.** Done in `adapters/server`. It is not on Render until the owner follows `online-setup.md`.
5. **"Play online" on the screen.** Done. Versus → Online. A refresh keeps the same seat. A missed move snaps to the server's board.

## What the owner needs to set up (step 4, still open)
- **An account on Render**, following `docs/online-setup.md`. A free plan that sleeps when nobody plays is fine for testing with friends. The first person to join after a sleep waits up to a minute.
- Rooms live in the server's memory. A restart wipes a match in progress. That is accepted for these first games.

Settled by the owner (2026-10-16): 5 minutes each, no per-turn limit, a mirrored 32×32 map with fog of war, one fixed purple set per class, and running out of the 5 minutes loses.
