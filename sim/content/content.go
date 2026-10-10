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
	c := &Content{
		Terrain:           map[byte]Terrain{},
		Blocked:           set(letters(mf.BlockedLetters)...),
		Flying:            set(mf.FlyingClasses...),
		FlyingCost:        mf.FlyingCost,
		Mounted:           set(mf.MountedClasses...),
		MountedForestCost: mf.MountedForestCost,
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
	return nil
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
