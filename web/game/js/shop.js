// The shop before every 7th floor: Buy, Sell and Combine tabs, like the old
// game's shop [old renderShop L766-770]. Prices, stock and what can be
// combined all come from the referee; every button is an action.
import * as referee from './referee.js';
import { url } from './assets.js';
import { icon } from './icons.js';

const $ = (id) => document.getElementById(id);
let screen = null, tab = 'buy';

export const wire = (s) => { screen = s; };

function act(a) {
  const r = referee.act(a);
  if (!r || r.refused) return screen.log(r ? r.refused.reason : 'The referee did not answer.');
  screen.show(r.view);
  if (r.view.climb && r.view.climb.phase === 'shop') render();
  else { $('shopWin').hidden = true; screen.after(); }
}

const row = (it, label, f, off) => {
  const d = document.createElement('div');
  d.className = 'srow';
  d.innerHTML = `<img class="ic" src="${icon(it)}" alt=""><span><b style="color:${it.color}">${it.name}</b> <small>(${it.kind})</small>${it.text ? `<br><small class="afx">${it.text}</small>` : ''}</span>`;
  const b = document.createElement('button');
  b.className = 'b';
  b.textContent = label;
  b.disabled = !!off;
  b.onclick = f;
  d.append(b);
  return d;
};

export function open() {
  tab = 'buy';
  render();
}

function render() {
  const c = screen.st.view.climb, w = $('shopWin');
  w.innerHTML = `<div class="panel box wide"><div class="ttl">Shop · before floor ${c.floor}</div>
    <p class="gold"><img src="${url('ui', 'coin')}" alt=""> ${c.gold} gold</p>
    <div class="tabs">${[['buy', 'Buy'], ['sell', 'Sell'], ['combine', 'Combine']].map(([k, l]) => `<button class="${tab === k ? 'on' : ''}" data-t="${k}">${l}</button>`).join('')}</div>
    <div class="list" id="shopList"></div><button class="b" id="shopLeave">Leave shop</button></div>`;
  w.hidden = false;
  const L = $('shopList');
  if (tab === 'buy') {
    if (!(c.shop || []).length) L.textContent = 'Sold out.';
    (c.shop || []).forEach((it, i) => L.append(row(it, `Buy ${it.price}g`, () => act({ kind: 'buy', item: i }), c.gold < it.price)));
  } else if (tab === 'sell') {
    if (!(c.bag || []).length) L.textContent = 'Nothing to sell. Worn gear is not listed.';
    (c.bag || []).forEach((it, i) => L.append(row(it, `Sell ${it.price}g`, () => act({ kind: 'sell', item: i }))));
  } else {
    if (!(c.pairs || []).length) L.textContent = 'Combine two copies of the same item (same rarity) into one a rarity higher. You have no pairs yet.';
    (c.pairs || []).forEach(([a, b]) => L.append(row(c.bag[a], 'Combine 2', () => act({ kind: 'combine', item: a, second: b }))));
  }
  w.querySelectorAll('[data-t]').forEach((b) => (b.onclick = () => { tab = b.dataset.t; render(); }));
  $('shopLeave').onclick = () => act({ kind: 'continue' });
}
