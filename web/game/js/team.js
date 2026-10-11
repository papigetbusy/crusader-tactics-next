// The team picker: cards for all ten heroes, like the old game's
// "Choose 5 heroes" screen. Choosing sends one action; the referee places
// the team on the map.
import * as referee from './referee.js';
import { url } from './assets.js';

const $ = (id) => document.getElementById(id);

export const heroPicUrl = (h) =>
  (h.stage >= 3 && h.ascension && url('heroes', h.class + ' ' + h.ascension)) ||
  url('heroes', h.class + ' form' + (Math.min(2, h.stage || 0) + 1)) ||
  url('heroes-first', h.class);

// open shows the picker. done(view) gets the referee's answer.
export function open(view, done, log) {
  const c = view.climb, w = $('team'), grid = $('teamGrid');
  const chosen = c.heroes.filter((h) => h.inTeam).map((h) => h.class);
  const count = () => {
    $('teamCount').textContent = `${chosen.length} / ${c.slots} chosen`;
    $('teamGo').disabled = chosen.length !== c.slots;
  };
  grid.innerHTML = '';
  for (const h of c.heroes) {
    const b = document.createElement('button');
    b.className = 'card' + (chosen.includes(h.class) ? ' on' : '');
    b.innerHTML = `<img src="${heroPicUrl(h)}" alt=""><b>${h.class}</b><small>Lv ${h.level} · ${h.form}</small>
      <small>HP ${h.hp} · ATK ${h.atk} · DEF ${h.def} · MOV ${h.mov}</small>
      <small class="weap">⚔ ${h.weapons}</small>${h.ability ? `<small class="abil">✦ ${h.ability}: ${h.abilityText}</small>` : ''}`;
    b.onclick = () => {
      const i = chosen.indexOf(h.class);
      if (i >= 0) chosen.splice(i, 1);
      else if (chosen.length < c.slots) chosen.push(h.class);
      else return log(`Only ${c.slots} heroes fit. Tap a chosen hero to take them out first.`);
      b.classList.toggle('on', i < 0);
      count();
    };
    grid.append(b);
  }
  $('teamTitle').textContent = `Choose ${c.slots} heroes`;
  count();
  w.hidden = false;
  $('teamBack').onclick = () => { w.hidden = true; };
  $('teamGo').onclick = () => {
    const r = referee.act({ kind: 'team', heroes: chosen.slice() });
    if (!r || r.refused) return log(r ? r.refused.reason : 'The referee did not answer.');
    w.hidden = true;
    done(r.view);
  };
}
