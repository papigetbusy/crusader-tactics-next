package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Ascension is one of the three Lv 30 paths of a class (old SPEC table):
// its element, its stat changes, its ultimate, and the numbers of its passive.
type Ascension struct {
	Name     string             `json:"name"`
	Class    string             `json:"class"`
	Element  string             `json:"element"`           // the hero's element after ascending
	HP       float64            `json:"hp,omitempty"`      // max HP multiplier (0 = unchanged)
	Atk      float64            `json:"atk,omitempty"`     // attack multiplier
	Def      float64            `json:"def,omitempty"`     // defense multiplier
	DefAdd   int                `json:"defAdd,omitempty"`  // defense added
	Mov      int                `json:"mov,omitempty"`     // move added
	Range    int                `json:"range,omitempty"`   // attack range added
	Heal     float64            `json:"heal,omitempty"`    // healing multiplier
	Ultimate string             `json:"ultimate"`          // the ultimate it teaches
	Bonus    string             `json:"bonus,omitempty"`   // the stat line, in words
	Passive  string             `json:"passive,omitempty"` // what the passive does, in words
	Numbers  map[string]float64 `json:"numbers,omitempty"` // the passive's numbers (see each rule)
}

// N is one of the passive's numbers.
func (a Ascension) N(key string) float64 { return a.Numbers[key] }

// oldAscensions reads SPEC and SPECS. The passives' numbers were written
// into the old code (dcalc, hurt, act and hitCh); they are listed here with
// the line they came from.
func oldAscensions(tables fs.FS) (map[string]Ascension, map[string][]string, error) {
	var spec map[string]struct {
		C   string             `json:"c"`
		El  string             `json:"el"`
		Buf map[string]float64 `json:"buf"`
		Bt  string             `json:"bt"`
		P   string             `json:"p"`
		Ult string             `json:"ult"`
	}
	var specs map[string][]string
	for name, v := range map[string]any{"SPEC.json": &spec, "SPECS.json": &specs} {
		b, err := fs.ReadFile(tables, name)
		if err != nil {
			return nil, nil, fmt.Errorf("old table %s is missing: %w", name, err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			return nil, nil, fmt.Errorf("old table %s: %w", name, err)
		}
	}
	out := map[string]Ascension{}
	for name, s := range spec {
		a := Ascension{Name: name, Class: s.C, Element: s.El, Ultimate: s.Ult, Bonus: s.Bt, Passive: s.P,
			HP: s.Buf["hp"], Atk: s.Buf["atk"], Def: s.Buf["def"], DefAdd: int(s.Buf["defA"]), Mov: int(s.Buf["mov"]),
			Range: int(s.Buf["rng"]), Heal: s.Buf["heal"], Numbers: oldPassiveNumbers[name]}
		if len(a.Numbers) == 0 {
			a.Numbers = nil // passives written fully in words have no numbers
		}
		out[name] = a
	}
	return out, specs, nil
}

// oldPassiveNumbers are the numbers of each ascension's passive, as written
// in the old code (index.html line in brackets).
var oldPassiveNumbers = map[string]map[string]float64{
	"Berserker":      {"lowHP": 0.5, "lowHPDamage": 1.4, "killHeal": 0.15},      // dcalc 447, hurt 1707
	"Paladin":        {"aegis": 0.8, "lifeSteal": 0.2},                          // dcalc 449, hurt 1699
	"Duelist":        {"riposte": 0.6},                                          // hurt 1709
	"Juggernaut":     {},                                                        // hurt 1703: immune to stun, snare, break, shock, lock, ivy
	"Guardian":       {"bodyguard": 0.7},                                        // hurt 1688
	"Warlord":        {"aura": 1.15, "auraRange": 2},                            // dcalc 447
	"Pyromancer":     {"fireDamage": 1.4, "fireBleed": 2},                       // dcalc 447, hurt 1702
	"Cryomancer":     {"snaredDamage": 1.25, "iceSnare": 1},                     // dcalc 447, hurt 1702
	"Stormcaller":    {"arc": 0.5, "arcRange": 2},                               // act 1726
	"Saint":          {"shieldCap": 0.5, "choirShieldCap": 1, "choirShield": 2}, // act 1719
	"Oracle":         {"haste": 2},                                              // act 1719
	"Plague Doctor":  {"bleed": 3, "healShare": 0.5},                            // hurt 1700, 1702
	"Sniper":         {"stillDamage": 1.35, "defense": 0.7},                     // dcalc 447
	"Trapper":        {"markedDamage": 1.3, "snare": 2},                         // dcalc 447-448, hurt 1702
	"Tempest Archer": {},                                                        // act 1731: hit and run
	"Phantom":        {"evasion": 30},                                           // hitCh 1060
	"Viper":          {"bleed": 3, "bleedingDamage": 1.3},                       // dcalc 447, hurt 1702
	"Blade Dancer":   {},                                                        // act 1733: a kill lets you act again
	"Outlaw":         {"shots": 2, "shotPower": 0.6},                            // act 1722
	"Bounty Hunter":  {"bigDamage": 1.5, "gold": 2},                             // dcalc 447, hurt 1707
	"Trickshot":      {"arc": 0.6, "arcRange": 3},                               // act 1726
	"Necromancer":    {},                                                        // hurt 1706: souls
	"Blood Witch":    {"damage": 1.4, "lifeSteal": 0.2, "hpCost": 0.05},         // dcalc 447, hurt 1699, act 1721
	"Inferno Rider":  {"fireDamage": 1.4, "fireBleed": 2},                       // dcalc 447, hurt 1702
	"Frost Wyrm":     {"snare": 1},                                              // hurt 1702-1703
	"Storm Dragoon":  {},                                                        // act 1730: Sky Dive has no cooldown
	"Hex Queen":      {"cursedDamage": 1.25},                                    // dcalc 448, hurt 1703
	"Templar":        {"chargeTiles": 3, "chargeHeal": 0.15},                    // act 1725
	"Black Knight":   {"singleDamage": 1.4},                                     // dcalc 448, hurt 1705
	"Storm Rider":    {"chargeMax": 1.2, "chargePerTile": 0.24},                 // dcalc 448
}
