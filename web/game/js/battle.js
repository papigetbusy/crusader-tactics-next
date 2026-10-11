// The battle controls. The page keeps only the referee's last answer (the
// View) and what the player is pointing at; every move, attack and turn is
// an action sent to the referee, and its answer is drawn.
import * as referee from './referee.js';
import { tileAt, unitAt, closeUp } from './camera.js';
import { feet } from './board.js';

const $ = (id) => document.getElementById(id);
let screen = null; // { st, show, kick, log }

export function wire(s) {
  screen = s;
  const c = $('map');
  c.addEventListener('click', tap);
  $('endTurn').onclick = endTurn;
  $('startBattle').onclick = () => send({ kind: 'start' }, 'Your turn. Tap a hero.');
  $('summaryGo').onclick = () => send({ kind: 'continue' });
}

const st = () => screen.st;
const view = () => st().view;
const phase = () => (view().climb ? view().climb.phase : 'battle');
const ready = (u) => u && u.team === 0 && !u.done && !u.companion && !u.decoy && !u.pet;
const key = (x, y) => x + ',' + y;

// ---------- sending an action ----------

// send asks the referee to do something and draws its answer.
function send(action, message) {
  const r = referee.act(action);
  if (!r) return false;
  if (r.refused) {
    screen.log(r.refused.reason);
    return false;
  }
  sparks(r.events || []);
  clear();
  screen.show(r.view);
  if (message) screen.log(message);
  after();
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
  $('endTurn').hidden = p !== 'battle' || !!v.outcome;
  decision();
  summary();
  if (p === 'prepare') screen.log(`Floor ${v.climb.floor}: tap a hero, then a blue tile to place them. Tap Start battle when ready.`);
  if (p === 'shop') send({ kind: 'continue' }, 'The shop is not on this screen yet: on to the floor.');
  const readyHeroes = st().units.filter(ready);
  $('endTurn').classList.toggle('pulse', p === 'battle' && !readyHeroes.length);
  screen.kick();
}

// ---------- tapping the map ----------

function tap(e) {
  const s = st(), v = view();
  if (!v || $('decision').hidden === false || $('summary').hidden === false) return;
  const t = tileAt(e.clientX, e.clientY), u = unitAt(s.units, e.clientX, e.clientY);
  const sel = s.sel;
  if (phase() === 'prepare') return place(t, u);
  if (phase() !== 'battle' || v.outcome) return;
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
    return send({ kind: 'move', unit: sel.id, x: t[0], y: t[1] }) && reselect(sel.id);
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
  if (!u.acted) {
    u.skills.forEach((sk, i) => {
      const n = sk.wait ? 0 : referee.targets(u.id, i).length;
      const name = (sk.ultimate ? '★ ' : '') + (sk.name === 'Strike' ? 'Attack' : sk.name);
      row(name + (sk.wait ? ` (${sk.wait})` : ''), () => aimWith(i), !n);
    });
    if (u.potions > 0 && u.hp < u.maxHp) row(`Item: Potion ×${u.potions}`, () => send({ kind: 'potion', unit: u.id }));
  }
  if (u.canUndo) row('Undo Move', () => send({ kind: 'undo', unit: u.id }) && select(st().units.find((x) => x.id === u.id)));
  if (!u.acted) row('Defend', () => send({ kind: 'defend', unit: u.id }));
  row('Wait', () => send({ kind: 'wait', unit: u.id }));
  row('Cancel', () => { clear(); screen.kick(); });
  m.hidden = false;
  screen.kick();
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

function endTurn() {
  const left = st().units.filter(ready).length;
  if (left && !confirm(`End turn? ${left} hero${left > 1 ? "es haven't" : " hasn't"} acted yet.`)) return;
  send({ kind: 'endturn' }, 'Your turn.');
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

// ---------- sparks: numbers that float up from the referee's events ----------

const pops = [];
function sparks(events) {
  const now = performance.now();
  let delay = 0;
  for (const e of events) {
    const who = st().units.find((u) => u.id === (e.target || e.unit));
    if (!who) continue;
    const at = feet(who.x, who.y);
    const add = (txt, col, big) => pops.push({ x: at.x, y: at.y - 72, t0: now + delay, txt, col, big });
    if (e.kind === 'hit') add(e.crit ? 'CRIT ' + e.amount : e.amount + (e.mult > 1 ? '!' : ''), e.crit ? '#ff9a2e' : e.mult > 1 ? '#ffd84a' : e.mult < 1 ? '#c9c9d6' : '#fff', e.crit || e.mult > 1);
    else if (e.kind === 'missed') add('MISS', '#c9c9d6');
    else if (e.kind === 'healed' || e.kind === 'potion') add('+' + e.amount, '#6ee7a0');
    else if (e.kind === 'ticked') add(String(e.amount), '#ff9a9a');
    else continue;
    delay += 120;
  }
}

// drawSparks draws the floating numbers (called by the scene each frame).
export function drawSparks(g, now) {
  for (let i = pops.length; i--; ) if (now > pops[i].t0 + 1000) pops.splice(i, 1);
  for (const o of pops) {
    const p = (now - o.t0) / 1000;
    if (p < 0) continue;
    g.save();
    g.globalAlpha = p > 0.7 ? (1 - p) / 0.3 : 1;
    g.font = `bold ${o.big ? 32 : 26}px ui-monospace,Menlo,monospace`;
    g.textAlign = 'center';
    g.lineWidth = 6;
    g.strokeStyle = '#1a0f2a';
    g.lineJoin = 'round';
    const y = o.y - p * 30 + (p < 0.15 ? (1 - p / 0.15) * 10 : 0);
    g.strokeText(o.txt, o.x, y);
    g.fillStyle = o.col;
    g.fillText(o.txt, o.x, y);
    g.restore();
  }
  return pops.length > 0;
}
