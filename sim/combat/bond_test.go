package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestCompanions(t *testing.T) {
	t.Run("Defeating a companion spirit bonds it to the hero; a hero that has one chooses which to keep", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Ally, f.Kind, f.Element, f.HP = true, "Wolf", "dark", 1
		Hurt(c, b, hits(), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if h.Bond == nil || h.Bond.Kind != "Wolf" || len(b.Decisions) == 0 || b.Decisions[0].Kind != state.DecideBond {
			t.Fatalf("bond %v decisions %v", h.Bond, b.Decisions)
		}
		b.Decisions = nil
		g := &state.Unit{ID: 3, Team: state.Enemies, Class: "Healer", X: 0, Y: 1, HP: 1, MaxHP: 20, Ally: true, Kind: "Slime", Status: map[string]int{}}
		b.Units = append(b.Units, g)
		Hurt(c, b, hits(), Attack{By: h, Skill: plain(1), Target: g}, nop)
		if h.Bond.Kind != "Wolf" || len(b.Decisions) == 0 || b.Decisions[0].Kind != state.DecideSwap || b.Decisions[0].Skill != "Slime" {
			t.Fatalf("bond %v decisions %v", h.Bond, b.Decisions)
		}
	})

	t.Run("A companion spirit starts the fight beside its hero, and once it dies it is gone for good", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Level, h.Bond, h.X, h.Y = 10, &state.Bond{Kind: "Wolf", Element: "dark"}, 0, 2
		b.NextID = 5
		SpawnCompanion(c, b, h)
		z := b.Unit(5)
		if z == nil || !z.Companion || z.Level != 8 || z.Kind != "Wolf" || Distance(z, h) > 1 {
			t.Fatalf("companion %+v", z)
		}
		z.HP = 1
		Hurt(c, b, hits(), Attack{By: f, Skill: plain(1), Target: z, Fixed: 99}, nop)
		if h.Bond != nil {
			t.Fatal("the hero still has its companion")
		}
	})
}
