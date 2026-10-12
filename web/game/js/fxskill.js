// Each skill's own drawn effect, from its entry in fxbook.js (base, colors,
// extras, count), with its stack of effect pictures on top. Pictures only.
// Ported from the old game's drawP [old index.html L922-966].
import { SKFX, BIMP } from './fxbook.js';
import { tools as T } from './fxpixel.js';
import { drawLayers } from './fxsprite.js';

let g = null;
const { pr, flash, streak, burst, star, sparkle, ring, ZZ, arrowP, tracer, muzzle, rune, grad, NEW } = T;

// drawP draws effect f (k = 'P:' + name) at p (0..1). f.d is its length in
// ms; t is the clock for flicker.
export function drawP(g2, f, p, t, calm) {
  if (p < 0 || p > 1) return;
  const D = SKFX[f.k.slice(2)];
  if (!D) return;
  g = g2;
  T.setG(g2);
  const b = D[0], pal = D[1].split(','), ex = D[2] || '', n = D[3] || 1, S = f.S || 1, { sx, sy, tx, ty } = f, c0 = pal[0], c1 = pal[1] || '#fff', c2 = pal[2] || c1, im = BIMP[b] / f.d;
  g.save();
  g.globalAlpha = p > 0.8 ? (1 - p) * 5 : 1;
  const dir = Math.sign(tx - sx) || 1, hx = sx + dir * 14 * S, hy = sy - 2;
  // The old game ignored the number given to after(); it always measures from
  // the landing moment. Kept as it was, since that is how the effects look.
  const after = () => Math.max(0, (p - im) / (1 - im));
  if (ex.includes('f') && p > im && p < im + 0.12) flash(0.35);
  switch (b) {
    case 'cut': for (let i = 0; i < n; i++) { const q = Math.max(0, p - i * 0.08); streak(tx + ((i * 13) % 17) - 8, ty + ((i * 7) % 11) - 5, -0.8 + (ex.includes('x') && i % 2 ? 1.6 : 0) + i * 0.35, 80 * S, q, 6 * S, c0, c1); } if (ex.includes('x') && n === 1) streak(tx, ty, 0.8, 80 * S, Math.max(0, p - 0.15), 6 * S, c0, c1); break;
    case 'thrust': { const q = Math.min(1, p / 0.45), x = sx + (tx - sx) * q, y = sy + (ty - sy) * q; for (let i = 0; i < 7; i++) pr(x - (tx - sx) * 0.03 * i - 2 * S, y - (ty - sy) * 0.03 * i - 2 * S, 4 * S, 4 * S, i < 2 ? '#fff' : c0); if (p > 0.4) star(tx, ty, after(0.4), c1, c0); break; }
    case 'smash': if (p < 0.3) { const y = ty - 90 * S + 80 * S * (p / 0.3); pr(tx - 12 * S, y, 24 * S, 14 * S, c0); pr(tx - 3 * S, y - 26 * S, 6 * S, 26 * S, '#8a5a2b'); } else { const q = (p - 0.3) / 0.7; star(tx, ty, q, '#fff', c1); ring(tx, ty + 20, q * 50 * S, c0, 4); for (let i = 0; i < 8; i++) pr(tx + Math.cos(i * 0.8) * q * 48 * S - 2, ty + Math.sin(i * 0.8) * q * 28 * S - q * 20, 5, 5, c0); } break;
    case 'claw': for (let i = 0; i < Math.max(3, n); i++) streak(tx - 12 + i * 8, ty - 6 + i * 3, -1.1 + (i % 2) * 0.3, 72 * S, Math.max(0, p - i * 0.06), 5 * S, '#fff', c0); break;
    case 'swirl': for (let k = 0; k < 2; k++) for (let i = 0; i < 14; i++) { const a = p * 12 + i * 0.45 + k * 3, r = (22 + k * 12) * S; pr(tx + Math.cos(a) * r - 3, ty + Math.sin(a) * r * 0.5 - 3, 6, 6, k ? c1 : c0); } break;
    case 'wave': case 'exec': case 'lance': case 'blizzard': case 'meteor': case 'revive': case 'chain': NEW[b](f, p); break;
    case 'quake': for (let i = 0; i < 3; i++) { const q = p * 1.5 - i * 0.22; if (q > 0 && q < 1) ring(tx, ty + 22, q * 70 * S, c0, 4); } burst(tx, ty + 14, p, [c0, c1, '#3b2a18'], 20, 46 * S); break;
    case 'buff': rune(sx, sy + 30, 30 * S, c0, t); for (let i = 0; i < 8; i++) { const x = sx - 24 * S + i * 7 * S, y = sy + 20 - ((p * 90 + i * 13) % 70); pr(x - 1, y, 3, 8, c1); pr(x - 3, y, 7, 3, c1); } if (ex.includes('r')) ring(sx, sy + 30, p * 80 * S, c0, 3); break;
    case 'debuff': rune(tx, ty + 26, 30 * S, c0, t, -1); for (let i = 0; i < 8; i++) { const x = tx - 24 * S + i * 7 * S, y = ty - 40 + ((p * 90 + i * 13) % 70); pr(x - 1, y, 3, 8, c0); pr(x - 3, y + 6, 7, 3, c0); } if (ex.includes('r')) ring(tx, ty + 26, (1 - p) * 70 * S, c1, 3); break;
    case 'orb': if (p < im) { const q = p / im; pr(sx + (tx - sx) * q - 7 * S, sy + (ty - sy) * q - 7 * S, 14 * S, 14 * S, c0); pr(sx + (tx - sx) * q - 3 * S, sy + (ty - sy) * q - 3 * S, 6 * S, 6 * S, c1); } else burst(tx, ty, after(), [c1, c0, c2], 18, 44 * S); break;
    case 'bolt': { const sd = Math.floor(t / 55); for (let j = 0; j < n; j++) { const ox = (j - (n - 1) / 2) * 40; if (p < 0.7) ZZ(tx + ox, ty - 150, tx + ox, ty, sd + j, c0, c1); } if (p > 0.2) burst(tx, ty, after(0.2), ['#fff', c0], 14, 36 * S); break; }
    case 'nova': g.fillStyle = c1 + '55'; g.beginPath(); g.arc(tx, ty, p * 50 * S, 0, 7); g.fill(); ring(tx, ty + 14, p * 100 * S, '#ffffff', 6); ring(tx, ty + 14, p * 72 * S, c0, 5); burst(tx, ty, p, ['#fff', c0, c1], 24, 70 * S); break;
    case 'rain': for (let i = 0; i < n; i++) { const st = i * 0.035, q = (p - st) / 0.4, x = tx - 56 * S + ((i * 47) % (112 * S)), gy = ty + 18 + ((i * 13) % 20) - 10; if (q > 0 && q < 1) { pr(x - 1, gy - 130 * (1 - q), 3, 12, c0); pr(x - 2, gy - 130 * (1 - q) + 10, 5, 4, c1); } else if (q >= 1 && q < 1.5) burst(x, gy, (q - 1) * 2, ['#fff', c1], 4, 10); } break;
    case 'shards': for (let i = 0; i < 9; i++) { const x = tx - 44 * S + i * 11 * S, h = 34 * S * Math.min(1, p * 2.2) * (1 - (i % 3) * 0.2); pr(x, ty + 22 - h, 7, h, c0); pr(x + 2, ty + 22 - h, 2, h, c1); } if (p > 0.3) burst(tx, ty, after(0.3), [c1, c0], 12, 40 * S); break;
    case 'heal': for (let k = 0; k < Math.max(1, n); k++) { const ox = [0, 34, -30][k % 3], oy = [0, -8, 10][k % 3]; for (let i = 0; i < 7; i++) sparkle(tx + ox + Math.sin(i * 2 + p * 3) * 16, ty + oy + 10 - p * 52 - (i % 4) * 8, i % 2 ? c1 : c0); } if (ex.includes('r')) ring(tx, ty + 22, p * 80 * S, c0, 4); break;
    case 'beam': { const a = Math.atan2(ty - sy, tx - sx), ex2 = tx + Math.cos(a) * 60, ey2 = ty + Math.sin(a) * 60, w = Math.sin(Math.min(1, p * 1.3) * Math.PI) * 9 * S; g.strokeStyle = c0; g.lineWidth = w + 3; g.beginPath(); g.moveTo(sx, sy); g.lineTo(ex2, ey2); g.stroke(); g.strokeStyle = c1; g.lineWidth = Math.max(1, w / 2.5); g.beginPath(); g.moveTo(sx, sy); g.lineTo(ex2, ey2); g.stroke(); if (p > 0.2) burst(tx, ty, after(0.2), ['#fff', c0, c1], 14, 36 * S); break; }
    case 'pillar': { const w = 48 * S * Math.sin(Math.min(1, p * 1.2) * Math.PI); grad(0, ty - 180, ty + 16, c0 + '00', c0 + 'dd'); g.fillRect(tx - w / 2, ty - 180, w, 196); ring(tx, ty + 22, p * 70 * S, c1, 4); for (let i = 0; i < 10; i++) sparkle(tx + Math.sin(i * 2.1 + p * 4) * 22 * S, ty + 10 - ((p * 90 + i * 17) % 110), i % 2 ? c0 : c1); if (ex.includes('e')) burst(tx, ty, p, [c0, c1], 16, 50 * S); break; }
    case 'arrow': for (let j = 0; j < n; j++) { const dy = (j - (n - 1) / 2) * 14, q = Math.min(1, Math.max(0, p - j * 0.05) / 0.5), x = hx + (tx - hx) * q, y = hy + (ty + dy * 0.4 - hy) * q, a = Math.atan2(ty - hy, tx - hx); if (q < 1) arrowP(x, y, a, 7, c0, c1, S); else burst(tx, ty + dy * 0.4, Math.min(1, (p - 0.5 - j * 0.05) * 2), ['#fff', c1], 6, 16 * S); } if (ex.includes('n') && p > 0.5) NEW.pin(f, p); if (ex.includes('c') && p > 0.5) { g.fillStyle = 'rgba(120,230,80,.4)'; g.beginPath(); g.arc(tx, ty, (p - 0.5) * 50, 0, 7); g.fill(); } break;
    case 'knives': for (let j = 0; j < n; j++) { const a0 = (j / n) * 6.28, q = Math.min(1, p / 0.5), x = sx + (tx + Math.cos(a0) * 30 - sx) * q, y = sy + (ty + Math.sin(a0) * 14 - sy) * q; pr(x - 3, y - 1, 7, 2, c0); pr(x + 3, y - 1, 2, 2, '#fff'); } if (p > 0.5) burst(tx, ty, after(0.5), ['#fff', c1], 10, 30 * S); break;
    case 'blink': if (p < 0.35) { g.globalAlpha = 0.6; pr(sx - 10, sy - 30, 20, 40, c0 + '55'); } else { for (let i = 0; i < 3; i++) streak(tx + i * 6 - 6, ty, -1.2 + i * 0.9, 70 * S, after(0.35), 5 * S, '#fff', c0); if (ex.includes('x')) streak(tx, ty, 0.9, 90 * S, after(0.45), 7 * S, c1, c0); } break;
    case 'cloud': { const q = Math.min(1, p * 1.6); for (let i = 0; i < 7; i++) { const a = i * 0.9, r = (14 + (i % 3) * 6) * S * q; g.fillStyle = (i % 2 ? c0 : c1) + '66'; g.beginPath(); g.arc(tx + Math.cos(a) * 22 * S * q, ty + Math.sin(a) * 10 * S * q, r, 0, 7); g.fill(); } for (let i = 0; i < 8; i++) pr(tx - 20 + i * 6, ty + 10 - p * 40 - (i % 3) * 6, 4, 4, c1); break; }
    case 'shot': { const step = 0.5 / n; for (let j = 0; j < n; j++) { const p0 = j * step, q = (p - p0) / 0.18; if (p >= p0 && p < p0 + 0.08) muzzle(hx, hy, S, c0); if (q > 0 && q < 1) tracer(hx, hy, tx + ((j * 7) % 9) - 4, ty + ((j * 5) % 7) - 3, q, c0, 3 * S); if (q >= 1 && q < 2.2) burst(tx, ty, (q - 1) / 1.2, ['#fff', c0, c1], 8, 20 * S); } if (p < 0.5) pr(hx - dir * 8 - p * 20 * dir, hy - 14 + p * 60, 3, 2, '#ffd84a'); if (ex.includes('e') && p > 0.3) burst(tx, ty, after(0.3), [c0, '#ff7a2e'], 14, 40 * S); break; }
    case 'scatter': if (p < 0.08) muzzle(hx, hy, S * 1.3, c0); for (let j = 0; j < 8; j++) { const q = Math.min(1, p / 0.25), ox = (j - 3.5) * 9 * S, oy = (((j * 5) % 7) - 3) * 5; tracer(hx, hy, tx + ox, ty + oy, q, c0, 2); if (p > 0.25) burst(tx + ox, ty + oy, Math.min(1, (p - 0.25) * 2), ['#fff', c1], 4, 10); } break;
    case 'ricochet': { const pts = [[hx, hy], [tx, ty], [tx + 44, ty - 14], [tx - 36, ty + 16], [tx + 20, ty + 26], [tx - 50, ty - 20]].slice(0, Math.max(2, n) + 1), segs = pts.length - 1, q = Math.min(0.999, p / 0.7) * segs, i = Math.floor(q), fq = q - i; if (p < 0.06) muzzle(hx, hy, S, c0); if (p < 0.7) tracer(pts[i][0], pts[i][1], pts[i + 1][0], pts[i + 1][1], fq, c0, 3 * S); for (let k = 1; k <= Math.min(segs, i + (p >= 0.7 ? 1 : 0)); k++) burst(pts[k][0], pts[k][1], Math.min(1, (q - k + 1) * 0.8), ['#fff', c0], 6, 16); break; }
    case 'flashbang': if (p < im) { const q = p / im, x = hx + (tx - hx) * q, y = hy + (ty - hy) * q - Math.sin(q * Math.PI) * 60; pr(x - 4, y - 5, 8, 10, '#8d97ad'); pr(x - 1, y - 8, 2, 3, '#ffd84a'); } else { const q = after(); g.fillStyle = `rgba(255,255,255,${0.8 * (1 - q)})`; g.beginPath(); g.arc(tx, ty, 30 + q * 60, 0, 7); g.fill(); ring(tx, ty + 18, q * 90 * S, '#ffffff', 5); star(tx, ty, q, '#fff', '#ffe45c'); } break;
    case 'gunrain': if (p < 0.3 && Math.floor(t / 60) % 2) muzzle(hx, hy - 10, S, c0); for (let i = 0; i < n; i++) { const st = i * 0.03, q = (p - st) / 0.25, x = tx - 60 + ((i * 43) % 120), gy = ty + ((i * 17) % 30) - 12; if (q > 0 && q < 1) tracer(hx, hy, x, gy, q, c0, 2); else if (q >= 1 && q < 1.6) burst(x, gy, (q - 1) / 0.6, ['#fff', c1], 5, 12); } break;
    case 'swarm': for (let i = 0; i < 26; i++) { const q = Math.min(1, p * 1.4), ph = i * 1.7, x = sx + (tx - sx) * q + Math.sin(t / 60 + ph) * 14 * (1 - q * 0.5), y = sy + (ty - sy) * q + Math.cos(t / 70 + ph) * 10; pr(x, y, 3, 3, i % 3 ? c0 : c1); } if (p > 0.7) burst(tx, ty, (p - 0.7) / 0.3, [c0, c1], 10, 26); break;
    case 'potion': if (p < im) { const q = p / im, x = hx + (tx - hx) * q, y = hy - 10 + (ty - hy + 10) * q - Math.sin(q * Math.PI) * 70, r = (t / 90) % 6.28; g.save(); g.translate(x, y); g.rotate(r); pr(-5, -4, 10, 9, c0); pr(-2, -9, 4, 5, '#cfeaff'); pr(-2, -11, 4, 2, '#8a5a2b'); g.restore(); } else { const q = after(); burst(tx, ty, q, ['#ffffff', '#cfeaff', c0, c1], 22, 50 * S); g.fillStyle = c0 + '55'; g.beginPath(); g.ellipse(tx, ty + 20, q * 50 * S, q * 18 * S, 0, 0, 7); g.fill(); for (let i = 0; i < 6; i++) pr(tx - 12 + i * 5, ty + 12 - q * 40 - (i % 3) * 5, 4, 4, c1); } break;
    case 'charge': { const q = Math.min(1, p / 0.4); for (let i = 0; i < 6; i++) { const x = sx + (tx - sx) * q - dir * i * 10, y = sy + (ty - sy) * q + 8; pr(x - 6, y - 2 + (i % 2) * 3, 12, 3, c1 + '88'); } if (p > 0.38) { star(tx, ty, after(0.4), '#fff', c0); ring(tx, ty + 20, after(0.4) * 60 * S, c0, 4); if (ex.includes('e')) burst(tx, ty, after(0.4), [c0, c1], 16, 44 * S); } break; }
    case 'breath': { for (let i = 0; i < 30; i++) { const q = (p * 1.6 + i * 0.033) % 1, x = hx + (tx - hx) * q + Math.sin(i * 3.1) * q * 18, y = hy - 6 + (ty - hy + 6) * q + Math.cos(i * 2.3) * q * 12, z = Math.max(2, (1 + q * 6) * S); if (p < 0.85) pr(x, y, z, z, [c1, c0, '#ffffff'][i % 3]); } if (p > im) burst(tx, ty, after(), [c0, c1], 14, 40 * S); break; }
    case 'dive': if (p < 0.45) { const q = p / 0.45; pr(tx - 8 * S, ty - 150 + 140 * q - 10, 16 * S, 20 * S, c0 + '99'); for (let i = 1; i < 5; i++) pr(tx - 3, ty - 150 + 140 * q - 10 - i * 14, 6, 10, c1 + '66'); } else { star(tx, ty, after(0.45), '#fff', c0); ring(tx, ty + 22, after(0.45) * 70 * S, c1, 5); if (ex.includes('e')) burst(tx, ty, after(0.45), [c0, c1, '#fff'], 20, 56 * S); } break;
    default: NEW.whirl(f, p);
  }
  if (ex.includes('e') && b !== 'pillar' && b !== 'shot' && b !== 'charge' && b !== 'breath' && b !== 'dive' && p > im) burst(tx, ty, after(), [c0, c1, '#3a1a10'], 12, 50 * S);
  if (ex.includes('b') && p > 0.35) for (let i = 0; i < 5; i++) pr(tx - 14 + i * 7, ty + (p - 0.35) * 44, 3, 6, '#c8102e');
  if (ex.includes('c') && b !== 'arrow' && p > 0.4) { g.fillStyle = 'rgba(120,230,80,.35)'; g.beginPath(); g.arc(tx, ty, (p - 0.4) * 44, 0, 7); g.fill(); }
  if (ex.includes('t') && p > 0.4) for (let i = 0; i < 3; i++) { const a = t / 140 + i * 2.1; sparkle(tx + Math.cos(a) * 18, ty - 34 + Math.sin(a) * 6, '#ffe45c'); }
  drawLayers(g, f, p, f.abn || f.k.slice(2), b, calm);
  g.restore();
}
