// The camera: zoom, panning, pinch, the whole-map overview, and turning a
// finger's spot into a tile or a unit. Matches the old game's camera
// [old layout L1832-1835, v0.23 bigger tiles L1976-1983, v0.24 overview
// L2087-2095, pan and pinch L1773-1780].
import { map, pj, feet, inside, height, CORNERS } from './board.js';
import { TS } from './look.js';

export const cam = {
  x: 0, y: 0, tx: 0, ty: 0, // where the camera looks (now and target)
  z: 1, zt: 1, // zoom (now and target), 1 = the whole map fits
  k0: 1, k: 1, // scale for the whole map, and the scale now
  offX: 0, offY: 0, // screen spot of the map picture's corner
  overview: false, before: null, lock: false,
  dpr: 1, w: 390, h: 844, touch: 0,
};
export const TOUCH = matchMedia('(pointer:coarse)').matches || innerWidth < 700;
const big = () => (map.flat ? 0.84 : 0.62); // smallest comfortable tile on a phone
export const zoomFloor = () => Math.max(1, Math.min(3.4, big() / cam.k0));

let kick = () => {};
export const onMove = (f) => (kick = f);

// safeTop is the space the phone keeps at the top (the notch).
let safe = null;
export function safeTop() {
  if (!safe) {
    safe = document.createElement('div');
    safe.style.cssText = 'position:fixed;top:0;left:0;width:1px;height:env(safe-area-inset-top,0px);visibility:hidden;pointer-events:none';
    document.body.append(safe);
  }
  return safe.offsetHeight || 0;
}

// frame points the camera at the heroes, at a finger-friendly zoom. On a
// phone it starts 30% closer than the old game (owner's pick B, 2026-10-16).
export function frame(heroes, chosen) {
  cam.k0 = Math.min(cam.w / map.ww, Math.max(200, cam.h) / map.wh);
  cam.zt = cam.z = Math.max(1, Math.min(4, (TOUCH ? 0.8 : 0.7) / cam.k0));
  const h = chosen || heroes[Math.min(heroes.length - 1, (heroes.length / 2) | 0)];
  if (!h) {
    cam.x = cam.tx = map.ww / 2;
    cam.y = cam.ty = map.wh / 2;
    return;
  }
  const b = feet(h.x, h.y);
  cam.x = cam.tx = b.x;
  cam.y = cam.ty = b.y - 120;
  // Start with the map filling the screen: no empty band past its edge.
  const vh = cam.h / (cam.k0 * cam.z);
  if (vh < map.wh) cam.y = cam.ty = Math.max(vh / 2, Math.min(map.wh - vh / 2, cam.y));
  // As in the old game, the first look is worked out before the
  // finger-friendly zoom applies, which settles the camera's height.
  layout(heroes, chosen, true);
}

// layout eases the camera toward its target and works out the scale. It runs
// every frame.
export function layout(heroes = [], chosen = null, first = false) {
  const W = cam.w, H = cam.h;
  cam.k0 = Math.min(W / map.ww, H / map.wh);
  cam.z += (cam.zt - cam.z) * 0.25;
  if (Math.abs(cam.zt - cam.z) < 0.003) cam.z = cam.zt;
  if (cam.zt <= 1 && cam.z < 1.02) {
    cam.tx = map.ww / 2;
    cam.ty = map.wh / 2;
  }
  cam.x += (cam.tx - cam.x) * 0.25;
  cam.y += (cam.ty - cam.y) * 0.25;
  cam.k = cam.k0 * cam.z;
  const vw = W / cam.k, vh = H / cam.k;
  let cx = cam.x, cy = cam.y;
  const mx = vw * 0.42, my = vh * 0.42;
  cx = vw < map.ww ? Math.max(vw / 2 - mx, Math.min(map.ww - vw / 2 + mx, cx)) : map.ww / 2;
  cy = vh < map.wh ? Math.max(vh / 2 - my, Math.min(map.wh - vh / 2 + my, cy)) : map.wh / 2;
  if (Math.abs(cx - cam.x) > 1) cam.x = cam.tx = cx;
  if (Math.abs(cy - cam.y) > 1) cam.y = cam.ty = cy;
  cam.offX = W / 2 - cx * cam.k;
  cam.offY = H / 2 - cy * cam.k;
}


export const moving = () => Math.abs(cam.tx - cam.x) > 0.5 || Math.abs(cam.ty - cam.y) > 0.5 || Math.abs(cam.zt - cam.z) > 0.003;
export const gesturing = () => performance.now() - cam.touch < 220;

// fill eases the camera so the map covers the screen, with no empty band past
// its edge (after placing heroes, whose view leaves room for the buttons).
export function fill() {
  const k = cam.k0 * cam.zt, vw = cam.w / k, vh = cam.h / k;
  const top = (safeTop() + 70) / k + TS * 2.2, bottom = 110 / k; // room for the top bar (and the back row's heads) and End Turn
  if (vh < map.wh) cam.ty = Math.max(vh / 2 - top, Math.min(map.wh - vh / 2 + bottom, cam.ty));
  if (vw < map.ww) cam.tx = Math.max(vw / 2, Math.min(map.ww - vw / 2, cam.tx));
  kick();
}

// toEdge puts the camera at the top or bottom edge of the map, leaving room
// for the top bar (and the back row's heads) or End Turn: where each versus
// player's heroes start.
export function toEdge(atTop) {
  const k = cam.k0 * cam.zt, vh = cam.h / k;
  if (vh >= map.wh) return;
  cam.y = cam.ty = atTop ? vh / 2 - ((safeTop() + 70) / k + TS * 2.2) : map.wh - vh / 2 + 110 / k;
  kick();
}

// lookAt eases the camera to a map spot.
export function lookAt(x, y) {
  cam.tx = x;
  cam.ty = y;
  kick();
}

// focusOn centres a map spot in the space between the top and barTop.
export function focusOn(fx, fy, barTop) {
  const t0 = safeTop() + 70, sc = (t0 + barTop) / 2, y0 = cam.offY + cam.y * cam.k;
  lookAt(fx, fy + (y0 - sc) / cam.k);
}

// closeUp zooms in to at least a comfortable size (for the minimap and
// next hero) and looks at a tile.
export function closeUp(x, y) {
  if (cam.zt < 1.3) cam.zt = Math.min(3.4, Math.max(1.3, (TOUCH ? 0.86 : 0.75) / cam.k0));
  const b = feet(x, y);
  lookAt(b.x, b.y - 20);
}

function zoomRaw(cx, cy, z) {
  z = Math.max(1, Math.min(3.4, z));
  const wx = (cx - cam.offX) / cam.k, wy = (cy - cam.offY) / cam.k, nk = cam.k0 * z, base = cam.offY + cam.y * cam.k;
  cam.x = cam.tx = (cam.w / 2 - cx) / nk + wx;
  cam.y = cam.ty = (base - cy) / nk + wy;
  cam.z = cam.zt = z;
  cam.touch = performance.now();
  kick();
}

// zoomAt zooms around a screen spot, smoothly, from the whole map up to the closest view.
export function zoomAt(cx, cy, z) {
  zoomRaw(cx, cy, z);
}

// zoomStep is the − and + buttons: fixed zoom steps; on a phone, − at the
// smallest comfortable size shows the whole map [old zin/zout L1770-1771,
// overview L2093].
const STEPS = [1, 1.6, 2.4, 3.4];
export function zoomStep(dir) {
  if (dir > 0) {
    const z = STEPS.find((s) => s > cam.zt + 0.01);
    if (z) cam.zt = z;
  } else {
    const z = STEPS.filter((s) => s < cam.zt - 0.01).pop();
    cam.zt = z || 1;
  }
  cam.overview = false;
  kick();
}

// overview shows the whole map (on) or zooms back in around a map spot.
export function overview(on, at, told = () => {}) {
  if (on) {
    if (cam.overview) return;
    cam.overview = true;
    cam.before = cam.zt;
    cam.zt = 1;
    cam.tx = map.ww / 2;
    cam.ty = map.wh / 2;
    told(true);
  } else {
    cam.overview = false;
    const zm = zoomFloor();
    cam.zt = cam.before && cam.before > zm ? cam.before : zm;
    if (at) {
      cam.tx = at[0];
      cam.ty = at[1] - 20;
    }
    told(false);
  }
  kick();
}

// toMap turns a screen spot into a spot on the map picture.
export const toMap = (cx, cy) => [(cx - cam.offX) / cam.k, (cy - cam.offY) / cam.k];

// tileAt is the tile under a screen spot (or null).
export function tileAt(cx, cy) {
  const [px, py] = toMap(cx, cy);
  const list = [];
  for (let y = 0; y < map.n; y++) for (let x = 0; x < map.n; x++) list.push([x, y, pj(x, y).d]);
  list.sort((a, b) => b[2] - a[2]);
  for (const [x, y] of list) {
    const h = height(x, y) > 0 ? map.th : 0;
    const cs = CORNERS.map(([ox, oy]) => {
      const q = pj(x, y, ox, oy);
      return [q.x, q.y - h];
    });
    let pos = 0, neg = 0;
    for (let i = 0; i < 4; i++) {
      const a = cs[i], b = cs[(i + 1) % 4], c = (b[0] - a[0]) * (py - a[1]) - (b[1] - a[1]) * (px - a[0]);
      c >= 0 ? pos++ : neg++;
    }
    if (!pos || !neg) return inside(x, y) ? [x, y] : null;
  }
  return null;
}

// unitAt is the unit whose picture is under a screen spot (or null).
export function unitAt(units, cx, cy) {
  const [px, py] = toMap(cx, cy);
  const t = tileAt(cx, cy);
  if (t) {
    const on = units.find((u) => u.x === t[0] && u.y === t[1]);
    if (on) return on;
  }
  const list = units.map((u) => [u, pj(u.x, u.y)]).sort((a, b) => b[1].d - a[1].d);
  const hit = list.find(([u, p]) => {
    const by = p.y - (height(u.x, u.y) > 0 ? map.th : 0) + map.foot;
    const big = u.class === 'Witch Knight' || u.class === 'Dragon Knight';
    return Math.abs(px - p.x) <= (u.boss ? 66 : u.elite ? 48 : 40) && py >= by - (u.boss ? 165 : u.elite ? 128 : big ? 136 : 98) && py <= by + 24;
  });
  return hit ? hit[0] : null;
}

// ---------- fingers, mouse and keys ----------

// controls wires panning, pinching, the mouse wheel and the arrow keys to the
// canvas. keys gets the other keys (Escape, N for next hero).
export function controls(canvas, keys, onOverview) {
  const fingers = new Map();
  let pinch = null, drag = null;
  cam.dragged = false;
  canvas.addEventListener('pointerdown', (e) => {
    fingers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    drag = { x: e.clientX, y: e.clientY, cx: cam.x, cy: cam.y, moved: false };
    if (fingers.size === 2) {
      const [a, b] = [...fingers.values()];
      pinch = { d: Math.max(20, Math.hypot(a.x - b.x, a.y - b.y)), z: cam.zt };
      drag = null;
    }
  });
  addEventListener('pointermove', (e) => {
    if (fingers.has(e.pointerId)) fingers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (pinch && fingers.size === 2) {
      const [a, b] = [...fingers.values()], d = Math.hypot(a.x - b.x, a.y - b.y);
      zoomAt((a.x + b.x) / 2, (a.y + b.y) / 2, (pinch.z * d) / pinch.d, onOverview);
      drag = null;
      cam.dragged = true;
      return;
    }
    if (!drag || cam.held) return;
    const dx = e.clientX - drag.x, dy = e.clientY - drag.y;
    if (!drag.moved && Math.hypot(dx, dy) > 10) drag.moved = true;
    if (drag.moved) {
      cam.x = cam.tx = drag.cx - dx / cam.k;
      cam.y = cam.ty = drag.cy - dy / cam.k;
      cam.touch = performance.now();
      kick();
    }
  });
  const up = (e) => {
    fingers.delete(e.pointerId);
    if (fingers.size < 2) pinch = null;
    if (drag && drag.moved) cam.dragged = true;
    drag = null;
    cam.lock = false;
  };
  addEventListener('pointerup', up);
  addEventListener('pointercancel', up);
  canvas.addEventListener('wheel', (e) => {
    e.preventDefault();
    zoomAt(e.clientX, e.clientY, cam.zt * Math.exp(-e.deltaY * 0.0015), onOverview);
  }, { passive: false });
  addEventListener('keydown', (e) => {
    if (e.target && /INPUT|SELECT|TEXTAREA/.test(e.target.tagName)) return;
    const st = 70 / cam.k, k = e.key;
    let used = true;
    if (k === 'ArrowLeft' || k === 'a' || k === 'A') cam.tx -= st;
    else if (k === 'ArrowRight' || k === 'd' || k === 'D') cam.tx += st;
    else if (k === 'ArrowUp' || k === 'w' || k === 'W') cam.ty -= st;
    else if (k === 'ArrowDown' || k === 's' || k === 'S') cam.ty += st;
    else if (k === '+' || k === '=') zoomAt(cam.w / 2, cam.h / 2, cam.zt * 1.25, onOverview);
    else if (k === '-' || k === '_') zoomAt(cam.w / 2, cam.h / 2, cam.zt / 1.25, onOverview);
    else used = keys(k);
    if (used) {
      e.preventDefault();
      kick();
    }
  });
}

// resize matches the canvas to the screen.
export function resize(canvas) {
  cam.dpr = Math.min(devicePixelRatio || 1, 2);
  cam.w = innerWidth;
  cam.h = innerHeight;
  canvas.width = cam.w * cam.dpr;
  canvas.height = cam.h * cam.dpr;
  canvas.style.width = cam.w + 'px';
  canvas.style.height = cam.h + 'px';
}

export { TS };
