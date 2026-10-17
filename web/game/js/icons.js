// Item pictures, cut from the Raven Fantasy icon sheet the old game used
// [old RVW, RVT, RVU, RVA, RVC and icon(), L1322-1327]; the class
// legendaries have their own pictures (assets/legendary). Only looks.
import { pic, sheets } from './assets.js';

const WEAPON = { sword: [0, 1], axe: [1, 8], lance: [1, 4], bow: [4, 10], dagger: [0, 3], staff: [3, 3], wand: [3, 0], gun: [2, 8] };
const WEAPON_ROW = [90, 100, 95];
const UNIQUE = { Dawnbreaker: [105, 0], Voidreaver: [106, 4], Heavensfall: [107, 10], 'Gale Piercer': [105, 5], Nightfang: [105, 7], 'Staff of Eternity': [105, 10], 'Emberheart Scepter': [106, 15], Thunderhowl: [106, 2], 'Heart of the Stair': [136, 4], 'Phoenix Feather': [136, 0] };
const ARMOR = {
  L: [[[122, 8], [123, 9], [124, 4]], [[122, 1], [123, 1], [124, 1]]],
  M: [[[125, 2], [126, 2], [127, 2]], [[125, 7], [126, 11], [127, 10]]],
  H: [[[119, 0], [120, 0], [121, 0]], [[119, 6], [120, 6], [121, 6]]],
};
const CHARM = [[115, 2], [115, 4], [115, 6], [136, 4]];
const made = {};

// icon is a 48×48 picture (as a data URL) for an item from the referee's View.
// legendary is the top rarity number (it gets the gold look).
export function icon(it, legendary = 4) {
  const r = it.tier >= legendary ? 3 : Math.min(2, it.tier || 0);
  const k = it.slot + it.name + r + it.weaponType + it.weight;
  if (made[k]) return made[k];
  let rc;
  if (UNIQUE[it.name]) rc = UNIQUE[it.name];
  else if (it.slot === 'w' || it.slot.startsWith('w')) {
    const w = WEAPON[it.weaponType] || WEAPON.sword;
    rc = [WEAPON_ROW[Math.min(2, r)] + w[0], w[1]];
  } else if (it.slot === 'c') rc = CHARM[Math.min(3, r)];
  else {
    const set = ARMOR[it.weight] || ARMOR.L;
    rc = set[r >= 2 ? 1 : 0]['hab'.indexOf(it.slot)] || set[0][0];
  }
  const own = pic('legendary', it.name); // a class legendary's own picture (owner's picks, 2026-10-17)
  const i = sheets.itemIcons[rc[0] + ',' + rc[1]], sheet = pic('ui', 'raven');
  if (!own && (i == null || !sheet)) return '';
  const c = document.createElement('canvas');
  c.width = c.height = 48;
  const x = c.getContext('2d');
  x.fillStyle = '#0b1340';
  x.fillRect(0, 0, 48, 48);
  if (r === 3) {
    x.fillStyle = 'rgba(242,184,75,.25)';
    x.fillRect(0, 0, 48, 48);
  }
  if (own) {
    x.imageSmoothingEnabled = true;
    x.drawImage(own, 0, 0, 48, 48);
  } else {
    x.imageSmoothingEnabled = false;
    x.drawImage(sheet, (i % 16) * 32, ((i / 16) | 0) * 32, 32, 32, 0, 0, 48, 48);
  }
  x.strokeStyle = it.color || '#fff';
  x.lineWidth = 3;
  x.strokeRect(1.5, 1.5, 45, 45);
  return (made[k] = c.toDataURL());
}
