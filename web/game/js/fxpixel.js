// The drawn (pixel) part of every attack effect: the fireball that flies,
// the bullet line, arrows, falling meteors, rings and bursts. Pictures only.
// Ported from the old game: helpers [old L814-820, L861-863, L918-921], the
// older effects [drawFx0 L821-840, NEW L842-872] and the per-skill effect
// (the per-skill effect is in fxskill.js). f = { k, sx, sy, tx, ty, S }:
// from the user (sx, sy) to the target (tx, ty) in map pixels; p runs 0..1.

let g = null; // the canvas being drawn on

const pr = (x, y, w, h, c) => { g.fillStyle = c; g.fillRect(Math.round(x), Math.round(y), w, h); };
function flash(a) { g.save(); g.setTransform(1, 0, 0, 1, 0, 0); g.globalAlpha = a; g.fillStyle = '#fff'; g.fillRect(0, 0, g.canvas.width, g.canvas.height); g.restore(); }
function streak(x, y, a, len, p, w, c1, c2) { const n = 9, m = Math.floor(Math.min(1, p * 1.8) * n); for (let i = 0; i <= m; i++) { const f = i / n - 0.5, px = x + Math.cos(a) * len * f, py = y + Math.sin(a) * len * f; pr(px - w / 2, py - w / 2, w, w, c1); pr(px - w / 4, py - w / 4, w / 2, w / 2, c2); } }
function burst(x, y, p, pal, n, r) { for (let i = 0; i < n; i++) { const a = i * 2.4, sp = 0.5 + ((i * 37) % 10) / 14, d = p * r * sp, z = Math.max(2, Math.round(8 * (1 - p))); pr(x + Math.cos(a) * d, y + Math.sin(a) * d * 0.75 - p * 14, z, z, pal[Math.min(pal.length - 1, Math.floor(p * pal.length))]); } }
function star(x, y, p, c1, c2) { for (let i = 0; i < 8; i++) { const a = (i * Math.PI) / 4, d = 8 + p * 30; pr(x + Math.cos(a) * d - 3, y + Math.sin(a) * d - 3, 6, 6, i % 2 ? c2 : c1); pr(x + Math.cos(a) * d * 0.55 - 2, y + Math.sin(a) * d * 0.55 - 2, 4, 4, '#fff'); } }
const orb = (f, q, c1, c2) => { pr(f.sx + (f.tx - f.sx) * q - 6, f.sy + (f.ty - f.sy) * q - 6, 12, 12, c1); pr(f.sx + (f.tx - f.sx) * q - 3, f.sy + (f.ty - f.sy) * q - 3, 6, 6, c2); };
const sparkle = (x, y, c) => { pr(x - 1, y - 4, 3, 9, c); pr(x - 4, y - 1, 9, 3, c); };
const ar = (f, p, tr, dy, c, L = 6) => { const sx = f.sx, sy = f.sy + dy, tx = f.tx, ty = f.ty + dy * 0.4, q = Math.min(1, p / tr), x = sx + (tx - sx) * q, y = sy + (ty - sy) * q, a = Math.atan2(ty - sy, tx - sx); if (p < tr) { for (let i = 0; i < L; i++) pr(x - Math.cos(a) * i * 4 - 2, y - Math.sin(a) * i * 4 - 2, 4, 4, i > L - 3 ? '#6ee7a0' : c); pr(x - 3, y - 3, 7, 7, '#e8ecf5'); } else burst(tx, ty, (p - tr) / (1 - tr), ['#fff', '#ffd84a'], 6, 16); };
const ring = (x, y, r, c, w) => { g.strokeStyle = c; g.lineWidth = w; g.beginPath(); g.ellipse(x, y, Math.max(1, r), Math.max(1, r * 0.4), 0, 0, 7); g.stroke(); };
const ZZ = (x1, y1, x2, y2, seed, c1, c2) => { [[c1, 7], [c2, 3]].forEach(([c, w]) => { g.strokeStyle = c; g.lineWidth = w; g.beginPath(); g.moveTo(x1, y1); for (let i = 1; i < 5; i++) { const t = i / 5; g.lineTo(x1 + (x2 - x1) * t + ((i * 53 + seed * 17) % 15) - 7, y1 + (y2 - y1) * t + ((i * 31 + seed * 13) % 15) - 7); } g.lineTo(x2, y2); g.stroke(); }); };
function arrowP(x, y, a, L, c1, c2, S) { for (let i = 0; i < L; i++) pr(x - Math.cos(a) * i * 4 * S - 2 * S, y - Math.sin(a) * i * 4 * S - 2 * S, 4 * S, 3 * S, c1); pr(x - 3 * S, y - 3 * S, 7 * S, 7 * S, c2); }
function tracer(x1, y1, x2, y2, q, c, w) { const x = x1 + (x2 - x1) * q, y = y1 + (y2 - y1) * q, q0 = Math.max(0, q - 0.35); g.strokeStyle = c; g.lineWidth = w; g.beginPath(); g.moveTo(x1 + (x2 - x1) * q0, y1 + (y2 - y1) * q0); g.lineTo(x, y); g.stroke(); pr(x - 2, y - 2, 4, 4, '#fff'); }
function muzzle(x, y, S, c) { star(x, y, 0.4, c, '#ffffff'); pr(x - 5 * S, y - 5 * S, 10 * S, 10 * S, '#fff6a8'); }
function rune(x, y, r, c, t, dir = 1) { g.save(); g.strokeStyle = c; g.lineWidth = 2; g.beginPath(); g.ellipse(x, y, r, r * 0.42, 0, 0, 7); g.stroke(); g.beginPath(); g.ellipse(x, y, r * 0.7, r * 0.3, 0, 0, 7); g.stroke(); for (let i = 0; i < 8; i++) { const a = (t / 400) * dir + i * 0.785; pr(x + Math.cos(a) * r * 0.85 - 2, y + Math.sin(a) * r * 0.36 - 2, 4, 4, c); } g.restore(); }
const grad = (x, y0, y1, c0, c1) => { const gr = g.createLinearGradient(0, y0, 0, y1); gr.addColorStop(0, c0); gr.addColorStop(1, c1); g.fillStyle = gr; return x; };

// The older effects [drawFx0 L821-840].
function drawFx0(f, p, t) {
  const { k, sx, sy, tx, ty } = f;
  g.globalAlpha = p > 0.75 ? (1 - p) * 4 : 1;
  const sp = (pal, o) => { if (p < 0.35) orb(f, p / 0.35, o[0], o[1]); else burst(tx, ty, (p - 0.35) / 0.65, pal, 16, 40); };
  if (k === 'slash') streak(tx, ty, -0.8, 70, p, 6, '#fff', '#9fd8ff');
  else if (k === 'rend') { streak(tx, ty, -0.8, 70, p, 6, '#fff', '#ff4d4d'); if (p > 0.3) for (let i = 0; i < 6; i++) pr(tx - 14 + i * 5, ty - 8 + (p - 0.3) * 45 * (1 + (i % 3) * 0.3), 3, 6, '#c8102e'); }
  else if (k === 'heavy') { streak(tx, ty, -0.8, 100, p, 10, '#fff', '#ffd84a'); streak(tx, ty, 0.8, 100, Math.max(0, p - 0.15), 10, '#fff', '#ffd84a'); if (p < 0.25) flash(0.45 * (1 - p / 0.25)); }
  else if (k === 'bash') star(tx, ty, p, '#fff', '#ffd84a');
  else if (k === 'shield') { star(tx, ty, Math.min(1, p * 1.5), '#fff', '#9fd8ff'); for (let i = 0; i < 3; i++) { const a = t / 140 + i * 2.1; sparkle(tx + Math.cos(a) * 18, ty - 34 + Math.sin(a) * 6, '#ffe45c'); } }
  else if (k === 'fire') sp(['#fff6a8', '#ffb12e', '#e4572e', '#7a1f12'], ['#ffb12e', '#fff6a8']);
  else if (k === 'ice') sp(['#ffffff', '#bfeaff', '#7fd1ff', '#3a8fd0'], ['#7fd1ff', '#fff']);
  else if (k === 'light') { g.fillStyle = 'rgba(255,246,200,.55)'; g.beginPath(); g.arc(tx, ty, p * 36, 0, 7); g.fill(); burst(tx, ty, p, ['#fff', '#fff6c8', '#ffd84a'], 14, 40); }
  else if (k === 'dark') { g.fillStyle = 'rgba(60,20,100,.45)'; g.beginPath(); g.arc(tx, ty, p * 30, 0, 7); g.fill(); burst(tx, ty, p, ['#d9b8ff', '#8a5fd0', '#3b1b66'], 18, 40); }
  else if (k === 'bolt') { if (p < 0.2) flash(0.5 * (1 - p / 0.2)); if (p > 0.08 && p < 0.7) { const sd = Math.floor(t / 55); [['#ffd84a', 8], ['#fff', 3]].forEach(([c, w]) => { g.strokeStyle = c; g.lineWidth = w; g.beginPath(); g.moveTo(tx, ty - 140); for (let i = 1; i <= 6; i++) g.lineTo(tx + (i < 6 ? ((i * 53 + sd * 17) % 25) - 12 : 0), ty - 140 + i * 23); g.stroke(); }); } if (p > 0.2) burst(tx, ty, (p - 0.2) / 0.8, ['#fff', '#ffd84a', '#f2b84b'], 14, 34); }
  else if (k === 'heal') { for (let i = 0; i < 8; i++) sparkle(tx + Math.sin(i * 2 + p * 3) * 18, ty + 8 - p * 52 - (i % 4) * 9, i % 2 ? '#d8ffe4' : '#6ee7a0'); }
  else if (k === 'grace') { const w = 44 * Math.sin(Math.min(1, p * 1.2) * Math.PI); grad(0, ty - 170, ty + 12, 'rgba(255,240,160,0)', 'rgba(255,240,160,.85)'); g.fillRect(tx - w / 2, ty - 170, w, 182); for (let i = 0; i < 10; i++) sparkle(tx + Math.sin(i * 2.1 + p * 4) * 20, ty + 10 - ((p * 90 + i * 17) % 110), i % 2 ? '#fff6c8' : '#ffd84a'); }
  else {
    const tr = k === 'parrow' ? 0.55 : 0.5, q = Math.min(1, p / tr), x = sx + (tx - sx) * q, y = sy + (ty - sy) * q, a = Math.atan2(ty - sy, tx - sx), L = k === 'parrow' ? 9 : 6;
    if (p < tr) { for (let i = 0; i < L; i++) pr(x - Math.cos(a) * i * 4 - 2, y - Math.sin(a) * i * 4 - 2, 4, 4, i > L - 3 ? '#6ee7a0' : k === 'parrow' ? '#ffd84a' : '#8a5a2b'); pr(x - 3, y - 3, 7, 7, '#e8ecf5'); }
    else {
      const r = (p - tr) / (1 - tr), e = Math.min(1, r * 3);
      if (k === 'net') { g.strokeStyle = '#fff'; g.lineWidth = 2; for (let i = 0; i < 3; i++) { const b = i * 1.05; g.beginPath(); g.moveTo(tx - Math.cos(b) * 24 * e, ty - Math.sin(b) * 24 * e); g.lineTo(tx + Math.cos(b) * 24 * e, ty + Math.sin(b) * 24 * e); g.stroke(); } g.beginPath(); g.arc(tx, ty, 14 * e, 0, 7); g.stroke(); }
      else burst(tx, ty, r, k === 'parrow' ? ['#fff', '#ffd84a', '#e4572e'] : ['#fff', '#ffd84a'], k === 'parrow' ? 16 : 8, k === 'parrow' ? 36 : 20);
    }
  }
}

// The named older effects [NEW L842-872].
const NEW = {
  whirl: (f, p) => { const { tx, ty } = f; for (let k = 0; k < 2; k++) for (let i = 0; i < 14; i++) { const a = p * 12 + i * 0.45 + k * 3, r = 22 + k * 10; pr(tx + Math.cos(a) * r - 3, ty + Math.sin(a) * r * 0.5 - 3, 6, 6, k ? '#9fd8ff' : '#fff'); } if (p < 0.15) flash(0.25); },
  wave: (f, p) => { const { sx, sy, tx, ty } = f, q = Math.min(1, p / 0.55), x = sx + (tx - sx) * q, y = sy + (ty - sy) * q; if (p < 0.55) { g.strokeStyle = '#cfeaff'; g.lineWidth = 6; g.beginPath(); g.arc(x - 12, y, 20, -1, 1); g.stroke(); g.strokeStyle = '#fff'; g.lineWidth = 2; g.beginPath(); g.arc(x - 12, y, 20, -1, 1); g.stroke(); for (let i = 1; i < 5; i++) pr(x - i * 9, y + Math.sin(i) * 4, 4, 4, '#9fd8ff'); } else { streak(tx, ty, -0.6, 80, (p - 0.55) / 0.45, 8, '#fff', '#9fd8ff'); burst(tx, ty, (p - 0.55) / 0.45, ['#fff', '#9fd8ff'], 10, 36); } },
  exec: (f, p) => { const { tx, ty } = f; if (p < 0.4) { const h = 130 * Math.min(1, p / 0.25); pr(tx - 2, ty - h - 10, 5, h, '#fff'); pr(tx - 6, ty - h - 10, 3, h, '#ff6a6a66'); if (p > 0.3) flash(0.55 * (1 - (p - 0.3) / 0.1)); } else { const q = (p - 0.4) / 0.6; streak(tx, ty, -0.9, 110, q, 10, '#fff', '#ff4d4d'); streak(tx, ty, 0.9, 110, Math.max(0, q - 0.1), 10, '#fff', '#ff4d4d'); burst(tx, ty, q, ['#fff', '#ff4d4d', '#8a1010'], 20, 50); } },
  dslash: (f, p) => { const { tx, ty } = f; streak(tx, ty, -0.7, 140, p, 14, '#fff', '#ff7a2e'); streak(tx + 6, ty + 8, -0.7, 120, Math.max(0, p - 0.1), 8, '#ffd84a', '#e4572e'); burst(tx, ty, p, ['#fff6a8', '#ffb12e', '#e4572e', '#7a1f12'], 26, 64); if (p < 0.2) flash(0.4 * (1 - p / 0.2)); },
  quake: (f, p) => { const { tx, ty } = f; for (let i = 0; i < 3; i++) { const q = p * 1.5 - i * 0.22; if (q > 0 && q < 1) ring(tx, ty + 22, q * 66, 'rgba(255,225,170,' + (1 - q) + ')', 4); } for (let i = 0; i < 5; i++) { const a = i * 1.26; g.strokeStyle = '#3b2a18'; g.lineWidth = 3; g.beginPath(); g.moveTo(tx, ty + 22); g.lineTo(tx + Math.cos(a) * 40 * Math.min(1, p * 3) + 3, ty + 22 + Math.sin(a) * 14 * Math.min(1, p * 3)); g.stroke(); } burst(tx, ty + 14, p, ['#d9c08a', '#9a7b4f', '#5a452c'], 18, 44); },
  crush: (f, p) => { const { tx, ty } = f; if (p < 0.3) { const y = ty - 90 + 70 * (p / 0.3); pr(tx - 12, y, 24, 14, '#cfd6e6'); pr(tx - 3, y - 26, 6, 26, '#8a5a2b'); } else { const q = (p - 0.3) / 0.7; star(tx, ty, q, '#fff', '#ffd84a'); for (let i = 0; i < 8; i++) pr(tx + Math.cos(i * 0.8) * q * 50 - 2, ty + Math.sin(i * 0.8) * q * 30 - q * 20, 5, 5, '#aeb4c2'); if (q < 0.2) flash(0.35); } },
  slam: (f, p) => { const { tx, ty } = f; ring(tx, ty + 22, p * 95, 'rgba(255,255,255,' + (1 - p) + ')', 7); ring(tx, ty + 22, p * 70, 'rgba(255,200,120,' + (1 - p) + ')', 4); burst(tx, ty + 10, p, ['#d9c08a', '#8d99b8', '#3b4260'], 28, 80); },
  titan: (f, p) => { const { tx, ty } = f, w = 50 * Math.sin(Math.min(1, p * 1.2) * Math.PI); grad(0, ty - 150, ty + 20, 'rgba(255,230,140,0)', 'rgba(255,230,140,.9)'); g.fillRect(tx - w / 2, ty - 150, w, 170); ring(tx, ty + 22, p * 80, 'rgba(255,220,120,' + (1 - p) + ')', 6); burst(tx, ty, p, ['#fff', '#ffd84a', '#e4a21a'], 20, 50); },
  lance: (f, p) => { const { sx, sy, tx, ty } = f, q = Math.min(1, p / 0.4), x = sx + (tx - sx) * q, y = sy + (ty - sy) * q, a = Math.atan2(ty - sy, tx - sx); if (p < 0.4) { for (let i = 0; i < 10; i++) pr(x - Math.cos(a) * i * 4 - 2, y - Math.sin(a) * i * 4 - 2, 5, 5, i < 2 ? '#fff' : '#7fd1ff'); pr(x + 2, y - 4, 8, 8, '#fff'); } else { const r = (p - 0.4) / 0.6; burst(tx, ty, r, ['#fff', '#bfeaff', '#7fd1ff', '#3a8fd0'], 18, 46); for (let i = 0; i < 6; i++) { const a2 = i * 1.05, d = r * 40; pr(tx + Math.cos(a2) * d - 3, ty + Math.sin(a2) * d * 0.7 - 6, 6, 10, '#bfeaff'); } } },
  chain: (f, p) => { const { sx, sy, tx, ty } = f, sd = Math.floor(performance.now() / 60), pts = [[sx, sy], [tx, ty], [tx + 38, ty - 10], [tx - 34, ty + 14]], n = Math.min(2, Math.floor(p * 5)); if (p < 0.9) for (let i = 0; i <= n; i++) ZZ(pts[i][0], pts[i][1], pts[i + 1][0], pts[i + 1][1], sd + i, '#ffd84a', '#fff'); if (p > 0.2) for (let i = 1; i < 4; i++) burst(pts[i][0], pts[i][1], Math.min(1, (p - 0.2) * 1.2), ['#fff', '#ffd84a'], 7, 18); },
  meteor: (f, p) => { const { tx, ty } = f; if (p < 0.45) { const q = p / 0.45, x = tx + 110 * (1 - q), y = ty - 150 * (1 - q); for (let i = 0; i < 8; i++) pr(x + i * 7, y - i * 9, 12 - i, 12 - i, i < 2 ? '#fff6a8' : i < 5 ? '#ffb12e' : '#e4572e'); pr(x - 8, y - 8, 18, 18, '#ffb12e'); pr(x - 4, y - 4, 10, 10, '#fff6a8'); } else { const q = (p - 0.45) / 0.55; burst(tx, ty, q, ['#fff6a8', '#ffb12e', '#e4572e', '#7a1f12'], 34, 80); ring(tx, ty + 20, q * 70, 'rgba(255,160,60,' + (1 - q) + ')', 6); if (q < 0.15) flash(0.5); } },
  blizzard: (f, p) => { const { tx, ty } = f; for (let i = 0; i < 34; i++) { const x = tx - 46 + ((i * 37) % 92) + Math.sin(p * 6 + i) * 6, y = ty - 100 + ((p * 190 + i * 23) % 130); pr(x, y, 3, 3, '#fff'); } ring(tx, ty + 22, 50 * Math.min(1, p * 2), 'rgba(160,225,255,.8)', 4); for (let i = 0; i < 7; i++) { const x = tx - 40 + i * 13, h = 26 * Math.min(1, p * 2.2) * (1 - (i % 3) * 0.2); pr(x, ty + 22 - h, 6, h, '#bfeaff'); pr(x + 2, ty + 22 - h, 2, h, '#fff'); } },
  nova: (f, p) => { const { tx, ty } = f; g.fillStyle = 'rgba(200,170,255,' + 0.5 * (1 - p) + ')'; g.beginPath(); g.arc(tx, ty, p * 50, 0, 7); g.fill(); ring(tx, ty + 14, p * 100, 'rgba(255,255,255,' + (1 - p) + ')', 7); ring(tx, ty + 14, p * 72, 'rgba(190,140,255,' + (1 - p) + ')', 5); burst(tx, ty, p, ['#fff', '#d9b8ff', '#8a5fd0'], 24, 70); if (p < 0.15) flash(0.4); },
  cleanse: (f, p) => { const { tx, ty } = f; ring(tx, ty + 22, p * 50, 'rgba(255,255,255,' + (1 - p) + ')', 4); for (let i = 0; i < 10; i++) { const a = i * 0.63 + p * 8; sparkle(tx + Math.cos(a) * 16, ty + 16 - p * 52 - (i % 3) * 7, i % 2 ? '#bfeaff' : '#fff'); } },
  holy: (f, p) => { const { tx, ty } = f; if (p < 0.5) { g.strokeStyle = 'rgba(255,246,200,.7)'; g.lineWidth = 5; for (let i = -3; i <= 3; i++) { g.beginPath(); g.moveTo(tx + i * 9, ty - 130); g.lineTo(tx + i * 3, ty + 10); g.stroke(); } flash(0.2 * p * 2); } else burst(tx, ty, (p - 0.5) / 0.5, ['#fff', '#fff6c8', '#ffd84a'], 22, 50); sparkle(tx, ty - 24 - p * 18, '#fff'); },
  mheal: (f, p) => { const { tx, ty } = f; ring(tx, ty + 22, p * 80, 'rgba(110,231,160,' + (1 - p) + ')', 6); ring(tx, ty + 22, p * 55, 'rgba(216,255,228,' + (1 - p) + ')', 3); [[0, 0], [34, -8], [-30, 10]].forEach(([dx, dy], k) => { for (let i = 0; i < 4; i++) sparkle(tx + dx + Math.sin(i * 2 + k) * 10, ty + dy + 10 - p * 46 - i * 7, i % 2 ? '#d8ffe4' : '#6ee7a0'); }); },
  revive: (f, p) => { const { tx, ty } = f, w = 34 * Math.sin(Math.min(1, p * 1.1) * Math.PI); grad(0, ty - 130, ty + 20, 'rgba(255,250,200,0)', 'rgba(255,250,200,.9)'); g.fillRect(tx - w / 2, ty - 130, w, 150); if (p > 0.3) { const s = Math.min(1, (p - 0.3) / 0.3) * 36; g.fillStyle = 'rgba(255,255,255,.85)'; [-1, 1].forEach((d) => { g.beginPath(); g.moveTo(tx + d * 6, ty - 10); g.lineTo(tx + d * (6 + s), ty - 40); g.lineTo(tx + d * (6 + s * 0.8), ty + 6); g.closePath(); g.fill(); }); } for (let i = 0; i < 10; i++) sparkle(tx + Math.sin(i * 1.9) * 20, ty + 14 - ((p * 80 + i * 13) % 80), i % 2 ? '#fff6c8' : '#ffd84a'); },
  volley: (f, p) => { [-16, 0, 16].forEach((dy, i) => ar(f, Math.max(0, Math.min(1, p - i * 0.06)), 0.5, dy, '#8a5a2b')); },
  poison: (f, p) => { const { tx, ty } = f; ar(f, p, 0.45, 0, '#6fe04a'); if (p > 0.45) { const q = (p - 0.45) / 0.55; g.fillStyle = 'rgba(120,230,80,' + 0.45 * (1 - q * 0.6) + ')'; [[0, 0, 24], [-14, 6, 16], [14, 4, 18]].forEach(([dx, dy, r]) => { g.beginPath(); g.arc(tx + dx, ty + dy, r * (0.5 + q), 0, 7); g.fill(); }); for (let i = 0; i < 6; i++) pr(tx - 16 + i * 6, ty + 10 - q * 38 - (i % 3) * 5, 4, 4, '#c8ff9a'); } },
  pin: (f, p) => { const { tx, ty } = f; ar(f, p, 0.4, 0, '#8a5a2b'); if (p > 0.4) { const q = Math.min(1, (p - 0.4) * 3); ring(tx, ty + 18, 22 * q, '#c9ced8', 3); for (let i = 0; i < 4; i++) pr(tx - 24 + i * 16, ty + 14, 7, 5, '#aab'); burst(tx, ty, (p - 0.4) / 0.6, ['#fff', '#c9ced8'], 8, 20); } },
  pierce: (f, p) => { const { sx, sy, tx, ty } = f, a = Math.atan2(ty - sy, tx - sx), ex = tx + Math.cos(a) * 70, ey = ty + Math.sin(a) * 70, w = Math.sin(Math.min(1, p * 1.3) * Math.PI) * 8; g.strokeStyle = 'rgba(180,240,255,.9)'; g.lineWidth = w + 2; g.beginPath(); g.moveTo(sx, sy); g.lineTo(ex, ey); g.stroke(); g.strokeStyle = '#fff'; g.lineWidth = Math.max(1, w / 3); g.beginPath(); g.moveTo(sx, sy); g.lineTo(ex, ey); g.stroke(); if (p > 0.25) { burst(tx, ty, (p - 0.25) / 0.75, ['#fff', '#7fd1ff'], 12, 30); burst(ex, ey, (p - 0.25) / 0.75, ['#fff', '#7fd1ff'], 8, 24); } },
  rain: (f, p) => { const { tx, ty } = f; for (let i = 0; i < 16; i++) { const st = i * 0.04, q = (p - st) / 0.4, x = tx - 52 + ((i * 47) % 104), gy = ty + 18 + ((i * 13) % 20) - 10; if (q > 0 && q < 1) { pr(x - 1, gy - 130 * (1 - q), 3, 12, '#8a5a2b'); pr(x - 2, gy - 130 * (1 - q) + 10, 5, 4, '#e8ecf5'); } else if (q >= 1 && q < 1.5) burst(x, gy, (q - 1) * 2, ['#fff', '#ffd84a'], 4, 10); } },
  claw: (f, p) => { const { tx, ty } = f; for (let i = 0; i < 3; i++) streak(tx - 10 + i * 10, ty - 6 + i * 3, -1.1, 70, Math.max(0, p - i * 0.08), 6, '#fff', '#ff4d4d'); if (p > 0.3) for (let i = 0; i < 5; i++) pr(tx - 14 + i * 7, ty + (p - 0.3) * 40, 3, 6, '#c8102e'); },
  rampage: (f, p) => { const { tx, ty } = f; for (let i = 0; i < 5; i++) { const q = Math.max(0, Math.min(1, (p - i * 0.12) * 2)); if (q > 0) streak(tx + ((i * 29) % 30) - 15, ty + ((i * 17) % 24) - 12, -1.4 + i * 0.7, 70, q, 7, '#fff', '#ff7a2e'); } burst(tx, ty, p, ['#fff6a8', '#ff7a2e', '#a02010'], 14, 50); },
  spit: (f, p) => { const { sx, sy, tx, ty } = f, q = Math.min(1, p / 0.4); if (p < 0.4) { const x = sx + (tx - sx) * q, y = sy + (ty - sy) * q - Math.sin(q * 3.14) * 20; pr(x - 5, y - 5, 10, 10, '#6fe04a'); pr(x - 2, y - 2, 4, 4, '#c8ff9a'); } else { const r = (p - 0.4) / 0.6; g.fillStyle = 'rgba(120,230,80,' + 0.5 * (1 - r * 0.5) + ')'; g.beginPath(); g.arc(tx, ty, 10 + r * 22, 0, 7); g.fill(); burst(tx, ty, r, ['#c8ff9a', '#6fe04a'], 10, 30); } },
  bullet: (f, p) => { const { sx, sy, tx, ty } = f; if (p < 0.25) { const q = p / 0.25, x = sx + (tx - sx) * q, y = sy + (ty - sy) * q, q0 = Math.max(0, q - 0.35); g.strokeStyle = '#ffe45c'; g.lineWidth = 3; g.beginPath(); g.moveTo(sx + (tx - sx) * q0, sy + (ty - sy) * q0); g.lineTo(x, y); g.stroke(); pr(sx - 5, sy - 5, 10, 10, '#fff6a8'); } else burst(tx, ty, (p - 0.25) / 0.75, ['#fff', '#ffd84a', '#ff7a2e'], 10, 22); },
};

// The helpers, for the per-skill effects in fxskill.js (same canvas).
export const tools = { pr, flash, streak, burst, star, sparkle, ring, ZZ, arrowP, tracer, muzzle, rune, grad, NEW, setG: (c) => (g = c) };

// draw draws one effect f at p (0..1) on canvas g2 [old drawFx L874].
export function drawOlder(g2, f, p, t) {
  g = g2;
  g.save();
  const fn = NEW[f.k];
  if (fn) { g.globalAlpha = p > 0.8 ? (1 - p) * 5 : 1; fn(f, p, t); }
  else drawFx0(f, p, t);
  g.restore();
}
