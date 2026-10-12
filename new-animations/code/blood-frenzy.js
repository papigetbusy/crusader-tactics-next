// Blood Frenzy. Pictures only. The game does not load this file yet.
// Signed off 2026-10-12. One file for this attack. No version copies.
//
// The signed picture is ../gifs/warrior/blood-frenzy/in-game.gif.
// Berserker only. This is the lock for later sheets.
//
// The sheet is eight equal cells on #FF00FF, two rows.
// Row 1 is the body. Row 2 is the red aura. No hero in the aura row.
// Bodies play 1 through 8. Body 5 and body 6 are the same yell, on purpose.
// The last three frames are body 1. The glow and the aura are gone.
// He does not stay on the yell.
//
// The aura stays on him, at his feet, drawn behind him.
// It does not fly to the enemy. It does not cover his face.
// A hit skill still travels. This one does not, because it powers him up.
//
// The class ability spends HP and raises his damage. This file does not
// touch that rule.
export const BLOOD_FRENZY = {
  name: 'Blood Frenzy',
  class: 'Warrior',
  ascension: 'Berserker',
  fps: 10,
  where: 'on the caster',
  approved: true,
  inGame: '../gifs/warrior/blood-frenzy/in-game.gif',
  frames: [
    { body: 1, effect: 1, glow: 0.55 },
    { body: 2, effect: 2, glow: 0.7 },
    { body: 3, effect: 3, glow: 0.85 },
    { body: 4, effect: 4, glow: 1 },
    { body: 5, effect: 5, glow: 1 },
    { body: 6, effect: 6, glow: 1 },
    { body: 7, effect: 7, glow: 0.6 },
    { body: 8, effect: 8, glow: 0.25 },
    { body: 1, glow: 0 },
    { body: 1, glow: 0 },
    { body: 1, glow: 0 },
  ],
  files: {
    body: [1, 2, 3, 4, 5, 6, 7, 8].map((n) => `../gifs/warrior/blood-frenzy/frames/body-${n}.png`),
    effect: [1, 2, 3, 4, 5, 6, 7, 8].map((n) => `../gifs/warrior/blood-frenzy/frames/effect-${n}.png`),
  },
};
