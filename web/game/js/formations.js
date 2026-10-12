// Three saved starting formations. They ride in the save file (the referee
// ignores them). A spot is where a hero stood, as a share of the map width
// and how many rows up from the bottom, so it still fits a bigger floor.

let slots = [null, null, null];

export function snapshot() {
  return slots.map((s) => (s ? { spots: s.spots.map((p) => ({ class: p.class, xf: p.xf, yb: p.yb })) } : null));
}

export function restore(saved) {
  slots = [null, null, null];
  if (!Array.isArray(saved)) return;
  for (let i = 0; i < 3; i++) {
    const s = saved[i];
    if (!s || !Array.isArray(s.spots)) continue;
    const spots = s.spots.filter((p) => p && typeof p.class === 'string' && Number.isFinite(p.xf) && Number.isFinite(p.yb));
    if (spots.length) slots[i] = { spots };
  }
}

export const filled = (i) => !!(slots[i] && slots[i].spots.length);

// remember stores where the team is standing now.
export function remember(i, placed, n) {
  const w = Math.max(1, n - 1);
  slots[i] = {
    spots: placed.map((p) => ({ class: p.class, xf: p.tile.x / w, yb: w - p.tile.y })),
  };
}

// spotsFor is where those heroes should stand on this floor's blue tiles.
export function spotsFor(i, v) {
  const form = slots[i];
  if (!form || !v) return [];
  const starts = v.starts || [];
  const team = new Set((v.climb && v.climb.placed || []).map((p) => p.class));
  const used = new Set();
  const out = [];
  const w = Math.max(1, v.n - 1);
  for (const s of form.spots) {
    if (!team.has(s.class)) continue;
    const x = Math.round(s.xf * w), y = w - s.yb;
    let tile = starts.find((t) => t.x === x && t.y === y && !used.has(t.x + ',' + t.y));
    if (!tile) {
      let best = null, d = 1e9;
      for (const t of starts) {
        const k = t.x + ',' + t.y;
        if (used.has(k)) continue;
        const dd = Math.abs(t.x - x) + Math.abs(t.y - y);
        if (dd < d) { d = dd; best = t; }
      }
      tile = best;
    }
    if (!tile) continue;
    used.add(tile.x + ',' + tile.y);
    out.push({ class: s.class, x: tile.x, y: tile.y });
  }
  return out;
}
