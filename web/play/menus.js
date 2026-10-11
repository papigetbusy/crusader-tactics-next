// Crusader Tactics: the old screen's aiming, abilities, floor list, shop and
// saves, answered by the new referee (loaded after bridge.js).
(() => {
  'use strict';
  const J = (s) => JSON.parse(s);
  const CT = window.CT;
  const vunit = (id) => CT.V && CT.V.units.find((u) => u.id === id);
  const obj = (id) => U.find((o) => o.id === id);

  // ---------- picking targets and tiles, one tap at a time ----------
  // pick runs an aimed ability or a skill aimed in several taps: it asks the
  // referee what can be picked next, lets the player tap it, and sends the
  // action when nothing more is needed.
  function pick(u, a, ask, title, text) {
    const info = { n: title || 'Aim', d: text || '' };
    const next = () => {
      const o = ask(a);
      const units = o.units || [], tiles = o.tiles || [];
      if (!units.length && !tiles.length) return send();
      mode = 'abil'; reach = []; tg = []; MENUON = false;
      if (units.length) {
        tg = units.map(obj).filter(Boolean);
        ABT = { kind: 'unit', u, A: info, n: 1, range: 0, picked: [], fin: (v) => { if (a.target == null) a.target = v.id; else a.second = v.id; step(); } };
        log('Tap a highlighted unit.');
      } else {
        reach = tiles.slice();
        const far = tiles.map((q) => Math.abs(q[0] - u.x) + Math.abs(q[1] - u.y));
        ABT = { kind: 'tile', u, A: info, range: Math.max(...far), min: Math.min(...far), n: 1, empty: false, picked: [], fin: (pk) => {
          const t = pk[0];
          if (!tiles.some((q) => q[0] === t[0] && q[1] === t[1])) { ABT.picked = []; log('Not there: tap a blue tile.'); return; }
          a.tiles = (a.tiles || []).concat([t]); step();
        } };
        log('Tap a blue tile.');
      }
      ui(); draw();
    };
    const step = () => next();
    const send = () => {
      ABT = null; mode = null; reach = []; tg = [];
      const evs = CT.act(a);
      if (evs) { CT.playNow ? CT.playNow(evs) : 0; }
      setTimeout(() => { CT.sync(); CT.after(); ui(); draw(); }, 500);
      ui(); draw();
    };
    next();
  }
  CT.pick = pick;

  window.abOf = function (u) {
    const v = vunit(u.id);
    if (!v || !v.ability) return null;
    const ab = v.ability;
    return { n: ab.name, d: ab.text, free: !!ab.free, cd: 0, ok: () => (ab.wait > 0 ? 'Not ready yet.' : '') };
  };
  window.useAbility = function (u) {
    const v = vunit(u.id);
    if (!v || !v.ability || busy) return;
    if (v.ability.wait > 0) { log('Not ready yet.'); return; }
    if (!v.ability.target) {
      const evs = CT.act({ kind: 'ability', unit: u.id });
      if (evs && CT.playNow) CT.playNow(evs);
      setTimeout(() => { CT.sync(); CT.after(); ui(); draw(); }, 400);
      return;
    }
    const n = v.ability.target.n || 1;
    const a = { kind: 'ability', unit: u.id, tiles: [] };
    if (v.ability.target.kind === 'unit') return pick(u, a, () => (a.target == null ? { units: J(ctAim(u.id)).units || [] } : {}), v.ability.name, v.ability.text);
    pick(u, a, () => ((a.tiles || []).length < n ? { tiles: J(ctAim(u.id)).tiles || [] } : {}), v.ability.name, v.ability.text);
  };
  const oldAim = window.aimSkill;
  window.aimSkill = function (s, v) {
    if (sel && s._aim) {
      const i = sel.sks.indexOf(s);
      const a = { kind: 'attack', unit: sel.id, skill: i, tiles: [] };
      return pick(sel, a, (x) => J(ctSkillAim(sel.id, i, JSON.stringify(x))), s.n, s.d);
    }
    return oldAim(s, v);
  };

  // ---------- floor list ----------
  window.openFloors = function () {
    const c = CT.V && CT.V.climb;
    if (!c) return;
    const w = ovl('flw');
    const rows = c.floors.map((f) => `<button class="fl-r ${f.here ? 'cur' : ''}" data-f="${f.number}"><span class="fl-n"><b>Floor ${f.number}</b> <span class="muted">${f.biome || ''}</span></span><span class="fl-t">${f.here ? '<i class="t-c">here</i>' : ''}${f.top ? '<i class="t-top">top</i>' : ''}${f.cleared ? '<i class="t-ok">✓ cleared</i>' : ''}${f.boss ? `<i class="t-b">${f.bossLegendary ? 'boss · legendary taken' : 'boss'}</i>` : f.elite ? '<i class="t-e">elite</i>' : ''}${f.shop ? '<i class="t-s">shop</i>' : ''}${f.chestsLeft ? `<i class="t-ch">${f.chestsLeft} chest${f.chestsLeft > 1 ? 's' : ''} left</i>` : ''}</span></button>`);
    w.innerHTML = `<div class="cap-box dA-panel fl-box"><div class="cap-t">Floor travel</div><p class="muted">Floors keep the same map and monsters. Opened chests stay opened; replays pay half gold and full XP.</p><div class="fl-list">${rows.join('')}</div><button class="b" id="flclose">Close</button></div>`;
    w.hidden = false;
    $('flclose').onclick = () => { w.hidden = true; };
    w.querySelectorAll('.fl-r').forEach((b) => (b.onclick = () => {
      const f = +b.dataset.f;
      if (busy || f === c.floor) return;
      w.hidden = true;
      if (CT.act({ kind: 'travel', floor: f })) CT.arrive();
    }));
  };

  // ---------- shop ----------
  CT.shopUI = function () {
    const c = CT.V.climb;
    const w = ovl('ctshop');
    const tab = CT.shopTab || 'buy';
    let list = '';
    if (tab === 'buy') list = (c.shop || []).map((it, i) => row(it, `Buy ${it.price}g`, `data-buy="${i}"`, c.gold < it.price)).join('') || '<p class="muted">Sold out.</p>';
    else if (tab === 'sell') list = (c.bag || []).map((it, i) => row(it, `Sell ${it.price}g`, `data-sell="${i}"`)).join('') || '<p class="muted">Nothing to sell. Worn gear is not listed.</p>';
    else {
      const seen = {};
      (c.bag || []).forEach((it, i) => { const k = it.name + '|' + it.rarity; if (seen[k] != null && it.rarity !== 'legendary') list += row(it, 'Combine 2 →', `data-comb="${seen[k]}|${i}"`); else seen[k] = i; });
      if (!list) list = '<p class="muted">Combine needs 2 copies of the same item and rarity in the bag. The result is one rarity higher.</p>';
    }
    w.innerHTML = `<div class="cap-box dA-panel"><div class="cap-t">Shop · before floor ${c.floor}</div><p>💰 ${c.gold} gold</p>
      <div class="ct-tabs"><button class="b ${tab === 'buy' ? 'on' : ''}" data-tab="buy">Buy</button><button class="b ${tab === 'sell' ? 'on' : ''}" data-tab="sell">Sell</button><button class="b ${tab === 'comb' ? 'on' : ''}" data-tab="comb">Combine</button></div>
      <div class="ct-shopl">${list}</div><button class="b" id="ctshopx">Leave the shop</button></div>`;
    w.hidden = false;
    w.querySelectorAll('[data-tab]').forEach((b) => (b.onclick = () => { CT.shopTab = b.dataset.tab; CT.shopUI(); }));
    const after = () => CT.shopUI();
    w.querySelectorAll('[data-buy]').forEach((b) => (b.onclick = () => { if (CT.act({ kind: 'buy', item: +b.dataset.buy })) after(); }));
    w.querySelectorAll('[data-sell]').forEach((b) => (b.onclick = () => { if (CT.act({ kind: 'sell', item: +b.dataset.sell })) after(); }));
    w.querySelectorAll('[data-comb]').forEach((b) => (b.onclick = () => { const [i, j] = b.dataset.comb.split('|').map(Number); if (CT.act({ kind: 'combine', item: i, second: j })) after(); }));
    $('ctshopx').onclick = () => { w.hidden = true; if (CT.act({ kind: 'continue' })) CT.arrive(); };
  };
  function row(it, label, attr, off) {
    return `<div class="srow"><span style="color:${it.color}">${it.name} <small class="muted">(${it.kind})</small>${it.text ? `<br><small class="afx">${it.text}</small>` : ''}</span><button class="b" ${attr} ${off ? 'disabled' : ''}>${label}</button></div>`;
  }

  // ---------- saves: three slots and an autosave, in this browser ----------
  const KEY = (k) => 'ctn-save-' + k;
  const get = (k) => { try { return localStorage.getItem(KEY(k)); } catch (e) { return null; } };
  const info = (k) => { try { return JSON.parse(localStorage.getItem(KEY(k) + '-info') || 'null'); } catch (e) { return null; } };
  function put(k) {
    try {
      localStorage.setItem(KEY(k), ctSave());
      localStorage.setItem(KEY(k) + '-info', JSON.stringify({ floor, best, turn: TURN, when: new Date().toLocaleString() }));
      return true;
    } catch (e) { log('Could not save: ' + e.message); return false; }
  }
  window.slotData = (n) => info(n);
  window.save = () => {};
  CT.autosave = () => { if (CT.V && CT.V.climb) put('auto'); };
  function load(text, from) {
    if (!text) return;
    const s = crypto.getRandomValues(new Uint32Array(2));
    const r = J(ctLoad(text, s[0], s[1]));
    if (r.error) { alert('Could not load: ' + r.error); return; }
    CT.V = r.view;
    $('title').hidden = true;
    ['slw', 'ctsum', 'ctshop'].forEach((id) => { const w = $(id); if (w) w.hidden = true; });
    CT.arrive();
    log(`Loaded ${from}.`);
  }
  window.openSlots = async function (m) {
    await CT.boot;
    const w = ovl('slw');
    const rows = ['auto', 1, 2, 3].map((k) => {
      const i = info(k);
      return `<div class="srow"><span><b>${k === 'auto' ? 'Autosave' : 'Slot ' + k}</b><br><small class="muted">${i ? `Floor ${i.floor} (best ${i.best}) · turn ${i.turn} · ${i.when}` : 'empty'}</small></span>
        ${m === 'save' && k !== 'auto' && CT.V ? `<button class="b" data-sv="${k}">Save</button>` : ''}${i ? `<button class="b" data-ld="${k}">Load</button>` : ''}</div>`;
    }).join('');
    w.innerHTML = `<div class="cap-box dA-panel"><div class="cap-t">${m === 'save' ? 'Save your game' : 'Continue'}</div>${rows}
      <div class="ct-tabs">${CT.V ? '<button class="b" id="svexp">Export</button>' : ''}<button class="b" id="svimp">Import (paste)</button><label class="b">File<input type="file" id="svfile" accept=".json,.txt" hidden></label></div>
      <button class="b" id="slx">Close</button></div>`;
    w.hidden = false;
    w.querySelectorAll('[data-sv]').forEach((b) => (b.onclick = () => { if (put(b.dataset.sv)) { log('Saved to slot ' + b.dataset.sv + '.'); window.openSlots(m); } }));
    w.querySelectorAll('[data-ld]').forEach((b) => (b.onclick = () => load(get(b.dataset.ld), b.dataset.ld === 'auto' ? 'the autosave' : 'slot ' + b.dataset.ld)));
    $('slx').onclick = () => { w.hidden = true; };
    if ($('svexp')) $('svexp').onclick = () => {
      const t = ctSave();
      try { navigator.clipboard && navigator.clipboard.writeText(t); } catch (e) {}
      const a = document.createElement('a');
      a.href = URL.createObjectURL(new Blob([t], { type: 'application/json' }));
      a.download = 'crusader-tactics-save.json'; a.click();
      log('Save copied and downloaded.');
    };
    $('svimp').onclick = () => { const t = prompt('Paste a save:'); if (t) load(t.trim(), 'the pasted save'); };
    $('svfile').onchange = (e) => { const f = e.target.files[0]; if (f) f.text().then((t) => load(t.trim(), 'the file')); };
  };
  window.savTab = function (B) { B.innerHTML = '<p>Saves have their own window.</p>'; const b = mkb(B, 'Open saves', () => window.openSlots('save')); b.style.margin = '6px'; };

  const css = document.createElement('style');
  css.textContent = '.ct-tabs{display:flex;gap:6px;margin:8px 0}.ct-shopl{max-height:55vh;overflow:auto}.srow{display:flex;justify-content:space-between;align-items:center;gap:8px;padding:6px 0;border-bottom:1px solid rgba(255,255,255,.08)}';
  document.head.append(css);
})();
