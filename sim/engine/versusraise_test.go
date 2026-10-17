package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// necro makes a side's first hero a Necromancer witch.
func necro(g *Game, s int) *state.Unit {
	u := side(g, s)[0]
	u.Class, u.Ascension = "Witch", "Necromancer"
	return u
}

func raised(g *Game, team int) (n int) {
	for _, u := range g.B.Units {
		if u.Undead && u.Team == team {
			n++
		}
	}
	return
}

func TestVersusGraveBloom(t *testing.T) {
	t.Run("In versus, Grave Bloom raises the other player's fallen heroes, never your own", func(t *testing.T) {
		g := versus(t)
		nb := necro(g, 0)
		g.B.SideKills[0] = []state.Tile{{X: 3, Y: 20}}               // where a Blue hero fell
		g.B.SideKills[1] = []state.Tile{{X: 9, Y: 9}, {X: 10, Y: 9}} // where two Red heroes fell
		ab, ok := g.C.AbilityFor("Witch", 30, "Necromancer", false)
		if !ok || ab.Effect != "graveBloom" {
			t.Fatalf("ability %+v", ab)
		}
		if err := combat.UseAbility(g.C, g.B, g.Dice, nb, ab, nil, nil, func(action.Event) {}); err != nil {
			t.Fatal(err)
		}
		if raised(g, 0) != 2 || raised(g, 1) != 0 || len(g.B.SideKills[0]) != 1 || len(g.B.SideKills[1]) != 0 {
			t.Fatalf("Blue raised %d, Red %d; kills left %v", raised(g, 0), raised(g, 1), g.B.SideKills)
		}
	})
	t.Run("While Grave Bloom lasts, an enemy hero who falls rises for the Necromancer's side", func(t *testing.T) {
		g := versus(t)
		nb := necro(g, 0)
		nb.Status["gbloom"] = 3
		foe := side(g, 1)[0]
		foe.HP = 1
		emit := func(action.Event) {}
		combat.Hurt(g.C, g.B, g.Dice, combat.Attack{By: nb, Skill: content.Skill{Name: "Test", Power: 1, MaxRange: 99}, Target: foe, Sure: true, Fixed: 50}, emit)
		combat.FlushRaises(g.C, g.B, emit)
		if raised(g, 0) != 1 || raised(g, 1) != 0 {
			t.Fatalf("Blue raised %d, Red %d", raised(g, 0), raised(g, 1))
		}
	})
	t.Run("A raised monster is built the same for Blue and for Red", func(t *testing.T) {
		g := versus(t)
		b0, r0 := necro(g, 0), necro(g, 1)
		emit := func(action.Event) {}
		combat.RaiseSkeleton(g.C, g.B, b0, b0.X, b0.Y-2, false, emit)
		combat.RaiseSkeleton(g.C, g.B, r0, r0.X, r0.Y+2, false, emit)
		var a, b *state.Unit
		for _, u := range g.B.Units {
			if u.Undead && u.Team == 0 {
				a = u
			}
			if u.Undead && u.Team == 1 {
				b = u
			}
		}
		if a == nil || b == nil || a.Weapon != b.Weapon || a.PvP != b.PvP || a.MaxHP != b.MaxHP || a.Atk != b.Atk {
			t.Fatalf("Blue's %+v\nRed's %+v", a, b)
		}
	})
}
