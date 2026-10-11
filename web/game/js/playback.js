// Playing the referee's events as pictures: walking, lunges, effect sheets,
// numbers, hits and deaths. Nothing here decides anything: it keeps showing
// the previous answer while the events play, then the new answer is shown.
// Timings and effect choices match the old game [old startMove L788, anim
// L789-800, spawn L813, drawPop L1686, sprite effects L1305-1320].
import { feet } from './board.js';
import { pic, sheets } from './assets.js';
import { TS } from './look.js';
import { lookAt } from './camera.js';

const FX = {
  E1: 'fire_looping_004A', E2: 'epic_explosion_001', E3: 'stylized_explosion_002', E4: 'lightning_strike_001',
  E5: 'lightning_burst_002', E6: 'directional_impact_001', E7: 'symmetrical_impact_003', E8: 'spell_heal_001',
  E9: 'spell_attack_up_001', E10: 'spell_defense_up_001', E11: 'spell_haste_001', E12: 'spell_poison_001',
  E13: 'spell_death_001', E14: 'status_poison_001', E15: 'scifi_muzzle_flash_001', E16: 'burst_splatter_001',
  E17: 'stylized_skull_smoke_burst_001', E18: 'round_sparkle_burst_002',
};

// Which effects play for a skill, at the target ('t') or the user ('c').
const SKILL_FX = {};
const add = (e, at, names) => names.forEach((n) => (SKILL_FX[n] = SKILL_FX[n] || []).push([e, at]));
add('E1', 't', ['Fireball', 'Inferno', 'Dragonfire', 'Flame Breath', 'Fire Breath', 'Sacred Flame', 'Hellfire']);
add('E2', 't', ['Meteor', "Dragon's Fury", 'Cataclysm', 'Hellfire Dive', 'Greater Spell', 'Crash Down', 'Earthshatter']);
add('E3', 't', ['Arcane Nova', 'Soul Nova', 'Hex Storm', 'Grave Nova', 'Shadow Nova', 'Soul Bolt', 'Absolute Zero', 'Grand Hex', 'Frost Bolt', 'Ice Lance', 'Blizzard']);
add('E4', 't', ['Thunder', 'Thunderstorm', 'Lightning Charge', 'Thunder Lance']);
add('E5', 't', ['Chain Lightning', 'Ricochet Storm', 'Flash Bang']);
add('E6', 't', ['Slash', 'Rend', 'Power Strike', 'Heavy Slash', 'Lance Strike', 'Stab', 'Spear Thrust', 'Dragon Slash', 'Crescent Wave', 'Thousand Cuts', 'Crimson Waltz', 'Ambush', 'Backstab', 'Running Thrust', 'Charge', 'Grand Charge', 'Cavalry Breaker', 'Crusade', 'Wing Slash', 'Claw Swipe', 'Claw', 'Dragon Claw', 'Savage Claw', 'Impale', 'Piercing Shot', 'Shot', 'Multishot', 'Power Shot', 'Snaring Shot', 'Pinning Shot', 'Gale Volley', 'Deadeye', 'Quick Shot', 'Double Tap', 'Ricochet', 'Scatter Shot', 'Hollow Point', 'Bullet Hell', 'Six Shooter', 'Dead or Alive', 'Barrage', 'Venom Edge', 'Poison Blade', 'Strike']);
add('E7', 't', ['Bash', 'Shield Bash', 'Ground Pound', 'Crushing Blow', 'Fortress Slam', "Titan's Wrath", 'Smash', 'Club Smash', 'Flail Smash', 'Shell Slam', 'Stone Quake', 'Ground Slam', 'War Stomp', 'Glacier Fist', 'Tail Slam', 'Lance Break', 'Sky Dive', 'Dive Strike', 'Dive Claw', 'Tidal Crush', 'Rampage', 'Frost Nova']);
add('E8', 't', ['Mend Self', 'Heal Ally', 'Mass Heal', 'Renewal', 'Mass Mend', 'Holy Siphon', 'Sanctuary', 'Divine Grace']);
add('E9', 't', ['Blessing', 'War Cry', 'Valor', 'Dark Pact', 'Prism Aura']);
add('E10', 't', ['Ironhide', 'Bulwark']);
add('E11', 't', ['Time Warp']);
add('E12', 't', ['Toxic Bolt', 'Miasma', 'Poison Arrow', 'Plague Swarm', "Witch's Brew", 'Venom Storm', 'Pandemic', 'Toxic Spores', 'Plague Mist', 'Venom Wave', 'Venom Spit', 'Spore Bolt', 'Venom Edge']);
add('E13', 't', ['Execution', 'Assassinate', 'Doom Lance', 'Withering Curse', 'Death Bolt', 'Soul Rend', 'Grave Bolt', 'Blood Drain', 'Phantom Rush', 'Infernal Claw']);
add('E15', 'c', ['Quick Shot', 'Double Tap', 'Ricochet', 'Scatter Shot', 'Hollow Point', 'Bullet Hell', 'Six Shooter', 'Dead or Alive', 'Ricochet Storm']);
add('E16', 't', ['Blood Rage', 'Blood Moon', 'Death Blossom', 'Fan of Knives', 'Rend', 'Savage Claw']);
add('E17', 't', ['Hex', 'Curse of Frailty', 'Shackle', 'Constrict', 'Smoke Bomb', 'Cloak']);
add('E18', 't', ['Holy Judgment', 'Radiant Burst', 'Cleanse', 'Resurrection', 'Pillar']);
const ELEMENT_FX = { fire: 'E1', ice: 'E3', elec: 'E4', dark: 'E13', light: 'E18' };

// effectsFor picks the effects of a skill (by name, else by kind).
function effectsFor(name, skill) {
  if (SKILL_FX[name]) return SKILL_FX[name];
  if (!skill) return [['E6', 't']];
  if (skill.heal || skill.revive) return [['E8', 't']];
  if (skill.support) return [['E9', 't']];
  return [[ELEMENT_FX[skill.element] || 'E6', 't']];
}

// ---------- the timeline ----------

let run = null; // the events being played: { units: Map, items: [], end, done }
const pops = [];
const fx = [];
export const playing = () => !!run;
let skipped = false;
export const skip = () => { skipped = true; };

// play shows events one after another on the previous answer's units, then
// calls done (which shows the new answer). speed: 1, 2 or 3 (instant).
export function play(units, events, speed, done, onEnemyTurn) {
  skipped = false;
  const k = () => (skipped ? 40 : speed === 3 ? 12 : speed === 2 ? 2.2 : 1);
  const now = performance.now();
  const byId = new Map(units.map((u) => [u.id, u]));
  const spot = new Map(); // where each unit stands on screen now (it may have walked)
  const at = (id) => spot.get(id) || (byId.get(id) ? [byId.get(id).x, byId.get(id).y] : null);
  const items = [];
  let t = 0, busyUntil = 0, actor = null;
  // A new unit starts acting: wait until the last animation has finished,
  // then move the camera to it, then act (the camera never jumps away
  // mid-animation; owner's request, decision 10).
  const turnTo = (who) => {
    if (!who || who.id === actor) return;
    if (actor !== null) t = Math.max(t, busyUntil);
    if (who.team === 1) {
      items.push({ t, kind: 'cam', at: at(who.id) });
      t += 350;
    }
    actor = who.id;
  };
  for (const e of events) {
    const who = e.unit ? byId.get(e.unit) : null, target = e.target ? byId.get(e.target) : who;
    if (e.kind === 'moved' || e.kind === 'skill' || e.kind === 'ability') turnTo(who);
    switch (e.kind) {
      case 'moved': case 'undone': {
        if (!who || !e.path || e.path.length < 2) break;
        const dur = Math.min(1500, 150 * (e.path.length - 1));
        items.push({ t, kind: 'walk', id: who.id, path: e.path, dur, from: at(who.id) });
        busyUntil = Math.max(busyUntil, t + dur);
        spot.set(who.id, e.path[e.path.length - 1]);
        t += dur;
        break;
      }
      case 'skill': case 'ability': case 'riposte': case 'overwatch': {
        if (!who) break;
        const tg = target || who;
        const skill = (who.skills || []).find((s) => s.name === e.skill);
        items.push({ t, kind: 'lunge', id: who.id, to: e.target ? at(e.target) : null, from: at(who.id), cast: !e.target || (skill && (skill.heal || skill.support)) });
        busyUntil = Math.max(busyUntil, t + (e.target && !(skill && (skill.heal || skill.support)) ? 420 : 720), t + 160 + 650);
        for (const [code, where] of effectsFor(e.skill, skill)) {
          items.push({ t: t + (where === 'c' ? 0 : 160), kind: 'fx', sheet: FX[code], at: where === 'c' ? at(who.id) : at(tg.id), big: code === 'E2' ? 1.5 : code === 'E15' ? 0.7 : 1 });
        }
        t += 300;
        break;
      }
      case 'hit': case 'missed': case 'healed': case 'potion': case 'ticked': case 'revived': {
        const v = target || who;
        if (!v) break;
        const n = e.amount;
        const txt = e.kind === 'missed' ? 'MISS' : e.kind === 'hit' ? (e.crit ? 'CRIT ' + n : n + (e.mult > 1 ? '!' : '')) : e.kind === 'ticked' ? '' + n : '+' + n;
        const col = e.kind === 'missed' ? '#c9c9d6' : e.kind === 'hit' ? (e.crit ? '#ff9a2e' : e.mult > 1 ? '#ffd84a' : e.mult < 1 ? '#c9c9d6' : '#fff') : e.kind === 'ticked' ? '#ff9a9a' : '#6ee7a0';
        items.push({ t, kind: 'pop', at: at(v.id), txt, col, big: e.crit || e.mult > 1, id: v.id, hurt: e.kind === 'hit' || e.kind === 'ticked' });
        if (e.crit) items.push({ t, kind: 'fx', sheet: FX.E18, at: at(v.id), big: 1 });
        t += 120;
        break;
      }
      case 'died': case 'faded': case 'exploded':
        if (who) items.push({ t, kind: 'fade', id: who.id });
        t += 200;
        break;
      case 'turn-start':
        if (e.side === 1) { items.push({ t, kind: 'enemy' }); t += 300; }
        break;
      default:
    }
  }
  if (!items.length) return done();
  run = { now, k, items, units: byId, end: t, done, onEnemyTurn, last: now, clock: 0 };
}

// step moves the timeline on; the scene calls it every frame.
function step(now) {
  if (!run) return;
  run.clock += (now - run.last) * run.k();
  run.last = now;
  const c = run.clock;
  for (const i of run.items) {
    if (i.started || c < i.t) continue;
    i.started = now;
    if (i.kind === 'pop' && i.at) {
      const f = feet(...i.at);
      pops.push({ x: f.x, y: f.y - 72, t0: now, txt: i.txt, col: i.col, big: i.big });
      if (i.hurt) i.shake = now;
    }
    if (i.kind === 'fx' && i.at) {
      const f = feet(...i.at);
      fx.push({ sheet: i.sheet, x: f.x, y: f.y - 30, t0: now, d: 650 / Math.min(run.k(), 6), size: TS * 1.5 * i.big });
    }
    if (i.kind === 'cam' && i.at) {
      const f = feet(...i.at);
      lookAt(f.x, f.y - 40);
    }
    if (i.kind === 'enemy' && run.onEnemyTurn) run.onEnemyTurn();
  }
  if (c >= run.end + 450) {
    const done = run.done;
    run = null;
    done();
  }
}

// anim is how far a unit is drawn from its tile right now, in map pixels
// (x, y), plus a bounce (b), a tilt (r) and see-through (a).
export function anim(u) {
  const out = { x: 0, y: 0, b: 0, r: 0, a: 1 };
  if (!run) return out;
  const c = run.clock, now = performance.now(), base = feet(u.x, u.y);
  let place = null;
  for (const i of run.items) {
    if (i.id !== u.id) continue;
    if (i.kind === 'walk') {
      if (c >= i.t + i.dur) place = i.path[i.path.length - 1];
      else if (c >= i.t) {
        const q = ((c - i.t) / i.dur) * (i.path.length - 1), n = Math.min(i.path.length - 2, Math.floor(q)), f = q - n;
        const A = feet(...i.path[n]), B = feet(...i.path[n + 1]);
        out.x = A.x + (B.x - A.x) * f - base.x;
        out.y = A.y + (B.y - A.y) * f - base.y;
        out.b = -Math.abs(Math.sin(f * Math.PI)) * 7;
        out.r = Math.sin(now * 0.03) * 0.07;
        return out;
      }
    }
  }
  if (place) {
    const p = feet(...place);
    out.x = p.x - base.x;
    out.y = p.y - base.y;
  }
  for (const i of run.items) {
    if (i.id !== u.id || !i.started) continue;
    const p = (c - i.t) / (i.kind === 'lunge' ? (i.cast ? 720 : 420) : i.kind === 'fade' ? 500 : 320);
    if (p < 0 || p >= 1) {
      if (i.kind === 'fade' && p >= 1) out.a = 0;
      continue;
    }
    if (i.kind === 'lunge') {
      if (i.cast || !i.to) out.b -= Math.sin(p * Math.PI) * 9;
      else {
        const a = feet(...i.from), b = feet(...i.to), d = Math.hypot(b.x - a.x, b.y - a.y) || 1;
        const l = Math.sin(Math.min(1, p * 1.2) * Math.PI) * 16;
        out.x += ((b.x - a.x) / d) * l + Math.sin(now * 0.09) * 3 * (1 - p);
        out.y += ((b.y - a.y) / d) * l;
      }
    } else if (i.kind === 'pop' && i.hurt) out.x += Math.sin(now * 0.1) * 4 * (1 - p);
    else if (i.kind === 'fade') out.a = 1 - p;
  }
  return out;
}

// draw draws effect sheets and floating numbers; returns true while busy.
export function draw(g, now) {
  step(now);
  for (let i = fx.length; i--; ) if (now > fx[i].t0 + fx[i].d) fx.splice(i, 1);
  for (const f of fx) {
    const E = sheets.effects[f.sheet], im = pic('effects', f.sheet);
    if (!E || !im) continue;
    const p = (now - f.t0) / f.d, fr = E.fr[Math.min(E.fr.length - 1, (p * E.fr.length) | 0)], h = (f.size * E.h) / E.w;
    g.save();
    g.imageSmoothingEnabled = false;
    g.drawImage(im, fr[0], fr[1], E.w, E.h, f.x - f.size / 2, f.y - h / 2, f.size, h);
    g.restore();
  }
  for (let i = pops.length; i--; ) if (now > pops[i].t0 + 1000) pops.splice(i, 1);
  for (const o of pops) {
    const p = (now - o.t0) / 1000;
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
  return !!run || fx.length > 0 || pops.length > 0;
}
