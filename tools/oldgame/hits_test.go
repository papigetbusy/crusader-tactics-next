package oldgame

import (
	"strings"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		return x != ""
	}
	return false
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// unitAt turns a recorded unit into referee state, with the tracking the
// replay keeps (tiles moved this turn, skill uses).
func unitAt(h *HitUnit, m *Moment) *state.Unit {
	st := map[string]int{}
	for k, v := range h.St {
		st[k] = v
	}
	return &state.Unit{
		ID: h.ID, Team: h.Team, Class: h.Class, X: h.X, Y: h.Y,
		HP: h.HP, MaxHP: h.MaxHP, Atk: h.Atk, Def: h.Def, Mov: h.Mov, Level: h.Lv,
		Shield: h.Shield, Element: str(h.Element), Weapon: str(h.Weapon), RangeBonus: h.RB,
		Defending: truthy(h.Defend), Boss: truthy(h.Boss), Elite: truthy(h.Elite), Status: st,
		TilesMoved: m.TilesMoved[h.ID], SkillUses: m.Uses[h.ID],
	}
}

// TestOldGameHits runs every recorded hit and miss through Go's combat rules
// with the exact dice the old game rolled, and compares the outcome.
func TestOldGameHits(t *testing.T) {
	fights, err := Load(recordings)
	if err != nil {
		t.Fatal(err)
	}
	c := oldContent(t)
	var checked, hits, misses, kills, rolls int
	var bad problems
	Walk(c, fights, func(m *Moment, e Event) {
		if e.K != "hit" || e.A == nil || e.B == nil || e.R == nil {
			return
		}
		where := func() string {
			return strings.Join([]string{"fight", itoa(m.Fight), "step", itoa(m.Step) + ":",
				e.A.Class, "hits", e.B.Class, "with", e.S}, " ")
		}
		a, v := unitAt(e.A, m), unitAt(e.B, m)
		skill, classBase, ok := c.SkillFor(a.Team, a.Class, e.S)
		if !ok {
			bad.add("%s: Go does not know this skill", where())
			return
		}
		// The board as it was, with fresh copies of the two fighters.
		b := &state.Battle{N: m.B.N, Ground: m.B.Ground, Heights: m.B.Heights}
		for _, u := range m.B.Units {
			if u.ID != a.ID && u.ID != v.ID {
				cp := *u
				b.Units = append(b.Units, &cp)
			}
		}
		b.Units = append(b.Units, a, v)

		var values []uint32
		for i := e.G0; i < e.G1 && i < len(m.S.Dice.G); i++ {
			if strings.HasPrefix(m.F.Sites[m.S.Dice.S[i]], "hurt:") {
				values = append(values, m.S.Dice.G[i])
			}
		}
		feed := rng.NewFeed(values)
		var res combat.Result
		func() {
			defer func() {
				if r := recover(); r != nil {
					bad.add("%s: %v", where(), r)
				}
			}()
			res = combat.Hurt(c, b, feed, combat.Attack{By: a, Skill: skill, ClassBase: classBase, Target: v}, false)
		}()
		checked++
		rolls += feed.Used()
		switch {
		case res.Missed != (e.R.Miss != 0):
			bad.add("%s: old game missed=%v, Go missed=%v", where(), e.R.Miss != 0, res.Missed)
		case res.Damage != e.R.Dm:
			bad.add("%s: old game did %d damage, Go did %d (multiplier old %v, Go %v)", where(), e.R.Dm, res.Damage, e.R.M, res.Mult)
		case res.Mult != e.R.M:
			bad.add("%s: weak-spot multiplier old %v, Go %v", where(), e.R.M, res.Mult)
		case e.BHP != nil && res.HPAfter != *e.BHP:
			bad.add("%s: target HP after: old %d, Go %d", where(), *e.BHP, res.HPAfter)
		case feed.Used() != len(values):
			bad.add("%s: the old game rolled %d dice for this hit, Go rolled %d %v", where(), len(values), feed.Used(), feed.Labels())
		}
		switch {
		case res.Missed:
			misses++
		case res.Killed:
			kills++
			hits++
		default:
			hits++
		}
	})
	t.Logf("checked %d recorded hits: %d landed (%d of them kills), %d missed; %d dice replayed", checked, hits, kills, misses, rolls)
	if checked < 2500 {
		t.Errorf("expected about 2,550 recorded hits, checked only %d", checked)
	}
	for _, p := range bad {
		t.Error(p)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
