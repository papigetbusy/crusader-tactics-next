package engine

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func TestVersusSkills(t *testing.T) {
	c := content.MustDefault()
	t.Run("In versus, both players' heroes can use every skill and class ability without a hitch", func(t *testing.T) {
		used := [2]map[string]bool{{}, {}}
		abil := [2]int{}
		for seed := uint64(1); seed <= 30; seed++ {
			g, err := NewVersus(c, VersusSetup{Teams: [2][]Pick{picks(c, int(seed)), picks(c, int(seed)+4)}, MapSeed: uint32(seed)}, seed)
			if err != nil {
				t.Fatal(err)
			}
			dice := rng.New(seed, 77)
			for turns := 0; turns < 40 && g.B.Outcome == state.Ongoing; turns++ {
				answerAll(t, g)
				for _, u := range append([]*state.Unit(nil), g.B.Units...) {
					if !g.mine(u) || g.B.Unit(u.ID) == nil || u.Decoy || u.Companion || u.Summoned || g.B.Outcome != state.Ongoing {
						continue
					}
					if ab, ok := combat.AbilityOf(c, u); ok && u.AbilityCooldown == 0 && rng.IntN(dice, "soak", 3) == 0 {
						if useAbility(g, dice, u) {
							abil[u.Team]++
							used[u.Team]["✦"+ab.Name] = true
						}
						answerAll(t, g)
					}
					for tries := 0; tries < 3 && !u.Done && g.B.Unit(u.ID) != nil && g.B.Outcome == state.Ongoing; tries++ {
						slot := rng.IntN(dice, "soak", len(u.Skills))
						if u.Cooldowns[slot] == 0 && aimAndUse(t, g, dice, u, slot) {
							used[u.Team][u.Skills[slot].Name] = true
						}
						answerAll(t, g)
					}
				}
				answerAll(t, g)
				if g.B.Outcome != state.Ongoing {
					break
				}
				if _, err := g.Apply(action.Action{Kind: action.EndTurn}); err != nil {
					t.Fatalf("end turn: %v", err)
				}
			}
		}
		t.Logf("player 1 used %d different skills and abilities (%d abilities), player 2 %d (%d)", len(used[0]), abil[0], len(used[1]), abil[1])
		if len(used[1]) < len(used[0])/2 || abil[1] == 0 {
			t.Fatalf("player 2 could hardly act: %d vs %d", len(used[1]), len(used[0]))
		}
	})
}

// useAbility aims u's class ability at random from what the referee offers.
func useAbility(g *Game, dice rng.Dice, u *state.Unit) bool {
	a := action.Action{Kind: action.Ability, Unit: u.ID}
	if _, err := g.Apply(a); err == nil {
		return true
	}
	for _, v := range g.B.Units {
		a.Target = v.ID
		if _, err := g.Apply(a); err == nil {
			return true
		}
	}
	return false
}
