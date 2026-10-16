// Units on the map: the ring under each one, its picture with glows, its
// health bar and name, and status badges for the unit you are looking at.
// Matches the old game's unit drawing [old L1844-1880, L974-979].
import { pic, heroPic, sheets } from './assets.js';
import { feet, pj, map } from './board.js';
import { TS, ELEMENT, GLOW, AURA, BIG } from './look.js';
import { active, label, hue, STATUS } from './statuses.js';

const PAD = 34; // empty space around a hero picture [old PAD]
const boxed = new Map(); // hero pictures placed in their box, by picture

// heroBox puts a hero picture on a canvas with room around it, and remembers
// the drawing box (so every hero stands the same size) [old L2168-2170].
function heroBox(cls, im) {
  if (boxed.has(im)) return boxed.get(im);
  const b = sheets.heroBox[cls];
  if (!b) return im;
  const bw = Math.round(b[0]), bh = b[1];
  const cw = Math.max(bw, im.naturalWidth) + PAD * 2, ch = Math.max(bh, im.naturalHeight) + PAD;
  const c = document.createElement('canvas');
  c.width = cw;
  c.height = ch;
  const q = c.getContext('2d');
  q.imageSmoothingEnabled = false;
  q.drawImage(im, Math.round((cw - im.naturalWidth) / 2), ch - im.naturalHeight);
  Object.assign(c, { bw, bh, px: (cw - bw) / 2, py: ch - bh });
  boxed.set(im, c);
  return c;
}

const tinted = new Map();
// blue tints a companion spirit's picture [old tintImg].
function blue(im) {
  if (tinted.has(im)) return tinted.get(im);
  const c = document.createElement('canvas');
  c.width = im.naturalWidth || im.width;
  c.height = im.naturalHeight || im.height;
  const x = c.getContext('2d');
  x.drawImage(im, 0, 0);
  x.globalCompositeOperation = 'source-atop';
  x.fillStyle = 'rgba(70,150,255,.42)';
  x.fillRect(0, 0, c.width, c.height);
  tinted.set(im, c);
  return c;
}

const isHero = (u) => (u.team === 0 || u.pvp) && !u.companion && !u.decoy && !u.pet && !u.ally;
const monster = (u) => u.companion || u.pet || u.ally || (u.team === 1 && !u.pvp);

// picture is what a unit looks like.
export function picture(u) {
  if (isHero(u) || u.decoy) {
    const im = heroPic(u.class, u.stage || 0, u.ascension);
    return im && heroBox(u.class, im);
  }
  if (u.boss || u.elite) return pic('bosses', u.class) || pic('monsters', u.kind);
  if (monster(u)) return pic('monsters', u.kind) || pic('bosses', u.class) || pic('heroes-first', u.class);
  return pic('heroes-first', u.class);
}

// aura is a hero's glow color for its stage [old AUR].
function aura(u) {
  if (u.stage >= 3 && u.element) return ELEMENT[u.element];
  return (GLOW[u.class] && GLOW[u.class][u.stage]) || AURA[u.class] || '#ffd84a';
}

// sprite draws a unit's picture standing at (sx, by); returns its top.
function sprite(g, u, sx, by, s) {
  const hero = isHero(u) && (u.stage || 0) > 0 ? u : null;
  let im = picture(u);
  if (u.statuses && u.statuses.vamp > 0) {
    const pu = 0.45 + 0.3 * Math.sin(s.now / 180), gr = g.createRadialGradient(sx, by - 44, 4, sx, by - 44, 64);
    gr.addColorStop(0, `rgba(255,30,50,${pu})`);
    gr.addColorStop(1, 'rgba(255,30,50,0)');
    g.fillStyle = gr;
    g.beginPath();
    g.ellipse(sx, by - 44, 52, 66, 0, 0, 7);
    g.fill();
  }
  const spirit = u.ally || (u.companion && !u.undead);
  if (spirit && im) {
    const pu = 0.55 + 0.25 * Math.sin(s.now / 260), gr = g.createRadialGradient(sx, by, 2, sx, by, 44);
    gr.addColorStop(0, `rgba(79,168,255,${pu})`);
    gr.addColorStop(1, 'rgba(79,168,255,0)');
    g.fillStyle = gr;
    g.beginPath();
    g.ellipse(sx, by, 46, 18, 0, 0, 7);
    g.fill();
    im = blue(im);
  }
  if (!im) return by - 60;
  const w = im.bw || im.naturalWidth || im.width, h = im.bh || im.naturalHeight || im.height;
  const k = Math.min(1.9, (u.boss ? 165 : u.elite ? 128 : BIG[u.class] ? 140 : 104) / h, (u.boss ? 185 : u.elite ? 112 : BIG[u.class] ? 146 : 96) / w);
  const flash = u.flash && s.now >= u.flash && s.now < u.flash + 140;
  if (hero) {
    const ac = aura(hero);
    if (hero.stage >= 2) {
      const gr = g.createRadialGradient(sx, by, 2, sx, by, 42);
      gr.addColorStop(0, ac + '99');
      gr.addColorStop(1, ac + '00');
      g.fillStyle = gr;
      g.beginPath();
      g.ellipse(sx, by, 42, 17, 0, 0, 7);
      g.fill();
    }
    g.shadowColor = ac;
    g.shadowBlur = hero.stage >= 2 ? 10 : 6;
  }
  if (!hero && spirit) {
    g.shadowColor = '#4fa8ff';
    g.shadowBlur = 12;
  }
  const sh = hue(u) || (u.enraged ? '#ff2a2a' : null);
  if (sh) {
    g.shadowColor = sh;
    g.shadowBlur = 12;
  }
  const cs = u.statuses || {};
  if (cs.cloak > 0 || cs.blood > 0) { // cloaked: a soft white glow that pulses (owner's pick B, 2026-10-16)
    g.shadowColor = '#ffffff';
    g.shadowBlur = 20 + 8 * Math.sin(s.now / 300);
  }
  if (flash) g.filter = 'brightness(2.6)';
  if (im.bw) g.drawImage(im, sx - (w / 2 + im.px) * k, by + 8 - (h + im.py) * k, im.width * k, im.height * k);
  else g.drawImage(im, sx - (w * k) / 2, by + 8 - h * k, w * k, h * k);
  if (flash) g.filter = 'none';
  g.shadowBlur = 0;
  g.shadowColor = 'transparent';
  if (hero && hero.stage >= 3) {
    g.fillStyle = aura(hero);
    for (let i = 0; i < 6; i++) {
      const a = s.now / 450 + i * 1.047;
      g.fillRect(sx + Math.cos(a) * 30 - 2, by - 36 + Math.sin(a) * 12 - 2, 4, 4);
    }
  }
  return by + 8 - (h + (im.bw ? 12 : 0)) * k;
}

// loopFx draws a looping effect sheet (poison and burn on a unit).
function loopFx(g, name, x, y, size, now, speed = 70) {
  const E = sheets.effects[name], im = pic('effects', name);
  if (!E || !im) return;
  const fr = E.fr[((now / speed) | 0) % E.fr.length], h = (size * E.h) / E.w;
  g.save();
  g.imageSmoothingEnabled = false;
  g.drawImage(im, fr[0], fr[1], E.w, E.h, x - size / 2, y - h / 2, size, h);
  g.restore();
}

function shockArcs(g, cx, cy) {
  g.save();
  g.strokeStyle = '#ffe45c';
  g.lineWidth = 2;
  g.shadowColor = '#fff6a8';
  g.shadowBlur = 6;
  for (let k = 0; k < 3; k++) {
    g.beginPath();
    let a = Math.random() * 6.28, x = cx + Math.cos(a) * 22, y = cy + Math.sin(a) * 30;
    g.moveTo(x, y);
    for (let i = 0; i < 3; i++) {
      x += (Math.random() - 0.5) * 16;
      y += (Math.random() - 0.5) * 16;
      g.lineTo(x, y);
    }
    g.stroke();
  }
  g.restore();
}

const MONO = 'ui-monospace,Menlo,monospace';
function outlined(g, text, x, y, color, size = 12) {
  g.font = `bold ${size}px ${MONO}`;
  g.lineWidth = 3;
  g.strokeStyle = '#000';
  g.strokeText(text, x, y);
  g.fillStyle = color;
  g.fillText(text, x, y);
}

// badges stacks status badges above the unit you are looking at, then what
// the ground under it does (from the referee, owner 2026-10-16), in green.
function badges(g, u, cx, y, k) {
  const list = active(u).map((q) => ({ txt: label(u, q), bad: STATUS[q].bad, n: u.statuses[q] }))
    .concat((u.ground || []).map((e) => ({ txt: e.label, bad: !e.good, ground: true })));
  if (!list.length) return;
  const fs = Math.max(11, 10 / k), h = fs * 1.35;
  g.textAlign = 'center';
  let yy = y;
  list.slice(0, 5).forEach((q) => {
    g.font = `bold ${fs}px ${MONO}`;
    const w = g.measureText(q.txt).width + fs * 0.9;
    yy -= h + fs * 0.45;
    g.fillStyle = q.ground ? 'rgba(4,34,18,.9)' : q.bad ? 'rgba(50,0,6,.9)' : 'rgba(3,18,58,.9)';
    g.fillRect(cx - w / 2, yy, w, h);
    g.strokeStyle = q.ground ? '#5fd38a' : q.bad ? '#ff5d5d' : '#6ab8ff';
    g.lineWidth = 1.5;
    g.strokeRect(cx - w / 2, yy, w, h);
    g.fillStyle = q.ground ? '#9af0b8' : q.bad ? '#ff6a6a' : '#8fc8ff';
    g.fillText(q.txt, cx, yy + h * 0.76);
    if (q.n == null) return;
    const r = fs * 0.62, bx = cx + w / 2;
    g.fillStyle = '#fff';
    g.beginPath();
    g.arc(bx, yy, r, 0, 7);
    g.fill();
    g.fillStyle = '#000';
    g.font = `bold ${fs * 0.78}px ${MONO}`;
    g.fillText(q.n, bx, yy + r * 0.42);
  });
  if (list.length > 5) {
    g.font = `bold ${fs}px ${MONO}`;
    g.fillStyle = '#fff';
    g.fillText('+' + (list.length - 5), cx, yy - 4);
  }
}

// draw draws every unit, back to front. s carries: now, k (zoom), dpr,
// sel/look/aim (the selected, inspected and aimed-at units), shownHp(u) (the
// health bar's value while it slides), anim(u) (an offset while it moves or
// strikes), calm (fewer effects), low (low graphics).
export function draw(g, units, s) {
  const list = units
    .map((u) => {
      const an = s.anim(u), f = feet(u.x, u.y);
      return { u, an, sx: f.x, by: f.y, dep: f.y + an.y };
    })
    .sort((a, b) => a.dep - b.dep);
  for (const { u, an, sx, by } of list) {
    g.save();
    g.translate(an.x, an.y);
    const stt = u.statuses || {};
    const spent = u.done && (u.team === 0 || u.pvp);
    g.globalAlpha = (stt.cloak > 0 || stt.blood > 0 || stt.sky > 0 ? 0.45 : u.decoy ? 0.7 : 1) * (spent ? 0.8 : 1) * (an.a == null ? 1 : an.a);
    if (spent) g.filter = 'grayscale(1) brightness(0.55)';
    if (s.k * (map.flat ? TS : 96) < 15) {
      marker(g, u, sx, by, s.sel);
      g.restore();
      continue;
    }
    g.lineWidth = 3;
    g.strokeStyle = spent ? '#6a655c' : u === s.sel ? '#f2b84b' : u.ally || u.companion ? '#4fa8ff' : u.team ? '#ff5d5d' : '#5df0d0';
    g.beginPath();
    g.ellipse(sx, by, 26 * (u.boss ? 1.4 : 1), 13 * (u.boss ? 1.4 : 1), 0, 0, 7);
    g.stroke();
    g.save();
    g.translate(sx, by);
    g.rotate(an.r || 0);
    g.translate(-sx, -by + (an.b || 0));
    const tp = sprite(g, u, sx, by, s);
    g.restore();
    g.filter = 'none';
    if (!s.low && !s.calm) {
      if (stt.poison > 0) loopFx(g, 'poisonSt', sx, by - 46, TS * 0.95, s.now);
      if (stt.burn > 0) loopFx(g, 'fire4A', sx, by - 18, TS * 0.6, s.now, 55);
    }
    if (stt.shock > 0) shockArcs(g, sx, by - 34);
    healthBar(g, u, sx, tp, s);
    g.restore();
  }
}

function marker(g, u, sx, by, sel) {
  const rr = TS * 0.3;
  g.fillStyle = u.team ? (u.boss ? '#ff9bd8' : u.elite ? '#ffd84a' : '#ff4d4d') : u.companion || u.decoy ? '#7fc4ff' : '#5df0d0';
  g.beginPath();
  g.arc(sx, by - rr, rr, 0, 7);
  g.fill();
  g.lineWidth = 4;
  g.strokeStyle = u === sel ? '#fff' : '#0b1340';
  g.stroke();
  const fr = Math.max(0, Math.min(1, u.hp / u.maxHp));
  g.fillStyle = '#000';
  g.fillRect(sx - rr, by + 4, rr * 2, 9);
  g.fillStyle = u.team ? '#e5484d' : '#4cd97b';
  g.fillRect(sx - rr, by + 4, rr * 2 * fr, 9);
}

function healthBar(g, u, sx, tp, s) {
  const v = s.shownHp(u), hp = Math.max(0, Math.round(v));
  const bw = u.boss ? 76 : u.elite ? 58 : 48, bh = 12, bx = sx - bw / 2, yy = tp - bh - 4;
  const fr = Math.max(0, Math.min(1, v / u.maxHp));
  g.fillStyle = '#000';
  g.fillRect(bx - 2, yy - 2, bw + 4, bh + 4);
  g.fillStyle = '#2a2f55';
  g.fillRect(bx, yy, bw, bh);
  g.fillStyle = u.team ? '#e5484d' : fr > 0.5 ? '#4cd97b' : fr > 0.25 ? '#f0c43f' : '#e5484d';
  g.fillRect(bx, yy, bw * fr, bh);
  g.textAlign = 'center';
  outlined(g, String(hp), sx, yy + 10, '#fff', 11);
  if (u.boss || u.elite) outlined(g, u.class, sx, yy - 6, u.boss ? '#ff9bd8' : '#ffd84a');
  else if ((u.ally || u.companion || u.pet) && u.kind) {
    const lb = u.ally ? '✦ COMPANION SPIRIT' : (u.companion && !u.undead ? '✦ ' : '☠ ') + u.kind;
    outlined(g, lb, sx, yy - 6, u.pet ? '#9ae6b4' : '#7fc4ff');
  }
  if (u.team && u.element) {
    g.fillStyle = ELEMENT[u.element];
    g.beginPath();
    g.arc(bx - 9, yy + 6, 5, 0, 7);
    g.fill();
  }
  g.globalAlpha = 1;
  if (u === s.sel || u === s.look || u === s.aim) badges(g, u, sx, yy - (u.boss || u.elite || u.ally || u.companion || u.pet ? 16 : 3), s.k);
  const tag = s.marks && s.marks.get(u.id);
  if (tag) outlined(g, tag === 'weak' ? 'WEAK' : 'RESIST', sx, yy - 18, tag === 'weak' ? '#f2c14e' : '#9aa6c8', 12);
}

export { pj };
