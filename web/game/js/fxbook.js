// The attack-effect tables: for each skill, which drawn effect plays (its
// base), its colors, extras and count; how long each base lasts and when it
// lands; and the pose its user strikes. Pictures only, never rules.
// Ported from the old game [old index.html L801-803, L869-870, L876-916,
// L426 (FXN_PRE), L430 (CAUR), L1093, L1111]. The 53 skills added in the new
// game had no entry in the old game; theirs are marked NEW below.

// SKFX[name] = [base, 'color1,color2[,color3]', extras, count].
// Extras: f flash, s shake, x cross, e explosion, b blood, c poison cloud,
// t stars, r ring, n net [old drawP L922-966].
export const SKFX = {
  // Warrior
  Slash: ['cut', '#ffffff,#9fd8ff'], Rend: ['cut', '#ffffff,#ff4d4d', 'b'], Whirlwind: ['swirl', '#ffffff,#9fd8ff', ''], 'Crescent Wave': ['wave', '#cfeaff,#ffffff'], 'Power Strike': ['cut', '#fff6a8,#ffd84a', 'xs'], Execution: ['exec', '#ffffff,#ff4d4d', 'fs'], 'Dragon Slash': ['cut', '#fff6a8,#ff7a2e', 'xfse'],
  'Blood Rage': ['claw', '#ff4d4d,#7a0a1a', 'se', 5], 'Holy Judgment': ['pillar', '#fff6c8,#ffd84a', 'fs'], 'Thousand Cuts': ['cut', '#e8d0ff,#ffe45c', 'x', 5],
  // Tank
  Bash: ['smash', '#cfd6e6,#ffd84a'], 'Shield Bash': ['smash', '#9fd8ff,#ffffff', 't'], 'Ground Pound': ['quake', '#d9c08a,#5a452c', 's'], Ironhide: ['buff', '#9fd8ff,#ffffff'], 'Crushing Blow': ['smash', '#ff7a5a,#cfd6e6', 'sx'], 'Fortress Slam': ['quake', '#ffffff,#8d99b8', 'sf'], "Titan's Wrath": ['pillar', '#ffd84a,#ffffff', 'sf'],
  Earthshatter: ['quake', '#ff9a5a,#3b2a18', 'sfe'], Bulwark: ['buff', '#bfeaff,#7fd1ff', 'r'], 'War Cry': ['buff', '#ff7a5a,#ffd84a', 'r'],
  // Mage
  Fireball: ['orb', '#ffb12e,#fff6a8,#e4572e'], Thunder: ['bolt', '#ffd84a,#ffffff'], 'Frost Bolt': ['orb', '#7fd1ff,#ffffff,#3a8fd0'], 'Ice Lance': ['lance', '#bfeaff,#ffffff'], 'Chain Lightning': ['chain', '#ffe45c,#ffffff'], Blizzard: ['blizzard', '#ffffff,#bfeaff'], Meteor: ['meteor', '#ffb12e,#e4572e', 'sf'], 'Arcane Nova': ['nova', '#d9b8ff,#8a5fd0', 'sf'],
  Inferno: ['rain', '#ff7a2e,#ffd84a', 'fe', 22], 'Absolute Zero': ['shards', '#e8fbff,#7fd1ff', 'sf'], Thunderstorm: ['bolt', '#ffe45c,#ffffff', 'f', 3],
  // Healer
  'Mend Self': ['heal', '#6ee7a0,#d8ffe4'], 'Heal Ally': ['heal', '#6ee7a0,#ffffff'], Cleanse: ['buff', '#ffffff,#bfeaff'], 'Radiant Burst': ['beam', '#fff6c8,#ffd84a'], Blessing: ['buff', '#ffd84a,#fff6c8', 'r'], 'Holy Siphon': ['beam', '#fff6c8,#6ee7a0'], 'Mass Heal': ['heal', '#6ee7a0,#d8ffe4', 'r', 3], 'Prism Aura': ['buff', '#ff9bd8,#7fd1ff', 'r'], 'Sacred Flame': ['pillar', '#ffd84a,#ff9a2e', 'e'], Renewal: ['heal', '#9ae6b4,#ffffff', 'r', 3], 'Divine Grace': ['pillar', '#fff6c8,#ffffff', 'f'], Resurrection: ['revive', '#fff6c8,#ffd84a', 'f'],
  Sanctuary: ['pillar', '#fff6c8,#6ee7a0', 'fr'], 'Time Warp': ['nova', '#7fd1ff,#ffe45c', 'r'], Pandemic: ['cloud', '#8a5fd0,#6fe04a', 's'],
  // Ranger
  Shot: ['arrow', '#8a5a2b,#e8ecf5'], 'Snaring Shot': ['arrow', '#8a5a2b,#ffffff', 'n'], Multishot: ['arrow', '#8a5a2b,#e8ecf5', '', 3], 'Poison Arrow': ['arrow', '#6fe04a,#c8ff9a', 'c'], 'Pinning Shot': ['arrow', '#c9ced8,#ffffff', 't'], 'Piercing Shot': ['beam', '#bfeaff,#ffffff'], 'Power Shot': ['arrow', '#ffd84a,#ff7a2e', 'se'], 'Rain of Arrows': ['rain', '#8a5a2b,#e8ecf5', 's', 18],
  Deadeye: ['beam', '#ffffff,#7fd1ff', 'fs'], 'Snare Field': ['rain', '#c9ced8,#ffffff', 'n', 12], 'Gale Volley': ['arrow', '#7ff0c0,#ffffff', '', 5],
  // Assassin
  Stab: ['thrust', '#e8ecf5,#c79bff'], 'Venom Edge': ['thrust', '#6fe04a,#e8ecf5', 'c'], Ambush: ['blink', '#b48aff,#ffffff'], 'Fan of Knives': ['knives', '#e8ecf5,#b48aff', '', 6], 'Smoke Bomb': ['cloud', '#5a5a6a,#9a9aaa'], Assassinate: ['blink', '#ff4d4d,#ffffff', 'x'], 'Death Blossom': ['knives', '#ff6a8a,#ffffff', 's', 10],
  'Phantom Rush': ['blink', '#d9b8ff,#8a5fd0', 'f'], 'Venom Storm': ['cloud', '#6fe04a,#2a5a2a', 's'], 'Crimson Waltz': ['cut', '#ff4d4d,#ffd84a', 'x', 5],
  // Gunslinger
  'Quick Shot': ['shot', '#ffe45c,#ffffff'], 'Double Tap': ['shot', '#ffe45c,#ffffff', '', 2], Ricochet: ['ricochet', '#ffe45c,#ffffff', '', 3], 'Flash Bang': ['flashbang', '#ffffff,#c9ced8', 'f'], 'Scatter Shot': ['scatter', '#ffb12e,#ffffff'], 'Hollow Point': ['shot', '#ff7a2e,#ffd84a', 'se', 1], 'Bullet Hell': ['gunrain', '#ffe45c,#ff7a2e', 's', 16],
  'Six Shooter': ['shot', '#ffd84a,#ff7a2e', 's', 6], 'Dead or Alive': ['shot', '#ff4d4d,#ffffff', 'fse', 1], 'Ricochet Storm': ['ricochet', '#7fe3f0,#ffffff', 's', 5],
  // Witch
  'Toxic Bolt': ['orb', '#6fe04a,#c8ff9a,#2a5a2a'], Hex: ['debuff', '#c79bff,#5a2a8a'], Miasma: ['cloud', '#6fe04a,#3a8a3a'], 'Curse of Frailty': ['debuff', '#ff4d6a,#5a0a1a', 'r'], 'Plague Swarm': ['swarm', '#c8ff9a,#3a5a2a'], "Witch's Brew": ['potion', '#6fe04a,#c79bff'], 'Hex Storm': ['nova', '#8a5fd0,#6fe04a', 'sf'],
  'Grand Hex': ['debuff', '#bfeaff,#7fd1ff', 'fr'], 'Soul Nova': ['nova', '#9ae6b4,#1a3a2a', 'sf'], 'Blood Moon': ['pillar', '#c8102e,#ff4d4d', 'fs'],
  // Knight
  'Lance Strike': ['thrust', '#e8ecf5,#ffd84a'], Charge: ['charge', '#ffd84a,#ffffff'], 'Lance Break': ['smash', '#ff7a5a,#e8ecf5', 'x'], 'Running Thrust': ['charge', '#9fd8ff,#ffffff'], Valor: ['buff', '#ffd84a,#ffffff', 'r'], 'Grand Charge': ['charge', '#ffd84a,#ff7a2e', 's'], 'Cavalry Breaker': ['charge', '#ffffff,#ff4d4d', 'fse'],
  Crusade: ['charge', '#fff6c8,#ffd84a', 'fs'], 'Doom Lance': ['beam', '#5a1f1f,#ff4d4d', 'fs'], 'Lightning Charge': ['charge', '#ffe45c,#7fe3f0', 'fs'],
  // Dragon Knight
  'Spear Thrust': ['thrust', '#e8ecf5,#ff9a5a'], 'Dragon Claw': ['claw', '#ff7a5a,#ffffff', 'b'], Dragonfire: ['breath', '#ffb12e,#e4572e'], 'Wing Gust': ['swirl', '#cfeaff,#ffffff'], 'Sky Dive': ['dive', '#ffffff,#ff9a5a', 's'], Impale: ['beam', '#e8ecf5,#ff9a5a'], "Dragon's Fury": ['breath', '#ff7a2e,#ffd84a', 'sfe'],
  'Hellfire Dive': ['dive', '#ff7a2e,#ffd84a', 'sfe'], 'Glacial Breath': ['breath', '#bfeaff,#ffffff', 'sf'], 'Thunder Lance': ['beam', '#ffe45c,#7fe3f0', 'fs'],
  // monster tier / elites / bosses
  'Savage Claw': ['claw', '#ff4d4d,#ffffff', 'b'], Rampage: ['claw', '#ff7a2e,#ffd84a', 's', 5], 'Shell Slam': ['smash', '#9a8f84,#cfd6e6', 't'], 'Stone Quake': ['quake', '#9a7b4f,#3b2a18', 's'], 'Greater Spell': ['meteor', '#c79bff,#ff7a2e', 's'], Cataclysm: ['nova', '#ff4d4d,#8a5fd0', 'sf'], 'Venom Spit': ['potion', '#6fe04a,#c8ff9a'], Barrage: ['rain', '#c9ced8,#8a5a2b', '', 14], 'Mass Mend': ['heal', '#6ee7a0,#d8ffe4', 'r', 3], 'Toxic Spores': ['cloud', '#c8ff9a,#6fe04a'], 'Spore Bolt': ['orb', '#c8ff9a,#6fe04a,#3a5a2a'],
  'Heavy Slash': ['cut', '#ffffff,#8a8f99', 'x'], 'Shield Crush': ['smash', '#5a6584,#ffffff', 't'], Claw: ['claw', '#ffd84a,#ffffff'], 'Fire Breath': ['breath', '#ffb12e,#e4572e'], Backstab: ['blink', '#444455,#ffffff'], 'Poison Blade': ['thrust', '#6fe04a,#ffffff', 'c'], Smash: ['smash', '#6a8a4a,#ffd84a', 's'], 'Ground Slam': ['quake', '#d9c08a,#5a452c', 'st'],
  'Wing Slash': ['cut', '#ff4d6a,#ffffff'], 'Blood Drain': ['beam', '#c8102e,#ff9bd8'], 'Dive Strike': ['dive', '#ff4d6a,#ffffff', 's'], 'Soul Bolt': ['orb', '#c79bff,#ffffff,#5a2a8a'], 'Withering Curse': ['debuff', '#8a5fd0,#3b1b66', 'r'], 'Shadow Nova': ['nova', '#8a5fd0,#3b1b66', 'sf'], 'Claw Swipe': ['claw', '#ff7a2e,#ffffff'], 'Flame Breath': ['breath', '#ff7a2e,#ffd84a', 's'], 'Tail Slam': ['smash', '#8a2a2a,#ffd84a', 'st'],
  'Flail Smash': ['smash', '#5a3a6a,#cfd6e6', 's'], Shackle: ['debuff', '#c9ced8,#5a5a6a'], 'Club Smash': ['smash', '#8a5a2b,#ffd84a', 's'], 'War Stomp': ['quake', '#9a7b4f,#3b2a18', 'st'], 'Grave Bolt': ['orb', '#9ae6b4,#1a3a2a,#3a8a5a'], 'Plague Mist': ['cloud', '#6fe04a,#3a5a2a'], 'Dive Claw': ['dive', '#c9b98a,#ffffff'], 'Bone Storm': ['swirl', '#e8dcc0,#9a8f84', '', 1], 'Glacier Fist': ['smash', '#bfeaff,#ffffff', 's'], 'Frost Nova': ['shards', '#bfeaff,#ffffff', 's'],
  'Death Bolt': ['orb', '#8a5fd0,#ffffff,#3b1b66'], 'Soul Rend': ['beam', '#9ae6b4,#3a8a5a'], 'Grave Nova': ['nova', '#9ae6b4,#3b1b66', 'sf'], Constrict: ['debuff', '#3a5a4a,#9ae6b4', 't'], 'Venom Wave': ['cloud', '#6fe04a,#3a5a4a', 's'], 'Tidal Crush': ['shards', '#7fd1ff,#3a8fd0', 's'], 'Infernal Claw': ['claw', '#ff4d4d,#ffd84a', 's'], Hellfire: ['rain', '#ff7a2e,#ffd84a', 'fs', 20], 'Dark Pact': ['buff', '#c8102e,#5a0a1a', 'r'],
  // [L1093, L1111]
  Cloak: ['cloud', '#5a5a6a,#9a9aaa'],
  Strike_m: ['cut', '#ffffff,#c9ced8'], Strike_r: ['arrow', '#8a5a2b,#e8ecf5'], Strike_gun: ['shot', '#ffe45c,#ffffff'], Strike_g: ['orb', '#c79bff,#ffffff,#5a2a8a'],
};

// NEW: the skills the new game added. The old game drew a plain white slash
// for any skill without an entry; these were chosen to fit each skill.
Object.assign(SKFX, {
  'Shoulder Check': ['smash', '#cfd6e6,#ffd84a', 's'], 'Riposte Stance': ['buff', '#ffffff,#9fd8ff', 'r'], 'Leaping Cleave': ['dive', '#ffffff,#9fd8ff', 's'], Sunder: ['cut', '#ffd84a,#ff7a5a', 'x'], 'Battle Trance': ['buff', '#ff7a5a,#ffd84a', 'r'], Momentum: ['cut', '#ffffff,#ffd84a', '', 2],
  'Hook Chain': ['thrust', '#c9ced8,#8d99b8'], Bodyguard: ['buff', '#9fd8ff,#ffffff', 'r'], 'Iron Shell': ['buff', '#bfeaff,#8d99b8', 'r'], 'Taunting Roar': ['nova', '#ff7a5a,#ffd84a', 's'], Aftershock: ['quake', '#d9c08a,#5a452c', 's'], Unstoppable: ['charge', '#cfd6e6,#9fd8ff', 's'],
  Blink: ['blink', '#c79bff,#ffffff'], 'Mana Burn': ['orb', '#7fd1ff,#ffffff,#3a3aa0'], 'Static Field': ['bolt', '#ffe45c,#ffffff', '', 2], 'Gravity Well': ['nova', '#8a5fd0,#1a0a3a', 's'], Echo: ['buff', '#d9b8ff,#ffffff', 'r'],
  'Last Light': ['pillar', '#fff6c8,#ffd84a', 'fr'],
  'Steady Aim': ['buff', '#7be0a8,#ffffff'], Disengage: ['arrow', '#8a5a2b,#e8ecf5'], "Hunter's Mark": ['arrow', '#ff4d4d,#ffffff', 't'], 'Concussive Arrow': ['arrow', '#cfd6e6,#ffd84a', 's'], 'Flaming Volley': ['arrow', '#ff7a2e,#ffd84a', 'e', 3],
  Shadowstep: ['blink', '#5a5a6a,#c79bff'], Garrote: ['thrust', '#c9ced8,#5a2a8a'], 'Bait and Switch': ['blink', '#b48aff,#ffffff'], 'Crippling Cut': ['cut', '#ffffff,#ff4d4d', 'b'], 'Open Veins': ['claw', '#ff4d4d,#7a0a1a', 'b', 3], 'Twin Fangs': ['cut', '#e8ecf5,#b48aff', 'x'],
  'Boot Shot': ['shot', '#ffe45c,#ffffff'], 'Disarming Shot': ['shot', '#c9ced8,#ffffff'], 'Roll Out': ['shot', '#ffe45c,#ffffff', '', 2], 'Powder Keg': ['flashbang', '#ff7a2e,#ffd84a', 'fse'], 'Lucky Coin': ['buff', '#ffd84a,#ffffff', 'r'], 'High Noon': ['shot', '#ff4d4d,#ffd84a', 'fs', 1],
  'Tail Sweep': ['swirl', '#ff9a5a,#ffffff'], 'Dragon Scales': ['buff', '#ff9a5a,#ffd84a', 'r'], 'Wyrm Roar': ['breath', '#ffd84a,#ff7a2e', 's'], 'Ember Trail': ['charge', '#ff7a2e,#ffd84a', 'e'], Skewer: ['thrust', '#e8ecf5,#ff9a5a'], Wyrmbond: ['buff', '#ff7a2e,#ffd84a', 'r'],
  'Voodoo Link': ['debuff', '#c79bff,#5a2a8a', 'r'], Wither: ['debuff', '#8a5fd0,#3b1b66'], 'Puppet Strings': ['debuff', '#ff9bd8,#5a2a8a', 't'], 'Soul Siphon': ['beam', '#9ae6b4,#5a2a8a'], Contagion: ['cloud', '#6fe04a,#3a5a2a', 's'], 'Bog of Despair': ['cloud', '#3a5a4a,#6fe04a'],
  Trample: ['charge', '#ffd84a,#cfd6e6', 's'], 'Shield of Faith': ['buff', '#fff6c8,#ffd84a', 'r'], Unhorse: ['smash', '#e8ecf5,#ffd84a', 's'], 'Lance Wall': ['buff', '#e8ecf5,#ffd84a', 'r'], Rescue: ['buff', '#9fd8ff,#ffffff', 'r'], 'Relentless Charge': ['charge', '#ffd84a,#ff4d4d', 'fse'],
});

// Each class ability glows in its class's color [L430, L1093].
const CAUR = { Warrior: '#ffd45a', Tank: '#9fd8ff', Mage: '#c79bff', Healer: '#fff6c8', Ranger: '#7be0a8' };
for (const c of ['Warrior', 'Tank', 'Mage', 'Healer', 'Ranger', 'Assassin', 'Gunslinger', 'Witch', 'Knight', 'Dragon Knight']) SKFX['AB_' + c] = ['buff', (CAUR[c] || '#ffd84a') + ',#ffffff', 'r'];

// How long each base plays and when it lands, in ms [L913-914].
export const BDUR = { cut: 460, thrust: 430, smash: 620, claw: 480, swirl: 700, wave: 780, exec: 900, quake: 880, buff: 900, debuff: 950, orb: 820, bolt: 700, lance: 780, chain: 900, blizzard: 1050, meteor: 1100, nova: 1000, rain: 1150, shards: 900, heal: 950, beam: 760, pillar: 1150, revive: 1300, arrow: 640, knives: 700, blink: 620, cloud: 1050, shot: 520, ricochet: 900, flashbang: 900, scatter: 600, gunrain: 1100, swarm: 1100, potion: 1000, charge: 620, breath: 950, dive: 850, summon: 1200 };
export const BIMP = { cut: 150, thrust: 170, smash: 240, claw: 150, swirl: 150, wave: 430, exec: 360, quake: 150, buff: 200, debuff: 300, orb: 300, bolt: 200, lance: 320, chain: 200, blizzard: 300, meteor: 500, nova: 150, rain: 450, shards: 220, heal: 250, beam: 200, pillar: 450, revive: 400, arrow: 320, knives: 300, blink: 330, cloud: 300, shot: 130, ricochet: 150, flashbang: 450, scatter: 130, gunrain: 400, swarm: 450, potion: 500, charge: 260, breath: 300, dive: 420, summon: 500 };

// The user's pose for each base [L915]; anything else lunges.
export const ACT = { shot: 'recoil', scatter: 'recoil', ricochet: 'recoil', gunrain: 'recoil', flashbang: 'throw', arrow: 'recoil', potion: 'throw', orb: 'cast', bolt: 'cast', lance: 'cast', chain: 'cast', blizzard: 'cast', meteor: 'cast', nova: 'cast', rain: 'cast', shards: 'cast', heal: 'cast', beam: 'cast', pillar: 'cast', revive: 'cast', buff: 'cast', debuff: 'cast', cloud: 'cast', swarm: 'cast', summon: 'cast', charge: 'dash', blink: 'blink', breath: 'rear', dive: 'dash', knives: 'throw' };
// How long each pose lasts [old anim L791].
export const POSE_MS = { dash: 560, cast: 720, blink: 600 };

// The older effects, for skills without an entry [L802-803, L426, L869-870].
export const FXN = { 'Wing Slash': 'slash', 'Blood Drain': 'dark', 'Dive Strike': 'heavy', 'Soul Bolt': 'dark', 'Withering Curse': 'dark', 'Shadow Nova': 'dark', 'Claw Swipe': 'slash', 'Flame Breath': 'fire', 'Tail Slam': 'heavy', Claw: 'slash', 'Heavy Slash': 'heavy', 'Shield Crush': 'shield', 'Fire Breath': 'fire', Backstab: 'slash', 'Poison Blade': 'rend', Smash: 'heavy', 'Ground Slam': 'bash', Slash: 'slash', Rend: 'rend', 'Power Strike': 'heavy', Bash: 'bash', 'Shield Bash': 'shield', Fireball: 'fire', Thunder: 'bolt', 'Frost Bolt': 'ice', 'Snaring Shot': 'net', 'Power Shot': 'parrow', Shot: 'arrow', 'Mend Self': 'heal', 'Heal Ally': 'heal', 'Divine Grace': 'grace',
  Whirlwind: 'whirl', 'Crescent Wave': 'wave', Execution: 'exec', 'Dragon Slash': 'dslash', 'Ground Pound': 'quake', 'Crushing Blow': 'crush', 'Fortress Slam': 'slam', "Titan's Wrath": 'titan', Ironhide: 'cleanse', 'Ice Lance': 'lance', 'Chain Lightning': 'chain', Blizzard: 'blizzard', Meteor: 'meteor', 'Arcane Nova': 'nova', Cleanse: 'cleanse', 'Radiant Burst': 'holy', Blessing: 'cleanse', 'Holy Siphon': 'holy', 'Prism Aura': 'nova', 'Sacred Flame': 'holy', Renewal: 'mheal', 'Mass Heal': 'mheal', Resurrection: 'revive', Multishot: 'volley', 'Poison Arrow': 'poison', 'Pinning Shot': 'pin', 'Piercing Shot': 'pierce', 'Rain of Arrows': 'rain', 'Blood Rage': 'rampage', 'Holy Judgment': 'titan', 'Thousand Cuts': 'rampage', Earthshatter: 'slam', Bulwark: 'cleanse', 'War Cry': 'nova', Inferno: 'meteor', 'Absolute Zero': 'blizzard', Thunderstorm: 'chain', Sanctuary: 'grace', 'Time Warp': 'nova', Pandemic: 'nova', Deadeye: 'pierce', 'Snare Field': 'net', 'Gale Volley': 'volley', 'Savage Claw': 'claw', Rampage: 'rampage', 'Shell Slam': 'crush', 'Stone Quake': 'quake', 'Venom Spit': 'spit', Barrage: 'rain', 'Mass Mend': 'mheal' };
export const EE = { fire: 'fire', ice: 'ice', elec: 'bolt', light: 'light', dark: 'dark' };
export const DUR = { slash: 420, rend: 560, heavy: 650, bash: 450, shield: 750, fire: 850, bolt: 700, ice: 800, light: 750, dark: 800, heal: 950, grace: 1400, arrow: 650, parrow: 750, net: 950,
  whirl: 700, wave: 800, exec: 900, dslash: 800, quake: 900, crush: 800, slam: 900, titan: 1100, lance: 800, chain: 900, meteor: 1100, blizzard: 1100, nova: 1000, cleanse: 900, holy: 900, mheal: 1000, revive: 1300, volley: 800, poison: 1000, pin: 900, pierce: 800, rain: 1200, claw: 700, rampage: 900, spit: 900, bullet: 500 };
export const IMP = { slash: 160, rend: 160, heavy: 230, bash: 150, shield: 150, fire: 300, bolt: 200, ice: 280, light: 150, dark: 150, heal: 250, grace: 450, arrow: 330, parrow: 380, net: 330,
  whirl: 150, wave: 440, exec: 360, dslash: 150, quake: 150, crush: 240, slam: 150, titan: 300, lance: 320, chain: 200, meteor: 500, blizzard: 300, nova: 150, cleanse: 200, holy: 450, mheal: 200, revive: 400, volley: 330, poison: 350, pin: 330, pierce: 200, rain: 450, claw: 150, rampage: 200, spit: 360, bullet: 150 };
// Screen shake by effect [L801].
export const SHK0 = { heavy: 5, bolt: 3, bash: 3, fire: 2 };

// fxKey, durOf, impOf: the effect key of a skill and its timings [L916].
export const fxKey = (n) => (SKFX[n] ? 'P:' + n : null);
export const durOf = (k) => (k && k.startsWith('P:') ? BDUR[SKFX[k.slice(2)][0]] || 800 : DUR[k] || 800);
export const impOf = (k) => (k && k.startsWith('P:') ? BIMP[SKFX[k.slice(2)][0]] || 250 : IMP[k] || 250);
