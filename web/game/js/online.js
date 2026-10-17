// Online versus (online step 3, owner 2026-10-16): a room password shared by
// up to 4 friends. Two take Blue and Red and play; everyone watches; the rest
// wait in line, and the winner stays on. The server runs the referee and keeps
// the clocks. This phone keeps its own copy of the match only to answer "where
// can I walk" and the attack window; every move goes to the server and comes
// back on the room's stream for everyone.
import * as referee from './referee.js';
import * as playback from './playback.js';
import { cover } from './sysmenu.js';
import { setSeat, onlineSeat } from './side.js';
import { chooseTeam, charged } from './versus.js';
import { playOut, refresh } from './battle.js';

const $ = (id) => document.getElementById(id);
const NAMES = ['Blue', 'Red'], COL = ['#4fa8ff', '#ff5d5d'];
// The server: the page's own when it is served by the game server; from
// GitHub Pages, the Render service (or ?server=... to try another).
const SERVER = new URLSearchParams(location.search).get('server') ||
  (location.hostname.endsWith('github.io') ? 'https://crusader-tactics.onrender.com' : '');
let screen = null, me = null, room = null, es = null, inMatch = false, queue = [], busy = false, mine = null, fails = 0;
let seen = 0, inflight = false, clock = null, pendingClock = null;

export const wire = (s) => { screen = s; };
export const on = () => inMatch && onlineSeat() !== null;
// think is how much of the bank the server would charge right now.
export function think() {
  if (!clock) return 0;
  return Math.max(0, clock.thinkMs + (performance.now() - clock.at));
}
function noteClock(m) {
  if (m.thinkMs == null) return;
  if (m.seq != null && m.seq < seen) return;
  clock = { thinkMs: m.thinkMs, at: performance.now(), seq: m.seq || 0 };
}
const store = (k, v) => { try { if (v == null) localStorage.removeItem(k); else localStorage.setItem(k, v); } catch (e) { /* private mode */ } };
const read = (k) => { try { return localStorage.getItem(k) || ''; } catch (e) { return ''; } };

async function post(path, body) {
  try {
    const r = await fetch(SERVER + path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    return await r.json();
  } catch (e) {
    return { error: 'Could not reach the game server. Check your internet and try again.' };
  }
}

// ---------- joining ----------

export function lobby() {
  const w = cover('olJoin');
  w.className = 'cover olw';
  w.innerHTML = `<div class="panel box wide"><div class="ttl">Online</div>
    <p>Make up a room password and send it to your friends. Up to 4 people per room: two play, the others watch and wait their turn. The winner stays on.</p>
    <label class="olf">Room password<input id="olRoom" autocomplete="off" autocapitalize="none" maxlength="32" value="${read('ctn-room')}"></label>
    <label class="olf">Your name<input id="olName" maxlength="16" value="${read('ctn-name')}"></label>
    <div class="btns"><button class="b" id="olBack">Back</button><button class="b" id="olGo">Join room</button></div><p id="olMsg" class="why"></p></div>`;
  w.hidden = false;
  $('olBack').onclick = () => { w.hidden = true; $('vsMenu').hidden = false; };
  $('olGo').onclick = async () => {
    const pw = $('olRoom').value.trim(), name = $('olName').value.trim();
    if (pw.length < 3) return ($('olMsg').textContent = 'Use a room password of at least 3 letters or numbers.');
    store('ctn-room', pw);
    store('ctn-name', name);
    $('olGo').disabled = true;
    const slow = setTimeout(() => { $('olMsg').textContent = 'Waking the server up. The first join can take up to a minute.'; }, 2500);
    const r = await post('/api/join', { room: pw, name });
    clearTimeout(slow);
    $('olGo').disabled = false;
    if (r.error) return ($('olMsg').textContent = r.error);
    me = r.token;
    store('ctn-token', me);
    w.hidden = true;
    listen();
  };
}

function listen() {
  if (es) { es.onerror = null; es.close(); }
  es = new EventSource(SERVER + '/api/stream?token=' + me);
  es.onmessage = (e) => {
    fails = 0;
    let m;
    try { m = JSON.parse(e.data); } catch (err) { return; }
    if (m.type === 'clock') { if (!busy && !playback.playing()) noteClock(m); return; }
    if (m.type === 'result') queue.push(m);
    else { room = m; if (!busy && !playback.playing()) noteClock(m); }
    pump();
  };
  es.onerror = () => {
    if (es && es.readyState !== EventSource.CLOSED) return;
    if (++fails < 4) { setTimeout(listen, 1000); return; }
    lost();
  };
}

function lost() {
  if (es) es.close();
  const w = cover('olRoomWin');
  w.className = 'cover olw';
  w.innerHTML = `<div class="panel box"><div class="ttl">Connection lost</div><p>The room could not be reached. Your seat is still yours if you get back in.</p>
    <div class="btns"><button class="b" id="olRe">Try again</button><button class="b" id="olT">Title screen</button></div></div>`;
  w.hidden = false;
  $('olRe').onclick = () => { w.hidden = true; fails = 0; resume(); };
  $('olT').onclick = () => { store('ctn-token', null); location.reload(); };
}

// resume is the same person coming back after a refresh. A dead token is forgotten.
export function resume() {
  const token = read('ctn-token'), pw = read('ctn-room');
  if (!token || !pw) return;
  post('/api/join', { room: pw, name: read('ctn-name'), token }).then((r) => {
    if (r.error || !r.token) {
      if (r.error && /not in this room|full/i.test(r.error)) store('ctn-token', null);
      return;
    }
    me = r.token;
    store('ctn-token', me);
    fails = 0;
    listen();
  });
}

// ---------- what comes from the room, in order ----------

// pump handles the next message once the last one has finished playing.
function pump() {
  if (busy || playback.playing()) return;
  const r = queue.shift();
  if (r) return result(r);
  if (room) draw();
}

function result(r) {
  if (!inMatch || !r.match || (r.seq && r.seq <= seen)) { inflight = false; pump(); return; }
  const missed = r.seq > seen + 1;
  seen = r.seq || seen;
  inflight = false;
  sync(r.match.save);
  if (missed) {
    mine = null;
    noteClock(r);
    refresh(r.match.view);
    charged();
    pump();
    return;
  }
  const yours = mine && r.sideBefore === onlineSeat() ? mine : null;
  mine = null;
  pendingClock = { thinkMs: r.thinkMs || 0, seq: r.seq || seen };
  busy = true;
  charged();
  playOut({ events: r.events, view: r.match.view }, yours ? yours.message : '', yours ? yours.then : null, r.sideBefore);
}

// played is called by the battle once an answer has finished playing.
export function played(r) {
  busy = false;
  if (pendingClock) { noteClock(pendingClock); pendingClock = null; }
  charged();
  const v = r.view;
  if (v.outcome) {
    const win = v.versus.winner, time = (r.events || []).some((e) => e.kind === 'time-up');
    screen.log(`${NAMES[win]} wins${time ? ' on time' : ''}!`);
  } else if (v.side === onlineSeat()) screen.log('Your turn!');
  pump();
}

// sync loads the server's match into this phone's own copy, for the questions.
function sync(save) {
  const text = typeof save === 'string' ? save : JSON.stringify(save);
  const r = referee.load(text);
  if (r && r.error) screen.log('Could not follow the match: ' + r.error);
}

// ---------- the room ----------

function draw() {
  if (room.playing && room.match) {
    $('olRoomWin') && ($('olRoomWin').hidden = true);
    if (!inMatch || room.matches !== draw.n) {
      draw.n = room.matches;
      inMatch = true;
      seen = room.seq || 0;
      setSeat(room.yourSide);
      sync(room.match.save);
      screen.start(room.match.view);
      charged();
      screen.log(room.yourSide >= 0 ? `You are ${NAMES[room.yourSide]}. ${room.match.view.side === room.yourSide ? 'Your turn!' : 'Waiting for ' + NAMES[1 - room.yourSide] + '.'}` : 'You are watching. Your turn comes when you reach the front of the line.');
      return;
    }
    if ((room.seq || 0) > seen && !busy && !playback.playing()) {
      seen = room.seq;
      inflight = false;
      mine = null;
      sync(room.match.save);
      refresh(room.match.view);
      charged();
    }
    return;
  }
  if (inMatch && playback.playing()) return;
  roomPanel();
}

function roomPanel() {
  const w = cover('olRoomWin'), you = room.people.find((p) => p.you) || {}, side = room.yourSide;
  w.className = 'cover olw';
  const seat = (s) => {
    const p = room.people.find((x) => x.side === s);
    const who = p ? `<b>${esc(p.name)}</b><small>${p.ready ? 'Ready' : p.hasTeam ? 'Has heroes' : 'Choosing heroes'}${p.online ? '' : ' · away'}</small>` : '<small>Empty</small>';
    const btn = !p && side < 0 ? `<button class="b" data-sit="${s}">Take ${NAMES[s]}</button>` : '';
    return `<div class="olSeat" style="--c:${COL[s]}"><div class="olSide">${NAMES[s].toUpperCase()}${room.winner === s ? ' · winner stays on' : ''}</div>${who}${btn}</div>`;
  };
  const line = room.people.filter((p) => p.line > 0).sort((a, b) => a.line - b.line);
  let todo = '';
  if (side >= 0) {
    const yourTeam = room.yourTeam && room.yourTeam.length;
    if (you.ready) todo = `<p>Ready. Waiting for ${NAMES[1 - side]}…</p>`;
    else if (yourTeam) todo = `<div class="btns"><button class="b" id="olAgain">Play with the same heroes</button><button class="b" id="olPick">Change heroes</button></div>`;
    else todo = '<div class="btns"><button class="b" id="olPick">Choose your heroes</button></div>';
    todo += '<button class="b ghost" id="olStand">Stand up</button>';
  } else todo = `<p>${you.line ? `You are number ${you.line} in line.` : ''} Take a free side to play, or wait and watch.</p>`;
  w.innerHTML = `<div class="panel box wide"><div class="ttl">Room “${esc(room.room)}”</div>
    <p><small>${room.people.length} of ${room.max} people · send this password to friends so they can join</small></p>
    ${room.matches && room.winner >= 0 ? `<p class="olWin" style="color:${COL[room.winner]}">${NAMES[room.winner]} won the last match.</p>` : ''}
    <div class="olSeats">${seat(0)}${seat(1)}</div>${todo}
    <h4>Waiting in line</h4>${line.map((p) => `<p>${p.line}. ${esc(p.name)}${p.you ? ' (you)' : ''}${p.online ? '' : ' · away'}</p>`).join('') || '<p><small>Nobody.</small></p>'}
    <button class="b" id="olLeave">Leave the room</button></div>`;
  w.hidden = false;
  w.querySelectorAll('[data-sit]').forEach((b) => (b.onclick = () => call('/api/seat', { side: +b.dataset.sit })));
  if ($('olStand')) $('olStand').onclick = () => call('/api/seat', { side: -1 });
  if ($('olAgain')) $('olAgain').onclick = () => call('/api/again', {});
  if ($('olPick')) $('olPick').onclick = () => {
    w.hidden = true;
    chooseTeam(side, room.yourTeam || [], () => { w.hidden = false; }, (team) => { w.hidden = false; call('/api/team', { team }); }, 'Ready');
  };
  $('olLeave').onclick = async () => { store('ctn-token', null); if (es) es.close(); await post('/api/leave', { token: me }); location.reload(); };
}

async function call(path, body) {
  const r = await post(path, Object.assign({ token: me }, body));
  if (r.error) screen.log(r.error);
}

const esc = (t) => String(t).replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);

// ---------- moves ----------

// act sends this player's move to the server. Its answer arrives on the
// stream for everyone; message and then are used when this phone plays it.
export function act(action, message, then) {
  const v = screen.st.view;
  if (!v || v.side !== onlineSeat() || v.outcome) { screen.log('It is not your turn.'); return false; }
  if (inflight) return false;
  mine = { message, then };
  inflight = true;
  post('/api/act', { token: me, action }).then((r) => {
    if (r.error) { inflight = false; mine = null; screen.log(r.error); }
  });
  return true;
}
