// Package save wraps a game in a versioned save: a stamp that says what it
// is and which save format it uses, so later versions of the game can read
// older saves (and refuse what they can't read, in plain words).
package save

import (
	"encoding/json"
	"fmt"
)

// Game names the new game's saves (the old game's saves are not read).
const Game = "crusader-tactics-next"

// Version is the save format. Raise it when the saved data changes shape,
// and teach Open to read the older shapes.
const Version = 1

type envelope struct {
	Game    string          `json:"game"`
	Version int             `json:"version"`
	Kind    string          `json:"kind"` // "climb" or "fight"
	Data    json.RawMessage `json:"data"`
}

// Wrap makes a save of kind from v.
func Wrap(kind string, v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{Game: Game, Version: Version, Kind: kind, Data: data})
}

// Open reads a save: its kind, and its data into v.
func Open(b []byte, v any) (string, error) {
	var e envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return "", fmt.Errorf("this is not a Crusader Tactics save")
	}
	if e.Game != Game {
		return "", fmt.Errorf("this is not a save of the new Crusader Tactics")
	}
	if e.Version > Version {
		return "", fmt.Errorf("this save comes from a newer version of the game (save format %d)", e.Version)
	}
	if err := json.Unmarshal(e.Data, v); err != nil {
		return "", fmt.Errorf("this save is damaged: %v", err)
	}
	return e.Kind, nil
}
