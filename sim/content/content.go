// Package content reads the data files and hands the rules typed, checked values.
// No rule reads a file directly; they all receive a *Content.
package content

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/papigetbusy/crusader-tactics-next/data"
)

// Terrain is one kind of ground.
type Terrain struct {
	Letter      string  `json:"letter"`
	Name        string  `json:"name"`
	MoveCost    int     `json:"moveCost"`    // in half-steps; Plains = 2
	DamageTaken float64 `json:"damageTaken"` // 0.85 = 15% less damage taken here
	Heal        float64 `json:"heal"`        // share of max HP healed per turn
}

// Content is every game number the rules use.
type Content struct {
	Terrain           map[byte]Terrain
	Blocked           map[byte]bool   // ground nobody walks on (fliers cross it)
	Flying            map[string]bool // classes that fly
	FlyingCost        int             // what every tile costs a flier
	Mounted           map[string]bool // classes slowed by forest
	MountedForestCost int

	Skills     map[string]Skill  // hero skills by name
	SkillOrder []string          // hero skills in the old game's table order (it decides starting skills)
	Classes    map[string]Class  // hero classes and monster types by name
	Beats      map[string]string // element wheel: fire beats ice...
	Combat     Combat
	Turns      Turns
	AI         AI
	Progress   Progress

	Ascensions       map[string]Ascension // by name
	AscensionChoices map[string][]string  // class -> its three ascensions, in the order they are offered

	Abilities map[string]ClassAbility // class abilities by class
	CrashDown Ability                 // the Storm Dragoon's landing from Sky Lord
	Monsters  []Monster               // normal monster kinds, in the old game's order
}

// Default loads the live data files built into the program.
func Default() (*Content, error) { return Load(data.Files) }

// MustDefault is Default for places where missing data is a programming error.
func MustDefault() *Content {
	c, err := Default()
	if err != nil {
		panic(err)
	}
	return c
}

type terrainFile struct {
	Terrain []Terrain `json:"terrain"`
}

type movementFile struct {
	BlockedLetters    []string `json:"blockedLetters"`
	FlyingClasses     []string `json:"flyingClasses"`
	FlyingCost        int      `json:"flyingCost"`
	MountedClasses    []string `json:"mountedClasses"`
	MountedForestCost int      `json:"mountedForestCost"`
}

// Load reads the data files from fsys and checks them.
func Load(fsys fs.FS) (*Content, error) {
	var tf terrainFile
	var mf movementFile
	if err := readJSON(fsys, "terrain.json", &tf); err != nil {
		return nil, err
	}
	if err := readJSON(fsys, "rules/movement.json", &mf); err != nil {
		return nil, err
	}
	var skills struct {
		Skills []Skill `json:"skills"`
	}
	var classes struct {
		Classes []Class `json:"classes"`
	}
	var elements struct {
		Beats map[string]string `json:"beats"`
	}
	var combat Combat
	var turns Turns
	var ai AI
	var progress Progress
	var abil struct {
		Classes   map[string]ClassAbility `json:"classes"`
		CrashDown Ability                 `json:"crashDown"`
	}
	var mons struct {
		Monsters []Monster `json:"monsters"`
	}
	var asc struct {
		Ascensions []Ascension         `json:"ascensions"`
		Choices    map[string][]string `json:"choices"`
	}
	for name, v := range map[string]any{"skills.json": &skills, "classes.json": &classes,
		"elements.json": &elements, "rules/combat.json": &combat, "rules/turns.json": &turns, "rules/ai.json": &ai,
		"rules/progress.json": &progress, "ascensions.json": &asc, "abilities.json": &abil, "monsters.json": &mons} {
		if err := readJSON(fsys, name, v); err != nil {
			return nil, err
		}
	}
	c := &Content{
		Terrain:           map[byte]Terrain{},
		Blocked:           set(letters(mf.BlockedLetters)...),
		Flying:            set(mf.FlyingClasses...),
		FlyingCost:        mf.FlyingCost,
		Mounted:           set(mf.MountedClasses...),
		MountedForestCost: mf.MountedForestCost,
		Skills:            map[string]Skill{},
		Classes:           map[string]Class{},
		Beats:             elements.Beats,
		Combat:            combat,
		Turns:             turns,
		AI:                ai,
		Progress:          progress,
		Ascensions:        map[string]Ascension{},
		AscensionChoices:  asc.Choices,
		Abilities:         abil.Classes,
		CrashDown:         abil.CrashDown,
		Monsters:          mons.Monsters,
	}
	for _, a := range asc.Ascensions {
		c.Ascensions[a.Name] = a
	}
	for _, sk := range skills.Skills {
		c.Skills[sk.Name] = sk
		c.SkillOrder = append(c.SkillOrder, sk.Name)
	}
	for _, cl := range classes.Classes {
		c.Classes[cl.Name] = cl
	}
	for _, t := range tf.Terrain {
		c.Terrain[t.Letter[0]] = t
	}
	return c, c.Check()
}

// Check catches mistakes in the data files with a plain message.
func (c *Content) Check() error {
	if _, ok := c.Terrain['g']; !ok {
		return fmt.Errorf("terrain.json must define Plains (letter g): it is the fallback ground")
	}
	for l, t := range c.Terrain {
		if len(t.Letter) != 1 {
			return fmt.Errorf("terrain %q: letter must be a single character", t.Name)
		}
		if t.MoveCost <= 0 {
			return fmt.Errorf("terrain %q (%c): moveCost must be above 0", t.Name, l)
		}
		if t.DamageTaken <= 0 {
			return fmt.Errorf("terrain %q (%c): damageTaken must be above 0", t.Name, l)
		}
	}
	if c.FlyingCost <= 0 || c.MountedForestCost <= 0 {
		return fmt.Errorf("rules/movement.json: flyingCost and mountedForestCost must be above 0")
	}
	for name, sk := range c.Skills {
		if err := sk.check(); err != nil {
			return fmt.Errorf("skills.json, skill %q: %w", name, err)
		}
	}
	for name, cl := range c.Classes {
		for _, sk := range cl.Skills {
			if err := sk.check(); err != nil {
				return fmt.Errorf("classes.json, %s skill %q: %w", name, sk.Name, err)
			}
		}
	}
	if c.Combat.HitMin < 0 || c.Combat.HitMax > 100 || c.Combat.HitMin > c.Combat.HitMax {
		return fmt.Errorf("rules/combat.json: hit chance limits must be within 0-100")
	}
	for name, a := range c.Ascensions {
		if _, ok := c.Skills[a.Ultimate]; !ok {
			return fmt.Errorf("ascensions.json, %s: its ultimate %q is not in skills.json", name, a.Ultimate)
		}
	}
	for class, ca := range c.Abilities {
		if len(ca.Tiers) != len(abilityTierLevels) {
			return fmt.Errorf("abilities.json, %s: needs one ability per tier (%d)", class, len(abilityTierLevels))
		}
	}
	if err := c.Progress.check(); err != nil {
		return err
	}
	if len(c.Combat.SkillLevelUses) == 0 {
		return fmt.Errorf("rules/combat.json: skillLevelUses needs at least one entry")
	}
	return nil
}

func (s Skill) check() error {
	if s.Name == "" {
		return fmt.Errorf("has no name")
	}
	if s.Power < 0 || s.MinRange < 0 || s.MaxRange < s.MinRange || s.Cooldown < 0 {
		return fmt.Errorf("power, range or cooldown is impossible (range %d-%d)", s.MinRange, s.MaxRange)
	}
	return nil
}

// SkillFor finds the skill a unit uses by name. Heroes use the hero skill
// list. Monsters use their type's own skills first (a monster's Fireball can
// differ from a hero's), then skills unlocked by level. classBase reports a
// skill from the type's own list: the old game treats those specially for
// elements (index.html line 1704).
func (c *Content) SkillFor(team int, class, name string) (s Skill, classBase bool, ok bool) {
	if team != 0 {
		if cl, found := c.Classes[class]; found {
			for _, sk := range cl.Skills {
				if sk.Name == name {
					return sk, true, true
				}
			}
			for _, ls := range cl.LevelSkills {
				if ls.Skill.Name == name {
					return ls.Skill, false, true
				}
			}
			for _, ls := range cl.Schedule {
				if ls.Skill.Name == name {
					return ls.Skill, false, true
				}
			}
		}
	}
	s, ok = c.Skills[name]
	return s, false, ok
}

func readJSON(fsys fs.FS, name string, v any) error {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("data file %s is missing: %w", name, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("data file %s is not valid: %w", name, err)
	}
	return nil
}

func set[T comparable](xs ...T) map[T]bool {
	m := make(map[T]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func letters(xs []string) []byte {
	out := make([]byte, 0, len(xs))
	for _, x := range xs {
		if x != "" {
			out = append(out, x[0])
		}
	}
	return out
}
