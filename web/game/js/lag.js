// Lag watch. Off unless the page is opened with ?lag=1.
// It records slow frames and which part of the picture took the time.
// It does not change the rules, and it does not run at all during a normal game.
export const on = new URLSearchParams(location.search).get('lag') === '1';

const SLOW = 32;
const spikes = [];
let last = 0;
let frames = 0;
let drawn = 0;
let slow = 0;
let worst = 0;
let open = '';
let cursor = 0;
let bag = {};
let note = {};
let box = null;
let shown = 0;

function hud() {
  if (!box) {
    box = document.createElement('div');
    box.style.cssText = 'position:fixed;left:8px;bottom:8px;z-index:80;padding:4px 8px;background:rgba(0,0,0,.72);color:#f2c14e;font:12px/1.3 ui-monospace,monospace;pointer-events:none';
    document.body.append(box);
  }
  const now = performance.now();
  if (now - shown < 250) return;
  shown = now;
  box.textContent = `lag ${frames} frames · ${slow} slow · worst ${Math.round(worst)}ms`;
}

// mark starts timing the next named part of the picture.
export function mark(name) {
  const now = performance.now();
  if (open) bag[open] = (bag[open] || 0) + (now - cursor);
  open = name || '';
  cursor = now;
}

export function extra(more) {
  note = more || {};
}

export function begin(now) {
  if (!last) last = now;
}

export function end(info) {
  const now = performance.now();
  const gap = now - last;
  last = now;
  frames++;
  if (gap > worst) worst = gap;
  const drew = info && !info.skipped;
  if (drew) drawn++;
  if (gap >= SLOW) {
    slow++;
    const parts = {};
    for (const k of Object.keys(bag)) parts[k] = Math.round(bag[k] * 10) / 10;
    spikes.push({
      gap: Math.round(gap),
      floor: info && info.floor,
      n: info && info.n,
      zoom: info && info.zoom,
      units: info && info.units,
      moving: info && info.moving,
      skipped: !!(info && info.skipped),
      parts,
      tiles: note.tiles || 0,
      rebuilt: !!note.rebuilt,
    });
    if (spikes.length > 60) spikes.shift();
  }
  bag = {};
  open = '';
  note = {};
  hud();
}

export function reset() {
  spikes.length = 0;
  frames = 0;
  drawn = 0;
  slow = 0;
  worst = 0;
  last = performance.now();
}

export function report() {
  const gaps = spikes.map((s) => s.gap).sort((a, b) => a - b);
  const blame = {};
  for (const s of spikes) {
    let top = '';
    let ms = 0;
    for (const k of Object.keys(s.parts)) if (s.parts[k] > ms) { ms = s.parts[k]; top = k; }
    if (!top && s.rebuilt) top = 'ground';
    if (top) blame[top] = (blame[top] || 0) + 1;
    else blame.outside = (blame.outside || 0) + 1;
  }
  return {
    frames,
    drawn,
    slow,
    worst: Math.round(worst),
    p95: gaps.length ? gaps[Math.min(gaps.length - 1, Math.floor(gaps.length * 0.95))] : 0,
    blame,
    spikes: spikes.slice(-12),
  };
}
