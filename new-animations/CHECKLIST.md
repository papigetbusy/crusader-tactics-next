# New animations

This folder is not part of the game. Nothing here is loaded by the play page.

## Signed off
Not in the game yet. The code is waiting in `code/`.

- [x] Blood Moon (Blood Witch) — `code/blood-moon.js`
- [x] Grand Hex (Hex Queen) — `code/grand-hex.js`
- [x] Blood Frenzy (Berserker) — `code/blood-frenzy.js`

## How this works

1. You send the sprite sheet, or you drop a pile of them in `inbox/` and run `python3 new-animations/bot/make.py`.
2. A test GIF is made. The bot can make the first pass for many skills, then it stops. It does not write code and it does not sign off.
3. We go through them one by one and tweak until you say that version is the one.
4. Then one code file is written in `code/`, named after the attack. Test GIFs never get their own code.
5. An in-game GIF is made from that code: the move hitting an enemy. The effect plays with the sprite, and it is not allowed to cover or cut the attacker's sprite.
6. The code and that GIF stay here. The play page does not load them.
7. When every row below is signed off, they all go into the game in one pass. Not before.

Look, from 2026-10-12 on. Blood Frenzy is the lock. The prompt is in `PROMPT.md`. Eight cells, two rows, the hero called with @, every pose and every effect step written out. The review uses that sheet's own effect. A power-up stays on him. A hit travels to the enemy. Either way the effect is behind him and does not cover his face. The last three frames are body 1 with the glow gone. He does not stay on the attack pose. Blood Moon was signed before this rule, so its ending is still the swing until it is updated.

A shared skill can be used by every form of that class, so it needs a signed GIF on each body (the base forms and the ascensions). An ascension skill needs a GIF only on the body that learns it. One code file still covers the attack. The extra GIFs are the same move on the other bodies.

Walk cycles are not on this list. They are a separate pass.

A counterattack uses the Strike or Defend GIF. It does not get its own animation unless you ask for one.

## Status words

- gif: none, or `vN made, not signed off`, or `signed off (vN)`
- code: waiting, or `written` once you have signed the GIF off

## The list

### Warrior

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Double Strike (the class ability; Double Strike+, Triple Strike use this same GIF unless we split them) — gif: none — code: waiting
- [x] Blood Frenzy (Berserker ability only) — approved. Code is in `code/blood-frenzy.js`. Not in the game.
- [ ] Flurry of Blades (Duelist ability only) — gif: none — code: waiting
- [ ] Oath of Protection (Paladin ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Power Strike — gif: bot review made, not signed off — code: waiting
- [ ] Rend — gif: none — code: waiting
- [ ] Slash — gif: none — code: waiting
- [ ] Whirlwind — gif: none — code: waiting
- [ ] Crescent Wave — gif: none — code: waiting
- [ ] Shoulder Check — gif: none — code: waiting
- [ ] Execution — gif: bot review made, not signed off — code: waiting
- [ ] Riposte Stance — gif: none — code: waiting
- [ ] Leaping Cleave — gif: none — code: waiting
- [ ] Dragon Slash — gif: bot review made, not signed off — code: waiting
- [ ] Sunder — gif: none — code: waiting
- [ ] Battle Trance — gif: none — code: waiting
- [ ] Momentum — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Blood Rage (Berserker only) — gif: bot review made, not signed off — code: waiting
- [ ] Thousand Cuts (Duelist only) — gif: bot review made, not signed off — code: waiting
- [ ] Holy Judgment (Paladin only) — gif: bot review made, not signed off — code: waiting

### Knight

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Vow of Honor (the class ability; Vow of Honor+, Vow of Honor II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Oath of Ruin (Black Knight ability only) — gif: none — code: waiting
- [ ] Thunderous Oath (Storm Rider ability only) — gif: none — code: waiting
- [ ] Sacred Vow (Templar ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Charge — gif: none — code: waiting
- [ ] Grand Charge — gif: none — code: waiting
- [ ] Lance Strike — gif: none — code: waiting
- [ ] Lance Break — gif: none — code: waiting
- [ ] Running Thrust — gif: none — code: waiting
- [ ] Valor — gif: none — code: waiting
- [ ] Trample — gif: none — code: waiting
- [ ] Shield of Faith — gif: none — code: waiting
- [ ] Unhorse — gif: none — code: waiting
- [ ] Cavalry Breaker — gif: none — code: waiting
- [ ] Lance Wall — gif: none — code: waiting
- [ ] Rescue — gif: none — code: waiting
- [ ] Relentless Charge — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Doom Lance (Black Knight only) — gif: none — code: waiting
- [ ] Lightning Charge (Storm Rider only) — gif: none — code: waiting
- [ ] Crusade (Templar only) — gif: none — code: waiting

### Dragon Knight

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Beast Mode (the class ability; Beast Mode+, Beast Mode II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Glacial Scales (Frost Wyrm ability only) — gif: none — code: waiting
- [ ] Inferno Beast (Inferno Rider ability only) — gif: none — code: waiting
- [ ] Sky Lord (Storm Dragoon ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Dragon Claw — gif: none — code: waiting
- [ ] Impale — gif: none — code: waiting
- [ ] Spear Thrust — gif: none — code: waiting
- [ ] Dragonfire — gif: none — code: waiting
- [ ] Wing Gust — gif: none — code: waiting
- [ ] Sky Dive — gif: none — code: waiting
- [ ] Tail Sweep — gif: none — code: waiting
- [ ] Dragon Scales — gif: none — code: waiting
- [ ] Wyrm Roar — gif: none — code: waiting
- [ ] Dragon's Fury — gif: none — code: waiting
- [ ] Ember Trail — gif: none — code: waiting
- [ ] Skewer — gif: none — code: waiting
- [ ] Wyrmbond — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Glacial Breath (Frost Wyrm only) — gif: none — code: waiting
- [ ] Hellfire Dive (Inferno Rider only) — gif: none — code: waiting
- [ ] Thunder Lance (Storm Dragoon only) — gif: none — code: waiting

### Assassin

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Bloodlust (the class ability; Bloodlust+, Bloodlust II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Crimson Bloodlust (Blade Dancer ability only) — gif: none — code: waiting
- [ ] Shadow Clones (Phantom ability only) — gif: none — code: waiting
- [ ] Death Mark (Viper ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Assassinate — gif: none — code: waiting
- [ ] Stab — gif: none — code: waiting
- [ ] Venom Edge — gif: none — code: waiting
- [ ] Ambush — gif: none — code: waiting
- [ ] Fan of Knives — gif: none — code: waiting
- [ ] Shadowstep — gif: none — code: waiting
- [ ] Smoke Bomb — gif: none — code: waiting
- [ ] Garrote — gif: none — code: waiting
- [ ] Bait and Switch — gif: none — code: waiting
- [ ] Death Blossom — gif: none — code: waiting
- [ ] Crippling Cut — gif: none — code: waiting
- [ ] Open Veins — gif: none — code: waiting
- [ ] Twin Fangs — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Crimson Waltz (Blade Dancer only) — gif: none — code: waiting
- [ ] Phantom Rush (Phantom only) — gif: none — code: waiting
- [ ] Venom Storm (Viper only) — gif: none — code: waiting

### Ranger

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Hunter's Eye (the class ability; Hunter's Eye+, Hunter's Eye II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Deadeye Focus (Sniper ability only) — gif: none — code: waiting
- [ ] Storm Eye (Tempest Archer ability only) — gif: none — code: waiting
- [ ] Snare Field (Trapper ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Power Shot — gif: none — code: waiting
- [ ] Shot — gif: none — code: waiting
- [ ] Snaring Shot — gif: none — code: waiting
- [ ] Multishot — gif: none — code: waiting
- [ ] Poison Arrow — gif: none — code: waiting
- [ ] Pinning Shot — gif: none — code: waiting
- [ ] Piercing Shot — gif: none — code: waiting
- [ ] Steady Aim — gif: none — code: waiting
- [ ] Disengage — gif: none — code: waiting
- [ ] Rain of Arrows — gif: none — code: waiting
- [ ] Hunter's Mark — gif: none — code: waiting
- [ ] Concussive Arrow — gif: none — code: waiting
- [ ] Flaming Volley — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Deadeye (Sniper only) — gif: none — code: waiting
- [ ] Gale Volley (Tempest Archer only) — gif: none — code: waiting
- [ ] Snare Field (Trapper only) — gif: none — code: waiting

### Gunslinger

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Quick Draw (the class ability; Quick Draw+, Quick Draw II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Wanted Poster (Bounty Hunter ability only) — gif: none — code: waiting
- [ ] Dead Man's Hand (Outlaw ability only) — gif: none — code: waiting
- [ ] Overwatch (Trickshot ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Double Tap — gif: none — code: waiting
- [ ] Hollow Point — gif: none — code: waiting
- [ ] Quick Shot — gif: none — code: waiting
- [ ] Ricochet — gif: none — code: waiting
- [ ] Flash Bang — gif: none — code: waiting
- [ ] Scatter Shot — gif: none — code: waiting
- [ ] Boot Shot — gif: none — code: waiting
- [ ] Disarming Shot — gif: none — code: waiting
- [ ] Roll Out — gif: none — code: waiting
- [ ] Bullet Hell — gif: none — code: waiting
- [ ] Powder Keg — gif: none — code: waiting
- [ ] Lucky Coin — gif: none — code: waiting
- [ ] High Noon — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Dead or Alive (Bounty Hunter only) — gif: none — code: waiting
- [ ] Six Shooter (Outlaw only) — gif: none — code: waiting
- [ ] Ricochet Storm (Trickshot only) — gif: none — code: waiting

### Mage

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Inner Element (the class ability; Inner Element+, Inner Element II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Absolute Freeze (Cryomancer ability only) — gif: none — code: waiting
- [ ] Firestorm Field (Pyromancer ability only) — gif: none — code: waiting
- [ ] Elemental Overload (Stormcaller ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Fireball — gif: none — code: waiting
- [ ] Frost Bolt — gif: none — code: waiting
- [ ] Meteor — gif: none — code: waiting
- [ ] Thunder — gif: none — code: waiting
- [ ] Ice Lance — gif: none — code: waiting
- [ ] Chain Lightning — gif: none — code: waiting
- [ ] Blizzard — gif: none — code: waiting
- [ ] Blink — gif: none — code: waiting
- [ ] Mana Burn — gif: none — code: waiting
- [ ] Arcane Nova — gif: none — code: waiting
- [ ] Static Field — gif: none — code: waiting
- [ ] Gravity Well — gif: none — code: waiting
- [ ] Echo — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Absolute Zero (Cryomancer only) — gif: none — code: waiting
- [ ] Inferno (Pyromancer only) — gif: none — code: waiting
- [ ] Thunderstorm (Stormcaller only) — gif: none — code: waiting

### Witch

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Poison Ivy (the class ability; Poison Ivy+, Poison Ivy II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Blood Pact (Blood Witch ability only) — gif: none — code: waiting
- [ ] Queen's Garden (Hex Queen ability only) — review GIF made, not signed off (`gifs/witch/queens-garden/review.gif`) — code: waiting
- [ ] Grave Bloom (Necromancer ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Hex — gif: none — code: waiting
- [ ] Toxic Bolt — gif: none — code: waiting
- [ ] Witch's Brew — gif: none — code: waiting
- [ ] Miasma — gif: none — code: waiting
- [ ] Curse of Frailty — gif: none — code: waiting
- [ ] Plague Swarm — gif: none — code: waiting
- [ ] Voodoo Link — gif: none — code: waiting
- [ ] Wither — gif: none — code: waiting
- [ ] Puppet Strings — gif: none — code: waiting
- [ ] Hex Storm — gif: none — code: waiting
- [ ] Soul Siphon — gif: none — code: waiting
- [ ] Contagion — gif: none — code: waiting
- [ ] Bog of Despair — gif: none — code: waiting

Ascension skills (one body only):

- [x] Blood Moon (Blood Witch only) — approved. Code is in `code/blood-moon.js`. Not in the game.
- [x] Grand Hex (Hex Queen only) — approved. Code is in `code/grand-hex.js`. Not in the game.
- [x] Soul Nova (Necromancer only) — approved. Code is in `code/soul-nova.js`. Not in the game.

### Healer

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Enlightenment (the class ability; Enlightenment+, Enlightenment II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Time Shift (Oracle ability only) — gif: none — code: waiting
- [ ] Plague Bloom (Plague Doctor ability only) — gif: none — code: waiting
- [ ] Divine Choir (Saint ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Divine Grace — gif: none — code: waiting
- [ ] Heal Ally — gif: none — code: waiting
- [ ] Mend Self — gif: none — code: waiting
- [ ] Radiant Burst — gif: none — code: waiting
- [ ] Blessing — gif: none — code: waiting
- [ ] Cleanse — gif: none — code: waiting
- [ ] Holy Siphon — gif: none — code: waiting
- [ ] Mass Heal — gif: none — code: waiting
- [ ] Prism Aura — gif: none — code: waiting
- [ ] Last Light — gif: none — code: waiting
- [ ] Sacred Flame — gif: none — code: waiting
- [ ] Resurrection — gif: none — code: waiting
- [ ] Renewal — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Time Warp (Oracle only) — gif: none — code: waiting
- [ ] Pandemic (Plague Doctor only) — gif: none — code: waiting
- [ ] Sanctuary (Saint only) — gif: none — code: waiting

### Tank

- [ ] Strike — gif: none — code: waiting
- [ ] Defend — gif: none — code: waiting
- [ ] Iron Maiden (the class ability; Iron Maiden+, Iron Maiden II use this same GIF unless we split them) — gif: none — code: waiting
- [ ] Frost Bulwark (Guardian ability only) — gif: none — code: waiting
- [ ] Unbroken Maiden (Juggernaut ability only) — gif: none — code: waiting
- [ ] War Banner (Warlord ability only) — gif: none — code: waiting

Shared skills (every form of this class):

- [ ] Bash — gif: none — code: waiting
- [ ] Crushing Blow — gif: none — code: waiting
- [ ] Shield Bash — gif: none — code: waiting
- [ ] Ground Pound — gif: none — code: waiting
- [ ] Ironhide — gif: none — code: waiting
- [ ] Hook Chain — gif: none — code: waiting
- [ ] Bodyguard — gif: none — code: waiting
- [ ] Fortress Slam — gif: none — code: waiting
- [ ] Iron Shell — gif: none — code: waiting
- [ ] Titan's Wrath — gif: none — code: waiting
- [ ] Taunting Roar — gif: none — code: waiting
- [ ] Aftershock — gif: none — code: waiting
- [ ] Unstoppable — gif: none — code: waiting

Ascension skills (one body only):

- [ ] Bulwark (Guardian only) — gif: none — code: waiting
- [ ] Earthshatter (Juggernaut only) — gif: none — code: waiting
- [ ] War Cry (Warlord only) — gif: none — code: waiting

## Also on every hero, not a class skill

- [ ] Cloak — gif: none — code: waiting

160 class skills are on this list, plus Strike, Defend and the abilities above. 3 already have a first GIF.
