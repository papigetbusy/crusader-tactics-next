package combat

import (
	"fmt"
	"math"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/board"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// GroundEffect is one thing the ground under a unit does for it right now,
// in words for the screen's badges (owner, 2026-10-16: show terrain like
// statuses). The numbers come from the same rules the fight uses.
type GroundEffect struct {
	Label string `json:"label"`
	Good  bool   `json:"good"`
}

// GroundEffects lists what the tile under u does: less damage taken, healing
// each turn, and high ground.
func GroundEffects(c *content.Content, b *state.Battle, u *state.Unit) []GroundEffect {
	if b == nil || u == nil {
		return nil
	}
	tr, ok := c.Terrain[board.GroundAt(b, u.X, u.Y)]
	if !ok {
		return nil
	}
	var out []GroundEffect
	name := strings.ToUpper(tr.Name)
	if tr.DamageTaken > 0 && tr.DamageTaken < 1 {
		out = append(out, GroundEffect{Label: fmt.Sprintf("%s −%d%% DMG TAKEN", name, int(math.Round((1-tr.DamageTaken)*100))), Good: true})
	}
	if tr.Heal > 0 {
		out = append(out, GroundEffect{Label: fmt.Sprintf("%s +%d HP/t", name, int(Round(float64(u.MaxHP)*tr.Heal))), Good: true})
	}
	if c.Combat.Height && high(b, u.X, u.Y) {
		out = append(out, GroundEffect{Label: fmt.Sprintf("HIGH GROUND +%d%% HIT · +%d%% DMG", c.Combat.HeightHit,
			int(math.Round((c.Combat.HeightDamage-1)*100))), Good: true})
	}
	return out
}
