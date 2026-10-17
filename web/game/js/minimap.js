// The minimap in the top-right corner: ground colors, units, chests and the
// camera's view box. Tap it to look there [old drawMini, miniTap L1267-1273].
import { map, ground, unproject } from './board.js';
import { cam, closeUp } from './camera.js';
import { MINI, biome } from './look.js';

export function draw(c, st, show) {
  const v = st.view;
  show = show && v && map.n >= 13;
  c.hidden = !show;
  if (!show) return;
  const n = map.n, S = Math.max(3, Math.floor(132 / n)), W = S * n;
  if (c.width !== W * 2) {
    c.width = c.height = W * 2;
    c.style.width = c.style.height = W + 'px';
  }
  const x = c.getContext('2d');
  x.setTransform(2, 0, 0, 2, 0, 0);
  x.clearRect(0, 0, W, W);
  const grass = biome(map.biome).g;
  const spot = (x, y) => (map.flip ? [n - 1 - x, n - 1 - y] : [x, y]);
  for (let y = 0; y < n; y++) {
    for (let xx = 0; xx < n; xx++) {
      const [sx, sy] = spot(xx, y);
      x.fillStyle = MINI[ground(xx, y)] || grass;
      x.fillRect(sx * S, sy * S, S, S);
    }
  }
  (v.fields || []).forEach((f) => {
    const [sx, sy] = spot(f.X, f.Y);
    x.fillStyle = f.Kind === 'banner' ? 'rgba(255,215,90,.35)' : 'rgba(255,110,30,.45)';
    x.fillRect((sx - f.Range) * S, (sy - f.Range) * S, (2 * f.Range + 1) * S, (2 * f.Range + 1) * S);
  });
  st.units.forEach((u) => {
    const [sx, sy] = spot(u.x, u.y);
    x.fillStyle = u.team ? (u.boss ? '#ff9bd8' : u.elite ? '#ffd84a' : '#ff4d4d') : u === st.sel ? '#ffffff' : u.companion || u.decoy ? '#7fc4ff' : '#5df0d0';
    x.beginPath();
    x.arc(sx * S + S / 2, sy * S + S / 2, Math.max(1.6, S * 0.42), 0, 7);
    x.fill();
  });
  (v.chests || []).forEach((ch) => {
    const [sx, sy] = spot(ch.x, ch.y);
    x.fillStyle = '#ffd84a';
    x.fillRect(sx * S + S * 0.3, sy * S + S * 0.3, S * 0.4, S * 0.4);
  });
  const vw = cam.w / cam.k, vh = cam.h / cam.k;
  const corners = [[cam.x - vw / 2, cam.y - vh / 2], [cam.x + vw / 2, cam.y - vh / 2], [cam.x + vw / 2, cam.y + vh / 2], [cam.x - vw / 2, cam.y + vh / 2]].map(([a, b]) => unproject(a, b));
  x.strokeStyle = '#fff';
  x.lineWidth = 1.5;
  x.beginPath();
  corners.forEach(([a, b], i) => {
    const [fx, fy] = map.flip ? [n - 1 - a, n - 1 - b] : [a, b];
    const px = (fx + 0.5) * S, py = (fy + 0.5) * S;
    i ? x.lineTo(px, py) : x.moveTo(px, py);
  });
  x.closePath();
  x.stroke();
  x.strokeStyle = 'rgba(255,255,255,.5)';
  x.lineWidth = 1;
  x.strokeRect(0.5, 0.5, W - 1, W - 1);
}

// tap looks at the tapped spot of the minimap.
export function tap(c, e) {
  const r = c.getBoundingClientRect(), S = r.width / map.n;
  let gx = Math.floor((e.clientX - r.left) / S), gy = Math.floor((e.clientY - r.top) / S);
  if (map.flip) { gx = map.n - 1 - gx; gy = map.n - 1 - gy; }
  if (gx < 0 || gy < 0 || gx >= map.n || gy >= map.n) return;
  closeUp(gx, gy);
  e.stopPropagation();
}
