package engine

import (
	"errors"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

var heroClasses = []string{"Warrior", "Tank", "Mage", "Healer", "Ranger", "Assassin", "Gunslinger", "Knight", "Dragon Knight", "Witch"}
var monsterClasses = []string{"Warrior", "Tank", "Mage", "Ranger", "Healer"}

// highLevel turns a practice setup into a Lv 28 fight with every class.
func highLevel(s Setup, shift int) Setup {
	s.Units = append([]UnitSetup(nil), s.Units...)
	h, m := shift, shift
	for i := range s.Units {
		u := &s.Units[i]
		u.Level, u.HP, u.Atk, u.Def, u.Mov, u.Equipped = 28, 160, 30, 8, 4, nil
		if u.Team == state.Heroes {
			u.Class = heroClasses[h%len(heroClasses)]
			h++
		} else {
			u.Class = monsterClasses[m%len(monsterClasses)]
			u.Boss, u.Elite = false, false
			m++
		}
	}
	return s
}

// aimAndUse picks at random from what SkillAim offers until the referee
// takes the action. It reports whether the skill was used, and fails the
// test if a full pick the screen offered is refused.
func aimAndUse(t *testing.T, g *Game, dice rng.Dice, u *state.Unit, slot int) bool {
	a := action.Action{Kind: action.Attack, Unit: u.ID, Skill: slot}
	for range 4 {
		units, tiles := g.SkillAim(u.ID, slot, a)
		switch {
		case len(units) > 0:
			id := units[rng.IntN(dice, "soak", len(units))]
			if a.Target == 0 {
				a.Target = id
			} else {
				a.Second = id
			}
		case len(tiles) > 0:
			a.Tiles = append(a.Tiles, tiles[rng.IntN(dice, "soak", len(tiles))])
		default:
			return false
		}
		_, err := g.Apply(a)
		if err == nil {
			return true
		}
		var r *Refused
		if !errors.As(err, &r) {
			t.Fatalf("%s %s: %v", u.Class, u.Skills[slot].Name, err)
		}
		if r.Code != "bad-target" {
			return false
		}
	}
	t.Fatalf("%s %s: the screen offered a pick the referee refused: %+v", u.Class, u.Skills[slot].Name, a)
	return false
}

func TestNewSkillsInFullFights(t *testing.T) {
	c := content.MustDefault()
	t.Run("Every new skill can be used by heroes and monsters in full fights without a hitch", func(t *testing.T) {
		used, monsterUsed := map[string]bool{}, map[string]bool{}
		for i, s := range practiceSetups(t) {
			for seed := uint64(1); seed <= 6; seed++ {
				g := NewFight(c, highLevel(s, int(seed)+i), seed)
				dice := rng.New(seed, 99)
				for turns := 0; turns < 30 && g.B.Outcome == state.Ongoing; turns++ {
					answerAll(t, g)
					for _, u := range append([]*state.Unit(nil), g.B.Units...) {
						if u.Team != state.Heroes || g.B.Unit(u.ID) == nil || u.Decoy || u.Companion || u.Summoned || g.B.Outcome != state.Ongoing {
							continue
						}
						// Try a different set of learned skills each turn.
						var normal []string
						for _, n := range u.Learned {
							if sk, ok := c.SkillNamed(n); ok && sk.Kind != "u" {
								normal = append(normal, n)
							}
						}
						rng.Shuffle(dice, "soak", normal)
						if len(normal) > 4 {
							normal = normal[:4]
						}
						u.Equipped = normal
						cds := u.Cooldowns
						u.Skills = c.Bar(u.Class, u.Equipped, u.Ultimate)
						u.Cooldowns = make([]int, len(u.Skills))
						for j := range u.Cooldowns {
							if j < len(cds) && turns%3 != 0 {
								u.Cooldowns[j] = cds[j]
							}
						}
						for tries := 0; tries < 3 && !u.Done && g.B.Unit(u.ID) != nil && g.B.Outcome == state.Ongoing; tries++ {
							slot := rng.IntN(dice, "soak", len(u.Skills))
							if u.Cooldowns[slot] == 0 && aimAndUse(t, g, dice, u, slot) {
								used[u.Skills[slot].Name] = true
							}
							answerAll(t, g)
						}
					}
					answerAll(t, g)
					if g.B.Outcome != state.Ongoing {
						break
					}
					ev, err := g.Apply(action.Action{Kind: action.EndTurn})
					if err != nil {
						t.Fatalf("end turn: %v", err)
					}
					for _, e := range ev {
						if e.Kind == action.SkillUsed {
							monsterUsed[e.Skill] = true
						}
					}
				}
			}
		}
		var missing []string
		for _, n := range c.SkillOrder {
			if sk := c.Skills[n]; sk.Class != "" && sk.Ascension == "" && !used[n] && combat.IsNew(sk) {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			t.Errorf("new skills that never came up in these fights: %v", missing)
		}
		var mm []string
		for _, class := range monsterClasses {
			for _, ls := range c.Classes[class].Schedule {
				if !monsterUsed[ls.Skill.Name] && ls.Level >= 18 { // a Lv 28 monster has its newest skills
					mm = append(mm, ls.Skill.Name)
				}
			}
		}
		if len(mm) > 0 {
			t.Errorf("Lv 28 monsters never used: %v", mm)
		}
	})
}
