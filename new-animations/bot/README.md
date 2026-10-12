# Animation bot

This builds the first review GIF for a pile of sheets. Then it stops.

It follows the Grand Hex sign-off (2026-10-12):

1. Body frames come from the sheet, in the order they sit on the page.
2. The power is an effect the game already has, recolored to that skill.
3. On a spell, the warp ring sits around the hero. The inside is black and fades out to the skill color. The color is a little different on every frame.
4. A sword swing does not get that ring. It gets the same color shift on the hero, and the spell picture still lands on the enemy.
5. The spell starts off the hero's body. It does not cover them.
6. The circle on a hit is centered on the impact, behind the spell.
7. The last three frames are the first body again. The glow is gone.
8. It writes `gifs/<class>/<skill>/review.gif` and a note. It does not write `code/`. It does not check the row off. It does not change the play page.
9. A skill that is already signed off is skipped.
10. `last-run.md` ends with STOP. The next step is the owner and Grok, one skill at a time.

Labels on the sheet are not read. If the art put frame 6 in the wrong cell, the GIF will show it in that spot. That is a review note, not something the bot "fixes" by guessing.

Run it from the repo root: `python3 new-animations/bot/make.py`
