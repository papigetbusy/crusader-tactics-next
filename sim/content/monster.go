package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Monster is one kind of normal monster (old MON and MNAMES tables): its
// name, the class it fights as, and its element ("" = rolled when it
// spawns). Elites and bosses are listed separately.
type Monster struct {
	Name    string `json:"name"`
	Class   string `json:"class"`
	Element string `json:"element,omitempty"`
}

// MonsterNamed finds a monster kind and its number in the list.
func (c *Content) MonsterNamed(name string) (Monster, int, bool) {
	for i, m := range c.Monsters {
		if m.Name == name {
			return m, i, true
		}
	}
	return Monster{}, -1, false
}

func oldMonsters(tables fs.FS) ([]Monster, error) {
	var mon [][]any
	var names []string
	for name, v := range map[string]any{"MON.json": &mon, "MNAMES.json": &names} {
		b, err := fs.ReadFile(tables, name)
		if err != nil {
			return nil, fmt.Errorf("old table %s is missing: %w", name, err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			return nil, fmt.Errorf("old table %s: %w", name, err)
		}
	}
	var out []Monster
	for i, m := range mon {
		k := Monster{Name: names[i]}
		k.Class, _ = m[0].(string)
		k.Element, _ = m[1].(string)
		out = append(out, k)
	}
	return out, nil
}
