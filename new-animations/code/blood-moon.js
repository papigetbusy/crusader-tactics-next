// Blood Moon. Pictures only. The game does not load this file yet.
// Signed off 2026-10-12. One file for this attack. No version copies.
//
// Keep every effect from the signed test. The red moon circle stays. It is
// drawn large, in the air, on the same spot the blast's moon uses. It is not
// shrunk onto the enemy. The pillar and the blast still land on him.
// An effect is removed only when the owner says to remove it.
//
// The small red spark is already painted on body frames 3 and 4.
// The effect the game already draws for this skill, from the codex.
// It plays on the enemy, with the new pictures. It is not a replacement.
// Blood Moon's entry is a pillar: a red beam, the ts_expl1 burst tinted
// red inside that beam, and a red ring that grows under his feet and fades.
export const GAME_EFFECT = {
  base: 'pillar',
  colors: ['#c8102e', '#ff4d4d'],
  sheet: 'effects/ts-expl1.png',
  tint: '#ff4d4d',
  ring: 'under the enemy, grows, then fades',
  beam: 'red light on the enemy',
};

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
    { body: 5, effect: 'moon', at: 'sky' },
    { body: 6, effect: 'moon', at: 'sky' },
    { body: 7, effect: 'pillar', at: 'enemy' },
    { body: 7, effect: 'blast', at: 'enemy' },
    { body: 8, effect: 'blast', at: 'enemy' },
    { body: 8, effect: 'blast', at: 'enemy' },
  ],
  files: {
    body: [1, 2, 3, 4, 5, 6, 7, 8].map((n) => `../gifs/witch/blood-moon/frames/body-${n}.png`),
    moon: '../gifs/witch/blood-moon/frames/moon.png',
    pillar: '../gifs/witch/blood-moon/frames/pillar.png',
    blast: '../gifs/witch/blood-moon/frames/blast.png',
  },
};

// placeEffect returns where the effect is drawn.
// 'sky' is the red moon circle, already in the air on the blast's moon.
// It is not shoved aside. 'enemy' is the pillar and the blast, on him.
// If an effect's pixels would cover her, move it toward the enemy. Do not
// move it just because it is near her.
export function placeEffect(enemyX, skyX, at) {
  return at === 'sky' ? skyX : enemyX;
}
