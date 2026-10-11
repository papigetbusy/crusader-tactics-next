// Crusader Tactics, the game screen. It starts the referee and the pictures,
// shows the title, and runs the battle screen. The screen keeps no rules and
// no game state of its own: everything comes from the referee's View.
import * as referee from './referee.js';
import * as assets from './assets.js';
import { map, setSize } from './board.js';
import { cam, controls, frame, resize, onMove, moving, closeUp, overview } from './camera.js';
import * as scene from './scene.js';
import * as minimap from './minimap.js';
import { options } from './options.js';
import * as battle from './battle.js';
import * as playback from './playback.js';
import * as sysmenu from './sysmenu.js';
import * as gear from './gear.js';

const $ = (id) => document.getElementById(id);

// st is what the screen is showing right now.
export const st = {
  view: null, units: [], sel: null, look: null, aim: null,
  skill: -1, reach: new Set(), targets: new Set(), starts: null, danger: null, zone: null, attack: null, picks: null, inReach: null,
  overlay: (g, now) => { if (playback.draw(g, now)) kick(); },
  opt: options,
  anim: playback.anim,
  shownHp: (u) => u.hp,
  unitAt: (x, y) => st.units.find((u) => u.x === x && u.y === y) || null,
};

// ---------- drawing ----------

let wake = 0;
export const kick = () => { wake = performance.now() + 600; };
onMove(kick);

function loop() {
  requestAnimationFrame(loop);
  if (!st.view || !$('title').hidden) return;
  const now = performance.now();
  // busy: draw every frame; quiet: a few frames a second for water and trees
  if (now > wake && !moving() && now - (loop.last || 0) < 120) return;
  loop.last = now;
  scene.draw($('map'), st);
  minimap.draw($('mini'), st, st.opt.minimap && !document.body.classList.contains('aim'));
}

// ---------- the referee's View on screen ----------

// show puts a new View on screen.
export function show(view, again) {
  const fresh = again || !st.view || st.view.n !== view.n || (st.view.floor || 0) !== (view.floor || 0);
  st.view = view;
  st.units = view.units;
  if (st.sel) st.sel = st.units.find((u) => u.id === st.sel.id) || null;
  if (st.look) st.look = st.units.find((u) => u.id === st.look.id) || null;
  map.ground = view.ground || [];
  map.heights = view.heights || [];
  map.biome = view.biome || 'Grasslands';
  if (fresh) {
    map.flat = st.opt.flat;
    setSize(view.n);
    frame(st.units.filter((u) => u.team === 0 && !u.companion));
  }
  chip();
  kick();
}

function chip() {
  const v = st.view, c = v.climb;
  $('chip').hidden = false;
  $('chipText').textContent = `Floor ${c ? c.floor : v.floor || 1} · Turn ${v.turn}`;
}

// log shows a message for 3 seconds, then it fades away (owner, 2026-10-11:
// messages must not stay on screen and clutter the battle).
let logTimer = 0;
export function log(text) {
  const l = $('log');
  clearTimeout(logTimer);
  l.classList.remove('gone');
  l.textContent = text || '';
  if (text) logTimer = setTimeout(() => { l.classList.add('gone'); logTimer = setTimeout(() => { l.textContent = ''; }, 400); }, 3000);
}

// ---------- starting ----------

export async function boot(build) {
  $('version').textContent = build === '__BUILD__' ? 'test' : build;
  const loading = $('loading'), text = $('loadingText');
  $('titleStart').hidden = true;
  loading.hidden = false;
  resize($('map'));
  addEventListener('resize', () => { resize($('map')); kick(); });
  try {
    await Promise.all([
      referee.start(build),
      assets.load(build, (d, n) => { text.textContent = `Loading pictures ${d} / ${n}`; }),
    ]);
  } catch (e) {
    text.textContent = 'The game did not start: ' + e.message;
    return;
  }
  loading.hidden = true;
  $('titleStart').hidden = false;
  $('newGame').onclick = () => start(referee.newClimb());
  $('startButton').onclick = () => { $('titleStart').hidden = true; $('titleMenu').hidden = false; $('continueGame').disabled = !sysmenu.anySave(); };
  $('continueGame').onclick = () => sysmenu.slots('load');
  controls($('map'), keys, (on) => log(on ? 'Map overview. Pinch in or tap + to zoom back in.' : ''));
  $('mini').addEventListener('pointerdown', (e) => minimap.tap($('mini'), e));
  $('nextHero').onclick = () => nextHero();
  const nextHero = () => {
    const hs = st.units.filter((u) => u.team === 0 && !u.done && !u.companion && !u.decoy && !u.pet);
    if (!hs.length) return log('Everyone has acted. Tap End Turn.');
    const h = hs[(hs.indexOf(st.sel) + 1) % hs.length];
    st.sel = h;
    closeUp(h.x, h.y);
  };
  battle.wire({ st, show, kick, log });
  gear.wire({ st, show, kick, log, after: battle.after });
  sysmenu.wire({ st, show, kick, log, after: battle.after, gear: () => gear.open(), endTurn: battle.endTurn, nextHero });
  requestAnimationFrame(loop);
}

function start(view) {
  $('title').hidden = true;
  show(view);
  battle.after();
}

function keys(k) {
  if (k === 'Escape') {
    st.sel = null;
    st.reach = new Set();
    kick();
    return true;
  }
  if (k === 'n' || k === 'N' || k === 'Tab') {
    $('nextHero').click();
    return true;
  }
  return false;
}

export { overview };
