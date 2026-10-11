// The look of the game: colors and sizes, chosen to match the old game
// exactly (old line numbers in brackets). Only looks live here, never rules.

// Biome colors for the minimap, water and cliff sides [old L982-988].
export const BIOMES = {
  Grasslands: { g: '#4f8a3a', sl: '#5a4a30', sr: '#4a3a24' },
  Forest: { g: '#2f6a32', sl: '#3a2a1a', sr: '#2a1e12' },
  Swamp: { g: '#4a5a2a', sl: '#3a3220', sr: '#2a2416' },
  Desert: { g: '#d9b46a', sl: '#9a6a3a', sr: '#7a5230' },
  Tundra: { g: '#e8f0f8', sl: '#7a90b0', sr: '#60789a' },
  Volcanic: { g: '#3a2a2a', sl: '#2a1a18', sr: '#1e1210' },
  Crypt: { g: '#5a5a6a', sl: '#2a2a38', sr: '#20202c' },
};
export const biome = (name) => BIOMES[name] ? Object.assign({ n: name }, BIOMES[name]) : Object.assign({ n: 'Grasslands' }, BIOMES.Grasslands);

// Minimap tile colors by ground letter [old L1268].
export const MINI = { r: '#b8975f', f: '#2f6a2a', h: '#8a8f6a', t: '#8d8f99', T: '#f2c14e', b: '#a8743a', w: '#2a6ab0', m: '#6b5f55' };

// Element colors [old L431].
export const ELEMENT = { fire: '#e4572e', ice: '#7fd1ff', elec: '#f2e14b', light: '#fff6c8', dark: '#8a5fd0' };

// Hero glow by class and stage, and each class's own aura [old L555, L1047].
export const GLOW = {
  Warrior: ['', '#7fc4ff', '#ffd45a'],
  Tank: ['', '#c9d6ff', '#5df0d0'],
  Mage: ['', '#7fd1ff', '#c79bff'],
  Healer: ['', '#bfffd0', '#fff2a0'],
  Ranger: ['', '#a8ff7a', '#4fd1ff'],
};
export const AURA = { Warrior: '#ffd45a', Tank: '#9fd8ff', Mage: '#c79bff', Healer: '#fff6c8', Ranger: '#7be0a8' };

// Chest colors [old CHT].
export const CHEST = {
  wood: { c: '#8a5a2b', b: '#5a3a1a', name: 'Wooden crate' },
  silver: { c: '#c9ced8', b: '#7a8090', name: 'Silver chest' },
  gold: { c: '#f2c14e', b: '#a07a1a', name: 'Gold chest' },
};

// Map units: the size of a tile, flat and isometric [old L787, L806].
export const TS = 72;
export const HW = 48;
export const HH = 24;
export const OY = 110;

// Classes drawn bigger, and the ones that fly [old L432].
export const BIG = { 'Witch Knight': 1, 'Dragon Knight': 1 };
export const FLY = { 'Dragon Knight': 1, 'Demon Lord': 1, 'Bone Gargoyle': 1, 'Bat Lord': 1, 'Golden Drake': 1 };
