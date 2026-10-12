// Playing the referee's events as pictures: walking, each attack's pose and
// effect, numbers, hits and deaths. Nothing here decides anything: it keeps
// showing the previous answer while the events play, then the new answer is
// shown. Timings, poses and effects match the old game [old startMove L788,
// anim L789-800, spawn L813, act L1711-1730, hurt L1700-1707, useAbility
// L1194, drawPop L1686]. The effects themselves are in fxbook.js,
// fxpixel.js, fxskill.js and fxsprite.js.
import { feet } from './board.js';
import { book } from './assets.js';
import { lookAt } from './camera.js';
import { SKFX, ACT, POSE_MS, FXN, EE, SHK0, fxKey, durOf, impOf } from './fxbook.js';
import { drawOlder } from './fxpixel.js';
import { drawP } from './fxskill.js';
import { setup } from './fxsprite.js';
import { options } from './options.js';

let ready = false;
const skillOf = (name) => (book.skills || {})[name] || {};

// The weapon kinds: melee, ranged, magic [old WCAT L1109].
const WCAT = { sword: 'm', axe: 'm', lance: 'm', dagger: 'm', bow: 'r', gun: 'r', staff: 'g', wand: 'g' };
// strikeKey: a Strike looks like the weapon in hand [old strikeKey L1112].
function strikeKey(u) {
  const g = u.gear;
  if (!g || !g.sets) return 'Strike_m';
  const i = g.sets.findIndex((s) => s.inHand), it = (g.worn || []).find((w) => w.slot === ['w', 'w2', 'w3'][Math.max(0, i)]);
  const t = it && it.weaponType;
  return t === 'gun' ? 'Strike_gun' : 'Strike_' + ((t && WCAT[t]) || 'm');
}

// look: the effect key, the pose and the size of a skill's effect for unit u
// [old act L1711-1712, fxScale L917].
function look(u, name) {
  const s = skillOf(name), strike = s.strike || name === 'Strike';
  const sk = strike ? strikeKey(u) : name, D9 = SKFX[sk];
  let k;
  if (strike) k = 'P:' + sk;
  else if (u.team && s.element && s.hero && s.class === u.class) k = EE[u.element] || 'fire'; // a monster using a hero class's spell
  else k = fxKey(name) || s.visual || FXN[name] || 'slash';
  const pose = s.charge ? 'dash' : D9 ? ACT[D9[0]] || 'lunge' : 'lunge';
  const S = u.boss ? 1.6 : u.elite ? 1.3 : u.team ? 1 : s.ascension ? 1.8 : s.kind === 'u' ? ((s.level || 0) >= 20 ? 1.7 : 1.35) : (s.level || 0) >= 12 ? 1.35 : (s.level || 0) >= 6 ? 1.15 : 1;
  return { k, pose, S };
}

// ---------- the timeline ----------

let run = null; // { items, units, end, done, clock }
const pops = [];
const fx = [];
export const shake = { amp: 0, t0: 0 };
export const playing = () => !!run;
let skipped = false;
export const skip = () => { skipped = true; };

// play shows events one after another on the previous answer's units, then
// calls done (which shows the new answer). speed: 1, 2 or 3 (instant).
export function play(units, events, speed, done, onEnemyTurn, calm) {
  if (!ready) { setup(book); ready = true; }
  skipped = false;
  const k = () => (skipped ? 40 : speed === 3 ? 12 : speed === 2 ? 2.2 : 1);
  const now = performance.now();
  const byId = new Map(units.map((u) => [u.id, u]));
  const spot = new Map(); // where each unit stands on screen now (it may have walked)
  const at = (id) => spot.get(id) || (byId.get(id) ? [byId.get(id).x, byId.get(id).y] : null);
  const items = [];
  let t = 0, busyUntil = 0, actor = null, cur = null;
  const busy = (x) => { busyUntil = Math.max(busyUntil, x); };
  // A new unit starts acting: wait until the last animation has finished,
  // then move the camera to it, then act (the camera never jumps away
  // mid-animation; owner's request, decision 10).
  const turnTo = (who) => {
    if (!who || who.id === actor) return;
    if (actor !== null) t = Math.max(t, busyUntil);
    if (who.team === 1) {
      items.push({ t, kind: 'cam', at: at(who.id) });
      t += 350;
    }
    actor = who.id;
  };
  // An effect from unit `from` to unit `to`, starting at time T.
  const effect = (T, k, from, to, S, abn) => {
    const d = durOf(k);
    items.push({ t: T, kind: 'fx', k, from: at(from), to: at(to), S, d, abn });
    const D = k.startsWith('P:') ? SKFX[k.slice(2)] : null, amp = SHK0[k] || (D && (D[2] || '').includes('s') ? 2.5 * S : 0);
    if (amp) items.push({ t: T + impOf(k), kind: 'shake', amp });
    busy(T + d);
  };
  const pop = (pt, id, txt, col, big, hurt) => {
    items.push({ t: pt, kind: 'pop', at: at(id), txt, col, big, id, hurt });
    busy(pt + 450);
  };
  for (const e of events) {
    const who = e.unit ? byId.get(e.unit) : null, target = e.target ? byId.get(e.target) : who;
    if (e.kind === 'moved' || e.kind === 'skill' || e.kind === 'ability' || e.kind === 'overwatch') turnTo(who);
    switch (e.kind) {
      case 'moved': case 'undone': {
        cur = null;
        if (!who || !e.path || e.path.length < 2) break;
        const dur = Math.min(1500, 150 * (e.path.length - 1));
        items.push({ t, kind: 'walk', id: who.id, path: e.path, dur });
        busy(t + dur);
        spot.set(who.id, e.path[e.path.length - 1]);
        t += dur;
        break;
      }
      case 'skill': case 'overwatch': case 'counter': {
        if (!who) break;
        t = Math.max(t, busyUntil);
        const tg = target || who, L = look(who, e.skill || 'Strike'), a = feet(...at(who.id)), b = feet(...at(tg.id));
        const dx = b.x - a.x, dy = b.y - a.y, l = Math.hypot(dx, dy);
        items.push({ t, kind: 'pose', id: who.id, k: L.pose, dx: l ? dx / l : 0, dy: l ? dy / l : -0.6, dur: POSE_MS[L.pose] || 420 });
        busy(t + (POSE_MS[L.pose] || 420));
        effect(t, L.k, who.id, tg.id, L.S);
        cur = { T: t, im: impOf(L.k), extra: 0, n: new Map(), last: t + impOf(L.k) };
        break;
      }
      case 'ability': {
        if (!who) break;
        t = Math.max(t, busyUntil);
        const k = 'P:AB_' + who.class;
        if (SKFX[k.slice(2)]) effect(t, k, who.id, who.id, 1, e.skill);
        cur = { T: t, im: impOf(k), extra: 0, n: new Map(), last: t + impOf(k) };
        pop(cur.last, who.id, '✦ ' + (e.skill || ''), '#ffd84a', true);
        break;
      }
      case 'riposte': // the Duelist strikes back after the hit lands [old hurt L1707]
        if (cur) cur.extra = 320;
        break;
      case 'hit': case 'missed': case 'healed': case 'revived': {
        const v = target || who;
        if (!v) break;
        let pt = t;
        if (cur) { const c = cur.n.get(v.id) || 0; cur.n.set(v.id, c + 1); pt = cur.T + cur.im + cur.extra + 130 * c; cur.last = Math.max(cur.last, pt); }
        else t += 120;
        const n = e.amount;
        const txt = e.kind === 'missed' ? 'MISS' : e.kind === 'hit' ? (e.crit ? 'CRIT ' + n : n + (e.mult > 1 ? '!' : '')) : '+' + n;
        const col = e.kind === 'missed' ? '#c9c9d6' : e.kind === 'hit' ? (e.crit ? '#ff9a2e' : e.mult > 1 ? '#ffd84a' : e.mult < 1 ? '#c9c9d6' : '#fff') : '#6ee7a0';
        pop(pt, v.id, txt, col, e.crit || e.mult > 1, e.kind === 'hit');
        break;
      }
      case 'potion': { // a potion heals with the heal sparkle [old usePotion L1806]
        if (!who) break;
        t = Math.max(t, busyUntil);
        effect(t, 'heal', who.id, who.id, 1);
        pop(t + impOf('heal'), who.id, '+' + e.amount, '#6ee7a0', false);
        t += 120;
        break;
      }
      case 'ticked': {
        if (!who) break;
        pop(t, who.id, '' + e.amount, '#ff9a9a', false, true);
        t += 120;
        break;
      }
      case 'died': case 'faded': case 'exploded': {
        if (!who) break;
        const ft = (cur ? cur.last : t) + 250; // [old hurt L1704: fades 250 ms after the hit]
        items.push({ t: ft, kind: 'fade', id: who.id, dur: 450 });
        busy(ft + 450);
        if (!cur) t += 200;
        break;
      }
      case 'turn-start':
        cur = null;
        if (e.side === 1) { t = Math.max(t, busyUntil); items.push({ t, kind: 'enemy' }); t += 300; }
        break;
      default:
    }
  }
  if (!items.length) return done();
  run = { k, items, units: byId, end: Math.max(t, busyUntil), done, onEnemyTurn, last: now, clock: 0, calm };
}

// step moves the timeline on; the scene calls it every frame.
function step(now) {
  if (!run) return;
  run.clock += (now - run.last) * run.k();
  run.last = now;
  const c = run.clock;
  for (const i of run.items) {
    if (i.started || c < i.t) continue;
    i.started = now;
    if (i.kind === 'pop' && i.at) {
      const f = feet(...i.at);
      pops.push({ x: f.x, y: f.y - 72, t0: now, txt: i.txt, col: i.col, big: i.big });
    }
    if (i.kind === 'fx' && i.from && i.to) {
      const a = feet(...i.from), b = feet(...i.to);
      fx.push({ k: i.k, sx: a.x, sy: a.y - 28, tx: b.x, ty: b.y - 30, S: i.S, d: i.d, abn: i.abn, t0: now, real: i.d / Math.min(run.k(), 6), calm: run.calm });
    }
    if (i.kind === 'shake') { shake.amp = i.amp; shake.t0 = now; }
    if (i.kind === 'cam' && i.at) {
      const f = feet(...i.at);
      lookAt(f.x, f.y - 40);
    }
    if (i.kind === 'enemy' && run.onEnemyTurn) run.onEnemyTurn();
  }
  if (c >= run.end + 450) {
    const done = run.done;
    run = null;
    done();
  }
}

// anim is how far a unit is drawn from its tile right now, in map pixels
// (x, y), plus a bounce (b), a tilt (r) and see-through (a) [old anim L789-800].
export function anim(u) {
  const out = { x: 0, y: 0, b: 0, r: 0, a: 1 };
  if (!run) return out;
  const c = run.clock, now = performance.now(), base = feet(u.x, u.y);
  let place = null;
  for (const i of run.items) {
    if (i.id !== u.id || i.kind !== 'walk') continue;
    if (c >= i.t + i.dur) place = i.path[i.path.length - 1];
    else if (c >= i.t) {
      const q = ((c - i.t) / i.dur) * (i.path.length - 1), n = Math.min(i.path.length - 2, Math.floor(q)), f = q - n;
      const A = feet(...i.path[n]), B = feet(...i.path[n + 1]);
      out.x = A.x + (B.x - A.x) * f - base.x;
      out.y = A.y + (B.y - A.y) * f - base.y;
      out.b = -Math.abs(Math.sin(f * Math.PI)) * 7;
      out.r = Math.sin(now * 0.03) * 0.07;
      return out;
    }
  }
  if (place) {
    const p = feet(...place);
    out.x = p.x - base.x;
    out.y = p.y - base.y;
  }
  for (const i of run.items) {
    if (i.id !== u.id || !i.started) continue;
    if (i.kind === 'pose') {
      const p = (c - i.t) / i.dur;
      if (p < 0 || p >= 1) continue;
      if (i.k === 'dash') { const l = Math.sin(Math.min(1, p * 1.1) * Math.PI) * 60; out.x += i.dx * l; out.y += i.dy * l; out.b -= Math.abs(Math.sin(p * 16)) * 4; }
      else if (i.k === 'recoil') { const l = p < 0.15 ? (-p / 0.15) * 9 : -(1 - (p - 0.15) / 0.85) * 9; out.x += i.dx * l; out.y += i.dy * l; out.r += p < 0.25 ? -0.07 : 0; }
      else if (i.k === 'cast') out.b -= Math.sin(p * Math.PI) * 9;
      else if (i.k === 'throw') { out.b -= Math.sin(Math.min(1, p * 2) * Math.PI) * 11; out.r += p < 0.4 ? -0.16 : 0; }
      else if (i.k === 'rear') { out.r += -0.2 * Math.sin(Math.min(1, p * 1.4) * Math.PI); out.b -= Math.sin(p * Math.PI) * 5; }
      else if (i.k === 'blink') out.a = p < 0.3 ? 1 - p / 0.3 : p < 0.6 ? 0 : (p - 0.6) / 0.4;
      else { const l = Math.sin(Math.min(1, p * 1.2) * Math.PI) * 16; out.x += i.dx * l + Math.sin(now * 0.09) * 3 * (1 - p); out.y += i.dy * l; out.r += Math.sin(now * 0.09) * 0.1 * (1 - p); }
    } else if (i.kind === 'pop' && i.hurt) {
      const p = (c - i.t) / 320;
      if (p >= 0 && p < 1) out.x += Math.sin(now * 0.1) * 4 * (1 - p);
    } else if (i.kind === 'fade') {
      const p = (c - i.t) / i.dur;
      if (p >= 1) out.a = 0;
      else if (p >= 0) out.a = 1 - p;
    }
  }
  return out;
}

// draw draws the effects and floating numbers; returns true while busy.
export function draw(g, now) {
  step(now);
  for (let i = fx.length; i--; ) if (now > fx[i].t0 + fx[i].real) fx.splice(i, 1);
  for (const f of fx) {
    const p = (now - f.t0) / f.real;
    if (f.k.startsWith('P:')) drawP(g, f, p, now, f.calm);
    else drawOlder(g, f, p, now);
  }
  for (let i = pops.length; i--; ) if (now > pops[i].t0 + 1000) pops.splice(i, 1);
  for (const o of pops) {
    // The damage-numbers option hides numbers, not words like MISS [old L1661].
    if (!options.damageNumbers && /^(CRIT )?[+−-]?\d/.test(o.txt)) continue;
    const p = (now - o.t0) / 1000;
    g.save();
    g.globalAlpha = p > 0.7 ? (1 - p) / 0.3 : 1;
    g.font = `bold ${o.big ? 32 : 26}px ui-monospace,Menlo,monospace`;
    g.textAlign = 'center';
    g.lineWidth = 6;
    g.strokeStyle = '#1a0f2a';
    g.lineJoin = 'round';
    const y = o.y - p * 30 + (p < 0.15 ? (1 - p / 0.15) * 10 : 0);
    g.strokeText(o.txt, o.x, y);
    g.fillStyle = o.col;
    g.fillText(o.txt, o.x, y);
    g.restore();
  }
  return !!run || fx.length > 0 || pops.length > 0;
}
