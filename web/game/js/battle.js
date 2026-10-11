// The battle controls. The page keeps only the referee's last answer (the
// View) and what the player is pointing at; every move, attack and turn is
// an action sent to the referee, and its answer is drawn.
import * as referee from './referee.js';
import { tileAt, unitAt, closeUp, focusOn } from './camera.js';
import { feet } from './board.js';
import * as team from './team.js';
import * as shop from './shop.js';
import * as playback from './playback.js';
import { autosave } from './sysmenu.js';

const $ = (id) => document.getElementById(id);
let screen = null; // { st, show, kick, log }
let seen = ''; // which floor and step the screen last showed (to frame a new floor once)

export function wire(s) {
  screen = s;
  const c = $('map');
  c.addEventListener('click', tap);
  $('endTurn').onclick = endTurn;
  $('startBattle').onclick = () => send({ kind: 'start' }, 'Your turn. Tap a hero.');
  $('summaryGo').onclick = () => send({ kind: 'continue' }, '', autosave);
  $('skip').onclick = () => { playback.skip(); $('skip').hidden = true; };
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
    screen.show(r.view);
    if (message) screen.log(message);
    after();
    if (then) then();
  };
  playback.play(st().units, r.events || [], st().opt.speed, finish, () => {
    screen.log('Enemy turn.');
    $('skip').hidden = st().opt.speed === 3;
  });
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
  s.reach = new Set();
  s.targets = new Set();
  s.zone = null;
  s.picks = null;
  s.picking = null;
  $('aimPanel').hidden = true;
  $('menu').hidden = true;
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
  seen = here;
  $('endTurn').hidden = p !== 'battle' || !!v.outcome;
  decision();
  summary();
  if (p === 'prepare') screen.log(`Floor ${v.climb.floor}: look at the map, tap Team to choose heroes, tap a hero then a blue tile to place them, then Start battle.`);
  if (p === 'shop' && $('shopWin').hidden) shop.open();
  const readyHeroes = st().units.filter(ready);
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
  if (sel && u && u.team !== sel.team) {
    const slot = sel.skills.findIndex((sk, i) => !sk.wait && !sk.heal && !sk.support && !sk.revive && referee.targets(sel.id, i).includes(u.id));
    if (slot >= 0) return aimWith(slot, u);
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
  s.reach = new Set((u.canMove ? referee.reach(u.id) : []).map(([x, y]) => key(x, y)));
  screen.log(u.canMove ? 'Tap a blue tile to move, tap an enemy to attack, or tap the hero again for commands.' : 'Choose a command.');
  if (!u.canMove) menu();
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

// look shows where a monster can walk next turn (the referee's answer).
function look(u) {
  const s = st();
  s.look = u;
  s.zone = { walk: new Set(referee.threat(u.id).map(([x, y]) => key(x, y))), hit: new Set() };
  screen.log(`${u.kind || u.class} · Lv ${u.level} · HP ${u.hp}/${u.maxHp} · ATK ${u.atk} · DEF ${u.def} · MOV ${u.mov}`);
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
  const s = st(), u = s.sel, m = $('menu');
  if (!u) return;
  m.innerHTML = '';
  const row = (text, f, off) => {
    const b = document.createElement('button');
    b.className = 'row';
    b.textContent = text;
    b.disabled = !!off;
    b.onclick = (e) => { e.stopPropagation(); f(); };
    m.append(b);
  };
  if (u.ability) {
    const ab = u.ability;
    row('✦ ' + ab.name + (ab.wait ? ` (${ab.wait})` : ''), () => ability(u), ab.wait > 0);
  }
  if (!u.acted) {
    u.skills.forEach((sk, i) => {
      const multi = MULTI.has(sk.aim);
      const ask = multi ? referee.skillAim(u.id, i, {}) : null;
      const n = sk.wait ? 0 : multi ? (ask.units || []).length + (ask.tiles || []).length : referee.targets(u.id, i).length;
      const name = (sk.ultimate ? '★ ' : '') + (sk.name === 'Strike' ? 'Attack' : sk.name);
      row(name + (sk.wait ? ` (${sk.wait})` : ''), () => (multi ? pick(u, sk.name, sk.text, { kind: 'attack', unit: u.id, skill: i, tiles: [] }, (a) => referee.skillAim(u.id, i, a)) : aimWith(i)), !n);
    });
    if (u.potions > 0 && u.hp < u.maxHp) row(`Item: Potion ×${u.potions}`, () => send({ kind: 'potion', unit: u.id }));
  }
  if (u.canUndo) row('Undo Move', () => send({ kind: 'undo', unit: u.id }, '', () => select(st().units.find((x) => x.id === u.id))));
  if (!u.acted) row('Defend', () => send({ kind: 'defend', unit: u.id }));
  row('Wait', () => send({ kind: 'wait', unit: u.id }));
  row('Cancel', () => { clear(); screen.kick(); });
  m.hidden = false;
  screen.kick();
}

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
  const ids = referee.targets(u.id, slot);
  if (!ids.length) return screen.log('No target in reach.');
  s.skill = slot;
  s.reach = new Set();
  s.targets = new Set(ids);
  s.aim = target || s.units.find((x) => x.id === ids[0]);
  $('menu').hidden = true;
  forecast();
}

// forecast shows what the attack will do, from the referee's preview.
function forecast() {
  const s = st(), u = s.sel, v = s.aim, sk = u.skills[s.skill], F = $('forecast');
  const f = referee.preview(u.id, s.skill, v.id);
  const name = v.kind || v.class;
  let line = '—';
  if (f) {
    if (sk.heal) line = `Heal +${f.heal || f.mid}`;
    else {
      const dmg = f.low === f.high ? f.mid : `${f.low}–${f.high}`;
      line = `Damage ${dmg} · Hit ${f.hitChance}%${f.mult > 1 ? ' · WEAK' : f.mult < 1 ? ' · RESIST' : ''}`;
    }
  }
  F.className = 'panel ' + (v.team !== u.team ? 'red' : 'green');
  F.innerHTML = `<div class="fh">${name} · Lv ${v.level}</div><div>HP <b>${v.hp}</b>/${v.maxHp} · DEF ${v.def}</div><div class="fd">${line}</div>
    <div class="fs">${sk.ultimate ? '★ ' : ''}${sk.name}${sk.text ? ' · ' + sk.text : ''}</div>
    <div class="fr"><button class="b" id="fNo">✕</button><button class="b red" id="fGo">${sk.heal ? '✚ Heal' : v.team !== u.team ? '⚔ Attack' : '✦ Cast'}</button></div>`;
  F.hidden = false;
  $('fNo').onclick = () => { clear(); screen.kick(); };
  $('fGo').onclick = attack;
  screen.log('Tap the target again or press the button.');
  screen.kick();
}

function attack() {
  const s = st();
  send({ kind: 'attack', unit: s.sel.id, skill: s.skill, target: s.aim.id });
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
