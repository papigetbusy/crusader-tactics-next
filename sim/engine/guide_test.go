package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

func TestGuide(t *testing.T) {
	t.Run("The screen's rules guide lists every kind of ground, the element wheel and the weapon triangle", func(t *testing.T) {
		c := content.MustDefault()
		g := RulesGuide(c)
		if len(g.Terrain) != len(c.Terrain) || g.Beats["fire"] != "ice" || g.WeaponBeats["sword"] != "lance" {
			t.Fatalf("%+v", g)
		}
	})
}
