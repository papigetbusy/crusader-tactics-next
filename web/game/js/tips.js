// First-time tips: off by default; Settings → "Show tips again" turns them on.
// Each tip shows once, the first time its window appears, and "Got it"
// keeps it away [old showTip L1573, its triggers L1582 and L1598-1600, the
// system-menu tip L1612, off by default since v0.25 L2149]. Only looks:
// what was seen is kept in this browser.
import { options } from './options.js';

const KEY = 'ctn-tips';
let seen = {};
try { seen = JSON.parse(localStorage.getItem(KEY) || '{}'); } catch (e) { seen = {}; }
const $ = (id) => document.getElementById(id);
const open = (id) => $(id) && !$(id).hidden;

const TEXT = {
  start: 'Drag a hero onto a blue tile to move (or tap the hero, then the tile). Tap an enemy in range to attack. Tap ☰ for the menu: End Turn, Gear, Save, Settings and map controls.',
  select: 'Drag or tap to a blue tile to move, or tap the hero again to act where they stand.',
  menu: 'Pick a command. <b>Skills ▸</b> lists skills that can reach a target; <b>✦</b> is your class ability, a free action.',
  aim: 'The red window shows the fight: target HP before → after, damage and hit chance. ◀ ▶ switch targets, ⚔ attacks.',
  enemy: 'Monster info: stats, weakness and statuses. Its reach is shaded purple, its attacks red.',
  sys: 'This is the system menu: End Turn, Gear, Save, Settings, map controls and floor travel.',
};

// show puts up tip k unless it was seen, another tip is up, or a big
// window covers the battle.
function show(k) {
  if (!options.tipsOn || seen[k] || ['gearWin', 'shopWin', 'summary', 'decision', 'team', 'slotsWin', 'settingsWin'].some(open)) return;
  let w = $('tipw');
  if (!w) {
    w = document.createElement('div');
    w.id = 'tipw';
    w.className = 'panel';
    document.body.append(w);
  }
  if (!w.hidden && w.dataset.k) return;
  w.dataset.k = k;
  w.innerHTML = `<div class="tip-h">Tip</div><div>${TEXT[k]}</div><button class="b" id="tipok">Got it</button>`;
  w.hidden = false;
  $('tipok').onclick = () => {
    seen[k] = 1;
    try { localStorage.setItem(KEY, JSON.stringify(seen)); } catch (e) { /* private browsing */ }
    w.hidden = true;
    w.dataset.k = '';
  };
}

// reset forgets every tip seen and turns tips on.
export function reset() {
  seen = {};
  try { localStorage.setItem(KEY, '{}'); } catch (e) { /* private browsing */ }
}

// watch checks a few times a second which window is up, like the old game
// did each time the screen changed. st is the screen state.
export function watch(st) {
  setInterval(() => {
    if (!options.tipsOn || !st.view || !open('chip') && !open('endTurn')) return;
    if (open('sys')) show('sys');
    else if (open('forecast')) show('aim');
    else if (open('einfo')) show('enemy');
    else if (open('menu')) show('menu');
    else if (st.sel && st.sel.team === 0) show('select');
    else if (!seen.start && st.view.side === 0 && !st.view.outcome) show('start');
  }, 400);
}
