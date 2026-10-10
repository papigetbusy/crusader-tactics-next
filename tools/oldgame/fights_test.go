package oldgame

import (
	"fmt"
	"slices"
	"sort"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// goUnit builds a referee unit from a recorded one, with its skill bar.
func goUnit(c *content.Content, su SnapUnit) *state.Unit {
	u := &state.Unit{
		ID: su.ID, Team: su.Team, Class: su.Class, X: su.X, Y: su.Y,
		HP: su.HP, MaxHP: su.MaxHP, Atk: su.Atk, Def: su.Def, Mov: su.Mov, Level: su.Lv,
		Element: su.Elem, Boss: su.Boss != 0, Elite: su.Elite != 0, Defending: su.Def2 != 0, Shield: su.Sh,
		Moved: su.Moved != 0, Acted: su.Acted != 0, Done: su.Done != 0, Status: map[string]int{},
	}
	for k, v := range su.St {
		u.Status[k] = v
	}
	if u.Team == state.Heroes {
		u.Skills = c.HeroSkills(u.Class, u.Level)
	} else {
		u.Skills = c.MonsterSkills(u.Class, u.Level, u.Boss || u.Elite)
		u.Weapon = c.Classes[u.Class].Weapon
	}
	u.Cooldowns = make([]int, len(u.Skills))
	for i, cd := range su.Cd {
		if cd != nil && i < len(u.Cooldowns) {
			u.Cooldowns[i] = *cd
		}
	}
	return u
}

// resync makes Go's board equal to a recorded moment, keeping what the
// recording does not store (tiles walked this turn, skill uses).
func resync(c *content.Content, f *Fight, g *engine.Game, s Snap) {
	old := map[int]*state.Unit{}
	b := f.Battle(Snap{Turn: s.Turn})
	if g.B != nil {
		for _, u := range g.B.Units {
			old[u.ID] = u
		}
		b.Side, b.Outcome = g.B.Side, g.B.Outcome
	}
	for _, su := range s.Units {
		u := goUnit(c, su)
		if o := old[su.ID]; o != nil {
			u.TilesMoved, u.SkillUses, u.LastSkill = o.TilesMoved, o.SkillUses, o.LastSkill
		}
		b.Units = append(b.Units, u)
	}
	g.B = b
}

// actionsFor turns one recorded player input into referee actions.
func actionsFor(step Step, b *state.Battle) []action.Action {
	switch step.Kind {
	case "wait":
		return []action.Action{{Kind: action.Wait, Unit: step.ID}}
	case "endturn":
		return []action.Action{{Kind: action.EndTurn}}
	case "move", "attack":
		var out []action.Action
		for _, e := range step.Ev {
			if e.K == "mv" && e.ID == step.ID {
				out = append(out, action.Action{Kind: action.Move, Unit: e.ID, X: e.To[0], Y: e.To[1]})
			}
			if e.K == "act" && e.ID == step.ID {
				u := b.Unit(e.ID)
				slot := -1
				if u != nil {
					slot = slices.IndexFunc(u.Skills, func(s content.Skill) bool { return s.Name == e.S })
				}
				out = append(out, action.Action{Kind: action.Attack, Unit: e.ID, Skill: slot, Target: e.V})
			}
		}
		return out
	}
	return nil
}

// arrivals adds reinforcements before the enemy phase. Where they appear is a
// floor rule ported later, so the recording says where.
func arrivals(c *content.Content, g *engine.Game, prev Snap, step Step) int {
	had := map[int]bool{}
	for _, u := range prev.Units {
		had[u.ID] = true
	}
	n := 0
	for _, su := range step.Snap.Units {
		if had[su.ID] || g.B.Unit(su.ID) != nil {
			continue
		}
		u := goUnit(c, su)
		u.HP, u.Status, u.Done, u.Moved = su.MaxHP, map[string]int{}, false, false
		for i := range u.Cooldowns {
			u.Cooldowns[i] = 0
		}
		for _, e := range step.Ev {
			if e.K == "mv" && e.ID == su.ID {
				u.X, u.Y = e.From[0], e.From[1]
				break
			}
		}
		g.B.Units = append(g.B.Units, u)
		n++
	}
	return n
}

// differences describes how Go's board differs from a recorded moment.
func differences(g *engine.Game, s Snap) []string {
	var out []string
	rec := map[int]SnapUnit{}
	for _, su := range s.Units {
		rec[su.ID] = su
		u := g.B.Unit(su.ID)
		if u == nil {
			out = append(out, fmt.Sprintf("%s %d is alive in the old game but gone in Go", su.Class, su.ID))
			continue
		}
		if u.X != su.X || u.Y != su.Y {
			out = append(out, fmt.Sprintf("%s %d stands at %d,%d in Go but %d,%d in the old game", su.Class, su.ID, u.X, u.Y, su.X, su.Y))
		}
		if u.HP != su.HP {
			out = append(out, fmt.Sprintf("%s %d has %d HP in Go but %d in the old game", su.Class, su.ID, u.HP, su.HP))
		}
		if a, b := statusText(u.Status), statusText(su.St); a != b {
			out = append(out, fmt.Sprintf("%s %d statuses: Go %s, old game %s", su.Class, su.ID, a, b))
		}
		for i, cd := range su.Cd {
			want := 0
			if cd != nil {
				want = *cd
			}
			if i < len(u.Cooldowns) && u.Cooldowns[i] != want {
				out = append(out, fmt.Sprintf("%s %d skill %s cooldown: Go %d, old game %d", su.Class, su.ID, u.Skills[i].Name, u.Cooldowns[i], want))
			}
		}
		if u.Done != (su.Done != 0) {
			out = append(out, fmt.Sprintf("%s %d finished-for-the-turn: Go %v, old game %v", su.Class, su.ID, u.Done, su.Done != 0))
		}
	}
	for _, u := range g.B.Units {
		if _, ok := rec[u.ID]; !ok {
			out = append(out, fmt.Sprintf("%s %d is alive in Go but gone in the old game", u.Class, u.ID))
		}
	}
	return out
}

func statusText(m map[string]int) string {
	var keys []string
	for k, v := range m {
		if v > 0 {
			keys = append(keys, fmt.Sprintf("%s:%d", k, v))
		}
	}
	sort.Strings(keys)
	return fmt.Sprint(keys)
}

// TestOldGameFights replays every recorded fight in Go from its first turn.
// The player's inputs are given to Go as actions; everything else (enemy
// decisions, damage, statuses, turn effects) is worked out by Go with the old
// game's dice, and compared with the old game after every step.
func TestOldGameFights(t *testing.T) {
	fights, err := Load(recordings)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var steps, matched, perfect, levelUps, reinforced, enemyTurns, rolls int
	var bad problems
	usable := 0
	for fi := range fights {
		f := &fights[fi]
		if !f.Usable() {
			continue
		}
		usable++
		g := &engine.Game{C: c}
		resync(c, f, g, f.Init.Snap)
		clean := true
		prev := f.Init.Snap
		for si := range f.Steps {
			step := f.Steps[si]
			steps++
			feed := newSiteFeed(f, &step)
			g.Dice = feed
			if step.Kind == "endturn" {
				reinforced += arrivals(c, g, prev, step)
			}
			var problem string
			func() {
				defer func() {
					if r := recover(); r != nil {
						problem = fmt.Sprint(r)
					}
				}()
				for _, a := range actionsFor(step, g.B) {
					evs, err := g.Apply(a)
					if err != nil {
						problem = fmt.Sprintf("Go refused %s: %v", a.Kind, err)
						return
					}
					for _, e := range evs {
						if e.Kind == action.SkillUsed && g.B.Unit(e.Unit) != nil && g.B.Unit(e.Unit).Team == state.Enemies {
							enemyTurns++
						}
					}
				}
				if left := feed.leftover(); left != "" {
					problem = "the old game made a roll Go never made, at " + left
				}
			}()
			rolls += feed.used
			// Experience and levels arrive in Phase 6: take level-ups from the recording.
			for _, su := range step.Snap.Units {
				if u := g.B.Unit(su.ID); u != nil && u.Team == state.Heroes && u.Level != su.Lv {
					u.Level, u.Atk, u.Def, u.MaxHP, u.Mov, u.HP = su.Lv, su.Atk, su.Def, su.MaxHP, su.Mov, su.HP
					levelUps++
				}
			}
			if problem == "" {
				if d := differences(g, step.Snap); len(d) > 0 {
					problem = d[0]
					if len(d) > 1 {
						problem += fmt.Sprintf(" (and %d more differences)", len(d)-1)
					}
				}
			}
			if problem != "" {
				if clean {
					bad.add("fight %d step %d (%s): %s", fi+1, si+1, step.Kind, problem)
				}
				clean = false
				resync(c, f, g, step.Snap)
			} else {
				matched++
			}
			prev = step.Snap
		}
		if clean {
			perfect++
		}
	}
	t.Logf("%d of %d fights replayed perfectly; %d of %d steps matched; %d monster skill uses decided by Go; %d dice used; %d level-ups and %d reinforcements taken from the recording",
		perfect, usable, matched, steps, enemyTurns, rolls, levelUps, reinforced)
	for _, p := range bad {
		t.Error(p)
	}
}
