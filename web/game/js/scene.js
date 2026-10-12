// One picture of the battle each frame: ground, tile highlights, chests,
// units, effects and the minimap. Reads the screen state (st); never rules.
import { map, tiles, drawWater, drawTile, objects, vignette, poly, top, turnStep, pj, height } from './board.js';
import { cam, layout } from './camera.js';
import * as units from './units.js';
import { CHEST, TS } from './look.js';
import * as iso from './iso.js';

// st: { view, sel, look, aim, reach:Set, targets:Set(ids), starts:Set, danger:Set,
//       zone:{walk:Set, hit:Set}, attack:Set, picks:{tiles:Set, all:Set}, opt,
//       anim(u), shownHp(u), fx:[], pops:[] }
export function draw(canvas, st) {
  const g = canvas.getContext('2d'), v = st.view;
  if (!v) return;
  turnStep();
  const heroes = v.units.filter((u) => u.team === 0 && !u.companion);
  layout(heroes, st.sel);
  const now = performance.now();
  g.setTransform(1, 0, 0, 1, 0, 0);
  g.clearRect(0, 0, canvas.width, canvas.height);
  let sx = 0, sy = 0;
  if (st.shake && st.shake.amp) {
    const q = (now - st.shake.t0) / 260;
    if (q >= 0 && q < 1) {
      sx = Math.sin(q * 60) * st.shake.amp * (1 - q);
      sy = Math.cos(q * 47) * st.shake.amp * (1 - q);
    }
  }
  const K = cam.k, D = cam.dpr;
  g.setTransform(D * K, 0, 0, D * K, D * (cam.offX + sx * K), D * (cam.offY + sy * K));
  const list = tiles(), calm = st.opt.lowGraphics;
  if (map.flat) drawWater(g, list, now, calm);
  else iso.water(g, list, now, calm);
  for (const [x, y, p] of list) {
    if (map.flat) drawTile(g, x, y, p);
    else iso.tile(g, x, y, p);
    marks(g, x, y, p, st, now);
  }
  const obj = map.flat ? objects(list, now, calm, !!v.throne && !v.outcome) : iso.objects(list, now, calm);
  obj.sort((a, b) => a[0] - b[0]);
  g.save();
  g.imageSmoothingEnabled = false;
  obj.forEach((o) => o[1](g));
  g.restore();
  if (!st.opt.lowGraphics) vignette(g, canvas.width, canvas.height);
  (v.chests || []).forEach((c) => chest(g, c, now));
  units.draw(g, st.units, {
    now, k: K, dpr: D, sel: st.sel, look: st.look, aim: st.aim,
    anim: st.anim, shownHp: st.shownHp, calm: st.opt.calm, low: st.opt.lowGraphics, marks: st.marks,
  });
  if (st.overlay) st.overlay(g, now);
}

// marks shades a tile: where a hero can walk, what is in reach, the danger
// zone, a looked-at monster's range, aiming, fields, snares, kegs, banners
// [old draw L1858-1866].
function marks(g, x, y, p, st, now) {
  const key = x + ',' + y, t = top(x, y), v = st.view;
  if ((st.look && st.look.team && st.look.x === x && st.look.y === y) || (st.aim && st.aim.x === x && st.aim.y === y)) {
    g.save();
    g.strokeStyle = '#f2c14e';
    g.lineWidth = 4;
    g.shadowColor = '#f2c14e';
    g.shadowBlur = 12;
    g.beginPath();
    t.forEach((q, i) => (i ? g.lineTo(q[0], q[1]) : g.moveTo(q[0], q[1])));
    g.closePath();
    g.stroke();
    g.restore();
  }
  g.strokeStyle = 'rgba(0,0,0,.08)';
  g.lineWidth = 1;
  g.beginPath();
  g.moveTo(t[0][0], t[0][1]);
  t.slice(1).forEach((q) => g.lineTo(q[0], q[1]));
  g.closePath();
  g.stroke();
  const cb = st.opt.colorblind;
  if (st.zone) {
    if (st.zone.walk.has(key)) poly(g, t, 'rgba(190,120,255,.38)');
    else if (st.zone.hit.has(key)) poly(g, t, cb ? 'rgba(255,215,60,.32)' : 'rgba(255,60,60,.3)');
  }
  if (st.picks) {
    if (st.picks.chosen.has(key)) poly(g, t, 'rgba(255,215,0,.65)');
    else if (st.picks.all.has(key)) poly(g, t, 'rgba(255,215,0,.22)');
  }
  (v.fields || []).forEach((f) => {
    const d = f.Square ? Math.max(Math.abs(x - f.X), Math.abs(y - f.Y)) : Math.abs(x - f.X) + Math.abs(y - f.Y);
    if (d <= f.Range) poly(g, t, f.Kind === 'banner' ? 'rgba(255,215,90,.16)' : 'rgba(255,110,30,' + (0.25 + 0.1 * Math.sin(now / 200 + x + y)).toFixed(2) + ')');
  });
  const h = height(x, y) > 0 ? map.th : 0;
  if ((v.trapTiles || []).some((q) => q.x === x && q.y === y)) {
    g.fillStyle = '#9ae6b4';
    g.font = 'bold 18px sans-serif';
    g.textAlign = 'center';
    g.fillText('✳', p.x, p.y - h + 6);
  }
  if ((v.kegs || []).some((q) => q.X === x && q.Y === y)) {
    g.font = 'bold 22px sans-serif';
    g.textAlign = 'center';
    g.fillText('💣', p.x, p.y - h + 4);
  }
  if ((v.fields || []).some((f) => f.Kind === 'banner' && f.X === x && f.Y === y)) {
    rect(g, p.x - 2, p.y - h - 40, 4, 40, '#5a3a1a');
    rect(g, p.x + 2, p.y - h - 40, 20, 13, '#c8102e');
    rect(g, p.x + 6, p.y - h - 37, 6, 6, '#f2c14e');
  }
  if (st.danger && st.danger.has(key) && !st.zone) poly(g, t, cb ? 'rgba(255,215,60,.2)' : 'rgba(200,30,70,.22)');
  if (st.starts && st.starts.has(key)) poly(g, t, 'rgba(110,210,255,.5)');
  if (st.reach.has(key)) poly(g, t, 'rgba(110,210,255,.5)');
  else if (st.attack && st.attack.has(key)) poly(g, t, cb ? 'rgba(255,150,40,.32)' : 'rgba(255,60,60,.3)');
  const o = st.unitAt(x, y);
  if (o && o === st.sel) poly(g, t, 'rgba(242,184,75,.4)');
  if (o && o === st.aim) poly(g, t, 'rgba(255,215,0,.55)');
  if (o && st.inReach && st.inReach.has(o.id) && !st.targets.has(o.id)) poly(g, t, 'rgba(255,90,90,.35)');
  if (o && st.targets.has(o.id)) poly(g, t, st.sel && o.team === st.sel.team ? (cb ? 'rgba(255,255,255,.55)' : 'rgba(110,255,150,.55)') : 'rgba(255,90,90,.55)');
}

function rect(g, x, y, w, h, c) {
  g.fillStyle = c;
  g.fillRect(Math.round(x), Math.round(y), w, h);
}

// chest draws a closed chest on its tile [old drawChest L1002].
function chest(g, c, now) {
  const p = pj(c.x, c.y), h = height(c.x, c.y) > 0 ? map.th : 0, x = p.x, y = p.y - h - 4 + (map.flat ? 10 : 0);
  const T = CHEST[c.kind] || CHEST.wood;
  g.fillStyle = 'rgba(0,0,0,.35)';
  g.beginPath();
  g.ellipse(x, y + 6, 16, 6, 0, 0, 7);
  g.fill();
  rect(g, x - 14, y - 14, 28, 18, T.b);
  rect(g, x - 13, y - 13, 26, 16, T.c);
  rect(g, x - 14, y - 20, 28, 8, T.b);
  rect(g, x - 13, y - 19, 26, 6, T.c);
  rect(g, x - 2, y - 16, 4, 6, '#1a1226');
  rect(g, x - 14, y - 8, 28, 2, T.b);
  if (Math.sin(now / 300 + c.x) > 0.6) sparkle(g, x + 8, y - 22, '#fff');
}

function sparkle(g, x, y, c) {
  rect(g, x - 1, y - 4, 3, 9, c);
  rect(g, x - 4, y - 1, 9, 3, c);
}

export { TS };
