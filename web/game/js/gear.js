// The gear sheet: a hero's numbers, weapon sets, worn gear, the bag, a link
// to its skills (the Skills window) and its ascension. Every change is an
// action; the sheet is redrawn from the referee's answer.
import * as referee from './referee.js';
import { mine } from './side.js';
import { heroPicUrl } from './team.js';
import { icon } from './icons.js';

const $ = (id) => document.getElementById(id);
const SLOT = { w: 'Set 1', w2: 'Set 2', w3: 'Set 3', h: 'Helm', a: 'Armor', b: 'Boots', c: 'Charm' };
let screen = null, cls = null, sortBy = 'new';

export const wire = (s) => { screen = s; };

const heroes = () => screen.st.units.filter((u) => mine(u, screen.st.view) && u.gear && !u.companion && !u.decoy && !u.pet);

// open shows the sheet for a hero (the chosen one, else the first).
export function open(className) {
  const list = heroes();
  if (!list.length) return screen.log('No hero to show.');
  cls = className || (screen.st.sel && screen.st.sel.gear ? screen.st.sel.class : cls) || list[0].class;
  if (!list.some((u) => u.class === cls)) cls = list[0].class;
  render();
}

function act(a) {
  const r = referee.act(a);
  if (!r || r.refused) {
    screen.log(r ? r.refused.reason : 'The referee did not answer.');
    return false;
  }
  screen.show(r.view);
  screen.after();
  render();
  return true;
}

function row(it, buttons, extra) {
  return `<div class="srow"><img class="ic" src="${icon(it)}" alt=""><span><b style="color:${it.color}">${it.name}</b> <small>(${it.kind})</small>${it.text ? `<br><small class="afx">${it.text}</small>` : ''}${it.lore ? `<br><small><i>${it.lore}</i></small>` : ''}${extra || ''}</span><span class="btns">${buttons || ''}</span></div>`;
}

function render() {
  const v = screen.st.view, c = v.climb, w = $('gearWin');
  const u = heroes().find((x) => x.class === cls);
  if (!u) { w.hidden = true; return; }
  const g = u.gear, phase = c ? c.phase : 'battle';
  const locked = u.done || v.side !== 0 || (phase !== 'battle' && phase !== 'prepare');
  const dis = locked ? 'disabled' : '';
  let h = `<div class="panel box wide sheet"><button class="xclose" aria-label="Close">✕</button><div class="tabs">${heroes().map((x) => `<button class="${x.class === cls ? 'on' : ''}" data-h="${x.class}">${x.class}</button>`).join('')}</div>
    <div class="who"><img src="${heroPicUrl({ class: u.class, stage: u.stage, ascension: u.ascension })}" alt=""><div><div class="ttl">${u.class}</div>
    <small>${u.form || ''} · Lv ${u.level}${u.ascension ? ' · ' + u.ascension : ''}</small>${g.capped ? ' <span class="capt">CAPPED</span>' : ''}<br>
    <small>HP ${u.hp}/${u.maxHp} · ATK ${u.atk} · DEF ${u.def} · MOV ${u.mov}</small><br><small class="weap">${g.weapons}</small></div></div>`;
  h += `<h4>Weapon sets</h4><div class="btns">` + (g.sets || []).map((s, i) => `<button class="b ${s.inHand ? 'on' : ''}" data-set="${i + 1}" ${s.inHand || locked || (g.sets || []).length < 2 ? 'disabled' : ''}>Set ${i + 1}: ${s.weapons}${s.rank ? ' (' + s.rank + ')' : ''}</button>`).join('') + '</div>';
  h += '<h4>Worn</h4>' + ((g.worn || []).map((it) => row(Object.assign({}, it, { kind: SLOT[it.slot] + ' · ' + it.kind }), `<button class="b" data-off="${it.slot}" ${dis}>Take off</button>`)).join('') || '<p><small>Nothing worn.</small></p>');
  if (g.bonuses) h += `<p><small>Active bonuses: ${g.bonuses}</small></p>`;
  h += '<h4>Bag</h4><p><small>Putting gear on or taking it off in battle uses up the hero’s turn.</small></p>';
  const bag = v.bag || [], order = bag.map((_, i) => i), so = { w: 0, h: 1, a: 2, b: 3, c: 4 };
  if (sortBy === 'rar') order.sort((a, b) => bag[b].tier - bag[a].tier || b - a);
  else if (sortBy === 'slot') order.sort((a, b) => so[bag[a].slot] - so[bag[b].slot] || bag[b].tier - bag[a].tier);
  else order.reverse();
  if (bag.length) h += `<div class="mrow"><span>Sort</span><div class="seg" id="bagSort">${[['new', 'Newest'], ['rar', 'Rarity'], ['slot', 'Slot']].map(([k, l]) => `<button data-v="${k}" class="${sortBy === k ? 'on' : ''}">${l}</button>`).join('')}</div></div>`;
  h += order.map((i) => {
    const it = bag[i], fit = (g.bag || [])[i] || { ok: true };
    const eq = it.slot === 'w' && !(g.freeSlots || []).length
      ? (g.weaponSlots || []).map((s) => `<button class="b" data-eq="${i}" data-slot="${s}" ${dis || !fit.ok ? 'disabled' : ''}>Into ${SLOT[s]}</button>`).join('')
      : `<button class="b" data-eq="${i}" ${dis || !fit.ok ? 'disabled' : ''}>Equip</button>`;
    const diff = Object.entries(fit.diff || {}).map(([k, d]) => `<span class="${d > 0 ? 'up' : 'dn'}">${k.toUpperCase()} ${d > 0 ? '+' : ''}${d} ${d > 0 ? '▲' : '▼'}</span>`).join(' ');
    const note = !fit.ok ? `<br><small class="why">${fit.why}</small>` : diff ? `<br><small class="cmp">vs worn: ${diff}</small>` : `<br><small class="cmp">${fit.worn ? 'same stats as worn' : 'empty slot'}</small>`;
    return row(it, eq + `<button class="b" data-drop="${i}">Drop</button>`, note);
  }).join('') || '<p><small>The bag is empty.</small></p>';
  const bar = (u.skills || []).filter((s) => s.name !== 'Strike' && !s.ultimate), ult = (u.skills || []).find((s) => s.ultimate);
  h += `<h4>Skills</h4><p><small>${bar.map((s) => s.name).join(', ')}${ult ? ' · ★ ' + ult.name : ''}</small></p><button class="b" id="skOpen">Skills ▸</button>`;
  const hr = c && c.heroes.find((x) => x.class === cls);
  if (hr && hr.canAscend && phase !== 'battle') h += '<h4>Ascension</h4><p><small>Change once per floor: learned skills stay; the ultimate switches.</small></p><div class="btns">' + hr.canAscend.map((a) => `<button class="b" data-asc="${a}">${a}</button>`).join('') + '</div>';
  h += '<button class="b" id="gearClose">Close</button></div>';
  w.innerHTML = h;
  w.hidden = false;
  const q = (s) => w.querySelectorAll(s);
  q('#bagSort button').forEach((b) => (b.onclick = () => { sortBy = b.dataset.v; render(); }));
  q('[data-h]').forEach((b) => (b.onclick = () => { cls = b.dataset.h; render(); }));
  q('[data-set]').forEach((b) => (b.onclick = () => act({ kind: 'set', unit: u.id, set: +b.dataset.set })));
  q('[data-off]').forEach((b) => (b.onclick = () => act({ kind: 'unequip', unit: u.id, slot: b.dataset.off })));
  q('[data-eq]').forEach((b) => (b.onclick = () => act({ kind: 'equip', unit: u.id, item: +b.dataset.eq, slot: b.dataset.slot || '' })));
  q('[data-drop]').forEach((b) => (b.onclick = () => { const it = v.bag[+b.dataset.drop]; if (confirm(`Drop ${it.name}? It will be gone for good.`)) act({ kind: 'drop', item: +b.dataset.drop }); }));
  q('[data-asc]').forEach((b) => (b.onclick = () => act({ kind: 'ascend', hero: cls, choice: b.dataset.asc })));
  if ($('skOpen')) $('skOpen').onclick = () => { w.hidden = true; screen.skills(cls); };
  $('gearClose').onclick = () => { w.hidden = true; };
  w.querySelector('.xclose').onclick = () => $('gearClose').click();
}
