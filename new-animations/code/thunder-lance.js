// Thunder Lance. Pictures only. The game does not load this file yet.
// Signed off 2026-10-17 (owner: "That's actually perfect yes sign off").
// One file for this attack. No version copies.
//
// The signed picture is ../gifs/dragon-knight/thunder-lance/in-game.gif
// (the v2 test). Storm Dragoon only.
//
// Made differently from the sheet-based moves: there is no OpenArt sheet.
// - The body is four poses cut from the owner's mockup (magenta keyed out):
//   ../gifs/dragon-knight/thunder-lance/poses/pose-{stand,charge,windup,thrust}.png.
//   He stands, raises the sparking lance (charge), lowers it (wind-up),
//   thrusts, lowers it again and ends standing. The last frames are the
//   standing body with the glow gone.
// - The effects are the game's own sheets from web/assets/effects/,
//   recolored to the skill (dark blue, cyan #7fe3f0, yellow #ffe45c, white
//   by lightness): lightning-burst-002 on every enemy hit, lightning-strike-001
//   where the bolt ends.
// - The bolt is drawn in code: a jagged line from the thrust pose's lance tip
//   (its rightmost solid pixel) to the last tile of the line, chest high.
//   Cyan glow (16 px, blurred), yellow (7 px), white core (3 px), small forks,
//   a new shape every frame. Thunder Lance hits the target and 3 tiles behind
//   it, so the bolt runs to the end of that line.
// - The hero glow is a soft blur around his outline, yellow and cyan in turn.
// - The rings are the look rule's black center fading out to yellow, on the
//   ground under each enemy hit.
// The make script and the exact drawing are in
// ../gifs/dragon-knight/thunder-lance/make-v2.py.
//
// The effect lands on the enemies and does not cover him. The damage numbers
// in the review GIF are only for the preview: the game shows its own.
// This file does not touch hit chance, damage, range or any other rule.
export const THUNDER_LANCE = {
  name: 'Thunder Lance',
  class: 'Dragon Knight',
  ascension: 'Storm Dragoon',
  fps: 20, // 50 ms a frame; the last frame is held 700 ms in the review GIF
  where: 'from his lance tip along the line, through the target and 3 tiles behind',
  approved: true,
  inGame: '../gifs/dragon-knight/thunder-lance/in-game.gif',
  colors: { dark: '#284678', cyan: '#7fe3f0', yellow: '#ffe45c', white: '#ffffff' },
  // lunge: pixels toward the target (half as much down, on the iso map)
  // bolt: line width (1 full); burst and strike: frame of those sheets
  // hit: tint on every enemy hit; ring: alpha of the black-center ring
  frames: [
    { body: 'stand' },
    { body: 'stand' },
    { body: 'stand' },
    { body: 'charge', glow: 0.12, glowColor: 'yellow', squash: true },
    { body: 'charge', glow: 0.25, glowColor: 'cyan', squash: true },
    { body: 'charge', glow: 0.38, glowColor: 'yellow', squash: true },
    { body: 'charge', glow: 0.5, glowColor: 'cyan', squash: true },
    { body: 'charge', glow: 0.62, glowColor: 'yellow', squash: true },
    { body: 'charge', glow: 0.75, glowColor: 'cyan', squash: true },
    { body: 'charge', glow: 0.88, glowColor: 'yellow', squash: true },
    { body: 'windup', glow: 1, glowColor: 'cyan' },
    { body: 'thrust', lunge: 6, glow: 1, glowColor: 'yellow', flashBody: true },
    { body: 'thrust', lunge: 12, glow: 1, glowColor: 'cyan', flashBody: true, bolt: 1, burst: 0, strike: 0, hit: 'white', screenShake: 2, foeShake: 3, screenFlash: 0.12 },
    { body: 'thrust', lunge: 12, glow: 1, glowColor: 'yellow', bolt: 1, burst: 1, strike: 1, hit: 'yellow', screenShake: 2, foeShake: 3, screenFlash: 0.05 },
    { body: 'thrust', lunge: 12, glow: 1, glowColor: 'cyan', bolt: 1, burst: 2, strike: 2, hit: 'white', screenShake: 2, foeShake: 3 },
    { body: 'thrust', lunge: 10, glow: 1, glowColor: 'yellow', bolt: 1, burst: 3, strike: 3, hit: 'yellow', screenShake: 2, foeShake: 3 },
    { body: 'thrust', lunge: 8, glow: 1, glowColor: 'cyan', bolt: 1, burst: 4, strike: 4, hit: 'cyan', foeShake: 3, ring: 1 },
    { body: 'thrust', lunge: 6, glow: 1, glowColor: 'yellow', bolt: 1, burst: 5, strike: 5, hit: 'cyan', foeShake: 3, ring: 1 },
    { body: 'thrust', lunge: 4, glow: 1, glowColor: 'cyan', bolt: 0.6, burst: 6, strike: 6, hit: 'cyan', ring: 1 },
    { body: 'thrust', lunge: 3, glow: 1, glowColor: 'yellow', bolt: 0.6, burst: 7, strike: 7, hit: 'cyan', ring: 1 },
    { body: 'thrust', lunge: 2, glow: 0.88, glowColor: 'cyan', hit: 'cyan', ring: 1 },
    { body: 'windup', lunge: 1, glow: 0.75, glowColor: 'yellow', hit: 'cyan', ring: 0.83 },
    { body: 'windup', glow: 0.62, glowColor: 'cyan', hit: 'cyan', ring: 0.67 },
    { body: 'windup', glow: 0.5, glowColor: 'yellow', ring: 0.5 },
    { body: 'stand', glow: 0.38, glowColor: 'cyan', ring: 0.33 },
    { body: 'stand', glow: 0.25, glowColor: 'yellow', ring: 0.17 },
    { body: 'stand', glow: 0.12, glowColor: 'cyan', ring: 0 },
    { body: 'stand' },
    { body: 'stand' },
    { body: 'stand' },
    { body: 'stand' },
    { body: 'stand' },
  ],
  files: {
    body: {
      stand: '../gifs/dragon-knight/thunder-lance/poses/pose-stand.png',
      charge: '../gifs/dragon-knight/thunder-lance/poses/pose-charge.png',
      windup: '../gifs/dragon-knight/thunder-lance/poses/pose-windup.png',
      thrust: '../gifs/dragon-knight/thunder-lance/poses/pose-thrust.png',
    },
    burst: 'web/assets/effects/lightning-burst-002.png', // 32 px cells
    strike: 'web/assets/effects/lightning-strike-001.png', // 64 px cells
  },
};
