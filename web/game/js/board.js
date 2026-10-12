// The battle map: where each tile sits on screen, and the ground drawn with
// the Tiny Swords tiles (water, foam, coasts, roads, bridges, mountains,
// trees, houses, the castle). Matches the old game's flat view [old L1279-1303].
import { pic } from './assets.js';
import { TS, HW, HH, OY } from './look.js';

// The map's geometry. One shared object, changed by setSize/rotate.
export const map = {
  n: 10, rc: 4.5, flat: true, th: 0, foot: Math.round(TS * 0.3),
  ox: 480, ww: 960, wh: 620,
  ang: 0, target: 0, from: 0, t0: 0, ca: 1, sa: 0,
  ground: [], heights: [], biome: 'Grasslands',
};

export function setSize(n) {
  const m = map;
  m.n = n;
  m.rc = (n - 1) / 2;
  m.th = m.flat ? 0 : 14;
  m.foot = m.flat ? Math.round(TS * 0.3) : 0;
  if (m.flat) {
    m.ww = n * TS + 40;
    m.ox = m.ww / 2;
    m.wh = OY + n * TS + 30;
  } else {
    m.ox = n * HW;
    m.ww = 2 * n * HW;
    m.wh = OY + 2 * n * HH + m.th + 30;
  }
}

// rotate turns the map a quarter turn, smoothly.
export function rotate() {
  turnStep();
  map.from = map.ang;
  map.target += Math.PI / 2;
  map.t0 = performance.now();
}
export const turning = () => map.ang !== map.target;

// turnStep moves the rotation along (ease in-out over half a second).
export function turnStep() {
  const m = map;
  if (m.ang !== m.target) {
    const p = Math.min(1, (performance.now() - m.t0) / 500);
    m.ang = p >= 1 ? m.target : m.from + (m.target - m.from) * (p < 0.5 ? 2 * p * p : 1 - Math.pow(-2 * p + 2, 2) / 2);
  }
  m.ca = Math.round(Math.cos(m.ang) * 1e6) / 1e6;
  m.sa = Math.round(Math.sin(m.ang) * 1e6) / 1e6;
}

// pj is where tile (x, y) (plus an offset in tiles) is on the map picture.
export function pj(x, y, ox = 0, oy = 0) {
  const m = map, u = x - m.rc + ox, v = y - m.rc + oy;
  const rx = u * m.ca - v * m.sa, ry = u * m.sa + v * m.ca;
  if (m.flat) return { x: m.ox + rx * TS, y: OY + (m.rc + ry) * TS + TS / 2, d: ry * 100 + rx };
  return { x: m.ox + (rx - ry) * HW, y: OY + HH + (2 * m.rc + rx + ry) * HH, d: rx + ry };
}

// unproject turns a point on the map picture back into tile coordinates.
export function unproject(X, Y) {
  const m = map;
  let rx, ry;
  if (m.flat) {
    rx = (X - m.ox) / TS;
    ry = (Y - OY - TS / 2) / TS - m.rc;
  } else {
    const a = (X - m.ox) / HW, b = (Y - OY - HH) / HH - 2 * m.rc;
    rx = (a + b) / 2;
    ry = (b - a) / 2;
  }
  return [rx * m.ca + ry * m.sa + m.rc, -rx * m.sa + ry * m.ca + m.rc];
}

// feet is where a unit standing on (x, y) touches the ground.
export function feet(x, y) {
  const p = pj(x, y), h = height(x, y) > 0 ? map.th : 0;
  return { x: p.x, y: p.y - h + map.foot };
}

export const CORNERS = [[-0.5, -0.5], [0.5, -0.5], [0.5, 0.5], [-0.5, 0.5]];
export const inside = (x, y) => x >= 0 && y >= 0 && x < map.n && y < map.n;
export const ground = (x, y) => (map.ground[y] && map.ground[y][x]) || (height(x, y) === -1 ? 'w' : 'g');
export const height = (x, y) => (map.heights[y] ? map.heights[y][x] || 0 : 0);

// tiles lists every tile with its screen spot, back to front. The spots only
// change when the map size, the view or the rotation changes, so they are kept.
let tileCache = null;
export function tiles() {
  if (tileCache && tileCache.n === map.n && tileCache.flat === map.flat && tileCache.ang === map.ang) return tileCache.list;
  const out = [];
  for (let y = 0; y < map.n; y++) for (let x = 0; x < map.n; x++) out.push([x, y, pj(x, y)]);
  out.sort((a, b) => a[2].d - b[2].d);
  if (!turning()) tileCache = { n: map.n, flat: map.flat, ang: map.ang, list: out };
  return out;
}

// ---------- Tiny Swords ground ----------

const land = (x, y) => !inside(x, y) || !'wb'.includes(ground(x, y));
const mountain = (x, y) => inside(x, y) && ground(x, y) === 'm';
const road = (x, y) => inside(x, y) && 'rbT'.includes(ground(x, y));
const bridge = (x, y) => inside(x, y) && ground(x, y) === 'b';
export const hash = (x, y) => ((x * 73856093) ^ (y * 19349663)) >>> 0;

// Screen directions of map right and map down, after rotation.
function dirs() {
  const c = Math.round(map.ca), s = Math.round(map.sa);
  return { R: [c, -s], D: [s, c] };
}

// edge picks a tile-sheet cell from which neighbours are of the same kind:
// column 0 left edge, 1 middle, 2 right edge, 3 alone; the same for rows.
function edge(same, x, y) {
  const { R, D } = dirs();
  const l = same(x - R[0], y - R[1]), r = same(x + R[0], y + R[1]);
  const u = same(x - D[0], y - D[1]), d = same(x + D[0], y + D[1]);
  return [!l && !r ? 3 : !l ? 0 : !r ? 2 : 1, !u && !d ? 3 : !u ? 0 : !d ? 2 : 1, d];
}

function put(g, img, sx, sy, sw, sh, X, Y, W, H) {
  if (img) g.drawImage(img, sx, sy, sw, sh, X, Y, W + 0.6, H + 0.6);
}

function groundSheet() {
  return map.biome === 'Desert' ? [pic('terrain', 'flat'), 5] : [pic('ground', map.biome) || pic('ground', 'Grasslands'), 0];
}

// water fills every tile with water, then foam around the coasts.
export function drawWater(g, list, now, calm) {
  g.save();
  g.imageSmoothingEnabled = false;
  const water = pic('terrain', 'water'), foam = pic('terrain', 'foam');
  for (const [, , p] of list) put(g, water, 0, 0, 64, 64, p.x - TS / 2, p.y - TS / 2, TS, TS);
  for (const [x, y, p] of list) {
    if (!land(x, y)) continue;
    if (land(x + 1, y) && land(x - 1, y) && land(x, y + 1) && land(x, y - 1)) continue;
    const f = calm ? 0 : ((now / 130) | 0) + (hash(x, y) % 8);
    put(g, foam, (f % 8) * 192, 0, 192, 192, p.x - TS * 1.5, p.y - TS * 1.5, TS * 3, TS * 3);
  }
  g.restore();
}

// drawTile draws one tile's ground (grass with its coast edges, road, bridge).
export function drawTile(g, x, y, p) {
  const t = ground(x, y), X = p.x - TS / 2, Y = p.y - TS / 2;
  g.save();
  g.imageSmoothingEnabled = false;
  if (t === 'b') {
    const { R, D } = dirs(), br = pic('terrain', 'bridge');
    const across = land(x - R[0], y - R[1]) || bridge(x - R[0], y - R[1]);
    if (across && (land(x + R[0], y + R[1]) || bridge(x + R[0], y + R[1]))) {
      const L = bridge(x - R[0], y - R[1]), Rr = bridge(x + R[0], y + R[1]);
      put(g, br, (L && Rr ? 1 : L ? 2 : 0) * 64, 0, 64, 64, X, Y, TS, TS);
    } else {
      const U = bridge(x - D[0], y - D[1]), Dn = bridge(x + D[0], y + D[1]);
      put(g, br, 0, (U && Dn ? 2 : U ? 3 : 1) * 64, 64, 64, X, Y, TS, TS);
    }
  } else if (t !== 'w') {
    const [sheet, col] = groundSheet(), [c, r] = edge(land, x, y);
    put(g, sheet, (col + c) * 64, r * 64, 64, 64, X, Y, TS, TS);
    if (t === 'r') {
      const [c2, r2] = edge(road, x, y);
      put(g, pic('terrain', 'flat'), (5 + c2) * 64, r2 * 64, 64, 64, X, Y, TS, TS);
      if (map.biome === 'Desert') {
        g.fillStyle = 'rgba(110,70,30,.28)';
        g.fillRect(X, Y, TS, TS);
      }
    }
  }
  g.restore();
}

// objects lists what stands on the tiles (mountains, trees, houses, the
// castle, rocks, flowers) with a depth, to draw back to front.
export function objects(list, now, calm, throneGlow) {
  const out = [];
  for (const [x, y, p] of list) {
    const t = ground(x, y), hh = hash(x, y);
    if (t === 'm') {
      const [c, r, below] = edge(mountain, x, y);
      const sheet = map.biome === 'Desert' ? pic('ground', 'DesertP') : pic('ground', map.biome) || pic('ground', 'Grasslands');
      out.push([p.y, (g) => {
        put(g, sheet, (5 + c) * 64, r * 64, 64, 64, p.x - TS / 2, p.y - TS * 0.95, TS, TS);
        if (!below) put(g, sheet, (5 + c) * 64, 4 * 64, 64, 64, p.x - TS / 2, p.y + TS * 0.05, TS, TS * 0.62);
      }]);
    } else if (t === 'f') {
      const f = calm ? hh % 4 : ((now / 160) | 0) + (hh % 4);
      out.push([p.y + TS * 0.4, (g) => {
        const s = TS * 1.75;
        put(g, pic('terrain', 'tree'), (f % 4) * 192, 0, 192, 192, p.x - s / 2 + (((hh >> 3) % 9) - 4), p.y + TS * 0.5 - s * 0.92, s, s);
      }]);
    } else if (t === 't') {
      out.push([p.y + TS * 0.4, (g) => {
        const w = TS * 1.15, h = w * 1.5;
        put(g, pic('terrain', 'house'), 0, 0, 128, 192, p.x - w / 2, p.y + TS * 0.5 - h, w, h);
      }]);
    } else if (t === 'T') {
      out.push([p.y + TS * 0.45, (g) => {
        const w = TS * 2.5, h = w * 0.8;
        put(g, pic('terrain', 'castle'), 0, 0, 320, 256, p.x - w / 2, p.y + TS * 0.55 - h, w, h);
        if (throneGlow) {
          g.strokeStyle = 'rgba(255,215,0,' + (0.5 + 0.4 * Math.sin(now / 250)).toFixed(2) + ')';
          g.lineWidth = 4;
          g.strokeRect(p.x - TS * 0.46, p.y - TS * 0.46, TS * 0.92, TS * 0.92);
        }
      }]);
    } else if (t === 'h') {
      out.push([p.y, (g) => put(g, pic('terrain', 'rocks'), (hh % 4) * 64, 0, 64, 64, p.x - TS * 0.4, p.y - TS * 0.35, TS * 0.8, TS * 0.8)]);
    } else if (t === 'g' && hh % 11 === 0 && land(x, y)) {
      out.push([p.y, (g) => put(g, pic('terrain', 'deco'), (hh % 18) * 64, 0, 64, 64, p.x - TS * 0.4, p.y - TS * 0.4, TS * 0.8, TS * 0.8)]);
    }
  }
  return out;
}

// vignette darkens the screen's edges (the night grade) [old tsVig].
export function vignette(g, w, h) {
  g.save();
  g.setTransform(1, 0, 0, 1, 0, 0);
  const gr = g.createRadialGradient(w / 2, h * 0.45, Math.min(w, h) * 0.22, w / 2, h * 0.45, Math.max(w, h) * 0.72);
  gr.addColorStop(0, 'rgba(0,0,0,0)');
  gr.addColorStop(1, 'rgba(12,0,4,.66)');
  g.fillStyle = gr;
  g.fillRect(0, 0, w, h);
  g.restore();
}

// poly fills a shape of points.
export function poly(g, pts, fill) {
  g.beginPath();
  pts.forEach((p, i) => (i ? g.lineTo(p[0], p[1]) : g.moveTo(p[0], p[1])));
  g.closePath();
  g.fillStyle = fill;
  g.fill();
}

// top is a tile's four corners on screen.
export function top(x, y) {
  const h = height(x, y) > 0 ? map.th : !map.flat && ground(x, y) === 'm' ? map.th * 1.8 : 0;
  return CORNERS.map(([a, b]) => {
    const c = pj(x, y, a, b);
    return [c.x, c.y - h];
  });
}
