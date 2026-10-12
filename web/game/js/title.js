// The title screen's life: two candle flames that flicker and glow, now and
// then a lightning flash with a bolt, and mist drifting along the bottom.
// Same look as the old game [old v0.24 title effects, L2104-2146]. Only
// pictures; it stops drawing while the title is hidden.
import { options } from './options.js';

export function start() {
  const T = document.getElementById('title'), fx = document.getElementById('titleFx');
  if (!T || !fx) return;
  const x = fx.getContext('2d');
  const IW = 1008, IH = 1792, FL = [[181, 96], [831, 98]]; // the title picture's size and its two candle flames
  const calm = () => matchMedia('(prefers-reduced-motion: reduce)').matches || options.calm;
  let W = 0, H = 0, S = 1, OX = 0, dpr = 1;
  const size = () => {
    dpr = Math.min(2, devicePixelRatio || 1);
    W = T.clientWidth || innerWidth;
    H = T.clientHeight || innerHeight;
    fx.width = W * dpr;
    fx.height = H * dpr;
    S = Math.max(W / IW, H / IH);
    OX = (W - IW * S) / 2;
  };
  addEventListener('resize', size);
  size();
  const P = (ix, iy) => [OX + ix * S, iy * S];

  // candle flicker
  const fl = [0, 0].map(() => ({ v: 1, t: 1, seed: Math.random() * 99 }));
  const flick = (c, t) => {
    if (Math.random() < 0.08) c.t = 0.55 + Math.random() * 0.45;
    c.v += (c.t - c.v) * 0.18;
    return c.v * (0.92 + 0.08 * Math.sin(t * 0.011 + c.seed) * Math.sin(t * 0.0067 + c.seed * 3));
  };

  // lightning
  let nextL = performance.now() + 2500, L = null;
  const bolt = (sx, sy, ex, ey, depth) => {
    const pts = [[sx, sy]], n = 10;
    for (let i = 1; i <= n; i++) {
      const q = i / n;
      pts.push([sx + (ex - sx) * q + (i < n ? (Math.random() - 0.5) * 60 * S : 0), sy + (ey - sy) * q + (i < n ? (Math.random() - 0.5) * 16 * S : 0)]);
    }
    const br = [];
    if (depth < 1) for (let i = 2; i < n - 2; i++) if (Math.random() < 0.3) { const [p0, p1] = pts[i]; br.push(bolt(p0, p1, p0 + (Math.random() - 0.5) * 160 * S, p1 + (80 + Math.random() * 120) * S, depth + 1)); }
    return { pts, br };
  };
  const strike = (t) => {
    const v0 = Math.max(0, -OX / S) + 12, v1 = Math.min(IW, (W - OX) / S) - 12, left = Math.random() < 0.55;
    const ix = left ? v0 + Math.random() * Math.max(20, 250 - v0) : Math.max(840, v1 - 120) + Math.random() * Math.max(20, v1 - Math.max(840, v1 - 120)), iy = 640 + Math.random() * 80;
    const [sx, sy] = P(ix, iy), [ex, ey] = P(ix + (Math.random() - 0.5) * 120, iy + 220 + Math.random() * 140);
    L = { t0: t, b: bolt(sx, sy, ex, ey, 0), twice: Math.random() < 0.5 };
    nextL = t + 4500 + Math.random() * 6000;
  };
  const drawBolt = (B, w, a) => {
    x.beginPath();
    B.pts.forEach((p, i) => (i ? x.lineTo(p[0], p[1]) : x.moveTo(p[0], p[1])));
    x.strokeStyle = 'rgba(210,205,255,' + a + ')';
    x.lineWidth = w;
    x.stroke();
    B.br.forEach((c) => drawBolt(c, w * 0.55, a * 0.8));
  };
  const flashA = (dt, tw) => {
    const k = [[0, 0], [40, 0.55], [110, 0.08], [170, tw ? 0.42 : 0], [260, 0.05], [520, 0]];
    for (let i = 1; i < k.length; i++) if (dt < k[i][0]) { const [a0, v0] = k[i - 1], [a1, v1] = k[i]; return v0 + ((v1 - v0) * (dt - a0)) / (a1 - a0); }
    return 0;
  };

  const frame = (t) => {
    requestAnimationFrame(frame);
    if (T.hidden) return;
    if (fx.width !== Math.round((T.clientWidth || innerWidth) * Math.min(2, devicePixelRatio || 1))) size();
    x.setTransform(dpr, 0, 0, dpr, 0, 0);
    x.clearRect(0, 0, W, H);
    const quiet = calm();
    // mist drifting along the bottom
    x.globalCompositeOperation = 'source-over';
    for (let i = 0; i < 4; i++) {
      const yy = H * (0.8 + i * 0.045), xx = ((((t * 0.008 * (i % 2 ? 1 : -1) * (1 + i * 0.3)) % (W * 2)) + W * 2) % (W * 2)) - W * 0.5;
      const gr = x.createRadialGradient(xx, yy, 0, xx, yy, W * 0.7);
      gr.addColorStop(0, 'rgba(120,100,150,.09)');
      gr.addColorStop(1, 'rgba(120,100,150,0)');
      x.fillStyle = gr;
      x.fillRect(0, yy - W * 0.7, W, W * 1.4);
    }
    // lightning flash and bolt
    if (!quiet && !L && t > nextL) strike(t);
    if (L) {
      const dt = t - L.t0, a = flashA(dt, L.twice);
      if (dt > 520) L = null;
      else {
        x.globalCompositeOperation = 'lighter';
        const gr = x.createLinearGradient(0, 0, 0, H * 0.8);
        gr.addColorStop(0, 'rgba(150,140,230,' + a * 0.5 + ')');
        gr.addColorStop(1, 'rgba(150,140,230,0)');
        x.fillStyle = gr;
        x.fillRect(0, 0, W, H);
        if (dt < 200 || (L.twice && dt > 150 && dt < 240)) {
          x.shadowColor = 'rgba(190,180,255,.9)';
          x.shadowBlur = 14;
          x.lineCap = 'round';
          x.lineJoin = 'round';
          drawBolt(L.b, 2.6, 0.95);
          x.shadowBlur = 0;
        }
      }
    }
    // candles: a warm glow on the title and swaying flames
    x.globalCompositeOperation = 'lighter';
    FL.forEach(([ix, iy], k) => {
      const I = quiet ? 0.9 : flick(fl[k], t), [cx, cy] = P(ix, iy);
      let gr = x.createRadialGradient(cx, cy + 40 * S, 0, cx, cy + 40 * S, 330 * S);
      gr.addColorStop(0, 'rgba(255,140,50,' + 0.16 * I + ')');
      gr.addColorStop(0.5, 'rgba(200,70,20,' + 0.06 * I + ')');
      gr.addColorStop(1, 'rgba(160,40,10,0)');
      x.fillStyle = gr;
      x.fillRect(cx - 340 * S, cy - 300 * S, 680 * S, 700 * S);
      gr = x.createRadialGradient(cx, cy, 0, cx, cy, 46 * S);
      gr.addColorStop(0, 'rgba(255,220,140,' + 0.55 * I + ')');
      gr.addColorStop(1, 'rgba(255,140,40,0)');
      x.fillStyle = gr;
      x.fillRect(cx - 50 * S, cy - 50 * S, 100 * S, 100 * S);
      const sw = quiet ? 0 : Math.sin(t * 0.013 + k * 2) * 3 * S + (Math.random() - 0.5) * 1.5 * S, h = (30 + 8 * I) * S;
      x.beginPath();
      x.moveTo(cx - 6 * S, cy + 14 * S);
      x.quadraticCurveTo(cx - 7 * S, cy - h * 0.3, cx + sw, cy - h * 0.75);
      x.quadraticCurveTo(cx + 7 * S, cy - h * 0.3, cx + 6 * S, cy + 14 * S);
      x.closePath();
      x.fillStyle = 'rgba(255,190,90,' + (0.28 + 0.25 * I) + ')';
      x.fill();
    });
    x.globalCompositeOperation = 'source-over';
  };
  requestAnimationFrame(frame);
}
