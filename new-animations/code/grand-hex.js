// Grand Hex. Pictures only. The game does not load this file yet.
// Signed off 2026-10-12. One file for this attack. No version copies.
//
// The signed picture is ../gifs/witch/grand-hex/in-game.gif.
// Hex Queen only. Body 6 dropped the point, and body 8 lost the staff,
// so those two pictures are not used. The point is held on body 5,
// and she returns to body 1.
//
// The aura is the game's warp ring (effects/scifi-warp-001.png,
// frames 5, 6, 7, 8). The inside is black and fades out to green.
// The green is a different shade on every frame, so it flickers.
// It is drawn around her, behind her. It does not cover her face.
// The seal grows off her staff and lands on the enemy. The circle
// opens on that seal, centered, behind it.
// The last three frames are body 1. The glow and the ring are gone.
//
// The codex draws this skill as a light-blue debuff. The signed look
// is the black-to-green ring. The curse itself (weak, armor break,
// snare, everything within 2 tiles) is a rule. This file does not touch it.
export const GAME_EFFECT = {
  base: 'debuff',
  colors: ['#12781e', '#baff46', '#1c9630', '#e6ff96', '#0c5a1e', '#8cff3c'],
  aura: 'effects/scifi-warp-001.png',
  auraFrames: [5, 6, 7, 8, 7, 6],
  inner: 'black, fading out into that frame\'s green',
  seal: 'the green star seal, grows, then lands on the enemy',
  ring: 'opens centered on the seal, drawn behind it, then gone',
};

export const GRAND_HEX = {
  name: 'Grand Hex',
  class: 'Witch',
  ascension: 'Hex Queen',
  fps: 10,
  clearOfCaster: 8,
  approved: true,
  inGame: '../gifs/witch/grand-hex/in-game.gif',
  // glow is how strong the black-to-green ring is. 0 means it is gone.
  // colors index picks which green that frame uses.
  frames: [
    { body: 1, glow: 0.55, color: 0 },
    { body: 2, glow: 0.9, color: 1 },
    { body: 3, glow: 1, color: 2 },
    { body: 4, glow: 1, color: 3, effect: 'seal', seal: 3, at: 'travel', t: 0.55 },
    { body: 5, glow: 1, color: 4, effect: 'seal', seal: 4, at: 'enemy' },
    { body: 5, glow: 0.9, color: 5, effect: 'seal', seal: 7, at: 'enemy', ring: 52 },
    { body: 5, glow: 0.7, color: 0, effect: 'seal', seal: 7, at: 'enemy', ring: 96 },
    { body: 7, glow: 0.4, color: 1, effect: 'seal', seal: 6, at: 'enemy', ring: 124 },
    { body: 1, glow: 0 },
    { body: 1, glow: 0 },
    { body: 1, glow: 0 },
  ],
  files: {
    body: [1, 2, 3, 4, 5, 7].map((n) => `../gifs/witch/grand-hex/frames/body-${n}.png`),
    seal: [3, 4, 6, 7].map((n) => `../gifs/witch/grand-hex/frames/seal-${n}.png`),
  },
};

// The seal and its circle are drawn on the enemy. The aura stays on her,
// behind her, and is not this function.
export function placeEffect(enemyX) {
  return enemyX;
}
