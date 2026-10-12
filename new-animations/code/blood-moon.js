// Blood Moon. Pictures only. The game does not load this file yet.
// Signed off 2026-10-12. One file for this attack. No version copies.
//
// The body frames are the eight poses from the signed sheet. The small red
// spark is already on body frames 3 and 4, so it is not drawn again.
// The moon, the pillar and the blast are separate pictures, drawn on the
// enemy. clearOfCaster keeps the effect off her: if it would cover her, it
// is moved toward the enemy until her sprite is whole.

export const BLOOD_MOON = {
  name: 'Blood Moon',
  class: 'Witch',
  ascension: 'Blood Witch',
  fps: 10,
  clearOfCaster: 8,
  frames: [
    { body: 1 },
    { body: 2 },
    { body: 3 },
    { body: 4 },
    { body: 5, effect: 'moon' },
    { body: 6, effect: 'moon' },
    { body: 7, effect: 'pillar' },
    { body: 7, effect: 'blast' },
    { body: 8, effect: 'blast' },
    { body: 8, effect: 'blast' },
  ],
  files: {
    body: [1, 2, 3, 4, 5, 6, 7, 8].map((n) => `../gifs/witch/blood-moon/frames/body-${n}.png`),
    moon: '../gifs/witch/blood-moon/frames/moon.png',
    pillar: '../gifs/witch/blood-moon/frames/pillar.png',
    blast: '../gifs/witch/blood-moon/frames/blast.png',
  },
};

// placeEffect returns the effect's center x. enemyX is where the hit lands.
// casterRight is the right edge of her current pose. The effect never
// crosses that edge.
export function placeEffect(casterRight, effectWidth, enemyX) {
  let x = enemyX;
  const left = x - effectWidth / 2;
  const limit = casterRight + BLOOD_MOON.clearOfCaster;
  if (left < limit) x += limit - left;
  return x;
}
