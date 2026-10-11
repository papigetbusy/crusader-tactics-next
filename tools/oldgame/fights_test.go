package oldgame

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// differences describes how Go's board differs from a recorded moment.
func differences(g *engine.Game, s Snap, party []int, seen int) []string {
	var out []string
	for i, id := range party {
		if u := heroOf(g.B, id); u != nil && i < len(s.XP) && (u.Level != s.XP[i][0] || u.XP != s.XP[i][1]) {
			out = append(out, fmt.Sprintf("%s %d is level %d with %d XP in Go but level %d with %d XP in the old game",
				u.Class, id, u.Level, u.XP, s.XP[i][0], s.XP[i][1]))
		}
	}
	rec := map[int]SnapUnit{}
	var goClones, oldClones []string
	for _, u := range g.B.Units {
		if u.Decoy {
			goClones = append(goClones, fmt.Sprintf("%d,%d:%d", u.X, u.Y, u.HP))
		}
	}
	for _, su := range s.Units {
		if su.Dc != 0 {
			oldClones = append(oldClones, fmt.Sprintf("%d,%d:%d", su.X, su.Y, su.HP))
		}
	}
	slices.Sort(goClones)
	slices.Sort(oldClones)
	if !slices.Equal(goClones, oldClones) {
		out = append(out, fmt.Sprintf("shadow clones (tile:HP): Go %v, old game %v", goClones, oldClones))
	}
	for _, su := range s.Units {
		if su.Dc != 0 {
			continue // the recorder gave clones their Phantom's ID; compared above
		}
		rec[su.ID] = su
		u := g.B.Unit(su.ID)
		if u == nil {
			out = append(out, fmt.Sprintf("%s %d is alive in the old game but gone in Go", su.Class, su.ID))
			continue
		}
		if u.X != su.X || u.Y != su.Y {
			out = append(out, fmt.Sprintf("%s %d stands at %d,%d in Go but %d,%d in the old game", su.Class, su.ID, u.X, u.Y, su.X, su.Y))
		}
		if u.MaxHP != su.MaxHP || u.Atk != su.Atk || u.Def != su.Def || u.Mov != su.Mov {
			out = append(out, fmt.Sprintf("%s %d has max HP %d, attack %d, defense %d, move %d in Go but %d, %d, %d, %d in the old game",
				su.Class, su.ID, u.MaxHP, u.Atk, u.Def, u.Mov, su.MaxHP, su.Atk, su.Def, su.Mov))
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
		if u.AbilityCooldown != su.Ab {
			out = append(out, fmt.Sprintf("%s %d ability cooldown: Go %d, old game %d", su.Class, su.ID, u.AbilityCooldown, su.Ab))
		}
		if su.Sp != nil && u.Ascension != *su.Sp {
			out = append(out, fmt.Sprintf("%s %d ascension: Go %q, old game %q", su.Class, su.ID, u.Ascension, *su.Sp))
		}
		if su.Sk != nil {
			var bar []string
			for _, sk := range u.Skills {
				bar = append(bar, sk.Name)
			}
			if fmt.Sprint(bar) != fmt.Sprint(su.Sk) {
				out = append(out, fmt.Sprintf("%s %d skill bar: Go %v, old game %v", su.Class, su.ID, bar, su.Sk))
			}
		}
		if s.Inv != nil && su.Team == state.Heroes {
			if u.Weapon != deref(su.W) || u.WeaponSet != su.Ws {
				out = append(out, fmt.Sprintf("%s %d holds set %d (%q) in Go but set %d (%q) in the old game", su.Class, su.ID, u.WeaponSet+1, u.Weapon, su.Ws+1, deref(su.W)))
			}
			if a, b := statusText(u.Affix), statusText(su.Af); a != b {
				out = append(out, fmt.Sprintf("%s %d gear bonuses: Go %s, old game %s", su.Class, su.ID, a, b))
			}
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
	if s.FieldsAt != nil && (len(g.B.Fields) != len(s.FieldsAt) || len(g.B.Traps) != len(s.TrapsAt)) {
		out = append(out, fmt.Sprintf("Go has %d fields and %d traps, the old game %d and %d", len(g.B.Fields), len(g.B.Traps), len(s.FieldsAt), len(s.TrapsAt)))
	}
	out = append(out, lootDifferences(g, s)...)
	for _, u := range g.B.Units {
		if _, ok := rec[u.ID]; !ok && !u.Decoy {
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

// TestOldGameHeroes replays the Phase 6 recordings: high-level parties with
// ascensions, every class ability, and the choices made at level-ups.
func TestOldGameHeroes(t *testing.T) { replayAll(t, heroesRecordings, "heroes") }

// TestOldGameGear replays fights with geared heroes: weapons, armor, charms,
// sets and legendaries, weapon set choices and equipping mid-battle.
func TestOldGameGear(t *testing.T) { replayAll(t, gearRecordings, "gear") }

// TestOldGameLoot replays fights with loot: drops, chests, gold, bounties,
// companion spirits and taking gear off.
func TestOldGameLoot(t *testing.T) { replayAll(t, lootRecordings, "loot") }

const lootRecordings = "../../testdata/oldgame/recordings/loot6-recordings.json.gz"

const gearRecordings = "../../testdata/oldgame/recordings/gear6-recordings.json.gz"

const heroesRecordings = "../../testdata/oldgame/recordings/heroes6-recordings.json.gz"

const combatRecordings = "../../testdata/oldgame/recordings/combat5-recordings.json.gz"

func replayAll(t *testing.T, file, set string) {
	fights, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var steps, matched, perfect, levelUps, reinforced, enemyTurns, rolls, ended, raised int
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
		party := partyOf(f.Init.Snap)
		resync(c, f, g, f.Init.Snap, party)
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
		seen, clones := maxID(f.Init.Snap, -1), 500_000
		for si := range f.Steps {
			step := f.Steps[si]
			steps++
			feed := newSiteFeed(f, &step)
			feed.noChests = prev.ChestsAt == nil && opensChest(f, &step)
			if prev.Inv != nil && step.Snap.Inv != nil {
				feed.newLoot = newItems(prev.Inv, step.Snap.Inv)
			}
			g.Dice = feed
			g.B.NextID = seen + 1 // units Go creates get the IDs the recorder gave them
			restart := restarted(prev, step.Snap)
			if step.Kind == "endturn" && !restart {
				reinforced += arrivals(c, g, prev, step)
			}
			if prev.Inv != nil {
				g.B.Bag = bagOf(prev) // loot is ported in 6d
			}
			var problem string
			over := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						problem = fmt.Sprint(r)
					}
				}()
				if debugging(set, fi, si) {
					printUnits(g)
					g.Dice = loudDice{feed}
				}
				acts := actionsFor(step, g.B)
				if step.Kind == "decide" {
					if a, ok := decision(c, g, step, party); ok {
						acts = append(acts, a)
					}
				}
				for _, a := range acts {
					evs, err := g.Apply(a)
					debugStep(set, fi, si, step, evs)
					if r, ok := err.(*engine.Refused); ok && r.Code == "battle-over" && noFoes(step.Snap) {
						over = true // the battle was won; the recorder kept clicking until the next floor
						return
					}
					if err != nil && (a.Kind == action.Ability || a.Kind == action.Undo) {
						continue // the old game also refuses some uses (quietly); the board comparison tells
					}
					if err != nil {
						problem = fmt.Sprintf("Go refused %s: %v", a.Kind, err)
						return
					}
					for _, e := range evs {
						if e.Kind == action.LevelUp {
							levelUps++
						}
						if e.Kind == action.SkillUsed && g.B.Unit(e.Unit) != nil && g.B.Unit(e.Unit).Team == state.Enemies {
							enemyTurns++
						}
					}
				}
				for _, u := range g.B.Units {
					if u.Decoy && u.ID < 1_000_000 {
						// The recorder never gave clones IDs of their own; keep Go's
						// clones clear of the IDs it gives later arrivals.
						u.ID, clones = 1_000_000+clones, clones+1
					}
				}
				if set != "heroes" && set != "gear" && set != "loot" {
					// These recordings dismissed every level-up choice unanswered.
					g.B.Decisions = nil
				}
				if left := feed.leftover(); left != "" {
					problem = "the old game made a roll Go never made, at " + left
				}
			}()
			if over {
				matched++
				fr.Matched++
				ended++
				fr.Steps = si + 1
				break
			}
			if feed.noChests && step.Snap.Inv != nil {
				g.B.Bag = bagOf(step.Snap) // the chest's item, from the recording (it did not store chests)
			}
			rolls += feed.used
			raised += skeletonsFromRecording(c, g, step.Snap, seen)
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
				if d := differences(g, step.Snap, party, seen); len(d) > 0 {
					problem = d[0]
					if len(d) > 1 {
						problem += fmt.Sprintf(" (and %d more differences)", len(d)-1)
					}
				}
			}
			if problem != "" && timerRaised(prev, step) {
				raised++ // a Grave Bloom skeleton the old game raised on its timer mid-turn
				resync(c, f, g, step.Snap, party)
				if step.Kind == "endturn" {
					g.B.Side = state.Heroes // Go's enemy turn stopped part-way; the old one finished
				}
				problem = ""
			}
			if problem != "" {
				if clean {
					bad.add("fight %d step %d (%s): %s", fi+1, si+1, step.Kind, problem)
					fr.First = fmt.Sprintf("step %d (%s): %s", si+1, step.Kind, problem)
				}
				clean = false
				resync(c, f, g, step.Snap, party)
			} else {
				matched++
				fr.Matched++
			}
			prev = step.Snap
			seen = maxID(step.Snap, seen)
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
	t.Logf("%d of %d fights replayed perfectly; %d of %d steps matched; %d monster skill uses decided by Go; %d dice used; %d battles ended with the same result in Go; %d level-ups worked out by Go; %d reinforcements and %d Grave Bloom skeletons taken from the recording",
		perfect, usable, matched, steps, enemyTurns, rolls, ended, levelUps, reinforced, raised)
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
		if u.Team == state.Heroes && !had[u.ID] && u.Cp == 0 && u.Sm == 0 && u.Dc == 0 {
			return true
		}
	}
	return false
}

// maxID is the highest unit ID seen so far.
func maxID(s Snap, seen int) int {
	for _, u := range s.Units {
		seen = max(seen, u.ID)
	}
	for _, id := range s.Fallen {
		seen = max(seen, id)
	}
	return seen
}

// noFoes reports a recorded moment with no monsters left.
func noFoes(s Snap) bool {
	for _, u := range s.Units {
		if u.Team == state.Enemies {
			return false
		}
	}
	return true
}
