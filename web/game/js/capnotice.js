// The level-cap notice: when a hero at the level cap would have earned XP
// (it acted, or a monster fell), a window says so, once per hero for each
// cap [old capHit and showCap, L1463-1465]. The referee decides the cap
// (View.levelCap); this only tells the player.
const $ = (id) => document.getElementById(id);
const told = {}; // hero class -> the cap already told about
const GAP = 5; // levels above the best floor's monsters (data/rules/progress.json capAboveFloor), for the text only

export function check(events, view, before = []) {
  const cap = view.levelCap;
  if (!cap || !view.units) return;
  const heroes = view.units.filter((u) => u.team === 0 && u.gear && !u.companion && !u.decoy && !u.pet && u.level >= cap);
  if (!heroes.length) return;
  const acted = new Set(events.filter((e) => ['skill', 'ability', 'riposte', 'overwatch'].includes(e.kind)).map((e) => e.unit));
  const kill = events.some((e) => e.kind === 'died' && before.some((u) => u.id === e.unit && u.team === 1));
  const names = heroes.filter((u) => (acted.has(u.id) || kill) && told[u.class] !== cap).map((u) => { told[u.class] = cap; return u.class; });
  if (!names.length) return;
  const best = view.climb ? view.climb.best : view.floor, L = cap - GAP;
  let w = $('capw');
  if (!w) {
    w = document.createElement('div');
    w.id = 'capw';
    document.body.append(w);
  }
  w.innerHTML = `<div class="panel cap-box"><div class="cap-t">Level cap reached</div><p><b>${names.join(', ')}</b> ${names.length > 1 ? 'have' : 'has'} reached <b>Lv ${cap}</b>.</p>
    <p>Heroes can be at most <b>${GAP} levels above the monsters on your highest floor</b> (Floor ${best}: monsters Lv ${L}).</p>
    <p>${names.length > 1 ? 'They' : 'This hero'} won't earn XP until you clear Floor ${best}. Every new floor you clear raises the cap.</p><button class="b" id="capok">Understood</button></div>`;
  w.hidden = false;
  $('capok').onclick = () => { w.hidden = true; };
}
