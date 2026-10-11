package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Progress holds the numbers of experience, levels and evolutions (old gain,
// xpShare, calc and onLevel, index.html lines 439, 736-737 and 779-780, and
// the XP lines of hurt and act, 1689, 1707, 1717 and 1720).
type Progress struct {
	XPRate         float64 `json:"xpRate"`         // every XP gain is multiplied by this (rounded, at least 1)
	XPPerLevel     int     `json:"xpPerLevel"`     // XP to the next level = level x this
	KillShare      float64 `json:"killShare"`      // share of kill XP split among the other heroes on the board
	KillBase       int     `json:"killBase"`       // kill XP = (this + enemy level) x boss/elite multiplier
	KillBoss       int     `json:"killBoss"`       // kill XP multiplier for a boss
	KillElite      int     `json:"killElite"`      // kill XP multiplier for an elite
	MissXP         int     `json:"missXP"`         // XP for an attack that misses
	HitBase        int     `json:"hitBase"`        // hit XP = this + damage / hitPerDamage (+ hitStatusBonus)
	HitPerDamage   float64 `json:"hitPerDamage"`   //
	HitStatusBonus int     `json:"hitStatusBonus"` // extra hit XP when the skill puts statuses on
	SupportXP      int     `json:"supportXP"`      // XP for a buff skill
	HealBase       int     `json:"healBase"`       // heal XP = this + min(healMax, healed / healPer)
	HealPer        float64 `json:"healPer"`        //
	HealMax        int     `json:"healMax"`        //
	CapAboveFloor  int     `json:"capAboveFloor"`  // level cap = monster level of the best floor + this

	GrowthHP   float64             `json:"growthHP"`   // max HP per level after the first
	GrowthAtk  float64             `json:"growthAtk"`  // attack per level after the first
	GrowthDef  float64             `json:"growthDef"`  // defense per level after the first
	AtkBonus   float64             `json:"atkBonus"`   // flat attack every hero gets
	FormLevels []int               `json:"formLevels"` // level each form starts at (base, first and second evolution)
	Forms      []Form              `json:"forms"`      // stat changes of each form; the last one is the ascension
	AscendAt   int                 `json:"ascendAt"`   // level of the ascension choice
	FormNames  map[string][]string `json:"formNames"`  // class -> name of each form ("Warrior", "Vanguard", "Champion")
}

// Form is how much stronger a hero gets in one form (old TIER table).
type Form struct {
	HP  float64 `json:"hp"`  // max HP multiplier
	Atk float64 `json:"atk"` // attack multiplier
	Def int     `json:"def"` // defense added
	Mov int     `json:"mov"` // move added
}

func (p Progress) check() error {
	if p.XPPerLevel <= 0 || p.XPRate <= 0 {
		return fmt.Errorf("rules/progress.json: xpPerLevel and xpRate must be above 0")
	}
	if len(p.Forms) != len(p.FormLevels)+1 {
		return fmt.Errorf("rules/progress.json: there must be one form per form level, plus the ascension form")
	}
	if p.HitPerDamage <= 0 || p.HealPer <= 0 {
		return fmt.Errorf("rules/progress.json: hitPerDamage and healPer must be above 0")
	}
	return nil
}

// oldProgress reads TUNE, TIER and FORMS and adds the numbers that were
// written into the old code.
func oldProgress(tables fs.FS) (Progress, error) {
	var tune map[string]float64
	var tier [][4]float64
	p := Progress{
		XPPerLevel: 6, KillBase: 3, KillBoss: 4, KillElite: 2, // gain, line 780; hurt, line 1707
		MissXP: 1, HitBase: 2, HitPerDamage: 12, HitStatusBonus: 2, // hurt, lines 1689 and 1707
		SupportXP: 4, HealBase: 2, HealPer: 8, HealMax: 8, // act, lines 1717 and 1720
		GrowthHP: 3, GrowthAtk: 1, GrowthDef: 0.25, AtkBonus: 1, // calc, line 736
		FormLevels: []int{1, 5, 20}, AscendAt: 30, // onLevel and initP, lines 437-439
	}
	for name, v := range map[string]any{"TUNE.json": &tune, "TIER.json": &tier, "FORMS.json": &p.FormNames} {
		b, err := fs.ReadFile(tables, name)
		if err != nil {
			return p, fmt.Errorf("old table %s is missing: %w", name, err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			return p, fmt.Errorf("old table %s: %w", name, err)
		}
	}
	p.XPRate, p.KillShare, p.CapAboveFloor = tune["xp"], tune["xs"], int(tune["capGap"])
	for _, t := range tier {
		p.Forms = append(p.Forms, Form{HP: t[0], Atk: t[1], Def: int(t[2]), Mov: int(t[3])})
	}
	return p, nil
}
