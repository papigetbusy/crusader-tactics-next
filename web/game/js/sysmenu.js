// The ☰ system menu: End Turn, Gear, Save, Settings, map controls and the
// floor list; plus the settings window, the floor list and the save slots
// [old openSys L1506, refreshSys L1605, openFloors L1614, save slots L1886].
// Saves are the referee's own save text, kept in this browser.
import * as referee from './referee.js';
import { options, saveOptions } from './options.js';
import { cam, zoomAt, closeUp } from './camera.js';
import { rotate } from './board.js';
import { url } from './assets.js';

const $ = (id) => document.getElementById(id);
let screen = null; // { st, show, kick, log, after, gear, endTurn, nextHero }

export function wire(s) {
  screen = s;
  $('menuButton').onclick = (e) => { e.stopPropagation(); $('sys').hidden ? open() : close(); };
  document.addEventListener('pointerdown', (e) => {
    if (!$('sys').hidden && !$('sys').contains(e.target) && e.target !== $('menuButton')) close();
  }, true);
}

const close = () => { $('sys').hidden = true; };
const sw = (id, on) => `<button class="sw ${on ? 'on' : ''}" id="${id}"><i></i></button>`;
const seg = (id, opts, val) => `<div class="seg" id="${id}">${opts.map(([v, l]) => `<button data-v="${v}" class="${String(v) === String(val) ? 'on' : ''}">${l}</button>`).join('')}</div>`;

export function open() {
  const v = screen.st.view, c = v.climb, w = $('sys');
  const battle = !c || c.phase === 'battle';
  w.innerHTML = `<div class="sysh"><img src="${url('ui', 'coin')}" alt=""> ${c ? c.gold : v.gold} gold</div>
    ${battle ? '<button class="row" id="sEnd">End Turn</button>' : ''}
    <button class="row" id="sGear">Gear</button><button class="row" id="sSave">Save</button>
    <button class="row" id="sSet">Settings</button><button class="row" id="sClose">✕ Close</button>
    <div class="lbl">Map</div>
    <div class="mrow"><span>Zoom</span><input type="range" id="sZoom" min="1" max="3.4" step="0.05" value="${cam.zt.toFixed(2)}"></div>
    <div class="mrow"><span>View</span>${seg('sView', [[1, 'Flat'], [0, 'Isometric']], options.flat ? 1 : 0)}</div>
    <div class="mrow"><span>Minimap</span>${sw('sMini', options.minimap)}</div>
    <div class="mrow"><span>Danger zone</span>${sw('sDanger', options.danger)}</div>
    <div class="mrow"><span>Battle speed</span>${seg('sSpeed', [[1, '1×'], [2, '2×'], [3, 'Instant']], options.speed)}</div>
    <div class="mbtns"><button class="mb" id="sRot">⟳<span>Rotate</span></button><button class="mb" id="sNext">⇥<span>Next hero</span></button></div>
    ${c ? '<div class="lbl">Floors</div><button class="b wide" id="sFloors">Floor list</button>' : ''}`;
  w.hidden = false;
  const on = (id, f) => { if ($(id)) $(id).onclick = f; };
  on('sEnd', () => { close(); screen.endTurn(); });
  on('sGear', () => { close(); screen.gear(); });
  on('sSave', () => { close(); slots('save'); });
  on('sSet', () => { close(); settings(); });
  on('sClose', close);
  on('sRot', () => { rotate(); screen.kick(); });
  on('sNext', () => { close(); screen.nextHero(); });
  on('sFloors', () => { close(); floors(); });
  $('sZoom').oninput = (e) => zoomAt(innerWidth / 2, innerHeight / 2, +e.target.value);
  const flip = (id, k) => on(id, () => { options[k] = !options[k]; saveOptions(); screen.after(); open(); });
  flip('sMini', 'minimap');
  flip('sDanger', 'danger');
  $('sSpeed').querySelectorAll('button').forEach((b) => (b.onclick = () => { options.speed = +b.dataset.v; saveOptions(); open(); }));
  $('sView').querySelectorAll('button').forEach((b) => (b.onclick = () => { screen.setView(b.dataset.v === '1'); saveOptions(); open(); }));
}

// ---------- settings ----------

export function settings() {
  const w = cover('settingsWin');
  const row = (label, ctl, sub) => `<div class="mrow"><span>${label}${sub ? `<br><small>${sub}</small>` : ''}</span>${ctl}</div>`;
  w.innerHTML = `<div class="panel box wide"><div class="ttl">Settings</div>
    <h4>Battle</h4>${row('Battle speed', seg('oSpeed', [[1, '1×'], [2, '2×'], [3, 'Instant']], options.speed))}
    ${row('Auto end turn', sw('oAuto', options.autoEnd), 'Ends your turn once every hero has acted')}
    ${row('Ask before ending turn', sw('oAsk', options.confirmEnd), 'Warns if heroes haven’t acted yet')}
    <h4>Map</h4>${row('Map view', seg('oView', [[1, 'Flat'], [0, 'Isometric']], options.flat ? 1 : 0))}${row('Minimap', sw('oMini', options.minimap))}${row('Danger zone', sw('oDanger', options.danger), 'Shade every tile monsters can reach')}
    <h4>Accessibility &amp; performance</h4>${row('Reduce flashing effects', sw('oCalm', options.calm))}
    ${row('Low graphics', sw('oLow', options.lowGraphics), 'Still water and trees, no vignette; saves battery')}
    ${row('Colorblind-friendly tiles', sw('oCb', options.colorblind))}
    <div class="btns"><button class="b" id="oTitle">Title screen</button><button class="b" id="oClose">Close</button></div></div>`;
  w.hidden = false;
  const flip = (id, k) => ($(id).onclick = () => { options[k] = !options[k]; saveOptions(); screen.after(); settings(); });
  flip('oAuto', 'autoEnd'); flip('oAsk', 'confirmEnd'); flip('oMini', 'minimap'); flip('oDanger', 'danger');
  flip('oCalm', 'calm'); flip('oLow', 'lowGraphics'); flip('oCb', 'colorblind');
  $('oSpeed').querySelectorAll('button').forEach((b) => (b.onclick = () => { options.speed = +b.dataset.v; saveOptions(); settings(); }));
  $('oView').querySelectorAll('button').forEach((b) => (b.onclick = () => { screen.setView(b.dataset.v === '1'); saveOptions(); settings(); }));
  $('oClose').onclick = () => { w.hidden = true; };
  $('oTitle').onclick = () => { if (confirm('Go back to the title screen? Unsaved progress is lost (the autosave keeps your last turn).')) location.reload(); };
}

function cover(id) {
  let w = $(id);
  if (!w) {
    w = document.createElement('div');
    w.id = id;
    w.className = 'cover';
    document.body.append(w);
  }
  return w;
}

// ---------- the floor list ----------

export function floors() {
  const c = screen.st.view.climb, w = cover('floorsWin');
  const tag = (cls, t) => `<i class="${cls}">${t}</i>`;
  const rows = c.floors.map((f) => `<button class="fl ${f.here ? 'cur' : ''}" data-f="${f.number}"><b>Floor ${f.number}</b> <small>${f.biome || ''}</small><span>
    ${f.here ? tag('t-c', 'here') : ''}${f.top ? tag('t-top', 'top') : ''}${f.cleared ? tag('t-ok', '✓ cleared') : ''}
    ${f.boss ? tag('t-b', f.bossLegendary ? 'boss · legendary taken' : 'boss') : f.elite ? tag('t-e', 'elite') : ''}
    ${f.shop ? tag('t-s', 'shop') : ''}${f.chestsLeft ? tag('t-ch', `${f.chestsLeft} chest${f.chestsLeft > 1 ? 's' : ''} left`) : ''}</span></button>`);
  w.innerHTML = `<div class="panel box wide"><div class="ttl">Floor travel</div><p><small>Floors keep the same map and monsters. Opened chests stay opened; replays pay half gold and full XP.</small></p>
    <div class="list">${rows.join('')}</div><button class="b" id="flClose">Close</button></div>`;
  w.hidden = false;
  $('flClose').onclick = () => { w.hidden = true; };
  w.querySelectorAll('.fl').forEach((b) => (b.onclick = () => {
    const f = +b.dataset.f;
    if (f === c.floor) return;
    if (c.phase === 'battle' && !confirm(`Leave this battle and go to floor ${f}? What you earned is kept.`)) return;
    const r = referee.act({ kind: 'travel', floor: f });
    if (!r || r.refused) return screen.log(r ? r.refused.reason : '');
    w.hidden = true;
    screen.show(r.view);
    screen.after();
    autosave();
  }));
}

// ---------- saves: three slots and an autosave, in this browser ----------

const KEY = (k) => 'ctn-save-' + k;
const get = (k) => { try { return localStorage.getItem(KEY(k)); } catch (e) { return null; } };
const info = (k) => { try { return JSON.parse(localStorage.getItem(KEY(k) + '-info') || 'null'); } catch (e) { return null; } };
export const anySave = () => ['auto', 1, 2, 3].some((k) => get(k));

function put(k) {
  const v = screen.st.view, c = v.climb;
  const text = referee.save();
  if (!text || text.startsWith('{"error"')) return false;
  try {
    localStorage.setItem(KEY(k), text);
    localStorage.setItem(KEY(k) + '-info', JSON.stringify({ floor: c ? c.floor : v.floor, best: c ? c.best : 0, turn: v.turn, when: new Date().toLocaleString() }));
    return true;
  } catch (e) {
    screen.log('Could not save: ' + e.message);
    return false;
  }
}

// autosave keeps the climb after each turn and each floor (a climb only).
export function autosave() {
  if (screen && screen.st.view && screen.st.view.climb) put('auto');
}

function load(text, from) {
  if (!text) return;
  const r = referee.load(text);
  if (!r || r.error) return alert('Could not load: ' + (r ? r.error : 'no answer'));
  ['slotsWin', 'summary', 'team'].forEach((id) => { if ($(id)) $(id).hidden = true; });
  $('title').hidden = true;
  screen.show(r.view, true);
  screen.after();
  screen.log(`Loaded ${from}.`);
}

export function slots(mode) {
  const w = cover('slotsWin');
  w.style.zIndex = 25;
  const rows = ['auto', 1, 2, 3].map((k) => {
    const i = info(k);
    return `<div class="srow"><span><b>${k === 'auto' ? 'Autosave' : 'Slot ' + k}</b><br><small>${i ? `Floor ${i.floor} (best ${i.best}) · turn ${i.turn} · ${i.when}` : 'empty'}</small></span>
      ${mode === 'save' && k !== 'auto' ? `<button class="b" data-sv="${k}">Save</button>` : ''}${i ? `<button class="b" data-ld="${k}">Load</button>` : ''}</div>`;
  }).join('');
  const inGame = screen.st.view && $('title').hidden;
  w.innerHTML = `<div class="panel box wide"><div class="ttl">${mode === 'save' ? 'Save your game' : 'Continue'}</div>${rows}
    <div class="btns">${inGame ? '<button class="b" id="svExp">Export</button>' : ''}<button class="b" id="svImp">Import</button></div>
    <button class="b" id="svClose">Close</button></div>`;
  w.hidden = false;
  w.querySelectorAll('[data-sv]').forEach((b) => (b.onclick = () => { if (put(b.dataset.sv)) { screen.log('Saved to slot ' + b.dataset.sv + '.'); slots(mode); } }));
  w.querySelectorAll('[data-ld]').forEach((b) => (b.onclick = () => load(get(b.dataset.ld), b.dataset.ld === 'auto' ? 'the autosave' : 'slot ' + b.dataset.ld)));
  $('svClose').onclick = () => { w.hidden = true; };
  if ($('svExp')) $('svExp').onclick = () => {
    const t = referee.save();
    try { navigator.clipboard && navigator.clipboard.writeText(t); } catch (e) { /* not allowed: the download still works */ }
    const a = document.createElement('a');
    a.href = URL.createObjectURL(new Blob([t], { type: 'application/json' }));
    a.download = 'crusader-tactics-save.json';
    a.click();
    screen.log('Save copied and downloaded.');
  };
  $('svImp').onclick = () => { const t = prompt('Paste a save:'); if (t) load(t.trim(), 'the pasted save'); };
}

export { closeUp };
