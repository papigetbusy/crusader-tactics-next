// The player's settings. This browser remembers the last ones used. Each save
// file also carries its own copy, so loading a game puts that game's look back.

const KEY = 'ctn-options';
const defaults = {
  speed: 1, // 1×, 2×, 3 = instant
  autoEnd: false, // end the turn once every hero has acted
  confirmEnd: true, // ask before ending with heroes who haven't acted
  damageNumbers: true,
  danger: false, // shade every tile monsters can reach
  flat: true, // flat or isometric map
  minimap: true,
  groundTips: true, // numbers on tiles that help or hurt (owner, 2026-10-16)
  autoZoom: false, // zoom in on the selected hero
  calm: false, // fewer flashes
  lowGraphics: false, // still water and trees, no vignette
  colorblind: false,
  tipsOn: false, // first-time tips are off until "Show tips again" (old v0.25)
};

function read() {
  try {
    return JSON.parse(localStorage.getItem(KEY) || '{}');
  } catch (e) {
    return {};
  }
}

export const options = Object.assign({}, defaults, read());

let onSaved = () => {};
// onOptionsSaved runs after a setting changes, so the current game can store them.
export const onOptionsSaved = (f) => { onSaved = f; };

export function saveOptions(quiet) {
  try {
    localStorage.setItem(KEY, JSON.stringify(options));
  } catch (e) {
    // private browsing: settings last until the page closes
  }
  if (!quiet) onSaved();
}

// screenSettings is the choices a save file should remember.
export function screenSettings() {
  const o = {};
  for (const k of Object.keys(defaults)) o[k] = options[k];
  return o;
}

// applyOptions puts a save file's choices on screen. Unknown keys are ignored.
// Returns true when the file actually carried settings.
export function applyOptions(saved) {
  if (!saved || typeof saved !== 'object') return false;
  let any = false;
  for (const k of Object.keys(defaults)) {
    if (typeof saved[k] === typeof defaults[k]) {
      options[k] = saved[k];
      any = true;
    }
  }
  if (any) saveOptions(true);
  return any;
}

// withSettings copies the current choices onto a save file. The referee ignores
// that part; it only reads the game.
export function withSettings(text) {
  let o;
  try { o = JSON.parse(text); } catch (e) { return text; }
  if (!o || typeof o !== 'object' || !o.game) return text;
  o.screen = screenSettings();
  return JSON.stringify(o);
}
