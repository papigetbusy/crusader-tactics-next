# Attack code, held back on purpose

Nothing in this folder is wired into the game.

How every attack is made, in this order. Do not skip a step.

1. The owner sends a sprite sheet.
2. A test GIF is made from it. It is not put in this folder yet.
3. He watches it and says what to change. New tries stay tests until he says that one is the one.
4. Only then is one code file written here, named after the attack. The test GIFs do not get their own code.
5. The code only says which pictures play, how fast, and where the effect lands. It does not change hit chance, damage or any other rule.
6. An in-game GIF is rendered from that code: she uses the move on an enemy, on the battle map. The new pictures play, and so does the effect that skill already has in the codex (its beam, its ring, its effect picture). The effect is not allowed to cover or cut her sprite. Every effect from the signed test stays. Nothing is dropped unless the owner said to drop it.
7. That GIF and the code stay here. They are not loaded by the play page.
8. When every row of `../CHECKLIST.md` is signed off, they all go into the game together. Not one at a time.

## Look rules (owner, 2026-10-12)

Soul Nova's glow test is the pattern from here on. Blood Moon stays as it was signed.

- The body can be a small motion. Do not invent new poses to fake a bigger swing.
- The power is effects the game already has, recolored to that skill, plus the hero glow: a colored blur around the sprite, a pool of that color under the feet, and the little sparks a powered-up hero already gets.
- Recolor sheets from `web/assets/effects/`. Do not draw a new effect when one of those will do.
- The effect lands on the enemy. It does not cover or cut the attacker.
- The last frames are the idle body. The glow and the attack pictures are gone. She does not stay on the attack pose.

Blood Moon is the first signed file. `blood-moon.js` and `../gifs/witch/blood-moon/in-game.gif`. It was signed before the idle rule, so it still ends on the swing. It gets an idle ending before it goes into the game.
