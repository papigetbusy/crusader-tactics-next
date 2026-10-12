// The Settings window: Options, How to Play, What's New, Saves and a
// copyable bug Report, like the old game [old openSettings and
// renderSettings L1619-1660]. Options are kept in this browser only.
import { options, saveOptions } from './options.js';
import { cover, sw, seg, slots, importer, exportSave } from './sysmenu.js';
import { reset as resetTips } from './tips.js';

const $ = (id) => document.getElementById(id);
const BUILD = (/[?&]v=([^&]+)/.exec(import.meta.url) || [])[1] || 'local';
let screen = null, tab = 'opt';
export const wireSettings = (s) => { screen = s; };

// Recent errors, for the report.
const ERRS = [];
addEventListener('error', (e) => { ERRS.push(String(e.message).slice(0, 160)); ERRS.splice(0, ERRS.length - 5); });
addEventListener('unhandledrejection', (e) => { ERRS.push(String(e.reason).slice(0, 160)); ERRS.splice(0, ERRS.length - 5); });

export function settings(t) {
  tab = t || tab;
  const w = cover('settingsWin');
  const tabs = [['opt', 'Options'], ['how', 'How to Play'], ['new', 'What’s New'], ['sav', 'Saves'], ['rep', 'Report']];
  w.innerHTML = `<div class="panel box wide sets"><div class="ttl">Settings</div>
    <div class="tabs">${tabs.map(([k, l]) => `<button data-t="${k}" class="${tab === k ? 'on' : ''}">${l}</button>`).join('')}</div>
    <div id="setBody"></div><div class="btns"><button class="b" id="oClose">Close</button></div></div>`;
  w.hidden = false;
  w.querySelectorAll('[data-t]').forEach((b) => (b.onclick = () => settings(b.dataset.t)));
  $('oClose').onclick = () => { w.hidden = true; };
  ({ opt, how, new: news, sav, rep })[tab]($('setBody'));
}

function opt(B) {
  const row = (label, ctl, sub) => `<div class="mrow"><span>${label}${sub ? `<br><small>${sub}</small>` : ''}</span>${ctl}</div>`;
  B.innerHTML = `<h4>Battle</h4>${row('Battle speed', seg('oSpeed', [[1, '1×'], [2, '2×'], [3, 'Instant']], options.speed))}
    ${row('Auto end turn', sw('oAuto', options.autoEnd), 'Ends your turn once every hero has acted')}
    ${row('Ask before ending turn', sw('oAsk', options.confirmEnd), 'Warns if heroes haven’t acted yet')}
    ${row('Damage numbers', sw('oDmg', options.damageNumbers))}
    <h4>Map</h4>${row('Map view', seg('oView', [[1, 'Flat'], [0, 'Isometric']], options.flat ? 1 : 0))}${row('Minimap', sw('oMini', options.minimap))}${row('Auto-zoom on selected hero', sw('oZoom', options.autoZoom))}${row('Danger zone', sw('oDanger', options.danger), 'Shade every tile monsters can reach')}
    <h4>Accessibility &amp; performance</h4>${row('Reduce flashing effects', sw('oCalm', options.calm), 'Softer, single-layer attack animations')}
    ${row('Low graphics', sw('oLow', options.lowGraphics), 'Still water and trees, no vignette; saves battery')}
    ${row('Colorblind-friendly tiles', sw('oCb', options.colorblind))}
    <h4>Sound</h4>${row('Sound effects &amp; music', '<input type="range" disabled value="70" aria-label="Volume">', 'Arrive with the sound update')}
    <div class="btns"><button class="b" id="oTips">${options.tipsOn ? 'Show tips again' : 'Turn tips on'}</button><button class="b" id="oCodex">Codex</button><button class="b" id="oTitle">Title screen</button></div>`;
  const flip = (id, k) => ($(id).onclick = () => { options[k] = !options[k]; saveOptions(); screen.after(); settings(); });
  flip('oAuto', 'autoEnd'); flip('oAsk', 'confirmEnd'); flip('oDmg', 'damageNumbers'); flip('oMini', 'minimap'); flip('oDanger', 'danger');
  flip('oCalm', 'calm'); flip('oLow', 'lowGraphics'); flip('oCb', 'colorblind'); flip('oZoom', 'autoZoom');
  $('oSpeed').querySelectorAll('button').forEach((b) => (b.onclick = () => { options.speed = +b.dataset.v; saveOptions(); settings(); }));
  $('oView').querySelectorAll('button').forEach((b) => (b.onclick = () => { screen.setView(b.dataset.v === '1'); saveOptions(); settings(); }));
  $('oTips').onclick = () => { options.tipsOn = true; saveOptions(); resetTips(); $('oTips').textContent = 'Tips reset ✓'; };
  $('oCodex').onclick = () => window.open('codex.html', '_blank');
  $('oTitle').onclick = () => { if (confirm('Go back to the title screen? Unsaved progress is lost (the autosave keeps your last turn).')) location.reload(); };
}

function how(B) {
  B.innerHTML = `<p><a class="b" href="codex.html" target="_blank" rel="noopener">Open the full Codex ↗</a></p>
  <h4>Turns</h4><p>Drag a hero onto a blue tile to move, or tap the hero and then the tile (tap the hero again to act in place). Tap an enemy to attack it; if it is out of reach the hero walks closer first. A wooden menu opens beside the hero: Attack, Skills ▸, ✦ class ability, Item, Move, Undo, Defend, Wait. Attacks show a forecast window beside the target. Pinch out or use − in the ☰ menu for a whole-map overview. Monsters move after you end your turn.</p>
  <h4>Your team</h4><p>All ten heroes are yours from the start. Each floor, look at the map, choose who to bring (5, plus one more at floors 25, 50 and 75), then place them on the blue tiles. Only the heroes who fight earn the floor's XP.</p>
  <h4>Weapons &amp; elements</h4><p>Sword beats Axe, Axe beats Lance, Lance beats Sword (+15% hit and ×1.5 damage with the edge; −15% and ×0.75 against). Fire beats Ice, Ice beats Thunder, Thunder beats Fire; Light and Dark beat each other (WEAK ×1.5, RESIST ×0.75). Weapon ranks rise E→S with level, and rarer items need higher ranks.</p>
  <h4>Counterattacks</h4><p>A unit that survives a single-target hit strikes back once with a full Strike, if the attacker is within its Strike range (melee 1 tile; bows, guns, staffs and wands 2–6). No counter while stunned, frozen, rooted or cloaked, and never against area skills, ultimates, heals or buffs. The forecast shows the counter, or why there is none.</p>
  <h4>Terrain</h4><p>Forest −15% damage taken and costs 2 move; hills −10%, and attacking from a hill onto lower ground deals +10% damage with +10% hit (uphill: −10% hit); climbing onto a hill costs 1 more move except for fliers, Rangers and Knights; forts and thrones −25% and heal 10% a turn; roads are fast; water and mountains block all but fliers.</p>
  <h4>Floors</h4><p>Elites every 5th floor, a boss every 10th, the shop before every 7th, and a free legendary every 100. Floors are fixed per run: returning gives the same map and monsters, opened chests stay opened, and replays pay half gold and full XP. A defeat costs nothing: fight the floor again, or go back and grow stronger. The game autosaves after every turn.</p>
  <h4>Skills</h4><p>Every class learns a skill at Lv 1, 3, 6 … 27, with ultimates at Lv 1, 20 and 30. Change the skills a hero brings once per floor, from the Gear sheet.</p>
  <h4>Level cap</h4><p>Heroes can be at most 5 levels above the monsters on your highest floor. Capped heroes earn no XP until you clear that floor.</p>
  <h4>Statuses</h4><p>Badges only show over the unit you select or target. Tap a monster for its full status list.</p>`;
}

function news(B) {
  B.innerHTML = `<p><b>Crusader Tactics, rebuilt</b> <small>(build ${BUILD})</small></p><ul>
  <li>The whole game runs on a new rules engine, built so online play can come later.</li>
  <li>All ten heroes from the start: choose your team for each floor and place them yourself (5 slots, one more at floors 25, 50 and 75).</li>
  <li>Every class learns on one even schedule: 53 new skills, and the Healer's new ultimate, Last Light.</li>
  <li>A purple (epic) tier between rare and legendary, with its own bonuses; legendaries drop once a hero reaches Lv 30.</li>
  <li>The shop comes up before every 7th floor; combine two copies of an item into one a rarity higher.</li>
  <li>Change skills once per floor; no penalty for a defeat.</li>
  <li>Three save slots, an autosave every turn, and saves you can export and import from a file.</li>
  <li>The forecast now shows exactly the damage you will deal.</li></ul>
  <p><a href="codex.html#changes" target="_blank" rel="noopener">Every change from the old game ↗</a></p>`;
}

function sav(B) {
  B.innerHTML = `<h4>Export</h4><p>Copy and download your current game so you can back it up or move it to another device.</p>
  <div class="btns"><button class="b" id="svE">Copy &amp; download</button></div>
  <h4>Import</h4><p>Choose a save file or paste a save code, pick a slot, then tap Import. Importing replaces that slot (it asks first).</p>
  <div class="btns"><button class="b" id="svI">Import a save</button><button class="b" id="svS">Save slots</button></div><p id="svMsg"></p>`;
  const inGame = screen.st.view && $('title').hidden;
  $('svE').disabled = !inGame;
  $('svE').onclick = () => { exportSave(); $('svMsg').textContent = 'Save copied and downloaded.'; };
  $('svI').onclick = () => { $('settingsWin').hidden = true; importer(inGame ? 'save' : 'load'); };
  $('svS').onclick = () => { $('settingsWin').hidden = true; slots(inGame ? 'save' : 'load'); };
}

function rep(B) {
  const v = screen.st.view || {}, c = v.climb;
  const party = (screen.st.units || []).filter((u) => u.team === 0 && !u.companion && !u.pet && !u.decoy).map((u) => `${u.class} Lv ${u.level}${u.ascension ? ' ' + u.ascension : ''}`).join('; ');
  const log = $('log') ? $('log').textContent : '';
  const t = `Crusader Tactics (rebuilt) build ${BUILD}\nDate: ${new Date().toISOString()}\nDevice: ${navigator.userAgent}\nScreen: ${innerWidth}x${innerHeight} @${devicePixelRatio}\n` +
    `Floor ${c ? c.floor : v.floor || 1} (best ${c ? c.best : '-'}) · Turn ${v.turn || 1} · Gold ${c ? c.gold : v.gold || 0} · Items ${(v.bag || []).length}\nParty: ${party || 'none'}\n` +
    `View: ${options.flat ? 'flat' : 'isometric'} · Options: ${JSON.stringify(options)}\nRecent errors: ${ERRS.length ? ERRS.join(' | ') : 'none'}\nLast message: ${log}\nWhat happened: `;
  B.innerHTML = '<p>Copy this report and send it along with what happened.</p><textarea id="repTxt" rows="11"></textarea><div class="btns"><button class="b" id="repCopy">Copy game report</button></div>';
  $('repTxt').value = t;
  $('repCopy').onclick = async () => {
    let ok = false;
    try { await navigator.clipboard.writeText($('repTxt').value); ok = true; } catch (e) { try { $('repTxt').select(); ok = document.execCommand('copy'); } catch (e2) { ok = false; } }
    $('repCopy').textContent = ok ? 'Copied ✓' : 'Select the text above and copy it';
  };
}
