// The battle controls. The page keeps only the referee's last answer (the
// View) and what the player is pointing at; every move, attack and turn is
// an action sent to the referee, and its answer is drawn.
import * as referee from './referee.js';
import { tileAt, unitAt, closeUp, focusOn, lookAt, cam, safeTop } from './camera.js';
import { TS } from './look.js';
import { active, label, STATUS } from './statuses.js';
import { picture } from './units.js';
import { feet, map as map0, top as top0 } from './board.js';
import * as team from './team.js';
import * as shop from './shop.js';
import * as playback from './playback.js';
import { check as capCheck } from './capnotice.js';
import { autosave } from './sysmenu.js';

const $ = (id) => document.getElementById(id);
let screen = null; // { st, show, kick, log }
let seen = '';
let rules = null; // the fixed rules guide (terrain, element wheel, weapon triangle), for words on screen
let lastReady = 0; // ready heroes last time, to move the camera on after one acts // which floor and step the screen last showed (to frame a new floor once)

export function wire(s) {
  screen = s;
  const c = $('map');
  c.addEventListener('click', tap);
  dragging(c);
  $('endTurn').onclick = endTurn;
  $('startBattle').onclick = () => send({ kind: 'start' }, 'Your turn. Tap a hero.');
  $('summaryGo').onclick = () => send({ kind: 'continue' }, '', autosave);
  $('skip').onclick = () => { playback.skip(); $('skip').hidden = true; };
  rules = referee.guide();
  $('teamButton').onclick = () => team.open(view(), (v) => { clear(); screen.show(v); after(); }, screen.log);
}

const st = () => screen.st;
const view = () => st().view;
const phase = () => (view().climb ? view().climb.phase : 'battle');
const ready = (u) => u && u.team === 0 && !u.done && !u.companion && !u.decoy && !u.pet;
const key = (x, y) => x + ',' + y;

// ---------- sending an action ----------

// send asks the referee to do something; its events play, then its answer
// is shown.
function send(action, message, then) {
  if (playback.playing()) return false;
  const r = referee.act(action);
  if (!r) return false;
  if (r.refused) {
    screen.log(r.refused.reason);
    return false;
  }
  clear();
  $('endTurn').hidden = true;
  const finish = () => {
    $('skip').hidden = true;
    const before = st().units;
    screen.show(r.view);
    if (message) screen.log(message);
    after();
    capCheck(r.events || [], r.view, before);
    if (then) then();
  };
  playback.play(st().units, r.events || [], st().opt.speed, finish, () => {
    screen.log('Enemy turn.');
    $('skip').hidden = st().opt.speed === 3;
  }, st().opt.calm);
  screen.kick();
  return true;
}

// clear forgets what the player was pointing at.
function clear() {
  const s = st();
  s.sel = null;
  s.look = null;
  s.aim = null;
  s.skill = -1;
  s.set = 0;
  s.reach = new Set();
  s.targets = new Set();
  s.zone = null;
  s.attack = null;
  s.picks = null;
  s.picking = null;
  $('aimPanel').hidden = true;
  $('menu').hidden = true;
  $('sub').hidden = true;
  $('einfo').hidden = true;
  stopDrag();
  $('terr').hidden = true;
  if (view()) $('endTurn').hidden = phase() !== 'battle' || !!view().outcome || playback.playing();
  $('forecast').hidden = true;
}

// after shows whatever the new View asks for: a choice, the summary, the
// start of a floor, or the next turn.
export function after() {
  const v = view(), p = phase();
  const s = st();
  s.starts = p === 'prepare' ? new Set((v.starts || []).map((t) => key(t.x, t.y))) : null;
  $('prepare').hidden = p !== 'prepare';
  // a new floor: show the blue starting tiles above the buttons
  const here = p + ':' + (v.climb ? v.climb.floor + ':' + v.turn : '');
  if (p === 'prepare' && seen !== here && (v.starts || []).length) {
    const xs = v.starts.map((t) => t.x), ys = v.starts.map((t) => t.y);
    const f = feet((Math.min(...xs) + Math.max(...xs)) / 2, (Math.min(...ys) + Math.max(...ys)) / 2);
    requestAnimationFrame(() => focusOn(f.x, f.y - 40, $('prepare').getBoundingClientRect().top - 20));
  }
  if (p === 'battle' && seen.startsWith('prepare')) {
    const foes = st().units.filter((u) => u.team === 1), lv = Math.max(...foes.map((u) => u.level), 1);
    screen.log(`${v.throne ? 'Seize the throne or defeat every enemy' : 'Defeat every enemy'} to climb. ${foes.length} enemies · Lv ${lv} · ${v.n}×${v.n} ${v.biome || ''}${(v.chests || []).length ? ` · ${v.chests.length} chest${v.chests.length > 1 ? 's' : ''} on the map: walk onto them!` : ''}`);
  }
  seen = here;
  $('endTurn').hidden = p !== 'battle' || !!v.outcome;
  decision();
  summary();
  if (p === 'prepare') screen.log(`Floor ${v.climb.floor}: look at the map, tap Team to choose heroes, tap a hero then a blue tile to place them, then Start battle.`);
  if (p === 'shop' && $('shopWin').hidden) shop.open();
  const readyHeroes = st().units.filter(ready);
  if (p === 'battle' && !v.outcome && lastReady > readyHeroes.length && readyHeroes.length && !s.sel) {
    const h = readyHeroes[0];
    const f = feet(h.x, h.y);
    lookAt(f.x, f.y - 40);
    screen.log(`Next: ${h.class}. Tap them to give orders.`);
  }
  lastReady = readyHeroes.length;
  $('endTurn').classList.toggle('pulse', p === 'battle' && !readyHeroes.length);
  // the danger zone: every tile a monster could walk to next turn
  s.danger = s.opt.danger && p === 'battle' && !v.outcome
    ? new Set(st().units.filter((u) => u.team === 1).flatMap((u) => referee.threat(u.id).map(([x, y]) => key(x, y))))
    : null;
  if (s.opt.autoEnd && p === 'battle' && !v.outcome && !readyHeroes.length && st().units.some((u) => u.team === 1))
    setTimeout(() => { if (!playback.playing() && !st().units.some(ready)) send({ kind: 'endturn' }, 'Your turn.', autosave); }, 800);
  screen.kick();
}

// ---------- tapping the map ----------

function tap(e) {
  if (cam.dragged) { cam.dragged = false; return; } // the end of a pan, not a tap
  const s = st(), v = view();
  if (!v || playback.playing() || $('decision').hidden === false || $('summary').hidden === false) return;
  const t = tileAt(e.clientX, e.clientY), u = unitAt(s.units, e.clientX, e.clientY);
  const sel = s.sel;
  if (phase() === 'prepare') return place(t, u);
  if (phase() !== 'battle' || v.outcome) return;
  if (s.picking) return tapPick(t, u);
  // aiming: tap a highlighted target
  if (sel && s.skill >= 0) {
    if (u && s.targets.has(u.id)) {
      if (s.aim === u) return attack();
      s.aim = u;
      return forecast();
    }
    clear();
    return screen.kick();
  }
  // a hero is chosen: walk to a blue tile, or tap an enemy in reach
  if (sel && t && s.reach.has(key(t[0], t[1])) && !u) {
    return send({ kind: 'move', unit: sel.id, x: t[0], y: t[1] }, '', () => reselect(sel.id));
  }
  if (sel && u && u.team !== sel.team && !u.ally) {
    const plan = referee.plan(sel.id, u.id);
    if (plan && plan.walk) {
      const id = sel.id, foe = u.id;
      return send({ kind: 'move', unit: id, x: plan.walk.x, y: plan.walk.y }, '', () => {
        const me = st().units.find((x) => x.id === id), them = st().units.find((x) => x.id === foe);
        if (me && them) { st().sel = me; aimWith(plan.slot, them); }
      });
    }
    if (plan) return aimWith(plan.slot, u);
  }
  if (u && u === sel) return menu();
  if (ready(u)) return select(u);
  clear();
  if (u) look(u);
  screen.kick();
}

function select(u) {
  clear();
  const s = st();
  s.sel = u;
  $('endTurn').hidden = true; // as in the old game, End Turn hides while a hero is chosen
  s.attack = new Set(referee.attackTiles(u.id).map(([x, y]) => key(x, y)));
  if (s.opt.autoZoom) closeUp(u.x, u.y);
  s.reach = new Set((u.canMove ? referee.reach(u.id) : []).map(([x, y]) => key(x, y)));
  screen.log(u.canMove ? 'Tap a blue tile to move, tap an enemy to attack, or tap the hero again for commands.' : 'Choose a command.');
  if (!u.canMove) menu();
  terrainBox(u);
  closeUp(u.x, u.y);
  screen.kick();
}

// reselect keeps a hero chosen after it moved, with its commands open.
function reselect(id) {
  const u = st().units.find((x) => x.id === id);
  if (!u || u.done) return;
  st().sel = u;
  menu();
  screen.kick();
}

// look shows a monster's info window and where it can walk next turn
// (purple) [old einfo L1521-1522].
function look(u) {
  const s = st();
  s.look = u;
  s.zone = { walk: new Set(referee.threat(u.id).map(([x, y]) => key(x, y))), hit: new Set(referee.threatHit(u.id).map(([x, y]) => key(x, y))) };
  const hp = u.team && !u.ally ? '#e5484d' : '#4cd97b', im = picture(u);
  const src = im ? (im.toDataURL ? im.toDataURL() : im.src) : '';
  const E = { fire: 'Fire', ice: 'Ice', elec: 'Thunder', light: 'Light', dark: 'Dark' }, W = { sword: 'Sword', axe: 'Axe', lance: 'Lance' };
  const weak = [];
  const by = (m, x) => Object.keys(m).find((k) => m[k] === x);
  if (u.element && rules) weak.push(`${E[u.element]} type (weak to ${E[by(rules.beats, u.element)] || '?'})`);
  if (u.weapon && W[u.weapon] && rules) weak.push(`${W[u.weapon]} (weak to ${W[by(rules.weaponBeats, u.weapon)] || '?'})`);
  const st2 = active(u);
  $('einfo').innerHTML = `<div class="ei-top"><div class="ei-por" style="background-image:url(${src})"></div><div class="ei-mid">
    <div class="ei-nm">${u.kind || u.class} · Lv ${u.level}${u.boss ? ' · BOSS' : u.elite ? ' · ELITE' : ''}</div>
    <div class="ei-hb"><i style="width:${(100 * u.hp) / u.maxHp}%;background:${hp}"></i></div>
    <div>HP ${u.hp}/${u.maxHp} · ATK ${u.atk} · DEF ${u.def} · MOV ${u.mov}</div></div></div>
    ${weak.length ? `<div class="ei-wk">${weak.join(' · ')}</div>` : ''}
    ${st2.length ? `<div class="pills">${st2.map((q) => `<span class="pill ${STATUS[q].bad ? 'd' : 'b'}">${label(u, q)} · ${u.statuses[q]}</span>`).join('')}</div>` : ''}
    ${u.team ? `<div class="ei-sk">${(u.skills || []).map((k) => k.name).join(' · ')}</div>` : ''}`;
  $('einfo').hidden = false;
}

// terrainBox shows what the ground under the chosen hero does [old terr L1523-1524].
function terrainBox(u) {
  const g = (view().ground[u.y] || '')[u.x] || 'g', T = rules && rules.terrain.find((t) => t.letter === g);
  if (!T) return;
  $('terr').innerHTML = `<div class="tn">${T.name}</div><div>${T.damageTaken < 1 ? 'DMG −' + Math.round((1 - T.damageTaken) * 100) + '%' : 'No bonus'}${T.heal ? ' · heals' : ''}</div>`;
  $('terr').hidden = false;
}

function place(t, u) {
  const s = st();
  if (u && u.team === 0) {
    s.sel = u;
    screen.log(`Tap a blue tile to place the ${u.class}.`);
    return screen.kick();
  }
  if (s.sel && t && s.starts.has(key(t[0], t[1]))) {
    const id = s.sel.class;
    send({ kind: 'place', hero: id, x: t[0], y: t[1] });
    s.starts = new Set((view().starts || []).map((q) => key(q.x, q.y)));
    return;
  }
  s.sel = null;
  screen.kick();
}

// ---------- the command menu ----------

function menu() {
  const s = st(), u = s.sel, m = $('menu'), sub = $('sub');
  if (!u) return;
  $('endTurn').hidden = true;
  m.innerHTML = '';
  sub.innerHTML = '';
  const row = (box, text, f, off, cls) => {
    const b = document.createElement('button');
    b.className = 'row' + (cls ? ' ' + cls : '');
    b.textContent = text;
    b.disabled = !!off;
    b.onclick = (e) => { e.stopPropagation(); f(b); };
    box.append(b);
    return b;
  };
  // [old dAbuild L1513-1518]: Attack, Skills ▸, ✦ ability, Item, Move, Undo Move, Defend, Wait, Cancel
  if (!u.acted) {
    u.skills.forEach((sk, i) => {
      const multi = MULTI.has(sk.aim);
      const ask = multi && !sk.wait ? referee.skillAim(u.id, i, {}) : null;
      const n = sk.wait ? 0 : multi ? (ask.units || []).length + (ask.tiles || []).length : referee.targets(u.id, i).length;
      const go = () => (multi ? pick(u, sk.name, sk.text, { kind: 'attack', unit: u.id, skill: i, tiles: [] }, (a) => referee.skillAim(u.id, i, a)) : aimWith(i));
      if (sk.name === 'Strike') return row(m, 'Attack', go, !n);
      row(sub, (sk.ultimate ? '★ ' : '') + sk.name + (sk.wait ? ` (${sk.wait})` : ''), go, !n);
    });
    if (sub.children.length) row(m, 'Skills ▸', (b) => { sub.hidden = !sub.hidden; b.classList.toggle('on', !sub.hidden); place2(); });
  }
  if (u.ability) row(m, '✦ ' + u.ability.name + (u.ability.wait ? ` (${u.ability.wait})` : ''), () => ability(u), u.ability.wait > 0, 'abl');
  if (!u.acted && u.potions > 0 && u.hp < u.maxHp) row(m, `Item: Potion ×${u.potions}`, () => send({ kind: 'potion', unit: u.id }));
  if (u.canMove && !s.reach.size) row(m, 'Move', () => select(u));
  if (u.canUndo) row(m, 'Undo Move', () => send({ kind: 'undo', unit: u.id }, '', () => select(st().units.find((x) => x.id === u.id))));
  if (!u.acted) row(m, 'Defend', () => send({ kind: 'defend', unit: u.id }));
  row(m, 'Wait', () => send({ kind: 'wait', unit: u.id }));
  row(m, 'Cancel', () => { clear(); screen.kick(); });
  m.hidden = false;
  sub.hidden = true;
  s.reach = new Set();
  s.attack = new Set(referee.attackTiles(u.id).map(([x, y]) => key(x, y)));
  place2();
  screen.kick();
}

// place2 puts the menu beside the hero, and the skills on the other side
// [old dApos L1528-1530].
function place2() {
  const s = st(), m = $('menu'), sub = $('sub');
  if (!s.sel || m.hidden) return;
  const b = feet(s.sel.x, s.sel.y), x = cam.offX + b.x * cam.k, y = cam.offY + b.y * cam.k;
  const W = innerWidth, H = innerHeight, w = m.offsetWidth || 150, h = m.offsetHeight || 200, off = TS * cam.k * 0.62, top0 = safeTop() + 52;
  let left = x + off, side = 1, top;
  if (left + w > W - 6) { left = x - off - w; side = -1; }
  if (left < 6) {
    side = 0;
    left = Math.max(6, Math.min(W - w - 6, x - w / 2));
    top = y - TS * cam.k * 1.15 - h - 6;
    if (top < top0) top = y + TS * cam.k * 0.15 + 6;
  } else top = y - TS * cam.k * 0.5 - h / 2;
  top = Math.max(top0, Math.min(H - h - 8, top));
  m.style.left = left + 'px';
  m.style.top = top + 'px';
  if (!sub.hidden) {
    const w2 = sub.offsetWidth || 170, h2 = sub.offsetHeight || 200;
    let l2 = side > 0 ? left + w + 6 : side < 0 ? left - w2 - 6 : left + w + 6 + w2 <= W - 6 ? left + w + 6 : left - w2 - 6;
    if (l2 + w2 > W - 6 || l2 < 6) l2 = side > 0 ? Math.max(6, left - w2 - 6) : Math.min(W - w2 - 6, left + w + 6);
    sub.style.left = l2 + 'px';
    sub.style.top = Math.max(top0, Math.min(H - h2 - 8, top)) + 'px';
  }
}
export const placeMenu = place2;

// ---------- aiming in several taps ----------

// Skills aimed at tiles, or at more than one thing, one tap at a time.
const MULTI = new Set(['tile', 'line', 'rollThenFoe', 'foeThenTile', 'twoFoes', 'echo']);

// ability uses the class ability: at once, or aimed in taps.
function ability(u) {
  const ab = u.ability, t = ab.target;
  if (!t) return send({ kind: 'ability', unit: u.id });
  const a = { kind: 'ability', unit: u.id, tiles: [] };
  const ask = t.kind === 'unit'
    ? () => (a.target ? {} : { units: referee.abilityAim(u.id).units || [] })
    : () => ((a.tiles || []).length < (t.n || 1) ? { tiles: referee.abilityAim(u.id).tiles || [] } : {});
  pick(u, ab.name, ab.text, a, ask);
}

// pick asks the referee what can be picked next, lets the player tap it, and
// sends the action once nothing more is needed.
function pick(u, title, text, action, ask) {
  const s = st();
  $('menu').hidden = true;
  const next = () => {
    const o = ask(action) || {}, units = o.units || [], tiles = (o.tiles || []).filter(([x, y]) => !(action.tiles || []).some((q) => q[0] === x && q[1] === y));
    if (!units.length && !tiles.length) {
      $('aimPanel').hidden = true;
      // nothing picked yet and nothing to pick: say so instead of asking the referee
      if (!action.target && !(action.tiles || []).length && first) { clear(); screen.kick(); return screen.log('No target in reach.'); }
      return send(action);
    }
    first = false;
    s.reach = new Set();
    s.targets = new Set(units);
    s.picks = tiles.length ? { all: new Set(tiles.map(([x, y]) => key(x, y))), chosen: new Set((action.tiles || []).map(([x, y]) => key(x, y))) } : null;
    s.picking = { action, next };
    const P = $('aimPanel');
    P.innerHTML = `<div class="fh">✦ ${title}</div><div class="fs">${units.length ? 'Tap a highlighted unit.' : 'Tap a highlighted tile (gold).'} ${text || ''}</div><div class="fr"><button class="b" id="aimNo">✕ Cancel</button></div>`;
    P.hidden = false;
    $('aimNo').onclick = () => { clear(); screen.kick(); };
    screen.kick();
  };
  let first = true;
  s.sel = u;
  next();
}

// tapPick handles a tap while picking.
function tapPick(t, u) {
  const s = st(), p = s.picking, a = p.action;
  if (u && s.targets.has(u.id)) {
    if (a.target) a.second = u.id;
    else a.target = u.id;
    return p.next();
  }
  if (t && s.picks && s.picks.all.has(key(t[0], t[1]))) {
    a.tiles = (a.tiles || []).concat([[t[0], t[1]]]);
    return p.next();
  }
  screen.log(s.picks ? 'Tap a gold tile.' : 'Tap a highlighted unit.');
}

function aimWith(slot, target) {
  const s = st(), u = s.sel;
  const ids = referee.targets(u.id, slot, s.set || 0).length ? referee.targets(u.id, slot, s.set || 0) : referee.targets(u.id, slot);
  if (!ids.length) return screen.log('No target in reach.');
  if (!referee.targets(u.id, slot, s.set || 0).length) s.set = 0;
  s.skill = slot;
  s.reach = new Set();
  s.attack = null;
  $('sub').hidden = true;
  s.targets = new Set(ids);
  s.aim = (target && ids.includes(target.id) && target) || s.units.find((x) => x.id === ids[0]);
  $('menu').hidden = true;
  forecast();
}

// forecast shows what the attack will do, from the referee's preview: the
// target before → after, damage, hit chance, a skill switcher, ◀ ▶ to change
// target, ⇄ to change weapon set [old dAforecast L1532-1549, v24strip L2081].
function forecast() {
  const s = st(), u = s.sel, v = s.aim, sk = u.skills[s.skill], F = $('forecast');
  const ids = [...s.targets], idx = ids.indexOf(v.id);
  const f = referee.preview(u.id, s.skill, v.id, s.set || 0);
  const foe = v.team !== u.team, dmg = !(sk.heal || sk.support || sk.revive);
  let after = '', line = '—';
  if (f && dmg) {
    const lo = Math.max(0, v.hp - f.high), hi = Math.max(0, v.hp - f.low);
    after = lo === hi ? '' + lo : lo + '–' + hi;
    line = `Dmg <b class="fd">${f.low === f.high ? f.mid : f.low + '–' + f.high}</b> · Hit <b>${f.hitChance}%</b> ${f.mult > 1 ? '<b class="weak">WEAK</b>' : f.mult < 1 ? '<b class="resist">RESIST</b>' : ''}`;
  } else if (f) line = `${sk.heal ? 'Heal' : 'Effect'} <b class="fd">${sk.heal ? '+' + (f.heal || f.mid) : sk.name}</b>`;
  const ko = after && /^0(–|$)/.test(after) ? '<b class="ko">KO</b>' : '';
  const sets = (u.gear && u.gear.sets) || [], hand = sets.findIndex((x) => x.inHand), set = s.set || hand + 1;
  const weapons = sets.length ? (sets[set - 1] || sets[0]).weapons : '';
  const pills = active(v).map((q) => `<span class="pill ${STATUS[q].bad ? 'd' : 'b'}">${label(v, q)} · ${v.statuses[q]}</span>`).join('');
  F.className = 'panel ' + (foe ? 'red' : 'green');
  F.innerHTML = `<div class="fh">${v.kind || v.class} · Lv ${v.level}${v.boss ? ' · Boss' : v.elite ? ' · Elite' : ''}</div>
    <div>HP <b>${v.hp}</b>${after ? ` → <b class="af">${after}</b>` : ''} ${ko}<span class="dim"> · DEF ${v.def}</span></div>
    <div>${line}</div><div class="fs">${sk.ultimate ? '★ ' : ''}${sk.name === 'Strike' ? 'Attack' : sk.name}${weapons ? ' · ' + weapons : ''}${sk.text ? ' · ' + sk.text : ''}</div>
    <div class="you">${u.class} · HP ${u.hp}/${u.maxHp}</div>${pills ? `<div class="pills">${pills}</div>` : ''}
    <div class="skills" id="fSkills"></div><div class="fr" id="fRow"></div>`;
  const sw = $('fSkills');
  u.skills.forEach((k, i) => {
    if (k.wait || MULTI.has(k.aim) || !(foe ? !(k.heal || k.support || k.revive) : k.heal || k.support || k.revive)) return;
    if (!referee.targets(u.id, i).length) return;
    const b = document.createElement('button');
    b.textContent = (k.ultimate ? '★ ' : '') + (k.name === 'Strike' ? 'Attack' : k.name);
    b.className = i === s.skill ? 'on' : '';
    b.onclick = () => { if (i !== s.skill) aimWith(i, v); };
    sw.append(b);
  });
  const row = $('fRow'), btn = (t, fn, cls, off) => {
    const b = document.createElement('button');
    b.className = 'b' + (cls ? ' ' + cls : '');
    b.innerHTML = t;
    b.disabled = !!off;
    b.onclick = fn;
    row.append(b);
  };
  const cycle = (d) => { s.aim = st().units.find((x) => x.id === ids[(idx + d + ids.length) % ids.length]); const p = feet(s.aim.x, s.aim.y); lookAt(p.x, p.y - 40); forecast(); };
  if (ids.length > 1) { btn('◀', () => cycle(-1)); btn('▶', () => cycle(1)); }
  if (dmg && sets.length > 1) btn('⇄' + set, () => {
    for (let k = 1; k <= sets.length; k++) {
      const n = ((set - 1 + k) % sets.length) + 1;
      if (referee.targets(u.id, s.skill, n).length) { s.set = n; break; }
    }
    s.targets = new Set(referee.targets(u.id, s.skill, s.set));
    if (!s.targets.has(s.aim.id)) s.aim = st().units.find((x) => s.targets.has(x.id));
    forecast();
  });
  btn('✕', () => { clear(); screen.kick(); });
  btn(sk.heal ? '✚ Heal' : sk.support ? '✦ Cast' : sk.revive ? '✦ Revive' : '⚔ Attack', attack, 'red go');
  F.hidden = false;
  const a = feet(u.x, u.y), b = feet(v.x, v.y);
  requestAnimationFrame(() => focusOn((a.x + b.x) / 2, (a.y + b.y) / 2 - 40, F.getBoundingClientRect().top));
  screen.log('Tap the target again or press ⚔ Attack.');
  screen.kick();
}

function attack() {
  const s = st();
  send({ kind: 'attack', unit: s.sel.id, skill: s.skill, target: s.aim.id, set: s.set || 0 });
}

export function endTurn() {
  const left = st().units.filter(ready).length;
  if (left && !confirm(`End turn? ${left} hero${left > 1 ? "es haven't" : " hasn't"} acted yet.`)) return;
  send({ kind: 'endturn' }, 'Your turn.', autosave);
}

// ---------- choices and the floor summary ----------

function decision() {
  const d = view().decision, w = $('decision');
  w.hidden = !d;
  if (!d) return;
  w.innerHTML = `<div class="ttl">${d.skill || d.kind}</div>${d.text ? `<p>${d.text}</p>` : ''}`;
  d.options.forEach((o) => {
    const b = document.createElement('button');
    b.className = 'b';
    b.textContent = o.label;
    b.onclick = () => send({ kind: 'decide', choice: o.choice, skill: o.slot });
    w.append(b);
  });
}

function summary() {
  const c = view().climb, s = c && c.summary, w = $('summary');
  w.hidden = !(s && (c.phase === 'won' || c.phase === 'lost'));
  if (w.hidden) return;
  $('summaryText').innerHTML = `<div class="ttl">${s.won ? `Floor ${s.floor} cleared` : `Defeated on floor ${s.floor}`}</div>
    <p>${s.won ? (s.newBest ? 'New highest floor!' : 'Replay of a cleared floor (half gold).') : 'Everything earned is kept. Level up, gear up, try again.'} · ${s.turns} turns</p>
    ${s.heroes.map((h) => `<p>${h.class} · Lv ${h.from}${h.to > h.from ? ' → ' + h.to + ' ▲' : ''} · +${h.xp} XP</p>`).join('')}
    ${s.loot && s.loot.length ? `<p>Loot: ${s.loot.map((i) => i.name).join(', ')}</p>` : ''}<p>Gold +${s.gold}</p>`;
  $('summaryGo').textContent = `Continue to Floor ${c.floor}`;
}

// ---------- dragging a hero to a tile [old v0.23 L1985-2019] ----------

let drag = null, eatClick = 0;
function stopDrag() { if (drag) { drag = null; cam.held = false; st().overlayDrag = null; } }

function dragging(c) {
  const canDrag = (u) => ready(u) && u.canMove && phase() === 'battle' && !view().outcome;
  c.addEventListener('pointerdown', (e) => {
    if (drag || playback.playing() || st().skill >= 0 || st().picking || (e.pointerType === 'mouse' && e.button !== 0)) return;
    const u = unitAt(st().units, e.clientX, e.clientY);
    if (!canDrag(u)) return;
    drag = { u, id: e.pointerId, x0: e.clientX, y0: e.clientY, x: e.clientX, y: e.clientY, on: false, t: null };
    cam.held = true; // the map doesn't pan while a hero might be dragged
  });
  addEventListener('pointermove', (e) => {
    if (!drag || e.pointerId !== drag.id) return;
    drag.x = e.clientX;
    drag.y = e.clientY;
    if (!drag.on && Math.hypot(drag.x - drag.x0, drag.y - drag.y0) > 12) {
      const D = drag;
      if (st().sel !== D.u || !st().reach.size) select(D.u); // (select clears, so keep the drag)
      drag = D;
      cam.held = true;
      cam.tx = cam.x; // keep the map still under the finger
      cam.ty = cam.y;
      drag.on = true;
      $('menu').hidden = true;
      requestAnimationFrame(edge);
    }
    if (drag.on) { drag.t = tileAt(drag.x, drag.y); screen.kick(); }
  });
  const end = (e, cancel) => {
    if (!drag || e.pointerId !== drag.id) return;
    const D = drag;
    drag = null;
    cam.held = false;
    if (!D.on) return;
    eatClick = performance.now() + 350;
    const u = D.u, t = cancel ? null : tileAt(e.clientX, e.clientY);
    if (t && t[0] === u.x && t[1] === u.y) return menu();
    if (t && st().reach.has(key(t[0], t[1])) && !st().unitAt(t[0], t[1])) {
      return send({ kind: 'move', unit: u.id, x: t[0], y: t[1] }, '', () => reselect(u.id));
    }
    if (!cancel) screen.log('They can’t reach that tile. Drag to a blue tile.');
    screen.kick();
  };
  addEventListener('pointerup', (e) => end(e, false));
  addEventListener('pointercancel', (e) => end(e, true));
  addEventListener('click', (e) => { if (performance.now() < eatClick) { e.stopImmediatePropagation(); e.preventDefault(); eatClick = 0; } }, true);
  st().dragLine = (g) => drawDrag(g);
}

// edge scrolls the map while a dragged hero is near the screen's edge.
function edge() {
  if (!drag || !drag.on) return;
  if (Math.hypot(drag.x - drag.x0, drag.y - drag.y0) >= 40) {
    const m = 56, sp = 8 / cam.k;
    let dx = 0, dy = 0;
    if (drag.x < m) dx = -sp; else if (drag.x > innerWidth - m) dx = sp;
    if (drag.y < m) dy = -sp; else if (drag.y > innerHeight - m) dy = sp;
    if (dx || dy) { cam.tx += dx; cam.ty += dy; cam.x += dx; cam.y += dy; drag.t = tileAt(drag.x, drag.y); screen.kick(); }
  }
  requestAnimationFrame(edge);
}

// drawDrag draws the dashed line from the hero and the tile under the finger.
function drawDrag(g) {
  if (!drag || !drag.on || !drag.t) return;
  const u = drag.u, t = drag.t, ok = st().reach.has(key(t[0], t[1])) && !st().unitAt(t[0], t[1]), home = t[0] === u.x && t[1] === u.y;
  const a = feet(u.x, u.y), b = feet(t[0], t[1]), col = ok ? '#d9c27a' : home ? '#ffffff' : '#e5484d';
  g.save();
  g.setLineDash([7, 6]);
  g.lineCap = 'round';
  g.beginPath();
  g.moveTo(a.x, a.y - map0.foot);
  g.lineTo(b.x, b.y - map0.foot);
  g.lineWidth = 3;
  g.strokeStyle = ok ? 'rgba(217,194,122,.9)' : 'rgba(255,255,255,.45)';
  g.stroke();
  g.setLineDash([]);
  const cs = top0(t[0], t[1]);
  g.beginPath();
  cs.forEach((p, i) => (i ? g.lineTo(p[0], p[1]) : g.moveTo(p[0], p[1])));
  g.closePath();
  g.fillStyle = ok ? 'rgba(217,194,122,.32)' : home ? 'rgba(255,255,255,.16)' : 'rgba(229,72,77,.3)';
  g.fill();
  g.lineWidth = 3;
  g.strokeStyle = col;
  g.stroke();
  g.restore();
}
