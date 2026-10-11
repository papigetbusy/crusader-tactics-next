package content

// UserElement marks a skill whose element is the user's own element
// (the old game wrote el:1).
const UserElement = "user"

// Skill is one attack, heal or buff. Field names are plain words; the old
// game's short names are noted after each.
type Skill struct {
	Name          string         `json:"name"`                    // n
	Class         string         `json:"class,omitempty"`         // cls
	Kind          string         `json:"kind,omitempty"`          // t: "n" normal, "u" ultimate
	Level         int            `json:"level,omitempty"`         // lv: level it is learned at
	Ascension     string         `json:"ascension,omitempty"`     // spec: ultimate of this ascension
	Power         float64        `json:"power"`                   // pw
	MinRange      int            `json:"minRange"`                // min
	MaxRange      int            `json:"maxRange"`                // max
	Cooldown      int            `json:"cooldown,omitempty"`      // cd
	Element       string         `json:"element,omitempty"`       // el ("user" = user's element)
	Magic         bool           `json:"magic,omitempty"`         // mag
	Physical      bool           `json:"physical,omitempty"`      // phys: an elemental weapon art, not magic
	Effects       map[string]int `json:"effects,omitempty"`       // fx: status -> turns, put on the target
	Area          int            `json:"area,omitempty"`          // aoe: also hits enemies within N of the target
	ExtraTargets  int            `json:"extraTargets,omitempty"`  // multi: hits up to N targets in range
	Hits          int            `json:"hits,omitempty"`          // hits: strikes N times
	Line          int            `json:"line,omitempty"`          // line: pierces N tiles behind the target
	Execute       bool           `json:"execute,omitempty"`       // exec: more damage to targets under 40% HP
	IgnoreDefense bool           `json:"ignoreDefense,omitempty"` // idef
	Ambush        bool           `json:"ambush,omitempty"`        // amb: more damage to targets at full HP
	Charge        bool           `json:"charge,omitempty"`        // chg: more damage per tile moved this turn
	Heal          bool           `json:"heal,omitempty"`          // heal
	Support       bool           `json:"support,omitempty"`       // sup: a buff for allies
	Buffs         map[string]int `json:"buffs,omitempty"`         // buff: status -> turns, put on allies
	Revive        bool           `json:"revive,omitempty"`        // rev
	Cleanse       bool           `json:"cleanse,omitempty"`       // cleanse
	AllAllies     bool           `json:"allAllies,omitempty"`     // all
	ResetCooldown bool           `json:"resetCooldowns,omitempty"`
	Drain         float64        `json:"drain,omitempty"`  // heals the user; the old value 1 meant 60%
	Siphon        bool           `json:"siphon,omitempty"` // heals the most hurt ally by the damage
	LockSelf      bool           `json:"lockSelf,omitempty"`
	HPCost        float64        `json:"hpCost,omitempty"` // hpc: share of max HP paid
	Cloak         bool           `json:"cloak,omitempty"`
	Strike        bool           `json:"strike,omitempty"` // the basic weapon attack
	Splash        float64        `json:"splashHeal,omitempty"`
	Summon        *int           `json:"summon,omitempty"`
	Visual        string         `json:"visual,omitempty"` // vfx: presentation only
	Text          string         `json:"text,omitempty"`   // d

	// What the skills of the even skill schedule (approved 2026-10-10) do.
	Aim        string  `json:"aim,omitempty"`      // how it is aimed (the Aim... names below); "" = a unit, as before
	Push       int     `json:"push,omitempty"`     // shoves the target this many tiles away
	Slam       bool    `json:"slam,omitempty"`     // +50% and a 1-turn stun when the shove is blocked
	Pull       bool    `json:"pull,omitempty"`     // drags the target next to the user
	Gather     int     `json:"gather,omitempty"`   // pulls enemies within N of the target one step toward it
	Leap       int     `json:"leap,omitempty"`     // first jumps up to N tiles to land next to the target
	Behind     bool    `json:"behind,omitempty"`   // first teleports to the tile behind the target
	Away       int     `json:"away,omitempty"`     // afterwards leaps up to N tiles away from the target
	Dash       int     `json:"dash,omitempty"`     // AimRoll: how far the roll reaches
	Around     bool    `json:"around,omitempty"`   // Area counts around the user instead of the target
	NoDamage   bool    `json:"noDamage,omitempty"` // only statuses and movement; never misses
	Dispel     bool    `json:"dispel,omitempty"`   // strips the target's buffs and shield
	Shield     float64 `json:"shield,omitempty"`   // gives a shield worth this share of the user's max HP
	Free       bool    `json:"free,omitempty"`     // does not use up the turn
	Still      bool    `json:"still,omitempty"`    // only before moving, and no moving afterwards
	Field      string  `json:"field,omitempty"`    // leaves ground of this kind: "fire", "static", "swamp"
	FieldRange int     `json:"fieldRange,omitempty"`
	FieldTurns int     `json:"fieldTurns,omitempty"`
	Trick      string  `json:"trick,omitempty"` // a one-off effect, named in the combat package
}

// How skills are aimed.
const (
	AimUnit  = ""            // a unit, as the old game's skills
	AimAlly  = "ally"        // another unit on your side
	AimAny   = "any"         // any other unit
	AimSelf  = "self"        // nobody: it works on or around the user
	AimTile  = "tile"        // one tile in range
	AimLine  = "line"        // a tile in a straight line, up to MaxRange away
	AimRoll  = "rollThenFoe" // a tile to roll to (Dash), then an enemy in range from there
	AimShove = "foeThenTile" // an enemy, then a tile within 2 of it
	AimTwo   = "twoFoes"     // two different enemies in range
	AimEcho  = "echo"        // whoever the user's last spell could hit
)

// Friendly reports skills that deal no damage to an aimed enemy: heals,
// buffs, revives, summons, and the new ally, self, tile and line skills
// without power.
func (s Skill) Friendly() bool {
	if s.Heal || s.Support || s.Revive || s.Summon != nil {
		return true
	}
	switch s.Aim {
	case AimAlly, AimAny, AimSelf, AimTile, AimLine:
		return s.Power == 0
	}
	return false
}

// IsMagic is the old isMag rule (index.html line 1096) for units with no
// weapon sets equipped: magic flag, or an element that is not a weapon art.
func (s Skill) IsMagic() bool { return s.Magic || (s.Element != "" && !s.Physical) }

// NoRoll reports skills that never miss (heals, buffs, revives, summons).
func (s Skill) NoRoll() bool { return s.Heal || s.Support || s.Revive || s.Summon != nil }

func oldSkill(m map[string]any) Skill {
	num := func(k string) float64 {
		if v, ok := m[k].(float64); ok {
			return v
		}
		return 0
	}
	str := func(k string) string { v, _ := m[k].(string); return v }
	flag := func(k string) bool {
		switch v := m[k].(type) {
		case bool:
			return v
		case float64:
			return v != 0
		}
		return false
	}
	counts := func(k string) map[string]int {
		raw, _ := m[k].(map[string]any)
		if len(raw) == 0 {
			return nil
		}
		out := map[string]int{}
		for q, v := range raw {
			out[q] = int(v.(float64))
		}
		return out
	}
	s := Skill{
		Name: str("n"), Class: str("cls"), Kind: str("t"), Level: int(num("lv")), Ascension: str("spec"),
		Power: num("pw"), MinRange: int(num("min")), MaxRange: int(num("max")), Cooldown: int(num("cd")),
		Magic: flag("mag"), Physical: flag("phys"), Effects: counts("fx"),
		Area: int(num("aoe")), ExtraTargets: int(num("multi")), Hits: int(num("hits")), Line: int(num("line")),
		Execute: flag("exec"), IgnoreDefense: flag("idef"), Ambush: flag("amb"), Charge: flag("chg"),
		Heal: flag("heal"), Support: flag("sup"), Buffs: counts("buff"), Revive: flag("rev"),
		Cleanse: flag("cleanse"), AllAllies: flag("all"), ResetCooldown: flag("reset"),
		Drain: num("drain"), Siphon: flag("siphon"), LockSelf: flag("lock"), HPCost: num("hpc"),
		Cloak: flag("cloak"), Strike: flag("strike"), Splash: num("splash"), Visual: str("vfx"), Text: str("d"),
	}
	switch v := m["el"].(type) {
	case string:
		s.Element = v
	case float64:
		if v == 1 {
			s.Element = UserElement
		}
	}
	if v, ok := m["summon"].(float64); ok {
		n := int(v)
		s.Summon = &n
	}
	return s
}
