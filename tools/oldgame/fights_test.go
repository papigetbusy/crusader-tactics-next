package oldgame

import (
	"encoding/json"
	"fmt"
	"os"
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
	if su.Pot != nil {
		u.Potions = *su.Pot
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
			u.TilesMoved, u.SkillUses, u.LastSkill, u.Undo = o.TilesMoved, o.SkillUses, o.LastSkill, o.Undo
		}
		b.Units = append(b.Units, u)
	}
	if len(s.Seized) == 2 {
		b.Throne = &state.Tile{X: s.Seized[0], Y: s.Seized[1]}
	}
	if g.B != nil {
		fallen := map[int]*state.Unit{}
		for _, u := range g.B.Fallen {
			fallen[u.ID] = u
		}
		for _, id := range s.Fallen {
			if u := fallen[id]; u != nil {
				b.Fallen = append(b.Fallen, u)
			}
		}
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
	case "defend":
		return []action.Action{{Kind: action.Defend, Unit: step.ID}}
	case "potion":
		return []action.Action{{Kind: action.Potion, Unit: step.ID}}
	case "undo":
		return []action.Action{{Kind: action.Undo, Unit: step.ID}}
	case "move", "attack", "skill":
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
		if su.Pot != nil && u.Potions != *su.Pot {
			out = append(out, fmt.Sprintf("%s %d potions: Go %d, old game %d", su.Class, su.ID, u.Potions, *su.Pot))
		}
		if u.Done != (su.Done != 0) {
			out = append(out, fmt.Sprintf("%s %d finished-for-the-turn: Go %v, old game %v", su.Class, su.ID, u.Done, su.Done != 0))
		}
	}
	if s.Fallen != nil {
		var ids []int
		for _, u := range g.B.Fallen {
			ids = append(ids, u.ID)
		}
		if fmt.Sprint(ids) != fmt.Sprint(s.Fallen) && !(len(ids) == 0 && len(s.Fallen) == 0) {
			out = append(out, fmt.Sprintf("fallen heroes: Go %v, old game %v", ids, s.Fallen))
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
func TestOldGameFights(t *testing.T) { replayAll(t, recordings, "core") }

// TestOldGameCombat replays the Phase 5 recordings: menu skills, heals and
// buffs, Defend, Potion, Undo, revives and seize-the-throne floors.
func TestOldGameCombat(t *testing.T) { replayAll(t, combatRecordings, "combat") }

const combatRecordings = "../../testdata/oldgame/recordings/combat5-recordings.json.gz"

func replayAll(t *testing.T, file, set string) {
	fights, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var steps, matched, perfect, levelUps, reinforced, enemyTurns, rolls, ended int
	var bad problems
	usable := 0
	type fightResult struct {
		Fight   int      `json:"fight"`
		Floor   int      `json:"floor"`
		Party   []string `json:"party"`
		Enemies []string `json:"enemies"`
		Steps   int      `json:"steps"`
		Matched int      `json:"matched"`
		First   string   `json:"firstDifference,omitempty"`
	}
	var results []fightResult
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
		fr := fightResult{Fight: fi + 1, Floor: f.Cfg.Floor, Steps: len(f.Steps)}
		for _, u := range f.Init.Snap.Units {
			if u.Team == state.Heroes {
				fr.Party = append(fr.Party, fmt.Sprintf("%s %d", u.Class, u.Lv))
			} else {
				fr.Enemies = append(fr.Enemies, u.Class)
			}
		}
		for si := range f.Steps {
			step := f.Steps[si]
			steps++
			feed := newSiteFeed(f, &step)
			g.Dice = feed
			restart := restarted(prev, step.Snap)
			if step.Kind == "endturn" && !restart {
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
			if restart {
				// The old game ended the battle in this step and started the floor
				// again (a defeat re-runs the floor). Go must agree the battle ended.
				if problem == "" && g.B.Outcome == state.Ongoing {
					problem = "the old game's battle ended here, but in Go it goes on"
				}
				if problem == "" {
					matched++
					fr.Matched++
					ended++
				} else if clean {
					bad.add("fight %d step %d (%s): %s", fi+1, si+1, step.Kind, problem)
					fr.First = fmt.Sprintf("step %d (%s): %s", si+1, step.Kind, problem)
					clean = false
				}
				fr.Steps = si + 1
				break
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
					fr.First = fmt.Sprintf("step %d (%s): %s", si+1, step.Kind, problem)
				}
				clean = false
				resync(c, f, g, step.Snap)
			} else {
				matched++
				fr.Matched++
			}
			prev = step.Snap
		}
		if clean {
			perfect++
		}
		results = append(results, fr)
	}
	if dir := os.Getenv("OLDGAME_REPORT"); dir != "" {
		b, _ := json.MarshalIndent(results, "", " ")
		if err := os.WriteFile(dir+"/match-"+set+".json", b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d of %d fights replayed perfectly; %d of %d steps matched; %d monster skill uses decided by Go; %d dice used; %d battles ended with the same result in Go; %d level-ups and %d reinforcements taken from the recording",
		perfect, usable, matched, steps, enemyTurns, rolls, ended, levelUps, reinforced)
	for _, p := range bad {
		t.Error(p)
	}
}

// restarted reports a recorded step after which the old game started a new
// battle: heroes appear that were not there before.
func restarted(prev, now Snap) bool {
	had := map[int]bool{}
	for _, u := range prev.Units {
		had[u.ID] = true
	}
	for _, id := range prev.Fallen { // a revived hero is not a new battle
		had[id] = true
	}
	for _, u := range now.Units {
		if u.Team == state.Heroes && !had[u.ID] {
			return true
		}
	}
	return false
}
