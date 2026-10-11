// The isometric view: the same Tiny Swords ground drawn on diamond tiles,
// with raised hills and mountains and their cliff sides [old L1468-1489].
// It only draws; the map and rules are the same as the flat view.
import { map, pj, ground, height, inside, hash, CORNERS } from './board.js';
import { pic } from './assets.js';
import { TS, biome } from './look.js';

const land = (x, y) => !inside(x, y) || !'wb'.includes(ground(x, y));
const mountain = (x, y) => inside(x, y) && ground(x, y) === 'm';
const road = (x, y) => inside(x, y) && 'rbT'.includes(ground(x, y));
const bridge = (x, y) => inside(x, y) && ground(x, y) === 'b';
const sheet = () => (map.biome === 'Desert' ? pic('ground', 'DesertP') : pic('ground', map.biome) || pic('ground', 'Grasslands'));

// lift is how high a tile's top is drawn: hills, and mountains (taller).
export const lift = (x, y) => (height(x, y) > 0 ? map.th : ground(x, y) === 'm' ? map.th * 1.8 : 0);

function edge(same, x, y) {
  const l = same(x - 1, y), r = same(x + 1, y), u = same(x, y - 1), d = same(x, y + 1);
  return [!l && !r ? 3 : !l ? 0 : !r ? 2 : 1, !u && !d ? 3 : !u ? 0 : !d ? 2 : 1, d];
}

// quad draws part of a picture stretched onto a tile's diamond.
function quad(g, img, sx, sy, sw, sh, cs, h) {
  if (!img) return;
  const a = (cs[1].x - cs[0].x) / sw, b = (cs[1].y - cs[0].y) / sw, c = (cs[3].x - cs[0].x) / sh, d = (cs[3].y - cs[0].y) / sh;
  g.save();
  g.transform(a, b, c, d, cs[0].x, cs[0].y - h);
  g.imageSmoothingEnabled = false;
  g.drawImage(img, sx, sy, sw, sh, 0, 0, sw + 0.7, sh + 0.7);
  g.restore();
}
const corners = (x, y, r) => [pj(x, y, -r, -r), pj(x, y, r, -r), pj(x, y, r, r), pj(x, y, -r, r)];

export function water(g, list, now, calm) {
  for (const [x, y] of list) quad(g, pic('terrain', 'water'), 0, 0, 64, 64, corners(x, y, 0.5), 0);
  for (const [x, y] of list) {
    if (!land(x, y)) continue;
    if (land(x + 1, y) && land(x - 1, y) && land(x, y + 1) && land(x, y - 1)) continue;
    const f = calm ? 0 : ((now / 130) | 0) + (hash(x, y) % 8);
    quad(g, pic('terrain', 'foam'), (f % 8) * 192, 0, 192, 192, corners(x, y, 1.5), 0);
  }
}

// tile draws a tile: its cliff sides if raised, then its top.
export function tile(g, x, y, p) {
  const h = lift(x, y), cs = CORNERS.map(([a, b]) => pj(x, y, a, b)), B = biome(map.biome);
  if (h) {
    for (let i = 0; i < 4; i++) {
      const a = cs[i], b = cs[(i + 1) % 4];
      g.beginPath();
      g.moveTo(a.x, a.y - h); g.lineTo(b.x, b.y - h); g.lineTo(b.x, b.y); g.lineTo(a.x, a.y);
      g.closePath();
      g.fillStyle = (a.x + b.x) / 2 < p.x ? B.sl : B.sr;
      g.fill();
      const sh = sheet();
      if (sh) {
        g.save();
        g.transform((b.x - a.x) / 64, (b.y - a.y) / 64, 0, h / 64, a.x, a.y - h);
        g.imageSmoothingEnabled = false;
        g.drawImage(sh, 6 * 64, 4 * 64, 64, 64, 0, 0, 64.6, 64.6);
        g.restore();
        if ((a.x + b.x) / 2 >= p.x) {
          g.beginPath();
          g.moveTo(a.x, a.y - h); g.lineTo(b.x, b.y - h); g.lineTo(b.x, b.y); g.lineTo(a.x, a.y);
          g.closePath();
          g.fillStyle = 'rgba(0,0,0,.28)';
          g.fill();
        }
      }
    }
  }
  const t = ground(x, y), put = (img, sx, sy) => quad(g, img, sx, sy, 64, 64, cs, h);
  if (t === 'b') {
    const across = (land(x - 1, y) || bridge(x - 1, y)) && (land(x + 1, y) || bridge(x + 1, y));
    if (across) {
      const L = bridge(x - 1, y), R = bridge(x + 1, y);
      put(pic('terrain', 'bridge'), (L && R ? 1 : L ? 2 : 0) * 64, 0);
    } else {
      const U = bridge(x, y - 1), D = bridge(x, y + 1);
      put(pic('terrain', 'bridge'), 0, (U && D ? 2 : U ? 3 : 1) * 64);
    }
    return;
  }
  if (t === 'w') return;
  if (t === 'm') {
    const [c, r] = edge(mountain, x, y);
    return put(sheet(), (5 + c) * 64, r * 64);
  }
  const desert = map.biome === 'Desert';
  const [c, r] = edge(land, x, y);
  put(desert ? pic('terrain', 'flat') : pic('ground', map.biome) || pic('ground', 'Grasslands'), ((desert ? 5 : 0) + c) * 64, r * 64);
  if (t === 'r') {
    const [c2, r2] = edge(road, x, y);
    put(pic('terrain', 'flat'), (5 + c2) * 64, r2 * 64);
  }
}

// objects lists trees, houses, the castle, rocks and flowers to draw back to front.
export function objects(list, now, calm) {
  const out = [];
  const draw = (img, sx, sy, sw, sh, X, Y, W, H) => (g) => img && g.drawImage(img, sx, sy, sw, sh, X, Y, W, H);
  for (const [x, y, p] of list) {
    const t = ground(x, y), hh = hash(x, y), py = p.y - (height(x, y) > 0 ? map.th : 0);
    if (t === 'f') {
      const f = calm ? hh % 4 : ((now / 160) | 0) + (hh % 4), s = TS * 1.6;
      out.push([p.y, draw(pic('terrain', 'tree'), (f % 4) * 192, 0, 192, 192, p.x - s / 2 + (((hh >> 3) % 9) - 4), py + TS * 0.3 - s * 0.92, s, s)]);
    } else if (t === 't') {
      const w = TS * 1.1, h = w * 1.5;
      out.push([p.y, draw(pic('terrain', 'house'), 0, 0, 128, 192, p.x - w / 2, py + TS * 0.3 - h, w, h)]);
    } else if (t === 'T') {
      const w = TS * 2.3, h = w * 0.8;
      out.push([p.y, draw(pic('terrain', 'castle'), 0, 0, 320, 256, p.x - w / 2, py + TS * 0.35 - h, w, h)]);
    } else if (t === 'h') {
      out.push([p.y, draw(pic('terrain', 'rocks'), (hh % 4) * 64, 0, 64, 64, p.x - TS * 0.35, py - TS * 0.45, TS * 0.7, TS * 0.7)]);
    } else if (t === 'g' && hh % 11 === 0 && land(x, y)) {
      out.push([p.y, draw(pic('terrain', 'deco'), (hh % 18) * 64, 0, 64, 64, p.x - TS * 0.35, py - TS * 0.5, TS * 0.7, TS * 0.7)]);
    }
  }
  return out;
}
