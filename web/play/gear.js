// Crusader Tactics: the gear sheet on the old screen (loaded after menus.js):
// what a hero wears, its weapon sets, the bag, its skills (once per floor)
// and its ascension (once per floor, between battles). Every change is a
// question to the referee.
(() => {
  'use strict';
  const CT = window.CT;
  const SLOTN = { w: 'Set 1', w2: 'Set 2', w3: 'Set 3', h: 'Helm', a: 'Armor', b: 'Boots', c: 'Charm' };
  let skillsOpen = false;

  function heroUnit(cls) { return CT.V.units.find((u) => u.class === cls && u.team === 0 && !u.companion && !u.decoy && !u.pet); }
  function row(it, btns) {
    return `<div class="srow"><span><b style="color:${it.color}">${it.name}</b> <small class="muted">(${it.kind})</small>${it.text ? `<br><small class="afx">${it.text}</small>` : ''}${it.lore ? `<br><small><i>${it.lore}</i></small>` : ''}</span><span>${btns || ''}</span></div>`;
  }
  function act(a) {
    const evs = CT.act(a);
    if (!evs) return false;
    CT.sync(); ui(); draw();
    return true;
  }

  window.openSheet = function (i) {
    const p = P[i];
    if (!p) return;
    CT.sheet = p.c;
    render();
  };
  function render() {
    const cls = CT.sheet, u = heroUnit(cls), w = ovl('ctgear');
    if (!u || !u.gear) { w.hidden = true; return; }
    const g = u.gear, V = CT.V, c = V.climb;
    const ph = c ? c.phase : 'battle', locked = u.done || V.side !== 0 || (ph !== 'battle' && ph !== 'prepare');
    const dis = locked ? 'disabled' : '';
    let h = `<div class="cap-box dA-panel ct-sheet"><div class="cap-t">${cls} · ${u.form || ''} · Lv ${u.level}</div>
      <p class="muted">HP ${u.hp}/${u.maxHp} · Attack ${u.atk} · Defense ${u.def} · Move ${u.mov}${u.ascension ? ' · ' + u.ascension : ''}</p><p class="muted">${g.weapons}</p>`;
    h += `<h3>Weapon sets</h3>` + (g.sets || []).map((s, i) => `<button class="b ${s.inHand ? 'on' : ''}" data-set="${i + 1}" ${s.inHand || locked || (g.sets || []).length < 2 ? 'disabled' : ''}>Set ${i + 1}: ${s.weapons}${s.rank ? ' (' + s.rank + ')' : ''}</button>`).join(' ');
    h += `<h3>Worn</h3>` + ((g.worn || []).map((it) => row(Object.assign({}, it, { kind: SLOTN[it.slot] + ' · ' + it.kind }), `<button class="b" data-off="${it.slot}" ${dis}>Take off</button>`)).join('') || '<p class="muted">Nothing worn.</p>');
    if (u.bond) h += `<p class="muted">Companion spirit: ${u.bond.kind}</p>`;
    if (g.bonuses) h += `<p class="muted">Active bonuses: ${g.bonuses}</p>`;
    h += `<h3>Bag</h3><p class="muted">Putting gear on or taking it off in battle uses up the hero's turn.</p>`;
    h += (V.bag || []).map((it, i) => {
      let b = '';
      if (it.slot === 'w') b = g.freeSlots && g.freeSlots.length ? `<button class="b" data-eq="${i}" ${dis}>Equip</button>` : (g.weaponSlots || []).map((s) => `<button class="b" data-eq="${i}" data-slot="${s}" ${dis}>Into ${SLOTN[s]}</button>`).join('');
      else b = `<button class="b" data-eq="${i}" ${dis}>Equip</button>`;
      return row(it, b + `<button class="b" data-drop="${i}">Drop</button>`);
    }).join('') || '<p class="muted">The bag is empty.</p>';
    // skills, once per floor
    const bar = (u.skills || []).filter((s) => s.name !== 'Strike' && !s.ultimate), ult = (u.skills || []).find((s) => s.ultimate);
    h += `<h3>Skills</h3><p class="muted">${bar.map((s) => s.name).join(', ')}${ult ? ' · ★ ' + ult.name : ''}</p>`;
    if (!u.swapped && (u.library || []).length && !locked) {
      h += `<button class="b" id="ctskb">${skillsOpen ? 'Hide' : 'Change skills (once per floor)'}</button>`;
      if (skillsOpen) {
        const normals = bar.map((s) => ({ name: s.name, on: 1 })).concat(u.library.filter((s) => !s.ultimate));
        const ults = (ult ? [{ name: ult.name, on: 1 }] : []).concat(u.library.filter((s) => s.ultimate));
        h += normals.map((s) => `<label class="ctrow"><input type="checkbox" class="skpick" value="${s.name}" ${s.on ? 'checked' : ''}> ${s.name}</label>`).join('');
        h += ults.map((s) => `<label class="ctrow"><input type="radio" name="ultpick" value="${s.name}" ${s.on ? 'checked' : ''}> ★ ${s.name}</label>`).join('');
        h += `<button class="b" id="ctskgo">Change skills</button>`;
      }
    } else if (u.swapped) h += `<p class="muted">Skills already changed on this floor.</p>`;
    // ascension, once per floor, between battles
    const hr = c && c.heroes.find((x) => x.class === cls);
    if (hr && hr.canAscend && ph !== 'battle') h += `<h3>Ascension</h3><p class="muted">Change once per floor: learned skills stay; the ultimate switches.</p>` + hr.canAscend.map((a) => `<button class="b" data-asc="${a}">${a}</button>`).join(' ');
    h += `<p><button class="b" id="ctgx">Close</button></p></div>`;
    w.innerHTML = h;
    w.hidden = false;
    const q = (s) => w.querySelectorAll(s);
    q('[data-set]').forEach((b) => (b.onclick = () => { act({ kind: 'set', unit: u.id, set: +b.dataset.set }); render(); }));
    q('[data-off]').forEach((b) => (b.onclick = () => { act({ kind: 'unequip', unit: u.id, slot: b.dataset.off }); render(); }));
    q('[data-eq]').forEach((b) => (b.onclick = () => { act({ kind: 'equip', unit: u.id, item: +b.dataset.eq, slot: b.dataset.slot || '' }); render(); }));
    q('[data-drop]').forEach((b) => (b.onclick = () => { const it = V.bag[+b.dataset.drop]; if (confirm(`Drop ${it.name}? It will be gone for good.`)) { act({ kind: 'drop', item: +b.dataset.drop }); render(); } }));
    q('[data-asc]').forEach((b) => (b.onclick = () => { if (act({ kind: 'ascend', hero: cls, choice: b.dataset.asc })) render(); }));
    q('.skpick').forEach((cb) => (cb.onchange = () => { if (q('.skpick:checked').length > 4) cb.checked = false; }));
    if ($('ctskb')) $('ctskb').onclick = () => { skillsOpen = !skillsOpen; render(); };
    if ($('ctskgo')) $('ctskgo').onclick = () => {
      const barPick = [...q('.skpick:checked')].map((x) => x.value), u2 = (w.querySelector('[name=ultpick]:checked') || {}).value || '';
      if (!barPick.length) { alert('Pick at least one skill.'); return; }
      if (confirm('Change skills now? You can only do this once on this floor.')) { skillsOpen = false; act({ kind: 'bring', unit: u.id, bar: barPick, choice: u2 }); render(); }
    };
    $('ctgx').onclick = () => { w.hidden = true; CT.sheet = null; };
  }
  CT.gearRender = render;

  const css = document.createElement('style');
  css.textContent = '.ct-sheet{max-height:86vh;overflow:auto}.ct-sheet h3{margin:10px 0 4px}';
  document.head.append(css);
})();
