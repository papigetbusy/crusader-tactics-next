// The referee: the only part of the screen that talks to the game rules.
//
// Today the referee runs inside this browser (the Go rules compiled to
// WebAssembly). For online play this one file is swapped for one that sends
// the very same questions to a server, and nothing else on the screen changes.
// Every answer is plain data (the View and the events), never rules.

let ready = null;

const json = (s) => (s == null || s === 'null' ? null : JSON.parse(s));

// start loads the referee once; every question waits for it.
export function start(build) {
  if (!ready) {
    ready = (async () => {
      await new Promise((ok, fail) => {
        const s = document.createElement('script');
        s.src = 'wasm_exec.js?v=' + build;
        s.onload = ok;
        s.onerror = () => fail(new Error('wasm_exec.js is missing'));
        document.head.append(s);
      });
      const go = new window.Go();
      const url = 'app.wasm?v=' + build;
      let r;
      try {
        r = await WebAssembly.instantiateStreaming(fetch(url), go.importObject);
      } catch (e) {
        const res = await fetch(url);
        if (!res.ok) throw new Error('app.wasm is missing');
        r = await WebAssembly.instantiate(await res.arrayBuffer(), go.importObject);
      }
      go.run(r.instance);
    })();
  }
  return ready;
}

const seeds = (n) => [...crypto.getRandomValues(new Uint32Array(n))];

// ---------- starting a game ----------

// newClimb starts a climb from floor 1 and returns its View.
export function newClimb() {
  const [run, hi, lo] = seeds(3);
  return json(window.ctClimb(run, hi, lo));
}

// practice starts one of the practice fights.
export function practice(i) {
  const [hi, lo] = seeds(2);
  return json(window.ctNew(i, hi, lo));
}

// save is the game as text, to keep; load starts from such a text with fresh
// dice. load returns {view} or {error}.
export const save = () => window.ctSave();
export function load(text) {
  const [hi, lo] = seeds(2);
  return json(window.ctLoad(text, hi, lo));
}

// ---------- doing things ----------

// act sends one action. It returns {events, view}, or {refused: {code, reason}}.
export const act = (a) => json(window.ctApply(JSON.stringify(a)));

// ---------- questions (they never change the game) ----------

export const reach = (unit) => json(window.ctReach(unit)) || [];
export const threat = (unit) => json(window.ctThreat(unit)) || [];
export const targets = (unit, slot, set = 0) => json(window.ctTargets(unit, slot, set)) || [];
export const preview = (unit, slot, target, set = 0) => json(window.ctPreview(unit, slot, target, set));
export const previewFrom = (unit, slot, target, x, y) => json(window.ctPreviewFrom(unit, slot, target, x, y));
export const abilityAim = (unit) => json(window.ctAim(unit)) || {};
export const skillAim = (unit, slot, picks) => json(window.ctSkillAim(unit, slot, JSON.stringify(picks))) || {};
export const info = () => json(window.ctInfo());
// plan is which skill a tap on an enemy uses, and where to walk first: {slot, walk} or null.
export const plan = (unit, target) => json(window.ctPlan(unit, target));
// guide is the fixed rules described in words (terrain, element wheel, weapon triangle).
export const guide = () => json(window.ctGuide());
