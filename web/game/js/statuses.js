// How each status looks on a unit's badge: bad (red) or good (blue), its
// glow color and its label. The old game's labels [old STD, L971-973 and
// later additions]; the new skills' statuses got labels in the same style.
// Amounts (HP lost or healed each turn) come from the referee (u.ticks).

const S = (bad, hue, label) => ({ bad, hue, label });
export const STATUS = {
  bleed: S(1, '#ff4d4d', (u) => `🩸−${tick(u, 'bleed')}/t`),
  poison: S(1, '#6fe04a', (u) => `☠−${tick(u, 'poison')}/t`),
  burn: S(1, '#ff7a2e', (u) => `🔥−${tick(u, 'burn')}/t`),
  stun: S(1, '#ffe45c', '💫STUN'),
  shock: S(1, '#ffe45c', '⚡SHOCK'),
  snare: S(1, '#7fd1ff', 'ROOTED'),
  lock: S(1, '#c9ced8', 'IMMOBILE'),
  brk: S(1, '#ff7a5a', '−30% DEF'),
  weak: S(1, '#c79bff', '−30% DMG'),
  ivy: S(1, '#6fe04a', '−30% ATK/DEF'),
  frozen: S(1, '#bfeaff', '❄ FROZEN'),
  dmark: S(1, '#ff4d6a', 'MARKED +50%'),
  tmark: S(1, '#9ae6b4', 'TRAPPED +30%'),
  grd: S(0, '#7fc4ff', '−40% DMG TAKEN'),
  rally: S(0, '#7fc4ff', '+25% DMG'),
  haste: S(0, '#7fc4ff', '+2 MOV'),
  prism: S(0, '#ff9bd8', 'ALL HITS ×1.5'),
  regen: S(0, '#6ee7a0', (u) => `+${tick(u, 'regen')} HP/t`),
  dbl: S(0, '#7fc4ff', 'NEXT HIT ×2'),
  imh: S(0, '#7fc4ff', '+DEF NEXT HIT'),
  ironm: S(0, '#7fc4ff', '+25% DEF'),
  inner: S(0, '#c79bff', '+30% · SURE HIT'),
  enl: S(0, '#6ee7a0', '+30% HEAL ALL'),
  heye: S(0, '#7fc4ff', '+3 RNG · MULTI'),
  blood: S(0, '#ff4d6a', 'CLOAK +3MOV 50%CRIT'),
  cloak: S(0, '#9a9aaa', 'CLOAKED'),
  qd: S(0, '#ffe45c', 'QUICK DRAW'),
  beast: S(0, '#ff7a2e', '+7 MOV +FIRE'),
  ivyA: S(0, '#6fe04a', 'POISON IVY'),
  vow: S(0, '#ffd84a', '+30% ATK +5 MOV'),
  flurry: S(0, '#7fc4ff', 'RIPOSTE ×2'),
  frenzy: S(0, '#ff4d4d', '+50% DMG · LEECH'),
  taunt: S(0, '#ffd84a', 'TAUNT · −30% DMG'),
  unb: S(0, '#9a9aaa', 'UNBROKEN'),
  bulwark: S(0, '#bfeaff', 'FROST BULWARK'),
  fstorm: S(0, '#ff7a2e', '+20% FIRE'),
  choir: S(0, '#fff6c8', 'DIVINE CHOIR'),
  pbloom: S(0, '#6fe04a', 'PLAGUE BLOOM'),
  deadeye: S(0, '#7fd1ff', 'DEADEYE'),
  scales: S(0, '#bfeaff', 'GLACIAL SCALES'),
  sky: S(0, '#ffe45c', 'IN THE SKY'),
  gbloom: S(0, '#9ae6b4', 'GRAVE BLOOM'),
  vamp: S(0, '#ff2a3a', 'BLOOD PACT'),
  svow: S(0, '#fff6c8', 'SACRED VOW'),
  ruin: S(0, '#5a1f1f', 'OATH OF RUIN'),
  clone: S(0, '#9a9aaa', 'CLONE'),
  // statuses of the new skills (2026-10-10)
  riposte: S(0, '#7fc4ff', 'RIPOSTE'),
  trance: S(0, '#ff7a5a', '+25% DMG +1 MOV'),
  dscale: S(0, '#ff9a5a', 'DRAGON SCALES'),
  lwall: S(0, '#7fc4ff', 'LANCE WALL'),
  wyrm: S(0, '#ff9a5a', 'WYRMBOND'),
  lastlight: S(0, '#fff6c8', 'LAST LIGHT'),
  aim: S(0, '#7fd1ff', 'STEADY AIM'),
  coin: S(0, '#ffd84a', 'LUCKY COIN'),
  cover: S(0, '#7fc4ff', 'COVER'),
  guarded: S(0, '#7fc4ff', 'GUARDED'),
  hmark: S(1, '#ff4d6a', "HUNTER'S MARK"),
  silence: S(1, '#c79bff', 'SILENCED'),
  slow: S(1, '#7fd1ff', '−MOV'),
  wither: S(1, '#8a5fd0', 'WITHER'),
  link: S(1, '#c79bff', 'VOODOO LINK'),
  taunted: S(1, '#ffd84a', 'TAUNTED'),
};
const tick = (u, k) => (u.ticks && u.ticks[k]) || 0;

// active lists a unit's statuses that have a badge.
export const active = (u) => Object.keys(u.statuses || {}).filter((k) => u.statuses[k] > 0 && STATUS[k]);

export const label = (u, k) => (typeof STATUS[k].label === 'function' ? STATUS[k].label(u) : STATUS[k].label);

// hue is the glow a unit gets from its statuses [old stHue].
export function hue(u) {
  const a = active(u);
  if (!a.length) return null;
  for (const k of ['vamp', 'frozen', 'shock', 'poison', 'stun', 'snare', 'bleed', 'weak', 'brk', 'lock']) if (a.includes(k)) return STATUS[k].hue;
  return a.some((k) => STATUS[k].bad) ? '#ff4d4d' : '#4fa8ff';
}
