package oldgame

import (
	"slices"
	"testing"
)

// TestOldGameSkillBars checks Go's skill bars against the recordings: every
// skill a unit used must be on its Go bar, and when the skill has a cooldown
// the old game must have recorded it in the same slot.
func TestOldGameSkillBars(t *testing.T) {
	fights, err := Load(recordings)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var bad problems
	checked := 0
	for fi, f := range fights {
		if !f.Usable() {
			continue
		}
		prev := f.Init.Snap
		for si, step := range f.Steps {
			byID := map[int]SnapUnit{}
			for _, u := range prev.Units {
				byID[u.ID] = u
			}
			after := map[int]SnapUnit{}
			for _, u := range step.Snap.Units {
				after[u.ID] = u
			}
			for _, e := range step.Ev {
				u, ok := byID[e.ID]
				if e.K != "act" || !ok {
					continue
				}
				var names []string
				bar := c.MonsterSkills(u.Class, u.Lv, u.Boss != 0 || u.Elite != 0)
				if u.Team == 0 {
					bar = c.HeroSkills(u.Class, u.Lv)
				}
				for _, s := range bar {
					names = append(names, s.Name)
				}
				i := slices.Index(names, e.S)
				checked++
				if i < 0 {
					bad.add("fight %d step %d: %s used %s, which is not on its Go skill bar %v", fi+1, si+1, u.Class, e.S, names)
					continue
				}
				if a, alive := after[u.ID]; alive && bar[i].Cooldown > 0 && (i >= len(a.Cd) || a.Cd[i] == nil) {
					bad.add("fight %d step %d: %s used %s (slot %d) but the old game put no cooldown there: %v", fi+1, si+1, u.Class, e.S, i, a.Cd)
				}
			}
			prev = step.Snap
		}
	}
	t.Logf("checked %d skill uses against the skill bars", checked)
	if checked < 2500 {
		t.Errorf("expected about 2,600 skill uses, checked %d", checked)
	}
	for _, p := range bad {
		t.Error(p)
	}
}
