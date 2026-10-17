package floor

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

// The owner's monster redraw added four bosses and six elites (Grok, 2026-10-17).
var (
	newBosses = []string{"Bone Hydra", "Fallen Seraph", "Brood Matron", "Ash Colossus"}
	newElites = []string{"Crowned Slime", "Goblin Captain", "Grave Knight", "Dire Wolf", "Iron Minotaur", "Blight Reaper"}
)

func TestNewMonsterRoster(t *testing.T) {
	c := content.MustDefault()

	t.Run("The four new bosses guard the boss floors after the old six, in order, then the list repeats", func(t *testing.T) {
		every := c.Floors.BossEvery
		boss := func(number int) string {
			for _, u := range Build(c, 7, number, team(5, 40), Memory{}, nil).Enemies {
				if u.Boss {
					return u.Class
				}
			}
			return ""
		}
		for i, name := range newBosses {
			if got := boss((7 + i) * every); got != name {
				t.Fatalf("boss floor %d has %q, want %q", (7+i)*every, got, name)
			}
		}
		if got := boss(11 * every); got != "Bat Lord" {
			t.Fatalf("boss floor %d has %q, want the list to start again with Bat Lord", 11*every, got)
		}
	})

	t.Run("Every new elite turns up on elite floors", func(t *testing.T) {
		seen := map[string]bool{}
		for run := uint32(1); run <= 400; run++ {
			for _, u := range Build(c, run, c.Floors.EliteEvery*3, team(5, 30), Memory{}, nil).Enemies {
				if u.Elite {
					seen[u.Class] = true
				}
			}
		}
		for _, name := range newElites {
			if !seen[name] {
				t.Errorf("%s never appeared", name)
			}
		}
	})

	t.Run("Every new boss and elite has fighting numbers and at least one attack", func(t *testing.T) {
		for _, name := range append(newBosses, newElites...) {
			u := Enemy(c, name, 30, "dark", false, false)
			if u.MaxHP <= 0 || u.Atk <= 0 || u.Mov <= 0 || len(u.Skills) == 0 {
				t.Errorf("%s: HP %d, ATK %d, move %d, %d skills", name, u.MaxHP, u.Atk, u.Mov, len(u.Skills))
			}
		}
	})
}
