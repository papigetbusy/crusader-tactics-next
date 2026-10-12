// The ☰ system menu: End Turn, Gear, Save, Settings, map controls and the
// floor list; plus the settings window, the floor list and the save slots
// [old openSys L1506, refreshSys L1605, openFloors L1614, save slots L1886].
// Saves are the referee's own save text, kept in this browser.
import * as referee from './referee.js';
import { options, saveOptions, onOptionsSaved, applyOptions, withSettings } from './options.js';
import { snapshot as formationSnapshot, restore as formationRestore } from './formations.js';
import { zoomStep, closeUp } from './camera.js';
import { rotate } from './board.js';
import { url } from './assets.js';
import { settings, wireSettings } from './settings.js';

const $ = (id) => document.getElementById(id);
let screen = null; // { st, show, kick, log, after, gear, endTurn, nextHero }

export function wire(s) {
  screen = s;
  // A changed setting is written into the autosave, so Continue comes back the same way.
  onOptionsSaved(() => {
    if (screen && screen.st.view && screen.st.view.climb) put('auto');
  });
  wireSettings(s);
  $('menuButton').onclick = (e) => { e.stopPropagation(); $('sys').hidden ? open() : close(); };
  const eye = $('mapEye');
  if (eye) eye.onclick = (e) => {
    e.stopPropagation();
    const bar = $('mapbar');
    const openBar = bar.hidden;
    if (openBar) mountMapbar();
    bar.hidden = !openBar;
    eye.classList.toggle('on', openBar);
  };
  document.addEventListener('pointerdown', (e) => {
    if (!$('sys').hidden && !$('sys').contains(e.target) && e.target !== $('menuButton')) close();
    const bar = $('mapbar'), eyeBtn = $('mapEye');
    if (bar && !bar.hidden && !bar.contains(e.target) && eyeBtn && !eyeBtn.contains(e.target)) {
      bar.hidden = true;
      eyeBtn.classList.remove('on');
    }
  }, true);
}

const close = () => { $('sys').hidden = true; };
// The menu's little line icons [old IC2, L1566].
const IC = {
  zout: '<circle cx="10" cy="10" r="6.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M15 15l6 6M7 10h6" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round"/>',
  zin: '<circle cx="10" cy="10" r="6.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M15 15l6 6M7 10h6M10 7v6" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round"/>',
  rot: '<path d="M19 12a7 7 0 1 1-2.1-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"/><path d="M17 3v4.5h-4.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>',
  next: '<path d="M4 12h12M12 7l5 5-5 5M20 5v14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>',
  map: '<path d="M3 6l6-2 6 2 6-2v14l-6 2-6-2-6 2z M9 4v14 M15 6v14" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/>',
  view: '<path d="M12 3l9 5-9 5-9-5z M3 12l9 5 9-5 M3 16l9 5 9-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/>',
  danger: '<path d="M12 3l10 18H2z M12 10v5 M12 18v.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"/>',
  speed: '<circle cx="12" cy="13" r="8" fill="none" stroke="currentColor" stroke-width="2"/><path d="M12 13l4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>',
  floors: '<path d="M4 20h16M6 16h12M8 12h8M10 8h4M12 4v4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>',
};
const ic = (n, s = 20) => `<svg viewBox="0 0 24 24" width="${s}" height="${s}" aria-hidden="true" style="flex:none">${IC[n]}</svg>`;
export const sw = (id, on) => `<button class="sw ${on ? 'on' : ''}" id="${id}"><i></i></button>`;
export const seg = (id, opts, val) => `<div class="seg" id="${id}">${opts.map(([v, l]) => `<button data-v="${v}" class="${String(v) === String(val) ? 'on' : ''}">${l}</button>`).join('')}</div>`;

export function open() {
  const v = screen.st.view, c = v.climb, w = $('sys');
  const battle = !c || c.phase === 'battle';
  w.innerHTML = `<div class="sysh"><img src="${url('ui', 'coin')}" alt=""> ${c ? c.gold : v.gold} gold</div>
    ${battle ? '<button class="row" id="sEnd">End Turn</button>' : ''}
    <button class="row" id="sGear">Gear</button><button class="row" id="sSave">Save</button>
    <button class="row" id="sSet">Settings</button><button class="row" id="sClose">✕ Close</button>
    ${c ? `<div class="lbl">Floors</div><button class="b wide" id="sFloors">${ic('floors', 18)} Floor list</button>` : ''}`;
  w.hidden = false;
  const on = (id, f) => { if ($(id)) $(id).onclick = f; };
  on('sEnd', () => { close(); screen.endTurn(); });
  on('sGear', () => { close(); screen.gear(); });
  on('sSave', () => { close(); slots('save'); });
  on('sSet', () => { close(); settings(); });
  on('sClose', close);
  on('sFloors', () => { close(); floors(); });
}

// The map controls are a narrow strip of icons on the edge of the screen,
// never a labeled panel over the map (owner, 2026-10-12). Tap toggles.
function mountMapbar() {
  const bar = $('mapbar');
  if (!bar || !screen) return;
  const b = (id, label, icon, on) => `<button type="button" class="mb2${on ? ' on' : ''}" id="${id}" aria-label="${label}" title="${label}">${ic(icon, 18)}</button>`;
  const speed = options.speed === 3 ? '⚡' : options.speed + '×';
  bar.innerHTML = [
    b('mbOut', 'Zoom out', 'zout', false),
    b('mbIn', 'Zoom in', 'zin', false),
    b('mbView', 'Flat or isometric', 'view', options.flat),
    b('mbMap', 'Minimap', 'map', options.minimap),
    b('mbDanger', 'Danger zone', 'danger', options.danger),
    `<button type="button" class="mb2${options.speed > 1 ? ' on' : ''}" id="mbSpeed" aria-label="Battle speed" title="Battle speed">${speed}</button>`,
    b('mbRot', 'Rotate', 'rot', false),
    b('mbNext', 'Next hero', 'next', false),
  ].join('');
  const on = (id, f) => { const el = $(id); if (el) el.onclick = (e) => { e.stopPropagation(); f(); }; };
  on('mbOut', () => zoomStep(-1));
  on('mbIn', () => zoomStep(1));
  on('mbView', () => { screen.setView(!options.flat); saveOptions(); mountMapbar(); });
  on('mbMap', () => { options.minimap = !options.minimap; saveOptions(); screen.after(); mountMapbar(); });
  on('mbDanger', () => { options.danger = !options.danger; saveOptions(); screen.after(); mountMapbar(); });
  on('mbSpeed', () => { options.speed = options.speed === 3 ? 1 : options.speed + 1; saveOptions(); mountMapbar(); });
  on('mbRot', () => { rotate(); screen.kick(); });
  on('mbNext', () => screen.nextHero());
}

export function mapOn(show) {
  const bar = $('mapbar'), eye = $('mapEye');
  if (bar) bar.hidden = true;
  if (eye) { eye.hidden = !show; eye.classList.remove('on'); }
  if (show) mountMapbar();
}

export function cover(id) {
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

// The purple-team test save ships with the game (web/assets/purple-team.json).
// Phones cannot download it, so the first empty slot — 3, then 2, then 1 —
// is filled on startup. A slot you have saved into yourself is never replaced.
// If every slot is already full, Continue still shows a Purple team row.
const PURPLE = 'purple-team';
let purpleText = '';
export async function seedPurpleTeam() {
  try {
    const res = await fetch('assets/purple-team.json');
    if (!res.ok) return;
    const text = (await res.text()).trim();
    const parsed = JSON.parse(text);
    if (!parsed || parsed.kind !== 'climb' || !parsed.data) return;
    purpleText = text;
    const run = parsed.data.run || {};
    const battle = parsed.data.battle || {};
    const meta = JSON.stringify({
      floor: run.floor || battle.Floor || 1,
      best: run.best || run.floor || 1,
      turn: battle.Turn || 1,
      when: 'Purple team',
      builtIn: PURPLE,
    });
    const held = [1, 2, 3].find((k) => (info(k) || {}).builtIn === PURPLE);
    const slot = held || [3, 2, 1].find((k) => !get(k));
    if (!slot) return;
    localStorage.setItem(KEY(slot), text);
    localStorage.setItem(KEY(slot) + '-info', meta);
  } catch (e) { /* storage blocked: the Continue row below still loads it */ }
}

function pack(text) {
  const t = withSettings(text);
  try {
    const o = JSON.parse(t);
    o.formations = formationSnapshot();
    return JSON.stringify(o);
  } catch (e) {
    return t;
  }
}
function put(k) {
  const v = screen.st.view, c = v.climb;
  const text = pack(referee.save());
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
  if (!text) return false;
  const r = referee.load(text);
  if (!r || r.error) { alert('Could not load: ' + (r ? r.error : 'no answer')); return false; }
  try {
    const o = JSON.parse(text);
    applyOptions(o.screen);
    formationRestore(o.formations);
  } catch (e) { /* a save with no settings keeps the ones already on screen */ }
  ['slotsWin', 'summary', 'team'].forEach((id) => { if ($(id)) $(id).hidden = true; });
  $('title').hidden = true;
  mapOn(true);
  screen.show(r.view, true);
  screen.after();
  screen.log(`Loaded ${from}.`);
  return true;
}

// importer: bring in a save from a file or pasted text, into a slot of the
// player's choice; it asks before replacing a slot that has a game [old
// Settings → Saves, savTab L1674-1684, importSave L1671]. The screen only
// reads the file; the referee checks and loads the save text.
export function importer(mode) {
  const w = cover('slotsWin');
  let slot = [1, 2, 3].find((k) => !get(k)) || 3, text = '', name = '';
  const draw = () => {
    w.innerHTML = `<div class="panel box wide"><div class="ttl">Import a save</div>
      <p><small>Choose a save file or paste a save code, pick a slot, then tap Import. <b>Importing replaces that slot.</b></small></p>
      <div class="mrow"><span>Slot</span>${seg('imSlot', [[1, '1'], [2, '2'], [3, '3']], slot)}</div>
      <input type="file" id="imFile" accept=".json,.txt,application/json,text/plain">
      <textarea id="imText" rows="3" placeholder="…or paste a save code here"></textarea>
      <p id="imMsg"><small>${name ? 'Ready: ' + name : ''}</small></p>
      <div class="btns"><button class="b" id="imGo">Import</button><button class="b" id="imBack">Back</button></div></div>`;
    w.hidden = false;
    const msg = (t, bad) => { $('imMsg').innerHTML = `<small style="color:${bad ? '#ff8a8a' : '#8ff0b0'}">${t}</small>`; };
    $('imSlot').querySelectorAll('button').forEach((b) => (b.onclick = () => { slot = +b.dataset.v; $('imSlot').querySelectorAll('button').forEach((x) => x.classList.toggle('on', x === b)); }));
    $('imFile').onchange = (e) => {
      const f = e.target.files && e.target.files[0];
      if (!f) return;
      if (f.size > 2e6) return msg('That file is too large to be a save.', true);
      f.text().then((t) => { text = t; name = f.name; msg(`Ready: ${f.name} (${Math.max(1, Math.round(f.size / 1024))} KB). Tap Import.`); }).catch(() => msg('Could not read that file.', true));
    };
    $('imBack').onclick = () => slots(mode);
    $('imGo').onclick = () => {
      const t = (text || $('imText').value).trim();
      if (!t) return msg('Choose a file or paste a save code first.', true);
      const i = info(slot);
      if (get(slot) && !confirm(`Replace slot ${slot}?\nSlot ${slot} already has a game${i ? ` (floor ${i.floor}, ${i.when})` : ''}. Importing replaces it.`)) return;
      if (load(t, name ? 'the save file' : 'the pasted save') && put(slot)) screen.log(`Imported save loaded into slot ${slot}.`);
    };
  };
  draw();
}

export function slots(mode) {
  const w = cover('slotsWin');
  w.style.zIndex = 25;
  const rows = ['auto', 1, 2, 3].map((k) => {
    const i = info(k);
    return `<div class="srow"><span><b>${k === 'auto' ? 'Autosave' : 'Slot ' + k}</b><br><small>${i ? `Floor ${i.floor} (best ${i.best}) · turn ${i.turn} · ${i.when}` : 'empty'}</small></span>
      ${mode === 'save' && k !== 'auto' ? `<button class="b" data-sv="${k}">Save</button>` : ''}${i ? `<button class="b" data-ld="${k}">Load</button>` : ''}</div>`;
  }).join('');
  const held = [1, 2, 3].some((k) => (info(k) || {}).builtIn === PURPLE);
  const extra = purpleText && !held
    ? `<div class="srow"><span><b>Purple team</b><br><small>Floor 40 · whole team in purple gear</small></span><button class="b" id="ldPurple">Load</button></div>`
    : '';
  const inGame = screen.st.view && $('title').hidden;
  w.innerHTML = `<div class="panel box wide"><div class="ttl">${mode === 'save' ? 'Save your game' : 'Continue'}</div>${extra}${rows}
    <div class="btns">${inGame ? '<button class="b" id="svExp">Export</button>' : ''}<button class="b" id="svImp">Import</button></div>
    <button class="b" id="svClose">Close</button></div>`;
  w.hidden = false;
  w.querySelectorAll('[data-sv]').forEach((b) => (b.onclick = () => { if (put(b.dataset.sv)) { screen.log('Saved to slot ' + b.dataset.sv + '.'); slots(mode); } }));
  w.querySelectorAll('[data-ld]').forEach((b) => (b.onclick = () => load(get(b.dataset.ld), b.dataset.ld === 'auto' ? 'the autosave' : 'slot ' + b.dataset.ld)));
  if ($('ldPurple')) $('ldPurple').onclick = () => load(purpleText, 'the purple team');
  $('svClose').onclick = () => { w.hidden = true; };
  if ($('svExp')) $('svExp').onclick = () => { exportSave(); screen.log('Save copied and downloaded.'); };
  $('svImp').onclick = () => importer(mode);
}

// exportSave copies the game's save text and downloads it as a file.
export function exportSave() {
  const t = pack(referee.save());
  try { navigator.clipboard && navigator.clipboard.writeText(t); } catch (e) { /* not allowed: the download still works */ }
  const a = document.createElement('a');
  a.href = URL.createObjectURL(new Blob([t], { type: 'application/json' }));
  const c = screen.st.view && screen.st.view.climb;
  a.download = 'crusader-save-floor' + (c ? c.best : 1) + '.json';
  a.click();
}

export { closeUp, settings };
