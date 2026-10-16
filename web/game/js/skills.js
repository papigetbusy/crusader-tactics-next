// The Skills window (owner, 2026-10-16): every skill each hero knows, with
// what it does, its range, power, cooldown and the level it was learned, and
// changing which skills a hero brings (once per floor). Every change is the
// referee's "bring" action; the window is redrawn from its answer.
import * as referee from './referee.js';
import { mine } from './side.js';
import * as online from './online.js';
import { heroPicUrl } from './team.js';
import { cover } from './sysmenu.js';
import { ELEMENT } from './look.js';

const $ = (id) => document.getElementById(id);
const ELN = { fire: 'Fire', ice: 'Ice', elec: 'Thunder', light: 'Light', dark: 'Dark' };
let screen = null, cls = null, editing = false, picked = null, ult = '';

export const wire = (s) => { screen = s; };

const heroes = () => screen.st.units.filter((u) => mine(u, screen.st.view) && u.gear && !u.companion && !u.decoy && !u.pet);

// open shows the window for a hero (the chosen one, else the first).
export function open(className) {
  const list = heroes();
  if (!list.length) return screen.log('No hero to show.');
  cls = className || (screen.st.sel && screen.st.sel.gear ? screen.st.sel.class : cls) || list[0].class;
  if (!list.some((u) => u.class === cls)) cls = list[0].class;
  editing = false;
  render();
}

// chips are a skill's numbers in short labels.
function chips(s) {
  const c = [];
  if (s.ability) c.push(`<i>${s.cooldown ? 'Cooldown ' + s.cooldown : 'No cooldown'}</i>`);
  if (s.ability) return `<div class="chips">${c.join('')}</div>`;
  if (s.maxRange) c.push(`<i>Range ${s.minRange && s.minRange !== s.maxRange ? s.minRange + '–' : ''}${s.maxRange}</i>`);
  else c.push('<i>Self</i>');
  if (s.power && !s.support) c.push(`<i>${s.heal ? 'Heal' : 'Power'} ×${s.power}</i>`);
  if (s.area) c.push(`<i>Area ${s.area}</i>`);
  if (s.element) c.push(`<i style="color:${ELEMENT[s.element] || '#fff'}">${ELN[s.element] || s.element}</i>`);
  if (s.revive) c.push('<i>Revive</i>');
  if (s.support) c.push('<i>Buff</i>');
  c.push(`<i>${s.cooldown ? 'Cooldown ' + s.cooldown : 'No cooldown'}</i>`);
  if (s.level) c.push(`<i>Lv ${s.level}</i>`);
  return `<div class="chips">${c.join('')}</div>`;
}

// card is one skill. pick is the edit-mode control, if any.
function card(s, extra, pick) {
  const name = (s.ultimate ? '★ ' : '') + s.name;
  return `<div class="skc${s.ultimate ? ' ult' : ''}${pick && pick.on ? ' on' : ''}"${pick ? ` data-pick="${s.name}" data-ult="${s.ultimate ? 1 : 0}"` : ''}>
    <div class="skn"><b>${name}</b>${extra || ''}${pick ? `<span class="tick">${pick.on ? '✓' : ''}</span>` : ''}</div>
    ${chips(s)}${s.text ? `<small>${s.text}</small>` : ''}</div>`;
}

function render() {
  const v = screen.st.view, c = v.climb, w = cover('skillsWin');
  const u = heroes().find((x) => x.class === cls);
  if (!u) { w.hidden = true; return; }
  const phase = c ? c.phase : 'battle';
  const locked = u.done || v.side !== 0 || (phase !== 'battle' && phase !== 'prepare');
  const bar = (u.skills || []).filter((s) => s.name !== 'Strike' && !s.ultimate), now = (u.skills || []).find((s) => s.ultimate);
  const lib = u.library || [];
  const canChange = !u.swapped && !locked && lib.length > 0;
  let h = `<div class="panel box wide sheet skw"><button class="xclose" aria-label="Close">✕</button><div class="ttl">Skills</div>
    <div class="tabs">${heroes().map((x) => `<button class="${x.class === cls ? 'on' : ''}" data-h="${x.class}">${x.class}</button>`).join('')}</div>
    <div class="who"><img src="${heroPicUrl({ class: u.class, stage: u.stage, ascension: u.ascension })}" alt=""><div><div class="ttl">${u.class}</div>
    <small>${u.form || ''} · Lv ${u.level}${u.ascension ? ' · ' + u.ascension : ''}</small><br>
    <small class="${u.swapped ? 'why' : 'cmp'}">${u.swapped ? 'Skills already changed on this floor.' : locked ? 'Skills can be changed on your turn, before the hero acts.' : 'You can change skills once on this floor.'}</small></div></div>`;
  if (u.ability) h += `<h4>Class ability</h4>${card({ name: '✦ ' + u.ability.name, text: u.ability.text, cooldown: u.ability.cooldown, ability: true }, u.ability.wait ? ` <span class="wait">${u.ability.wait} turn${u.ability.wait > 1 ? 's' : ''}</span>` : ' <span class="rdy">Ready</span>')}`;
  if (!editing) {
    const wait = (s) => (phase === 'battle' ? (s.wait ? ` <span class="wait">${s.wait} turn${s.wait > 1 ? 's' : ''}</span>` : ' <span class="rdy">Ready</span>') : '');
    h += `<h4>Brought · ${bar.length} of 4</h4>${bar.map((s) => card(s, wait(s))).join('')}`;
    if (now) h += `<h4>Ultimate</h4>${card(now, wait(now))}`;
    h += `<h4>Learned, not brought</h4>${lib.map((s) => card(s)).join('') || '<p><small>Nothing else learned yet.</small></p>'}`;
    if (canChange) h += '<button class="b" id="skEdit">Change skills</button>';
  } else {
    const normals = bar.map((s) => s).concat(lib.filter((s) => !s.ultimate));
    const ults = (now ? [now] : []).concat(lib.filter((s) => s.ultimate));
    h += `<h4>Pick up to 4 · ${picked.length} picked</h4>${normals.map((s) => card(s, '', { on: picked.includes(s.name) })).join('')}`;
    h += `<h4>Pick 1 ultimate</h4>${ults.map((s) => card(s, '', { on: ult === s.name })).join('')}`;
    h += '<div class="btns"><button class="b" id="skCancel">Cancel</button><button class="b" id="skGo">Save changes</button></div>';
  }
  h += '<button class="b" id="skClose">Close</button></div>';
  w.innerHTML = h;
  w.hidden = false;
  const q = (s) => w.querySelectorAll(s);
  q('[data-h]').forEach((b) => (b.onclick = () => { cls = b.dataset.h; editing = false; render(); }));
  q('[data-pick]').forEach((d) => (d.onclick = () => {
    const n = d.dataset.pick;
    if (d.dataset.ult === '1') ult = n;
    else if (picked.includes(n)) picked = picked.filter((x) => x !== n);
    else if (picked.length >= 4) return screen.log('A hero brings at most 4 skills. Untick one first.');
    else picked = picked.concat(n);
    render();
  }));
  const on = (id, f) => { if ($(id)) $(id).onclick = f; };
  on('skEdit', () => { editing = true; picked = bar.map((s) => s.name); ult = now ? now.name : ''; render(); });
  on('skCancel', () => { editing = false; render(); });
  on('skGo', () => {
    if (!picked.length) return screen.log('Pick at least one skill.');
    if (!confirm('Change skills now? You can only do this once on this floor.')) return;
    const a = { kind: 'bring', unit: u.id, bar: picked, choice: ult && (!now || ult !== now.name) ? ult : '' };
    if (online.on()) { w.hidden = true; editing = false; online.act(a); return; }
    const r = referee.act(a);
    if (!r || r.refused) return screen.log(r ? r.refused.reason : 'The referee did not answer.');
    editing = false;
    screen.show(r.view);
    screen.after();
    screen.log(`${u.class}'s skills changed.`);
    render();
  });
  on('skClose', () => { w.hidden = true; });
  w.querySelector('.xclose').onclick = () => { w.hidden = true; };
}
