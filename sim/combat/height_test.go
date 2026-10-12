package combat

import (
	"testing"
)

func TestHeight(t *testing.T) {
	t.Run("Attacking from a hill onto lower ground hits harder and surer", func(t *testing.T) {
		c, b, h, f := arena('g')
		flat := Preview(c, b, Attack{By: h, Skill: plain(1), Target: f})
		b.Heights = [][]int{{1, 0, 0}, {0, 0, 0}, {0, 0, 0}}
		up := Preview(c, b, Attack{By: h, Skill: plain(1), Target: f})
		if up.Mid <= flat.Mid || up.HitChance != min(100, flat.HitChance+10) {
			t.Fatalf("flat %d dmg %d%%, from the hill %d dmg %d%%", flat.Mid, flat.HitChance, up.Mid, up.HitChance)
		}
	})
	t.Run("Attacking uphill is less sure", func(t *testing.T) {
		c, b, h, f := arena('g')
		flat := Preview(c, b, Attack{By: f, Skill: plain(1), Target: h})
		b.Heights = [][]int{{1, 0, 0}, {0, 0, 0}, {0, 0, 0}}
		up := Preview(c, b, Attack{By: f, Skill: plain(1), Target: h})
		if up.HitChance != flat.HitChance-10 || up.Mid != flat.Mid {
			t.Fatalf("flat %d%%, uphill %d%%", flat.HitChance, up.HitChance)
		}
	})
	t.Run("Two units on hills fight as on flat ground", func(t *testing.T) {
		c, b, h, f := arena('g')
		flat := Preview(c, b, Attack{By: h, Skill: plain(1), Target: f})
		b.Heights = [][]int{{1, 1, 0}, {0, 0, 0}, {0, 0, 0}}
		both := Preview(c, b, Attack{By: h, Skill: plain(1), Target: f})
		if both != flat {
			t.Fatalf("flat %+v, both high %+v", flat, both)
		}
	})
}
