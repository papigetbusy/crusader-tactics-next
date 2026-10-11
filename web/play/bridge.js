// Crusader Tactics: the old game's screen, wired to the new referee.
//
// play.html is the old game's page (v0.26.1) with its art moved to files.
// This file runs after the old scripts. It starts the Go referee and swaps
// the old game's rules for questions to the referee: every move, attack and
// turn is decided by Go, and the old screen only draws what Go says.
(() => {
  'use strict';
  const CT = { V: null, objs: new Map(), party: new Map(), playing: false };
  window.CT = CT;
  const J = (s) => JSON.parse(s);
  const now = () => performance.now();
  const speed = () => (window.SPDX ? window.SPDX() : 1);
  const wait = (ms) => new Promise((r) => setTimeout(r, ms / speed()));

  // ---------- starting the referee ----------
  async function boot() {
    const go = new Go();
    const url = 'app.wasm?v=' + (window.CT_BUILD || '');
    let r;
    try { r = await WebAssembly.instantiateStreaming(fetch(url), go.importObject); }
    catch (e) { const res = await fetch(url); r = await WebAssembly.instantiate(await res.arrayBuffer(), go.importObject); }
    go.run(r.instance);
    CT.ready = true;
  }
  CT.boot = boot();

  // ---------- talking to the referee ----------
  function act(a) {
    const r = J(ctApply(JSON.stringify(a)));
    if (r.refused) { log(r.refused.reason); return null; }
    if (r.problem) console.warn(r.problem);
    CT.V = r.view;
    return r.events || [];
  }
  CT.act = act;
  const vunit = (id) => CT.V && CT.V.units.find((u) => u.id === id);
  const phase = () => (CT.V && CT.V.climb ? CT.V.climb.phase : 'battle');
  const obj = (id) => CT.objs.get(id);

  // ---------- the referee's state, in the old game's shapes ----------
  const RARN = { common: 0, uncommon: 1, rare: 2, epic: 3, legendary: 4 };
  function formOf(level, asc) { return asc ? 3 : level >= 20 ? 2 : level >= 5 ? 1 : 0; }
  function skillObj(s) {
    const old = (typeof SK !== 'undefined' && SK[s.name]) || {};
    return Object.assign({}, old, { n: s.name, d: s.text || old.d, min: s.minRange, max: s.maxRange, pw: s.power, el: s.element || old.el,
      heal: !!s.heal || undefined, sup: !!s.support || undefined, rev: !!s.revive || undefined, area: s.area || undefined,
      t: s.ultimate ? 'u' : 'n', _aim: s.aim || '', _free: !!s.free });
  }
  function partyObj(cls) {
    const row = CT.V.climb && CT.V.climb.heroes.find((h) => h.class === cls);
    let p = CT.party.get(cls);
    if (!p) { p = { c: cls, gear: {}, bag: { potion: 3 }, su: {}, learned: [], eq: [] }; CT.party.set(cls, p); }
    if (row) Object.assign(p, { lv: row.level, xp: row.xp, spec: row.ascension || null, form: formOf(row.level, row.ascension) });
    return p;
  }
  function sync() {
    const V = CT.V;
    if (!V) return;
    if (N !== V.n) { setN(V.n); startCam(); }
    TER = V.ground.map((r) => r.split(''));
    M = V.heights.map((r) => r.slice());
    BIO = BIOMES.find((b) => b.n === V.biome) || BIOMES[0];
    CHS = (V.chests || []).map((c) => ({ x: c.x, y: c.y, t: c.kind }));
    FIELDS = []; TRAPS = []; OBJ = null; TURN = V.turn;
    const c = V.climb;
    floor = c ? c.floor : V.floor || 1; best = c ? c.best : floor; gold = c ? c.gold : V.gold || 0;
    const seen = new Set(), list = [];
    for (const u of V.units) {
      let o = CT.objs.get(u.id);
      if (!o || o.c !== u.class) { o = { id: u.id }; CT.objs.set(u.id, o); }
      seen.add(u.id);
      const hero = u.team === 0 && !u.companion && !u.decoy && !u.pet;
      const m = u.kind && typeof MNAMES !== 'undefined' ? MNAMES.indexOf(u.kind) : -1;
      Object.assign(o, { c: u.class, team: u.team, x: u.x, y: u.y, hp: u.hp, max: u.maxHp, atk: u.atk, def: u.def, mov: u.mov, lv: u.level,
        e: u.element || '', st: Object.assign({}, u.statuses || {}), done: u.done, moved: !u.canMove && !u.done ? true : u.done && !u.canMove,
        acted: false, boss: !!u.boss, elite: !!u.elite, comp: !!u.companion, decoy: !!u.decoy, summ: !!u.pet, ally: !!u.ally,
        m: m >= 0 ? m : null, spec: u.ascension || null, sks: (u.skills || []).map(skillObj), cd: (u.skills || []).map((s) => s.wait || 0),
        shield: u.shield || 0, abcd: u.ability ? u.ability.wait : 0, undo: u.canUndo ? { x: u.x, y: u.y } : null });
      if (u.boss || u.elite) o.simg = (typeof IMX !== 'undefined' && IMX[u.class]) || o.simg;
      if (hero) { o.ref = partyObj(u.class); o.ref.bag.potion = u.potions; o.form = o.ref.form; }
      list.push(o);
    }
    for (const id of [...CT.objs.keys()]) if (!seen.has(id)) CT.objs.delete(id);
    U = list;
    P.length = 0;
    if (c) c.team.forEach((cls) => P.push(partyObj(cls)));
    else U.filter((o) => o.ref).forEach((o) => P.push(o.ref));
    if (sel && !U.includes(sel)) sel = null;
    if (insp && !U.includes(insp)) insp = null;
    if (sel && sel.done) { sel = null; mode = null; reach = []; tg = []; }
  }
  CT.sync = sync;

  // ---------- playing the referee's events with the old effects ----------
  const ELFX = { fire: 'fire', ice: 'ice', elec: 'bolt', light: 'light', dark: 'dark' };
  function fxFor(u, name) {
    const s = (typeof SK !== 'undefined' && SK[name]) || {};
    return fxKey(name) || s.vfx || FXN[name] || ELFX[s.el] || 'slash';
  }
  function pop(o, txt, col, big, delay = 0) {
    if (!o || !M[o.y]) return;
    const b = spos(o);
    POP.push({ x: b.x, y: b.y - 72, t0: now() + delay, txt, col, big });
  }
  // show plays one event and says how long to wait before the next one.
  function show(e, delay) {
    const t = now() + delay;
    switch (e.kind) {
      case 'moved': {
        const o = obj(e.unit);
        if (!o || !e.path || e.path.length < 2) return 0;
        const dur = Math.min(1500, 150 * (e.path.length - 1)) / speed();
        const end = e.path[e.path.length - 1];
        o.mv = { t0: now(), dur, path: e.path, n: e.path.length - 1 };
        o.x = end[0]; o.y = end[1];
        kick();
        return dur;
      }
      case 'skill': {
        const u = obj(e.unit), v = obj(e.target) || u;
        if (!u) return 0;
        const k = fxFor(u, e.skill);
        setA();
        const [sx, sy] = mvv(v.x - u.x, v.y - u.y), l = Math.hypot(sx, sy);
        u.at = { t0: t, dx: l ? sx / l : 0, dy: l ? sy / l : -0.6, k: 'lunge', c: '#c79bff' };
        if (v !== u) spawn(k, u, v, '', '#fff', 0, 1);
        return impOf(k);
      }
      case 'hit': {
        const v = obj(e.target);
        if (!v) return 0;
        let col = '#fff', big = 0, txt = '' + e.amount;
        if (e.mult > 1) { col = '#ffd84a'; big = 1; txt += '!'; } else if (e.mult && e.mult < 1) col = '#c9c9d6';
        if (e.crit) { txt = 'CRIT ' + e.amount; col = '#ff9a2e'; big = 1; }
        pop(v, txt, col, big, delay);
        v.sh = v.hp; v.sht = t; v.ht = t; v.fl = t;
        v.hp = Math.max(0, v.hp - (e.amount || 0));
        return 120;
      }
      case 'missed': pop(obj(e.target), 'MISS', '#c9c9d6', 0, delay); return 120;
      case 'healed': { const v = obj(e.target); if (v) { pop(v, '+' + e.amount, '#6ee7a0', 0, delay); v.hp = Math.min(v.max, v.hp + (e.amount || 0)); } return 120; }
      case 'ticked': { const v = obj(e.unit); if (v) { pop(v, '' + e.amount, '#c58bff', 0, delay); v.hp = Math.max(0, v.hp - (e.amount || 0)); } return 200; }
      case 'died': { const o = obj(e.unit); if (o) { GH.push({ u: o, t0: t, fade: 400 }); U = U.filter((z) => z !== o); } return 300; }
      case 'gold': pop(obj(e.unit), '+' + e.amount + 'g', '#ffd84a', 0, delay + 250); return 0;
      case 'loot': add('Loot: ' + e.skill + '!'); return 0;
      case 'turn-start': if (e.side === 1) log('Enemy turn.'); return 0;
      default: return 0;
    }
  }
  // playNow shows a player's action at once, its hits a moment later.
  function playNow(evs) {
    let d = 0;
    for (const e of evs) d += e.kind === 'hit' || e.kind === 'missed' || e.kind === 'healed' ? 0 : show(e, d);
    let hd = d;
    for (const e of evs) if (e.kind === 'hit' || e.kind === 'missed' || e.kind === 'healed') hd += show(e, d) * 0;
    kick();
    return hd;
  }
  CT.playNow = playNow;
  // playSlow shows events one after another (the enemy turn).
  async function playSlow(evs) {
    for (const e of evs) {
      const d = show(e, 0);
      kick();
      if (d) await wait(e.kind === 'moved' ? d * speed() : d);
    }
  }

  // ---------- the old game's rules, answered by the referee ----------
  window.bfs = function (u) {
    if (!u || !CT.V) return [];
    if (phase() === 'prepare') return (CT.V.starts || []).map((t) => [t.x, t.y]).filter(([x, y]) => !at(x, y) || at(x, y) === u);
    return J(ctReach(u.id)) || [];
  };
  window.bfsPath = (u, x, y) => [[u.x, u.y], [x, y]];
  window.tgt = function (u, s) {
    if (!u || !CT.V || !u.sks) return [];
    const i = u.sks.indexOf(s);
    if (i < 0) return [];
    return (J(ctTargets(u.id, i)) || []).map(obj).filter(Boolean);
  };
  window.est = function (u, s, v) {
    const i = u.sks.indexOf(s);
    const f = i >= 0 ? J(ctPreview(u.id, i, v.id)) : null;
    if (!f) return { t: '', h: 0, c: '#fff' };
    if (s.heal) return { t: '+' + f.mid, h: 100, c: '#6ee7a0' };
    return { t: f.low === f.high ? '' + f.mid : f.low + '–' + f.high, h: f.hitChance, c: f.mult > 1 ? '#ffd84a' : f.mult < 1 ? '#c9c9d6' : '#fff', d: f.mid };
  };
  window.startMove = function (u, x, y) {
    if (!u || (u.x === x && u.y === y)) return;
    if (phase() === 'prepare') {
      if (!act({ kind: 'place', hero: u.c, x, y })) return;
      const from = [u.x, u.y];
      sync();
      const o = U.find((z) => z.c === u.c && z.ref);
      if (o) { o.mv = { t0: now(), dur: 300, path: [from, [x, y]], n: 1 }; }
      sel = null; mode = null; reach = []; tg = [];
      kick();
      return;
    }
    const evs = act({ kind: 'move', unit: u.id, x, y });
    if (!evs) return;
    playNow(evs);
    sync();
  };
  window.undoMove = function () {
    MENUON = false; SUBOPEN = false;
    if (!sel || busy) return;
    const evs = act({ kind: 'undo', unit: sel.id });
    if (!evs) return;
    playNow(evs.map((e) => (e.kind === 'undone' ? Object.assign({}, e, { kind: 'moved' }) : e)));
    sync();
    if (sel) moveMode();
    log('Move undone. Drag or tap to a blue tile to move again.');
    ui(); draw();
  };
  window.act = function (u, s, v) {
    const i = u.sks.indexOf(s);
    const evs = act({ kind: 'attack', unit: u.id, skill: i, target: v && v.id ? v.id : 0 });
    if (!evs) { ui(); draw(); return; }
    const d = playNow(evs);
    setTimeout(() => { sync(); after(); ui(); draw(); }, Math.max(0, d) + 400 / speed());
    mode = null; cur = null; ptg = null; reach = []; tg = [];
    ui(); draw();
  };
  function simple(kind) {
    return function () {
      if (!sel || busy) return;
      const evs = act({ kind, unit: sel.id });
      if (!evs) return;
      playNow(evs);
      sync();
      sel = null; mode = null; reach = []; tg = [];
      after(); ui(); draw();
    };
  }
  window.defend = simple('defend');
  window.wait = simple('wait');
  window.usePotion = simple('potion');
  window.endPhase = () => {};
  window.startPhase = () => {};
  window.check = () => {};
  window.pairDone = () => {};
  window.openChest = () => {};
  window.gain = () => {};
  window.reinforce = () => {};
  window.enemyPhase = async function () {
    if (CT.playing) return;
    CT.playing = true; busy = true; ENEMYPH = true;
    sel = null; reach = []; tg = []; mode = null; ui();
    const evs = act({ kind: 'endturn' });
    if (evs) {
      sync0Positions(evs);
      await playSlow(evs);
    }
    sync();
    ENEMYPH = false; busy = false; CT.playing = false;
    log('Your turn.');
    if (CT.autosave) CT.autosave();
    after(); ui(); draw();
  };
  // sync0Positions keeps units where they stood before the enemy turn, so
  // their walks can be shown (the view already has where they ended).
  function sync0Positions() {}

  // ---------- after an action: choices, the end of a battle ----------
  function after() {
    const V = CT.V;
    if (!V) return;
    if (V.decision) return decide(V.decision);
    const c = V.climb;
    if (c && (c.phase === 'won' || c.phase === 'lost')) return summary();
    if (c && c.phase === 'shop') return shop();
  }
  CT.after = after;
  function decide(d) {
    const w = ovl('ctdec');
    w.innerHTML = `<div class="cap-box dA-panel"><div class="cap-t">${d.title || 'Choose'}</div><p class="muted">${d.text || ''}</p><div id="ctdeco"></div></div>`;
    w.hidden = false;
    const box = w.querySelector('#ctdeco');
    (d.options || []).forEach((o) => {
      const b = mkb(box, o.label, () => { w.hidden = true; const evs = act({ kind: 'decide', choice: o.choice, skill: o.slot || 0 }); if (evs) playNow(evs); sync(); after(); ui(); draw(); });
      if (o.text) { const p = document.createElement('div'); p.className = 'dsc'; p.textContent = o.text; box.append(p); }
      b.style.margin = '4px';
    });
  }
  function summary() {
    const c = CT.V.climb, s = c.summary || {};
    const w = ovl('ctsum');
    const rows = (s.heroes || []).map((x) => `<div class="sm-h"><div><b>${x.class}</b><br><span class="${x.to > x.from ? 'sm-up' : ''}">Lv ${x.from}${x.to > x.from ? ' → ' + x.to + ' ▲' : ''}</span> · +${x.xp} XP${x.capped ? ' · <span class="sm-cap">CAPPED</span>' : ''}</div></div>`).join('');
    w.innerHTML = `<div class="cap-box dA-panel sm-box"><div class="cap-t">${s.won ? 'Floor ' + s.floor + ' cleared' : 'Defeated on floor ' + s.floor}</div>
      <p class="muted">${s.won ? (s.newBest ? 'New highest floor!' : s.replayed ? 'Replay of a cleared floor (half gold).' : '') : 'Everything earned is kept. Level up, gear up, try again.'} · ${s.turns || 0} turns</p>
      <div class="sm-list">${rows}</div><div class="sm-loot"><b>Loot</b> ${(s.loot || []).map((i) => `<span style="color:${i.color}">${i.name}</span>`).join(', ') || '<span class="muted">none this floor</span>'}</div>
      <div class="sm-gold"><b>Gold</b> +${s.gold || 0} <span class="muted">(${c.gold} total)</span></div>${s.gift ? `<p>Gift: <b style="color:${s.gift.color}">${s.gift.name}</b></p>` : ''}
      <button class="b" id="ctgo">Continue to Floor ${c.floor}</button></div>`;
    w.hidden = false;
    w.querySelector('#ctgo').onclick = () => { w.hidden = true; act({ kind: 'continue' }); arrive(); };
  }
  function shop() { CT.shopUI && CT.shopUI(); }

  // ---------- a floor: look at the map, choose and place the team ----------
  function arrive() {
    if (phase() === 'shop') return shop();
    CT.objs.clear();
    sync();
    startCam();
    sel = null; insp = null; mode = null; reach = []; tg = [];
    $('party').innerHTML = '';
    buildChips();
    bar();
    if (CT.autosave) CT.autosave();
    log(`Floor ${floor}: choose your team, then drag heroes onto the blue tiles. Tap Start when ready.`);
    ui(); draw();
  }
  CT.arrive = arrive;
  function bar() {
    let b = $('ctbar');
    if (!b) { b = document.createElement('div'); b.id = 'ctbar'; document.body.append(b); }
    const p = phase();
    b.innerHTML = '';
    if (p === 'prepare') {
      mkb(b, 'Team', teamUI);
      mkb(b, '⚔ Start battle', () => { const evs = act({ kind: 'start' }); if (!evs) return; CT.objs.clear(); sync(); bar(); log('Your turn. Drag a hero to a blue tile, or tap it.'); ui(); draw(); }).classList.add('gold');
    }
    b.hidden = p !== 'prepare';
  }
  CT.bar = bar;
  function teamUI() {
    const c = CT.V.climb, w = ovl('ctteam');
    w.innerHTML = `<div class="cap-box dA-panel"><div class="cap-t">Choose ${c.slots} heroes</div><p class="muted">Only heroes who fight on a floor earn its XP.</p><div id="ctteaml"></div><button class="b" id="ctteamok">Use this team</button> <button class="b" id="ctteamx">Back</button></div>`;
    const l = w.querySelector('#ctteaml');
    c.heroes.forEach((h) => {
      const d = document.createElement('label');
      d.className = 'ctrow';
      d.innerHTML = `<input type="checkbox" value="${h.class}" ${h.inTeam ? 'checked' : ''}> <b>${h.class}</b> <span class="muted">${h.form} · Lv ${h.level}${h.ascension ? ' · ' + h.ascension : ''}</span>`;
      l.append(d);
    });
    l.querySelectorAll('input').forEach((cb) => (cb.onchange = () => { if (l.querySelectorAll('input:checked').length > c.slots) cb.checked = false; }));
    w.hidden = false;
    w.querySelector('#ctteamx').onclick = () => (w.hidden = true);
    w.querySelector('#ctteamok').onclick = () => {
      const t = [...l.querySelectorAll('input:checked')].map((x) => x.value);
      if (t.length !== c.slots) { alert(`Choose ${c.slots} heroes.`); return; }
      w.hidden = true;
      if (act({ kind: 'team', heroes: t })) arrive();
    };
  }

  // ---------- the title screen: a new climb, or load ----------
  window.newFloor = () => arrive();
  window.showEv = () => {};
  window.save = () => {};
  window.autoSave = () => {};
  window.startGame = async function () {
    await CT.boot;
    $('title').hidden = true;
    const s = crypto.getRandomValues(new Uint32Array(3));
    CT.V = J(ctClimb(s[2], s[0], s[1]));
    arrive();
  };
  CT.boot.then(() => {
    $('tnew').onclick = () => window.startGame();
  });

  // ---------- look ----------
  const css = document.createElement('style');
  css.textContent = `#ctbar{position:fixed;top:calc(env(safe-area-inset-top) + 56px);left:50%;transform:translateX(-50%);z-index:30;display:flex;gap:6px}
  #ctbar .b{padding:8px 14px}.ctrow{display:block;padding:6px 0;border-bottom:1px solid rgba(255,255,255,.08)}.ctrow input{transform:scale(1.3);margin-right:8px}`;
  document.head.append(css);
})();
