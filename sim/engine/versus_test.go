package engine

import (
	"slices"
	"sort"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// picks is 8 different classes, each with one of its ascensions (shift picks
// which ones).
func picks(c *content.Content, shift int) []Pick {
	var classes []string
	for k := range c.AscensionChoices {
		classes = append(classes, k)
	}
	sort.Strings(classes)
	var out []Pick
	for i := 0; i < c.Versus.Team; i++ {
		cl := classes[(i+shift)%len(classes)]
		ch := c.AscensionChoices[cl]
		out = append(out, Pick{Class: cl, Ascension: ch[(i+shift)%len(ch)]})
	}
	return out
}

func versus(t *testing.T) *Game {
	t.Helper()
	c := content.MustDefault()
	g, err := NewVersus(c, VersusSetup{Teams: [2][]Pick{picks(c, 0), picks(c, 2)}, MapSeed: 7}, 11)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func side(g *Game, s int) []*state.Unit {
	var out []*state.Unit
	for _, u := range g.B.Units {
		if u.Team == s {
			out = append(out, u)
		}
	}
	return out
}

func TestVersus(t *testing.T) {
	t.Run("A versus match starts with 8 heroes each at Lv 30, with their ascension's ultimate and purple gear", func(t *testing.T) {
		g := versus(t)
		for s := 0; s < 2; s++ {
			us := side(g, s)
			if len(us) != 8 {
				t.Fatalf("side %d has %d units", s, len(us))
			}
			for _, u := range us {
				a := g.C.Ascensions[u.Ascension]
				if !u.PvP || u.Level != 30 || u.Ultimate != a.Ultimate || u.HP != u.MaxHP || len(u.Gear) < 5 {
					t.Fatalf("%s: pvp %v level %d ultimate %q (want %q) hp %d/%d gear %d", u.Class, u.PvP, u.Level, u.Ultimate, a.Ultimate, u.HP, u.MaxHP, len(u.Gear))
				}
				for sl, it := range u.Gear {
					if it.Rarity != g.C.Gear.Epic() {
						t.Fatalf("%s %s is rarity %d, not purple", u.Class, sl, it.Rarity)
					}
				}
			}
		}
	})

	t.Run("A class always wears the same purple set, and wearing it makes the hero stronger", func(t *testing.T) {
		c := content.MustDefault()
		for cl := range c.AscensionChoices {
			a, b := PurpleSet(c, cl), PurpleSet(c, cl)
			if len(a) != len(b) || a["a"].Name != b["a"].Name || !slices.Equal(a["w"].Affixes, b["w"].Affixes) {
				t.Fatalf("%s: two different sets", cl)
			}
			bare, _ := statsOf(c, cl, nil)
			geared, _ := statsOf(c, cl, a)
			if geared.Atk <= bare.Atk || geared.MaxHP <= bare.MaxHP {
				t.Fatalf("%s: atk %d → %d, hp %d → %d", cl, bare.Atk, geared.Atk, bare.MaxHP, geared.MaxHP)
			}
		}
	})

	t.Run("A versus team must be 8 different classes, each with one of its own ascensions", func(t *testing.T) {
		c := content.MustDefault()
		ok := picks(c, 0)
		twice := slices.Clone(ok)
		twice[1] = twice[0]
		wrong := slices.Clone(ok)
		wrong[0].Ascension = "Nobody"
		for _, team := range [][]Pick{ok[:7], twice, wrong} {
			if _, err := NewVersus(c, VersusSetup{Teams: [2][]Pick{team, ok}}, 1); err == nil {
				t.Fatalf("team %v was allowed", team)
			}
		}
	})

	t.Run("Players take turns: after the first player ends their turn, only the second player's heroes can act", func(t *testing.T) {
		g := versus(t)
		p1, p2 := side(g, 0)[0], side(g, 1)[0]
		if _, err := g.Apply(action.Action{Kind: action.Wait, Unit: p2.ID}); code(err) != "no-such-hero" {
			t.Fatalf("player 2 acted on player 1's turn: %v", err)
		}
		if _, err := g.Apply(action.Action{Kind: action.EndTurn}); err != nil {
			t.Fatal(err)
		}
		if g.B.Side != state.Enemies || g.B.Turn != 1 {
			t.Fatalf("side %d turn %d", g.B.Side, g.B.Turn)
		}
		if _, err := g.Apply(action.Action{Kind: action.Wait, Unit: p1.ID}); code(err) != "no-such-hero" {
			t.Fatalf("player 1 acted on player 2's turn: %v", err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Wait, Unit: p2.ID}); err != nil {
			t.Fatal(err)
		}
		if _, err := g.Apply(action.Action{Kind: action.EndTurn}); err != nil || g.B.Side != state.Heroes || g.B.Turn != 2 {
			t.Fatalf("side %d turn %d: %v", g.B.Side, g.B.Turn, err)
		}
	})

	t.Run("The second player's heroes have their class abilities, like the first player's", func(t *testing.T) {
		g := versus(t)
		for _, u := range side(g, 1) {
			if _, ok := combat.AbilityOf(g.C, u); !ok {
				t.Fatalf("player 2's %s has no class ability", u.Class)
			}
		}
	})

	t.Run("A player's fallen heroes go to their own side, for their own healers to revive", func(t *testing.T) {
		g := versus(t)
		v := side(g, 1)[0]
		v.HP = 0
		combat.Hurt(g.C, g.B, g.Dice, combat.Attack{By: side(g, 0)[0], Skill: g.C.Skills["Strike"], Target: v, Sure: true, Fixed: 1}, func(action.Event) {})
		if len(g.B.FallenFoes) != 1 || g.B.FallenFoes[0] != v || len(g.B.Fallen) != 0 {
			t.Fatalf("fallen %d, fallen foes %d", len(g.B.Fallen), len(g.B.FallenFoes))
		}
	})

	t.Run("Versus gives no experience, gold or loot", func(t *testing.T) {
		g := versus(t)
		u, v := side(g, 0)[0], side(g, 1)[0]
		xp, lv := u.XP, u.Level
		v.HP = 1
		combat.Hurt(g.C, g.B, g.Dice, combat.Attack{By: u, Skill: g.C.Skills["Strike"], Target: v, Sure: true, Fixed: 5}, func(action.Event) {})
		if g.B.Unit(v.ID) != nil || u.XP != xp || u.Level != lv || g.B.Gold != 0 || len(g.B.Bag) != 0 {
			t.Fatalf("xp %d→%d, gold %d, bag %d", xp, u.XP, g.B.Gold, len(g.B.Bag))
		}
	})

	t.Run("A player with no heroes left loses the match", func(t *testing.T) {
		g := versus(t)
		g.B.Units = slices.DeleteFunc(g.B.Units, func(u *state.Unit) bool { return u.Team == state.Enemies })
		ev, err := g.Apply(action.Action{Kind: action.EndTurn})
		if err != nil || g.B.Outcome != state.Won {
			t.Fatalf("outcome %q: %v %v", g.B.Outcome, ev, err)
		}
	})

	t.Run("There is no per-turn limit: a long think does not end the turn", func(t *testing.T) {
		g := versus(t)
		v := g.B.Versus
		ev, err := g.Apply(action.Action{Kind: action.Clock, Ms: 60000})
		if err != nil || g.B.Side != state.Heroes || has(ev, action.TurnTimeUp) || g.B.Outcome != state.Ongoing {
			t.Fatalf("side %d outcome %q events %v: %v", g.B.Side, g.B.Outcome, ev, err)
		}
		if v.BankMs[0] != g.C.Versus.BankSeconds*1000-60000 || v.UsedMs != 60000 || v.TurnMs != 0 {
			t.Fatalf("bank %d used %d turn %d", v.BankMs[0], v.UsedMs, v.TurnMs)
		}
	})

	t.Run("Running out of your total time loses the match, and the action is not done", func(t *testing.T) {
		g := versus(t)
		u := side(g, 0)[0]
		g.B.Versus.BankMs[0] = 1000
		ev, err := g.Apply(action.Action{Kind: action.Wait, Unit: u.ID, Ms: 1500})
		if err != nil || g.B.Outcome != state.Lost || g.B.Versus.TimedOut != 0 || !has(ev, action.TimeUp) || u.Done {
			t.Fatalf("outcome %q timed out %d done %v events %v: %v", g.B.Outcome, g.B.Versus.TimedOut, u.Done, ev, err)
		}
	})

	t.Run("The screen is shown whose turn it is, both clocks, and the winner", func(t *testing.T) {
		g := versus(t)
		if _, err := g.Apply(action.Action{Kind: action.Clock, Ms: 4000}); err != nil {
			t.Fatal(err)
		}
		v := g.View()
		if v.Versus == nil || v.Side != 0 || v.Versus.UsedMs != 4000 || v.Versus.BankMs[0] != g.C.Versus.BankSeconds*1000-4000 || v.Versus.Winner != -1 {
			t.Fatalf("view %+v side %d", v.Versus, v.Side)
		}
		g.B.Units = slices.DeleteFunc(g.B.Units, func(u *state.Unit) bool { return u.Team == state.Heroes })
		_, _ = g.Apply(action.Action{Kind: action.EndTurn})
		if w := g.View().Versus.Winner; w != 1 {
			t.Fatalf("winner %d", w)
		}
		for _, u := range g.View().Units {
			if u.Team == 1 && u.Form == "" {
				t.Fatalf("player 2's %s is shown without its form", u.Class)
			}
		}
	})

	t.Run("A player who gives up loses at once, and the other player wins", func(t *testing.T) {
		g := versus(t)
		ev, err := g.Apply(action.Action{Kind: action.Forfeit, Side: 1})
		if err != nil || g.B.Outcome != state.Won || !has(ev, action.GaveUp) || !has(ev, action.Won) {
			t.Fatalf("outcome %q events %v: %v", g.B.Outcome, ev, err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Forfeit, Side: 0}); err == nil {
			t.Fatal("a finished match accepted a second give-up")
		}
	})

	t.Run("Time can't run backwards", func(t *testing.T) {
		g := versus(t)
		if _, err := g.Apply(action.Action{Kind: action.Clock, Ms: -5}); code(err) != "bad-clock" {
			t.Fatalf("%v", err)
		}
	})
}

// statsOf is a class's numbers at the versus level wearing gear.
func statsOf(c *content.Content, class string, gear map[string]content.Item) (stats.Numbers, int) {
	return stats.Geared(c, class, c.Versus.Level, "", gear)
}

func TestVersusRules(t *testing.T) {
	c := content.MustDefault()
	t.Run("The versus picker is told the team size, level, clocks and every class's three ascensions in words", func(t *testing.T) {
		r := VersusRules(c)
		if r.Team != 8 || r.Level != 30 || r.TurnSeconds != 0 || r.BankSeconds != 300 || r.GraceSeconds != 2 || len(r.Classes) != 10 {
			t.Fatalf("%+v", r)
		}
		for _, cl := range r.Classes {
			if len(cl.Ascensions) != 3 {
				t.Fatalf("%s: %d ascensions", cl.Class, len(cl.Ascensions))
			}
			for _, a := range cl.Ascensions {
				if a.Passive == "" || a.Ultimate == "" {
					t.Fatalf("%s %s has no words", cl.Class, a.Name)
				}
			}
		}
	})
}
