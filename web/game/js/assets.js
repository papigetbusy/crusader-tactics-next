// Pictures: loads assets/manifest.json (every picture by name) and
// assets/sheets.json (how sheets are cut into frames) and assets/skillbook.json
// (what each skill looks like, made by tools/skillbook). When the owner's new
// art arrives, it replaces files in assets/ and nothing here changes.

const pics = {}; // "group/name" -> Image
let manifest = {};
export let sheets = { effects: {}, itemIcons: {}, heroBox: {} };
export let book = { skills: {}, abilities: {} }; // assets/skillbook.json: what each skill looks like

const loadImage = (src) =>
  new Promise((ok) => {
    const im = new Image();
    im.onload = () => ok(im);
    im.onerror = () => ok(null); // a missing picture never stops the game
    im.src = src;
  });

// load fetches every picture. progress(done, total) is called as they arrive.
export async function load(build, progress = () => {}) {
  const get = (f) => fetch('assets/' + f + '?v=' + build).then((r) => r.json());
  [manifest, sheets, book] = await Promise.all([get('manifest.json'), get('sheets.json'), get('skillbook.json')]);
  const jobs = [];
  const files = {};
  for (const [group, names] of Object.entries(manifest)) {
    for (const [name, file] of Object.entries(names)) jobs.push([group + '/' + name, file]);
  }
  let done = 0;
  await Promise.all(
    jobs.map(async ([key, file]) => {
      if (!files[file]) files[file] = loadImage('assets/' + file);
      pics[key] = await files[file];
      progress(++done, jobs.length);
    }),
  );
}

// pic is a loaded picture, or null.
export const pic = (group, name) => pics[group + '/' + name] || null;

// url is a picture's address, for pictures used by the page itself.
export const url = (group, name) => (manifest[group] && manifest[group][name] ? 'assets/' + manifest[group][name] : '');

// heroPic is a hero's picture for its stage (0 base, 1-2 evolved, 3
// ascended, which shows the ascension's own picture).
export const heroPic = (cls, stage, ascension) =>
  (stage >= 3 && ascension && pic('heroes', cls + ' ' + ascension)) ||
  pic('heroes', cls + ' form' + (Math.min(2, stage || 0) + 1)) ||
  pic('heroes-first', cls);
