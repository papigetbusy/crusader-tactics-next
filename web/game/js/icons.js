// Item pictures the owner signed off on (2026-10-17). Painted pixel art.
// Common, uncommon, rare, and epic each have their own picture. A legendary
// that is not one of the class weapons uses the epic picture with a gold tint.
// The ten class legendaries, Heart of the Stair, and Phoenix Feather have
// their own pictures. Only looks.
import { pic } from './assets.js';

const WEAPONS = ['sword', 'axe', 'lance', 'bow', 'dagger', 'staff', 'wand', 'gun'];
const made = {};

// icon is a 48×48 picture (as a data URL) for an item from the referee's View.
// legendary is the top rarity number (it gets the gold look).
export function icon(it, legendary = 4) {
  const tier = it.tier || 0;
  const leg = tier >= legendary;
  const k = (it.slot || '') + it.name + tier + (it.weaponType || '') + (it.weight || '');
  if (made[k]) return made[k];
  const art = picture(it, tier);
  if (!art) return '';
  const c = document.createElement('canvas');
  c.width = c.height = 48;
  const x = c.getContext('2d');
  x.fillStyle = '#0b1340';
  x.fillRect(0, 0, 48, 48);
  if (leg) {
    x.fillStyle = 'rgba(242,184,75,.25)';
    x.fillRect(0, 0, 48, 48);
  }
  x.imageSmoothingEnabled = true;
  x.drawImage(art, 2, 2, 44, 44);
  x.strokeStyle = it.color || '#fff';
  x.lineWidth = 3;
  x.strokeRect(1.5, 1.5, 45, 45);
  return (made[k] = c.toDataURL());
}

function picture(it, tier) {
  const own = pic('legendary', it.name) || pic('items', it.name);
  if (own) return own;
  const slot = it.slot || '';
  if (slot === 'w' || slot.startsWith('w')) {
    const kind = WEAPONS.includes(it.weaponType) ? it.weaponType : 'sword';
    return pic('items', kind + '-' + Math.min(3, tier));
  }
  if (slot === 'c') return pic('items', 'charm-' + Math.min(2, tier));
  const weight = it.weight || 'L';
  const part = 'hab'.includes(slot) ? slot : 'a';
  return pic('items', weight + '-' + part + '-' + Math.min(3, tier));
}
