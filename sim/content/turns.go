package content

// Turns holds the numbers of turn start and turn end (old startPhase and
// endPhase, index.html lines 1738-1739).
type Turns struct {
	DamageOverTime    map[string]float64 `json:"damageOverTime"`    // status -> share of max HP lost at turn start
	DamageOverTimeMin int                `json:"damageOverTimeMin"` // never less than this
	RegenShare        float64            `json:"regenShare"`        // regen status heals this share at turn start
	Stunned           []string           `json:"stunned"`           // statuses that skip the turn
	CountdownAtEnd    []string           `json:"countdownAtEnd"`    // statuses that count down at turn end
}

// AI holds the enemy decision numbers (old aiTurn, index.html line 1740).
type AI struct {
	HealWhenAllyBelow  float64 `json:"healWhenAllyBelow"`  // healers heal when an ally is under this share of HP
	SpecialSkillChance float64 `json:"specialSkillChance"` // chance to use a special (non-basic) skill
	TooClosePenalty    float64 `json:"tooClosePenalty"`    // tile score penalty for standing inside minimum range
	SpreadWeight       float64 `json:"spreadWeight"`       // ranged units prefer distance, melee closeness

	// Smarter enemies as you climb (owner, 2026-10-11; Phase 9). 0 = never.
	SmartFrom       int     `json:"smartFrom,omitempty"`       // from this floor monsters weigh who to hit
	SmartBossFrom   int     `json:"smartBossFrom,omitempty"`   // the same for bosses and elites
	CautiousFrom    int     `json:"cautiousFrom,omitempty"`    // from this floor they avoid ending next to several heroes when they can't attack
	KillWeight      float64 `json:"killWeight,omitempty"`      // pull of a hero it can defeat this turn
	HealerWeight    float64 `json:"healerWeight,omitempty"`    // pull of a healer
	WeakWeight      float64 `json:"weakWeight,omitempty"`      // pull of a hero weak to its attack
	TileWeight      float64 `json:"tileWeight,omitempty"`      // how much that pull counts when choosing where to stand
	CautiousPenalty float64 `json:"cautiousPenalty,omitempty"` // tile score penalty for ending next to 2+ heroes without attacking
}

var oldTurns = Turns{
	DamageOverTime:    map[string]float64{"burn": 0.06, "bleed": 0.06, "poison": 0.07},
	DamageOverTimeMin: 3, RegenShare: 0.12,
	Stunned: []string{"stun", "frozen"},
	CountdownAtEnd: []string{"snare", "stun", "lock", "brk", "grd", "rally", "haste", "prism", "regen", "weak",
		"shock", "dbl", "imh", "ironm", "inner", "enl", "heye", "blood", "cloak", "qd", "beast", "ivyA", "vow",
		"ivy", "flurry", "frenzy", "taunt", "unb", "bulwark", "fstorm", "frozen", "choir", "pbloom", "deadeye",
		"dmark", "scales", "sky", "gbloom", "vamp", "svow", "ruin", "clone", "tmark"},
}

var oldAI = AI{HealWhenAllyBelow: 0.75, SpecialSkillChance: 0.45, TooClosePenalty: 30, SpreadWeight: 0.01}
