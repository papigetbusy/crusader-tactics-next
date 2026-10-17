# Putting the online server on Render (for the owner, from an iPhone)

The game's pages stay on GitHub Pages. Online matches need a small server that runs the same referee. Render runs it for free. A free server sleeps after 15 minutes with nobody on it. The first person to join after that waits up to a minute while it wakes up; the game says so.

## Once: make the server (about 5 minutes)
1. Open **render.com** in Safari and tap **Get Started**. Sign up with **GitHub** (the same account that owns the game).
2. When Render asks to see your repositories, allow **papigetbusy/crusader-tactics-next**.
3. In Render, tap **New +** → **Blueprint**.
4. Pick **crusader-tactics-next**. Render reads `render.yaml` and shows one web service called **crusader-tactics** on the **Free** plan. Tap **Apply**.
5. Wait for the build to finish (about 5 minutes; it shows **Live** in green).
6. Render shows the server's address, like `https://crusader-tactics.onrender.com`. Send that address to Claude. If the name was taken and Render added letters (like `crusader-tactics-abcd.onrender.com`), Claude points the game at it.

## Every time after
Nothing. Each push to GitHub rebuilds the server by itself (`autoDeploy`).

## Playing with friends
- **Easiest link:** the Render address plus `/play.html`, for example `https://crusader-tactics.onrender.com/play.html`. Everyone opens it, taps **Versus** → **Online**, and types the same room password.
- The GitHub Pages link works too: its **Online** button talks to the Render server.
- **A room:** up to 4 people.
  - Two tap **Take Blue** and **Take Red**, choose their heroes, and the match starts for everyone.
  - The others watch and wait in line. The winner stays on; the loser goes to the back of the line.
- **Leaving:** Give up ends the match now and the other player wins. Closing the phone keeps your seat for 20 seconds, so a refresh is safe. Longer than that, the other player wins and you are out of the room. A watcher can leave without ending the match. The free server forgets rooms when it sleeps or restarts; the next open says the match is over.
