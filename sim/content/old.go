package content

import (
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
	return c, c.Check()
}
