// Soul Nova. Pictures only. The game does not load this file yet.
// Signed off 2026-10-12. One file for this attack. No version copies.
//
// The signed picture is ../gifs/witch/soul-nova/in-game.gif.
// Her cast stays small. The power is a purple glow, effects the game
// already has recolored purple, and the black ball from the first GIF.
// The circle opens centered on that ball, behind it. It is not off to the side.
// The last three frames are her idle body. The glow is gone.
// The codex entry for this skill is green. The signed look is purple.
export const GAME_EFFECT = {
  base: 'nova',
  colors: ['#b06bff', '#8a5fd0'],
  glow: 'purple blur, a pool under her feet, and sparks',
  charge: 'effects/scifi-charge-up-001.png',
  ball: 'the black ball from the first GIF, on the enemy',
  ring: 'opens centered on the ball, drawn behind it, then gone',
};

export const SOUL_NOVA = {
  name: 'Soul Nova',
  class: 'Witch',
  ascension: 'Necromancer',
  fps: 10,
  clearOfCaster: 8,
  approved: true,
  inGame: '../gifs/witch/soul-nova/in-game.gif',
  // body 1 is idle. Bodies 2 to 8 are the small cast.
  // The last three steps return to body 1 with the glow off.
  frames: [
    { body: 1, glow: 0.35 },
    { body: 2, glow: 0.65 },
    { body: 3, glow: 0.95, effect: 'wisp', at: 'hand' },
    { body: 4, glow: 1, effect: 'wisp', at: 'travel', t: 0.22 },
    { body: 5, glow: 1, effect: 'orb', at: 'travel', t: 0.45 },
    { body: 6, glow: 1, effect: 'orb', at: 'travel', t: 0.7 },
    { body: 7, glow: 0.9, effect: 'orb', at: 'enemy' },
    { body: 8, glow: 0.85, effect: 'orb', at: 'enemy' },
    { body: 8, glow: 0.8, effect: 'ball', at: 'enemy', ring: 34 },
    { body: 8, glow: 0.75, effect: 'ball', at: 'enemy', ring: 62 },
    { body: 8, glow: 0.7, effect: 'ball', at: 'enemy', ring: 96 },
    { body: 8, glow: 0.6, effect: 'ball', at: 'enemy', ring: 128 },
    { body: 1, glow: 0.15 },
    { body: 1, glow: 0 },
    { body: 1, glow: 0 },
  ],
};

// The ring is drawn at the enemy, centered on the ball, before the ball,
// so the ball sits in front of it.
export function placeEffect(enemyX) {
  return enemyX;
}
