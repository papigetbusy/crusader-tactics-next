// The stack of effect pictures each skill plays on top of its drawn effect:
// two or three recolored sprite sheets picked from the skill's element,
// kind and colors, so no two skills look the same. Pictures only.
// Ported from the old game's recipes, the version that actually runs
// [old index.html L1331-1380, makeRecipe L1365; the "calm" option L1663].
// The later FX2 code in the old file (L1386-1453) was never drawn there.
import { SKFX, BIMP } from './fxbook.js';
import { pic, sheets } from './assets.js';
import { TS } from './look.js';

// FXP: the picture groups [L1336-1354].
const FXP = {
  melee: ['directional_impact_001', 'directional_impact_002', 'directional_impact_003', 'directional_impact_004', 'symmetrical_impact_001', 'scifi_spark_burst_001', 'symmetrical_impact_004'],
  heavy: ['symmetrical_impact_003', 'symmetrical_impact_004', 'symmetrical_impact_006', 'symmetrical_impact_002', 'symmetrical_explosion_001', 'symmetrical_explosion_002', 'stylized_explosion_001'],
  spin: ['symmetrical_impact_002', 'looping_smoke_burst_002', 'round_sparkle_burst_001', 'symmetrical_impact_001'],
  fire: ['fire_looping_004A', 'fire_looping_003B', 'fire_looping_004B', 'fire_looping_003A', 'stylized_explosion_001', 'symmetrical_explosion_004', 'ts_fireB', 'ts_fire3', 'ts_fire2'],
  ice: ['symmetrical_impact_002', 'round_sparkle_burst_001', 'scifi_warp_003', 'symmetrical_explosion_002', 'directional_impact_001', 'status_sparkling_001'],
  elec: ['lightning_burst_001', 'lightning_burst_002', 'lightning_burst_003', 'lightning_strike_001', 'scifi_spark_burst_001', 'scifi_charge_up_001'],
  holy: ['round_light_burst_001', 'status_sparkling_001', 'round_sparkle_burst_002', 'round_firework_burst_002', 'symmetrical_impact_006', 'scifi_charge_up_001'],
  dark: ['spell_death_001', 'stylized_explosion_002', 'spell_absorb_001', 'scifi_warp_002', 'stylized_skull_smoke_burst_001', 'looping_smoke_burst_001'],
  arcane: ['stylized_explosion_002', 'stylized_explosion_003', 'round_sparkle_burst_001', 'scifi_warp_003', 'round_firework_burst_001', 'spell_absorb_001'],
  poison: ['spell_poison_001', 'burst_splatter_003', 'directional_bubble_burst_001', 'round_firework_burst_001', 'scifi_warp_001', 'status_poison_001'],
  blood: ['burst_splatter_001', 'directional_splatter_001', 'directional_splatter_003', 'round_sparkle_burst_003'],
  big: ['epic_explosion_001', 'epic_explosion_002', 'ts_expl1', 'ts_expl2', 'ts_explo', 'stylized_explosion_003', 'symmetrical_explosion_004'],
  heal: ['spell_heal_001', 'round_light_burst_001', 'round_heart_burst_001', 'round_sparkle_burst_002', 'status_sparkling_001', 'round_firework_burst_002'],
  buff: ['spell_attack_up_001', 'spell_defense_up_001', 'spell_haste_001', 'scifi_charge_up_001', 'status_sparkling_001', 'round_firework_burst_002', 'round_light_burst_001'],
  debuff: ['spell_absorb_001', 'stylized_skull_smoke_burst_001', 'looping_smoke_burst_001', 'scifi_warp_002', 'spell_death_001', 'looping_smoke_burst_002'],
  water: ['ts_splash', 'symmetrical_impact_002', 'directional_bubble_burst_001'],
  dust: ['ts_dust1', 'ts_dust2', 'symmetrical_smoke_burst_001', 'directional_smoke_burst_001'],
  muzzle: ['scifi_muzzle_flash_001', 'scifi_muzzle_flash_003'],
};
const BASEG = { cut: 'melee', claw: 'melee', thrust: 'melee', blink: 'melee', knives: 'melee', exec: 'melee', charge: 'heavy', dive: 'heavy', smash: 'heavy', quake: 'heavy', swirl: 'spin', wave: 'spin', shot: 'melee', scatter: 'melee', ricochet: 'melee', gunrain: 'melee', arrow: 'melee', heal: 'heal', revive: 'heal', buff: 'buff', debuff: 'debuff', meteor: 'big', nova: 'big', pillar: 'holy', cloud: 'poison', swarm: 'poison', potion: 'poison', flashbang: 'holy' };
const ELN = { fire: 'Fire', ice: 'Ice', elec: 'Thunder', light: 'Light', dark: 'Dark' };

function fxHash(s) { let h = 2166136261; for (let i = 0; i < s.length; i++) h = Math.imul(h ^ s.charCodeAt(i), 16777619); return h >>> 0; }
function fxHsl(hex) { const m = /#?([0-9a-f]{6})/i.exec(hex || ''); if (!m) return [0, 0, 1]; const n = parseInt(m[1], 16), r = ((n >> 16) & 255) / 255, g9 = ((n >> 8) & 255) / 255, b = (n & 255) / 255, mx = Math.max(r, g9, b), mn = Math.min(r, g9, b), l = (mx + mn) / 2; let h = 0, s = 0; if (mx !== mn) { const d = mx - mn; s = l > 0.5 ? d / (2 - mx - mn) : d / (mx + mn); h = mx === r ? (g9 - b) / d + (g9 < b ? 6 : 0) : mx === g9 ? (b - r) / d + 2 : (r - g9) / d + 4; h *= 60; } return [h, s, l]; }
function fxShift(hex, deg) { const [h, s9, l] = fxHsl(hex); const H = (((h + deg) % 360) + 360) % 360, C = (1 - Math.abs(2 * l - 1)) * s9, X = C * (1 - Math.abs(((H / 60) % 2) - 1)), m = l - C / 2; let r = 0, g9 = 0, b = 0; if (H < 60) { r = C; g9 = X; } else if (H < 120) { r = X; g9 = C; } else if (H < 180) { g9 = C; b = X; } else if (H < 240) { g9 = X; b = C; } else if (H < 300) { r = X; b = C; } else { r = C; b = X; } const t = (v) => Math.round((v + m) * 255).toString(16).padStart(2, '0'); return '#' + t(r) + t(g9) + t(b); }

// sType is the old game's kind label for a hero skill [old L455].
const sType = (s) => (s.heal ? (s.cleanse ? 'Support' : 'Heal') : s.support ? 'Buff' : s.revive ? 'Revive' : s.element && s.element !== 'user' ? ELN[s.element] + (s.physical ? ' weapon art' : ' magic') : s.magic ? 'Arcane magic' : 'Physical');

let SK = {}, ABCLS = {}; // hero skills; ability -> class (from assets/skillbook.json)

// fxElem: which family of pictures a skill uses [L1358-1361].
function fxElem(n, pal, base) {
  const s = SK[n], t = s ? sType(s) : '';
  if (/Toxic|Venom|Poison|Plague|Miasma|Spore|Brew|Pandemic|Sludge|Acid/i.test(n)) return 'poison';
  if (/Hex|Curse|Shackle|Constrict|Smoke|Doom|Wither|Cloak/i.test(n)) return 'debuff';
  if (/Blood|Drain|Rend|Savage|Crimson/i.test(n)) return 'blood';
  if (/Tidal|Wave|Splash|Bubble/i.test(n) && !/Crescent/.test(n)) return 'water';
  if (t === 'Physical') return null;
  if (!s) {
    if (/Fire|Flame|Infernal|Hell|Burn|Ember|Magma|Lava/i.test(n)) return 'fire';
    if (/Frost|Ice|Glacier|Glacial|Blizzard|Freeze|Snow/i.test(n)) return 'ice';
    if (/Thunder|Lightning|Shock|Spark/i.test(n)) return 'elec';
    if (/Holy|Divine|Radiant|Sacred|Mend/i.test(n)) return n.includes('Mend') ? 'heal' : 'holy';
    if (/Shadow|Soul|Grave|Death|Doom|Bone|Cataclysm|Void|Dark/i.test(n)) return 'dark';
    if (/Greater Spell|Arcane/i.test(n)) return 'arcane';
    if (['smash', 'quake', 'cut', 'claw', 'thrust', 'charge', 'dive', 'swirl', 'wave'].includes(base)) return null;
  }
  if (/Fire/.test(t)) return 'fire'; if (/Ice/.test(t)) return 'ice'; if (/Thunder/.test(t)) return 'elec'; if (/Heal|Revive/.test(t)) return 'heal';
  if (/Light/.test(t)) return 'holy'; if (/Dark/.test(t)) return 'dark'; if (/Arcane/.test(t)) return 'arcane'; if (/Buff|Support/.test(t)) return 'buff';
  const [h, sa, l] = fxHsl((pal || [])[1] || (pal || [])[0]);
  if (sa < 0.2) return null; if (h < 14 || h > 345) return l < 0.5 ? 'blood' : 'fire'; if (h < 48) return 'fire'; if (h < 70) return 'holy'; if (h < 165) return 'poison'; if (h < 215) return 'ice'; if (h < 265) return 'elec'; return 'dark';
}

const RCP = {}, RUSED = new Set();
// makeRecipe picks a skill's stack of pictures [L1365-1376].
function makeRecipe(n) {
  const isAb = !SKFX[n] && ABCLS[n];
  const sf = SKFX[n] || (isAb ? SKFX['AB_' + ABCLS[n]] : null) || ['buff', '#ffd84a,#ffffff', ''], base = sf[0], pal = (sf[1] || '#ffffff,#ffffff').split(','), ex = sf[2] || '', s = SK[n], h = fxHash(n);
  const el = fxElem(n, pal, base), bg = BASEG[base], ult = (s && s.kind === 'u') || /meteor|nova/.test(base) || (s && s.power >= 1.8), phys = !el || bg === 'melee' || bg === 'heavy' || bg === 'spin';
  const pp = FXP[bg === 'heal' || bg === 'buff' || bg === 'debuff' ? bg : bg === 'big' || (ult && !phys) ? 'big' : phys && !['fire', 'ice', 'elec', 'holy', 'dark', 'poison'].includes(el || '') ? bg || 'melee' : el && FXP[el] ? el : bg || 'melee'] || FXP.melee;
  let sp = el && FXP[el] && pp !== FXP[el] ? FXP[el] : bg === 'heavy' || bg === 'spin' ? FXP.dust : ex.includes('b') ? FXP.blood : FXP.melee;
  if (el === 'blood') sp = FXP.blood;
  let tint = pal[1] && !/^#?f{6}$/i.test(pal[1]) ? pal[1] : pal[0] && !/^#?f{6}$/i.test(pal[0]) ? pal[0] : null;
  if (isAb && tint) tint = fxShift(tint, ((h % 7) - 3) * 22);
  const tint2 = pal[0] && pal[0] !== tint && !/^#?f{6}$/i.test(pal[0]) ? pal[0] : null;
  const mk = (k) => {
    const a = pp[(h + k) % pp.length];
    let b2 = sp[((h >> 5) + k * 3) % sp.length];
    if (b2 === a) b2 = sp[((h >> 5) + k * 3 + 1) % sp.length];
    const L = [{ e: a, col: /^(fire_|ts_fire|epic_|ts_expl|ts_explo)/.test(a) && el === 'fire' ? null : tint, sc: ult ? 1.75 : 1.25, at: 't', dl: 0 }];
    L.push({ e: b2, col: /^(ts_)/.test(b2) && el !== 'fire' ? tint2 : tint2 || tint, sc: ult ? 1.1 : 0.8, at: 't', dl: 0.12, ox: (((h >> 9) % 3) - 1) * 0.18 });
    if (ult) L.push({ e: FXP.holy.concat(FXP.big)[((h >> 3) + k) % 13], col: tint, sc: 0.6, at: 't', dl: 0.25, ring: 3 });
    if (/shot|scatter|ricochet|gunrain/.test(base) || n === 'Strike_gun') L.unshift({ e: FXP.muzzle[h % 2], col: null, sc: 0.7, at: 'c', dl: 0 });
    return L;
  };
  let k = 0, L, sig;
  do { L = mk(k); sig = L.map((q) => q.e + ':' + (q.col || '') + (q.ring ? 'r' : '')).join('+'); k++; } while (RUSED.has(sig) && k < pp.length * sp.length * 2);
  RUSED.add(sig);
  if (bg === 'buff' || bg === 'heal') L.forEach((q) => (q.up = 0.25));
  return { L, el, base };
}

// setup reads the skill book and picks every stack once, in name order. The
// old game picked them in the order skills were first used, so a skill could
// look different from one session to the next; here each always looks the
// same (listed in docs/changes-from-old-game.md).
export function setup(book) {
  SK = {};
  for (const [n, s] of Object.entries(book.skills || {})) if (s.hero) SK[n] = s;
  ABCLS = book.abilities || {};
  const names = new Set([...Object.keys(SKFX), ...Object.keys(book.skills || {}), ...Object.keys(ABCLS)]);
  for (const n of [...names].sort()) if (!RCP[n]) RCP[n] = makeRecipe(n);
}
export const recipe = (n) => RCP[n] || (RCP[n] = makeRecipe(n));

// Recolored copies of the sheets [fxTint L1333].
const TINT = {};
function fxTint(e, col) {
  const im = pic('effects', e);
  if (!col || !im || !im.naturalWidth) return im;
  const k = e + col;
  if (TINT[k]) return TINT[k];
  const c = document.createElement('canvas');
  c.width = im.naturalWidth; c.height = im.naturalHeight;
  const x = c.getContext('2d');
  x.drawImage(im, 0, 0);
  x.globalCompositeOperation = 'color'; x.fillStyle = col; x.fillRect(0, 0, c.width, c.height);
  x.globalCompositeOperation = 'destination-in'; x.drawImage(im, 0, 0);
  return (TINT[k] = c);
}

// efxDraw draws frame i of sheet e at (x, y), sz wide [L1334].
export function efxDraw(g, e, i, x, y, sz, col, rot, al) {
  const E = sheets.effects[e], im = fxTint(e, col);
  if (!E || !im) return;
  const n = E.fr.length, fr = E.fr[Math.max(0, Math.min(n - 1, i))], h = (sz * E.h) / E.w;
  g.save();
  g.imageSmoothingEnabled = false;
  if (al != null) g.globalAlpha *= al;
  g.translate(x, y);
  if (rot) g.rotate(rot);
  g.drawImage(im, fr[0], fr[1], E.w, E.h, -sz / 2, -h / 2, sz, h);
  g.restore();
}

// drawLayers draws skill n's stack for effect f at p [L1377-1380]. With the
// "fewer flashes" option, only the last picture, half see-through [L1663].
export function drawLayers(g, f, p, n, b, calm) {
  const R = recipe(n);
  if (!R) return;
  let L = R.L;
  if (calm) { L = L.filter((q) => !q.ring).slice(-1); g.save(); g.globalAlpha *= 0.55; }
  const im = (BIMP[b] || 250) / f.d, S = f.S || 1, dir = Math.sign(f.tx - f.sx) || 1, ang = Math.atan2(f.ty - f.sy, f.tx - f.sx);
  for (const q of L) {
    const E = sheets.effects[q.e];
    if (!E) continue;
    const st = q.at === 'c' ? 0 : Math.min(0.85, im + q.dl), en = q.at === 'c' ? Math.min(1, im + 0.3) : 1, t = (p - st) / (en - st);
    if (t < 0 || t >= 1) continue;
    const sz = TS * q.sc * S, fi = (t * E.fr.length) | 0;
    const x0 = q.at === 'c' ? f.sx + dir * 16 * S : f.tx + (q.ox || 0) * TS, y0 = (q.at === 'c' ? f.sy : f.ty) - (q.up || 0) * TS - (/lightning_strike/.test(q.e) ? sz * 0.25 : 0);
    if (q.ring) for (let r = 0; r < q.ring; r++) { const a2 = (r / q.ring) * Math.PI * 2 + p * 2; efxDraw(g, q.e, fi, x0 + Math.cos(a2) * TS * 0.42 * S, y0 + Math.sin(a2) * TS * 0.25 * S, sz, q.col, 0, 0.9); }
    else efxDraw(g, q.e, fi, x0, y0, sz, q.col, /^directional_/.test(q.e) ? ang : 0);
  }
  if (calm) g.restore();
}
