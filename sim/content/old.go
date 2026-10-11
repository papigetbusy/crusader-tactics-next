package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
)

// FromOld builds Content from the old game's saved tables
// (testdata/oldgame/tables). The old-game match checks always use this, never
// the live data files, so editing a number in data/ can never break them.
func FromOld(tables fs.FS) (*Content, error) {
	var tinfo map[string]struct {
		N    string  `json:"n"`
		C    int     `json:"c"`
		R    float64 `json:"r"`
		Heal float64 `json:"heal"`
	}
	b, err := fs.ReadFile(tables, "TINFO.json")
	if err != nil {
		return nil, fmt.Errorf("old table TINFO.json is missing: %w", err)
	}
	if err := json.Unmarshal(b, &tinfo); err != nil {
		return nil, err
	}
	c := &Content{
		Terrain: map[byte]Terrain{},
		// These were constants in the old code, not tables (index.html 1217-1219).
		Blocked:           set[byte]('w', 'm'),
		Flying:            set("Dragon Knight", "Demon Lord", "Bone Gargoyle", "Bat Lord", "Golden Drake"),
		FlyingCost:        2,
		Mounted:           set("Knight"),
		MountedForestCost: 6,
	}
	for l, t := range tinfo {
		c.Terrain[l[0]] = Terrain{Letter: l, Name: t.N, MoveCost: t.C, DamageTaken: t.R, Heal: t.Heal}
	}
	if err := oldSkillsAndClasses(tables, c); err != nil {
		return nil, err
	}
	if c.Progress, err = oldProgress(tables); err != nil {
		return nil, err
	}
	return c, c.Check()
}

// oldSkillsAndClasses reads SK, CL, MSK and EB, and sets the combat numbers
// that were constants in the old code.
func oldSkillsAndClasses(tables fs.FS, c *Content) error {
	var sk map[string]map[string]any
	var cl map[string]struct {
		W   string           `json:"w"`
		HP  int              `json:"hp"`
		Atk int              `json:"atk"`
		Def int              `json:"def"`
		Mov int              `json:"mov"`
		Col string           `json:"col"`
		Sk  []map[string]any `json:"sk"`
	}
	var msk map[string][][2]json.RawMessage
	for name, v := range map[string]any{"SK.json": &sk, "CL.json": &cl, "MSK.json": &msk, "EB.json": &c.Beats} {
		b, err := fs.ReadFile(tables, name)
		if err != nil {
			return fmt.Errorf("old table %s is missing: %w", name, err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			return fmt.Errorf("old table %s: %w", name, err)
		}
	}
	c.Skills = map[string]Skill{}
	b, _ := fs.ReadFile(tables, "SK.json")
	order, err := objectKeys(b)
	if err != nil {
		return err
	}
	for _, name := range order {
		s := oldSkill(sk[name])
		c.Skills[s.Name] = s
		c.SkillOrder = append(c.SkillOrder, s.Name)
	}
	c.Classes = map[string]Class{}
	for name, o := range cl {
		k := Class{Name: name, Weapon: o.W, HP: o.HP, Atk: o.Atk, Def: o.Def, Mov: o.Mov, Color: o.Col}
		for _, m := range o.Sk {
			k.Skills = append(k.Skills, oldSkill(m))
		}
		for _, pair := range msk[name] {
			var lv int
			var m map[string]any
			if err := json.Unmarshal(pair[0], &lv); err != nil {
				return err
			}
			if err := json.Unmarshal(pair[1], &m); err != nil {
				return err
			}
			k.LevelSkills = append(k.LevelSkills, LevelSkill{Level: lv, Skill: oldSkill(m)})
		}
		c.Classes[name] = k
	}
	c.Combat = oldCombat
	c.Turns = oldTurns
	c.AI = oldAI
	return nil
}

// oldCombat are the hit and damage numbers written into the old code.
var oldCombat = Combat{
	DamageRollLow: 0.9, DamageRollSpread: 0.2, MinDamage: 1, // hurt, line 1690; dcalc, line 450
	HitMagic: 95, HitRanged: 85, HitMelee: 90, // hitCh, line 1060
	HitHeroBonus: 5, HitBossBonus: 10, HitEliteBonus: 5, HitTriangle: 15, HitMin: 5, HitMax: 100,
	Strong: 1.5, Weak: 0.75, // mult, line 778
	BossEnrageBelow: 0.5, BossEnrage: 1.3, // dcalc, line 446
	SkillLevelBonus: 0.05, SkillLevelUses: []int{0, 8, 16, 24, 32}, // SKB, SKT, line 434
	ExecuteBelow: 0.4, Execute: 1.5, Ambush: 1.5, ChargePerTile: 0.12, ChargeMax: 0.6, // dcalc, line 448
	MagicDefense: 0.5, DefendDamage: 0.5, DrainDefault: 0.6, ElecShockTurns: 1, CritDamage: 1.5,
	PotionBase: 10, PotionPerLevel: 3, ReviveShare: 0.4, // pot, line 1806; act, line 1715
	RallyAttack: 1.25, WeakAttack: 0.7, BrokenDefense: 0.7, GuardDamage: 0.6, PrismAtLeast: 1.5,
	WeaponBeats:     map[string]string{"axe": "sword", "sword": "lance", "lance": "axe"}, // WB, line 316
	TriangleWeapons: []string{"sword", "axe", "lance"},                                   // TRI, line 440
}

// objectKeys lists a JSON object's keys in the order they are written.
func objectKeys(b []byte) ([]string, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	var keys []string
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}
		keys = append(keys, t.(string))
		var skip json.RawMessage
		if err := d.Decode(&skip); err != nil {
			return nil, err
		}
	}
	return keys, nil
}
