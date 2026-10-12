# Attack code, held back on purpose

Nothing in this folder is wired into the game.

How every attack is made, in this order. Do not skip a step.

1. The owner sends a sprite sheet.
2. A test GIF is made from it. It is not put in this folder yet.
3. He watches it and says what to change. New tries stay tests until he says that one is the one.
4. Only then is one code file written here, named after the attack. The test GIFs do not get their own code.
5. The code only says which pictures play, how fast, and where the effect lands. It does not change hit chance, damage or any other rule.
6. An in-game GIF is rendered from that code: she uses the move on an enemy, on the battle map. The effect plays with the sprite. The effect is not allowed to cover or cut her sprite. If it would, it is moved onto the enemy. Every effect from the signed test stays. Nothing is dropped unless the owner said to drop it.
7. That GIF and the code stay here. They are not loaded by the play page.
8. When every row of `../CHECKLIST.md` is signed off, they all go into the game together. Not one at a time.

Blood Moon is the pattern. `blood-moon.js` and `../gifs/witch/blood-moon/in-game.gif`.
