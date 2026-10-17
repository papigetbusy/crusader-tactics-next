// Versus on one phone (online step 2, owner signed off the mockup 2026-10-16):
// each player picks 8 heroes and their ascensions, then they take turns
// handing the phone over. The screen only shows the clocks and counts the
// time used; the referee keeps the real clocks and decides when time is up.
import * as referee from './referee.js';
import * as playback from './playback.js';
import { heroPicUrl } from './team.js';
import { cover } from './sysmenu.js';
import { isVersus, onlineSeat } from './side.js';
import { think } from './online.js';

const $ = (id) => document.getElementById(id);
const NAME = ['Player 1 · Blue', 'Player 2 · Red'], SHORT = ['BLUE', 'RED'], COL = ['#4fa8ff', '#ff5d5d'];
let screen = null, sendIt = null, rules = null, teams = [[], []];
let pending = 0, last = 0, asked = false, held = true;

export const wire = (s) => { screen = s; };
// useSend gets the battle's send, so a clock action plays like any other.
export const useSend = (f) => { sendIt = f; };

const view = () => screen && screen.st.view;
const live = () => { const v = view(); return isVersus(v) && !v.outcome; };

// ---------- the Versus menu on the title screen ----------

export function menu() {
  $('titleStart').hidden = true;
  $('vsMenu').hidden = false;
  $('vsLocal').onclick = () => { $('vsMenu').hidden = true; teams = [[], []]; pick(0); };
  $('vsOnline').onclick = () => { $('vsMenu').hidden = true; screen.lobby(); };
  $('vsBack').onclick = () => { $('vsMenu').hidden = true; $('titleStart').hidden = false; };
}

// ---------- picking 8 heroes and their ascensions ----------

function pick(side) {
  chooseTeam(side, teams[side], () => { if (side === 0) $('vsMenu').hidden = false; else pick(0); }, (team) => {
    teams[side] = team;
    if (side === 0) return handOver(1, 'Hand the phone to Player 2 to choose their heroes.', 'I\'m Player 2 · Choose', () => pick(1));
    begin();
  }, side === 0 ? 'Done → Player 2' : 'Start the match');
}

// chooseTeam is the picker: 8 heroes and an ascension each. done(team) gets
// the picks; back() leaves. Online uses it too.
export function chooseTeam(side, start, back, done, goText) {
  rules = rules || referee.versusRules();
  const w = cover('vsPick'), team = (start || []).map((p) => Object.assign({}, p));
  const draw = () => {
    const n = team.length, need = rules.team;
    const keep = (w.querySelector('.vsGrid') || {}).scrollTop || 0;
    w.innerHTML = `<div class="panel picker vsp"><div class="vsP" style="color:${COL[side]}">${NAME[side].toUpperCase()}</div>
      <h3>Choose ${need} heroes</h3><div class="vsCount">${n} / ${need} chosen · tap a hero, then its ascension</div>
      <div class="vsGrid">${rules.classes.map((c) => {
        const p = team.find((x) => x.class === c.class), a = p && c.ascensions.find((x) => x.name === p.ascension);
        const img = p ? heroPicUrl({ class: c.class, stage: 3, ascension: p.ascension }) : heroPicUrl({ class: c.class, stage: 0 });
        return `<button class="card${p ? ' on' : ''}" data-c="${c.class}"><img src="${img}" alt=""><b>${c.class}</b><small>Lv ${rules.level} · purple set</small>${p ? `<div class="asc">${c.ascensions.map((x) => `<i data-a="${x.name}" class="${x.name === p.ascension ? 'on' : ''}">${x.name}</i>`).join('')}</div><span class="pas">${a.passive}${a.bonus ? ` <em>${a.bonus}</em>` : ''}</span>` : ''}</button>`;
      }).join('')}</div>
      <div class="btns"><button class="b" id="vsPBack">Back</button><button class="b" id="vsPGo" ${n === need ? '' : 'disabled'}>${goText}</button></div></div>`;
    w.hidden = false;
    const grid = w.querySelector('.vsGrid');
    if (grid && keep) { grid.scrollTop = keep; requestAnimationFrame(() => { grid.scrollTop = keep; }); }
    w.querySelectorAll('[data-c]').forEach((b) => (b.onclick = () => {
      const cl = b.dataset.c, i = team.findIndex((x) => x.class === cl);
      if (i >= 0) team.splice(i, 1);
      else if (team.length >= need) return screen.log(`Only ${need} heroes. Tap a chosen hero to take them out first.`);
      else team.push({ class: cl, ascension: rules.classes.find((c) => c.class === cl).ascensions[0].name });
      draw();
    }));
    w.querySelectorAll('[data-a]').forEach((t) => (t.onclick = (e) => {
      e.stopPropagation();
      team.find((x) => x.class === t.closest('[data-c]').dataset.c).ascension = t.dataset.a;
      draw();
    }));
    $('vsPBack').onclick = () => { w.hidden = true; back(); };
    $('vsPGo').onclick = () => { w.hidden = true; done(team.slice()); };
  };
  draw();
}

function begin() {
  const v = referee.versus({ teams });
  if (!v || v.error) return screen.log('Could not start: ' + (v ? v.error : 'no answer'));
  pending = 0;
  screen.start(v);
  handOver(0, 'Player 1 moves first. Your heroes start at the bottom.', 'I\'m Player 1 · Start');
}

// ---------- handing the phone over ----------

function handOver(side, text, button, then) {
  const w = cover('vsHand'), v = view(), bank = v && v.versus ? v.versus.bankMs[side] : (rules ? rules.bankSeconds * 1000 : 0);
  held = true;
  w.className = 'cover vsHand';
  w.innerHTML = `<div class="vsWho" style="color:${COL[side]}">${NAME[side].toUpperCase()}</div><div class="vsBig">Your turn</div>
    <div class="vsTxt">${text}</div>${bank ? `<div class="vsBank">⏱ ${mmss(bank)} left</div><div class="vsSmall">Your clock starts when you tap.</div>` : ''}
    <button class="b" id="vsHGo">${button}</button><button class="b ghost" id="vsHTitle">Title screen</button>`;
  w.hidden = false;
  $('vsHGo').onclick = () => { w.hidden = true; held = false; last = performance.now(); if (then) then(); screen.kick(); };
  $('vsHTitle').onclick = () => { if (screen.title) screen.title(); };
}

// after looks at a played action: the match is over, or the other player's
// turn has come and the phone is handed over.
export function after(events, sideBefore) {
  const v = view();
  if (!isVersus(v)) return;
  if (v.outcome) return ended(events);
  if (v.side === sideBefore) return;
  const me = v.side, foe = 1 - me;
  screen.show(v, true); // the camera turns to this player's own heroes
  let dmg = 0;
  const lost = [];
  const byId = new Map(v.units.map((u) => [u.id, u]));
  for (const e of events) {
    if (e.kind === 'hit' && e.target && (byId.get(e.target) || {}).team === me) dmg += e.amount || 0;
    if (e.kind === 'died' && e.unit && !byId.get(e.unit)) lost.push(e.unit);
  }
  const timeOut = events.some((e) => e.kind === 'turn-time');
  const said = (timeOut ? `${SHORT[foe][0] + SHORT[foe].slice(1).toLowerCase()}'s turn time ran out. ` : '') +
    (dmg ? `Last turn your heroes took ${dmg} damage.` : 'Last turn nobody hit your heroes.') + (lost.length ? ` You lost ${lost.length} hero${lost.length > 1 ? 'es' : ''}.` : '');
  handOver(me, `Hand the phone to ${NAME[me].split(' · ')[0]}. ${said} Your heroes are at the bottom.`, `I'm ${NAME[me].split(' · ')[0]} · Start`);
}

function ended(events) {
  const v = view(), win = v.versus.winner, time = events.some((e) => e.kind === 'time-up');
  const w = cover('vsHand');
  held = true;
  w.className = 'cover vsHand';
  w.innerHTML = `<div class="vsWho" style="color:${COL[win]}">${NAME[win].toUpperCase()}</div><div class="vsBig">Wins!</div>
    <div class="vsTxt">${time ? `${NAME[1 - win].split(' · ')[0]} ran out of time.` : `${NAME[1 - win].split(' · ')[0]} has no heroes left.`} · ${v.turn} rounds</div>
    <button class="b" id="vsAgain">Rematch</button><button class="b" id="vsTitle">Title screen</button>`;
  w.hidden = false;
  $('vsAgain').onclick = () => { w.hidden = true; begin(); };
  $('vsTitle').onclick = () => { if (screen.title) screen.title(); else location.reload(); };
}

// ---------- the clocks ----------

// stamp adds the time used since the last accepted action.
export function stamp(a) {
  if (!live()) return a;
  return Object.assign({}, a, { ms: Math.round(pending) });
}
// charged is called once the referee took an action: that time is counted.
export function charged() { pending = 0; asked = false; }

const mmss = (ms) => { const s = Math.max(0, Math.ceil(ms / 1000)); return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`; };

// tick runs every frame: count the time while the player to move can play,
// draw the clocks, and tell the referee when a clock runs out.
export function tick(now) {
  const bar = $('vsClock'), v = view();
  if (!isVersus(v)) { if (bar) bar.hidden = true; return; }
  const net = onlineSeat() !== null; // online the server keeps time; this only shows it
  const counting = live() && (net || (!playback.playing() && !held && $('vsHand') && $('vsHand').hidden));
  if (counting && !net) pending += now - (last || now);
  last = now;
  const c = v.versus, side = v.side;
  const spend = net ? think() : pending;
  const leftBank = (s) => Math.max(0, s === side ? c.bankMs[s] - spend : c.bankMs[s]);
  if (bar) {
    bar.hidden = false;
    const chip = $('chip');
    if (chip) bar.style.right = Math.max(8, innerWidth - chip.getBoundingClientRect().left + 6) + 'px';
    bar.innerHTML = [0, 1].map((s) => {
      const on = s === side && !v.outcome, left = leftBank(s);
      const turn = c.turnMs > 0 && on ? `<div class="tb"><i style="width:${(100 * Math.max(0, c.turnMs - c.usedMs - spend)) / c.turnMs}%;${c.turnMs - c.usedMs - spend < 10000 ? 'background:#ff5d5d' : ''}"></i></div><div class="tt">turn ${mmss(Math.max(0, c.turnMs - c.usedMs - spend))}</div>` : `<div class="tt">${on ? 'your time' : 'waiting'}</div>`;
      return `<div class="vsc${on ? ' on' : ''}" style="--c:${COL[s]}"><div class="who">${SHORT[s]}</div><div class="t">${mmss(left)}</div>${turn}</div>`;
    }).join('');
  }
  if (counting && !net && !asked && sendIt && leftBank(side) <= 0) {
    asked = true;
    sendIt({ kind: 'clock' });
  }
}
