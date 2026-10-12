// The player's settings, kept in this browser only (they never change the
// game). Same choices and defaults as the old game's Settings → Options.

const KEY = 'ctn-options';
const defaults = {
  speed: 1, // 1×, 2×, 3 = instant
  autoEnd: false, // end the turn once every hero has acted
  confirmEnd: true, // ask before ending with heroes who haven't acted
  damageNumbers: true,
  danger: false, // shade every tile monsters can reach
  flat: true, // flat or isometric map
  minimap: true,
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

export function saveOptions() {
  try {
    localStorage.setItem(KEY, JSON.stringify(options));
  } catch (e) {
    // private browsing: settings last until the page closes
  }
}
