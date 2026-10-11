package content

// Class is a hero class or a monster type: its base numbers and the skills a
// monster of this type uses.
type Class struct {
	Name        string       `json:"name"`
	Weapon      string       `json:"weapon,omitempty"` // sword, axe, lance... (weapon triangle)
	HP          int          `json:"hp"`
	Atk         int          `json:"atk"`
	Def         int          `json:"def"`
	Mov         int          `json:"mov"`
	Color       string       `json:"color,omitempty"`       // presentation only
	Skills      []Skill      `json:"skills"`                // what a monster of this type uses
	LevelSkills []LevelSkill `json:"levelSkills,omitempty"` // extra monster skills from a level on
}

// LevelSkill is a monster skill unlocked at a level (the old MSK table).
type LevelSkill struct {
	Level int   `json:"level"`
	Skill Skill `json:"skill"`
}

// Combat holds the numbers of the hit and damage rules (old index.html
// lines 445-453, 778, 1060-1061, 1687-1690).
type Combat struct {
	DamageRollLow    float64 `json:"damageRollLow"`    // damage roll runs from this...
	DamageRollSpread float64 `json:"damageRollSpread"` // ...to this much higher (0.9 to 1.1)
	MinDamage        int     `json:"minDamage"`

	HitMagic      int `json:"hitMagic"` // base hit chance in %
	HitRanged     int `json:"hitRanged"`
	HitMelee      int `json:"hitMelee"`
	HitHeroBonus  int `json:"hitHeroBonus"`
	HitBossBonus  int `json:"hitBossBonus"`
	HitEliteBonus int `json:"hitEliteBonus"`
	HitTriangle   int `json:"hitTriangle"` // +/- for winning/losing the weapon triangle
	HitMin        int `json:"hitMin"`
	HitMax        int `json:"hitMax"`

	Strong float64 `json:"strong"` // damage x when winning the weapon triangle or element wheel
	Weak   float64 `json:"weak"`   // damage x when losing it

	BossEnrageBelow float64 `json:"bossEnrageBelow"` // bosses under this share of HP...
	BossEnrage      float64 `json:"bossEnrage"`      // ...hit this much harder

	SkillLevelBonus float64 `json:"skillLevelBonus"` // per skill level above 1
	SkillLevelUses  []int   `json:"skillLevelUses"`  // uses needed for levels 1, 2, 3...

	ExecuteBelow   float64 `json:"executeBelow"`
	Execute        float64 `json:"execute"`
	Ambush         float64 `json:"ambush"`
	ChargePerTile  float64 `json:"chargePerTile"`
	ChargeMax      float64 `json:"chargeMax"`
	MagicDefense   float64 `json:"magicDefense"`   // magic only faces this share of defense
	DefendDamage   float64 `json:"defendDamage"`   // damage x while defending
	DrainDefault   float64 `json:"drainDefault"`   // what an old drain value of 1 means
	ElecShockTurns int     `json:"elecShockTurns"` // lightning always shocks for this long
	CritDamage     float64 `json:"critDamage"`     // a critical hit does this many times the damage
	PotionBase     int     `json:"potionBase"`     // a potion heals this much...
	PotionPerLevel int     `json:"potionPerLevel"` // ...plus this much per level
	ReviveShare    float64 `json:"reviveShare"`    // a revived hero comes back with this share of max HP
	ReviveAnyDeath bool    `json:"reviveAnyDeath"` // heroes killed by poison or bleeding can be revived too
	TrapsOnPath    bool    `json:"trapsOnPath"`    // traps spring anywhere along a monster's walk, not only where it stops

	// Statuses that change damage.
	RallyAttack   float64 `json:"rallyAttack"`   // rally: attacker hits this much harder
	WeakAttack    float64 `json:"weakAttack"`    // weak: attacker hits this much softer
	BrokenDefense float64 `json:"brokenDefense"` // brk: target keeps this share of defense
	GuardDamage   float64 `json:"guardDamage"`   // grd: target takes this share of damage
	PrismAtLeast  float64 `json:"prismAtLeast"`  // prism: hits count as at least this effective

	WeaponBeats     map[string]string `json:"weaponBeats"`     // sword beats lance...
	TriangleWeapons []string          `json:"triangleWeapons"` // weapons in the triangle
}
