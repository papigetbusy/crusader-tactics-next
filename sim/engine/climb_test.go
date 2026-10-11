package engine

import (
	"slices"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/turn"
)

func climbGame() *Game { return NewClimb(content.MustDefault(), 4242, 7) }

// winNow ends the battle as won (every monster gone).
func winNow(g *Game) {
	g.B.Units = slices.DeleteFunc(g.B.Units, func(u *state.Unit) bool { return u.Team == state.Enemies })
	turn.Check(g.B, func(action.Event) {})
	g.finish()
}

func must(t *testing.T, g *Game, a action.Action) {
	t.Helper()
	if _, err := g.Apply(a); err != nil {
		t.Fatalf("%s: %v", a.Kind, err)
	}
}

func TestClimb(t *testing.T) {
	t.Run("A climb starts on floor 1 with all ten heroes at Lv 1 and 5 hero slots, looking at the map", func(t *testing.T) {
		g := climbGame()
		v := g.View().Climb
		if v.Floor != 1 || v.Slots != 5 || len(v.Heroes) != 10 || v.Phase != PhasePrepare || len(v.Placed) != 5 {
			t.Fatalf("%+v", v)
		}
	})

	t.Run("Choosing a team needs one different hero for every slot", func(t *testing.T) {
		g := climbGame()
		if _, err := g.Apply(action.Action{Kind: action.Team, Heroes: []string{"Mage", "Healer"}}); code(err) != "slots" {
			t.Fatalf("2 heroes: %v", err)
		}
		if _, err := g.Apply(action.Action{Kind: action.Team, Heroes: []string{"Mage", "Mage", "Tank", "Knight", "Witch"}}); code(err) != "bad-team" {
			t.Fatalf("a hero twice: %v", err)
		}
		must(t, g, action.Action{Kind: action.Team, Heroes: []string{"Mage", "Healer", "Tank", "Knight", "Witch"}})
		if g.B.Unit(g.B.Party[0]).Class != "Mage" || len(g.B.Party) != 5 {
			t.Fatal("the team is not on the map")
		}
	})

	t.Run("Heroes start only on blue tiles; placing on another hero's tile swaps the two", func(t *testing.T) {
		g := climbGame()
		p := g.climb.placed
		if _, err := g.Apply(action.Action{Kind: action.Place, Hero: p[0].Class, X: 0, Y: 0}); code(err) != "not-start" {
			t.Fatalf("top corner: %v", err)
		}
		a, b := p[0], p[1]
		must(t, g, action.Action{Kind: action.Place, Hero: a.Class, X: b.Tile.X, Y: b.Tile.Y})
		if g.climb.placed[0].Tile != b.Tile || g.climb.placed[1].Tile != a.Tile {
			t.Fatalf("placed %v", g.climb.placed)
		}
	})

	t.Run("Fighting waits until the battle starts; gear can be changed before", func(t *testing.T) {
		g := climbGame()
		if _, err := g.Apply(action.Action{Kind: action.Wait, Unit: g.B.Party[0]}); code(err) != "not-in-battle" {
			t.Fatalf("waiting before the start: %v", err)
		}
		must(t, g, action.Action{Kind: action.Start})
		must(t, g, action.Action{Kind: action.Wait, Unit: g.B.Party[0]})
	})

	t.Run("Winning the highest floor opens the next; only the heroes who fought gain XP", func(t *testing.T) {
		g := climbGame()
		must(t, g, action.Action{Kind: action.Start})
		fighter := g.B.Unit(g.B.Party[0])
		fighter.XP = 5
		bench := g.climb.run.Hero("Knight")
		winNow(g)
		v := g.View().Climb
		if v.Phase != PhaseWon || v.Best != 2 || v.Floor != 2 || !v.Summary.NewBest {
			t.Fatalf("%+v", v)
		}
		if g.climb.run.Hero(fighter.Class).XP != 5 || bench.XP != 0 || slices.ContainsFunc(v.Summary.Heroes, func(h HeroSummary) bool { return h.Class == "Knight" }) {
			t.Fatal("XP went to the wrong heroes")
		}
		must(t, g, action.Action{Kind: action.Continue})
		if g.climb.phase != PhasePrepare || g.B.Floor != 2 {
			t.Fatal("not on floor 2")
		}
	})

	t.Run("A defeat keeps what was earned and the floor is fought again", func(t *testing.T) {
		g := climbGame()
		must(t, g, action.Action{Kind: action.Start})
		g.B.Gold = 40
		g.B.Units = slices.DeleteFunc(g.B.Units, func(u *state.Unit) bool { return u.Team == state.Heroes })
		turn.Check(g.B, func(action.Event) {})
		g.finish()
		if g.climb.phase != PhaseLost || g.climb.run.Gold != 40 {
			t.Fatalf("phase %s, gold %d", g.climb.phase, g.climb.run.Gold)
		}
		must(t, g, action.Action{Kind: action.Continue})
		if g.B.Floor != 1 || g.climb.phase != PhasePrepare {
			t.Fatal("not back on floor 1")
		}
	})

	t.Run("The shop comes up before every 7th floor, then that floor is fought", func(t *testing.T) {
		g := climbGame()
		g.climb.run.Floor, g.climb.run.Best = 6, 6
		g.prepare()
		must(t, g, action.Action{Kind: action.Start})
		winNow(g)
		must(t, g, action.Action{Kind: action.Continue})
		v := g.View().Climb
		if v.Phase != PhaseShop || len(v.Shop) != 6 {
			t.Fatalf("phase %s, %d items", v.Phase, len(v.Shop))
		}
		must(t, g, action.Action{Kind: action.Continue})
		if g.climb.phase != PhasePrepare || g.B.Floor != 7 || len(g.B.Units) <= 5 {
			t.Fatal("floor 7 is not a fight")
		}
	})

	t.Run("Buying costs an item's price; selling pays half", func(t *testing.T) {
		g := climbGame()
		r := g.climb.run
		r.Floor, r.Best, r.Gold, g.climb.phase = 7, 7, 10000, PhaseShop
		stock := r.Stock(g.C, 7)
		it := stock[0]
		v := g.View().Climb
		must(t, g, action.Action{Kind: action.Buy, Item: 0})
		if r.Gold != 10000-v.Shop[0].Price || r.Bag[len(r.Bag)-1].Name != it.Name {
			t.Fatalf("gold %d", r.Gold)
		}
		gold := r.Gold
		sell := g.View().Climb.Bag[0].Price
		must(t, g, action.Action{Kind: action.Sell, Item: 0})
		if r.Gold != gold+sell || sell != (v.Shop[0].Price+1)/2 && sell != v.Shop[0].Price/2 {
			t.Fatalf("sold for %d (price %d)", sell, v.Shop[0].Price)
		}
		r.Gold = 0
		if _, err := g.Apply(action.Action{Kind: action.Buy, Item: 0}); code(err) != "shop" {
			t.Fatalf("buying with no gold: %v", err)
		}
	})

	t.Run("Two copies of the same item combine into one a rarity higher", func(t *testing.T) {
		g := climbGame()
		r := g.climb.run
		g.climb.phase = PhaseShop
		sw := content.Item{Name: "Steel Sword", Slot: "w", Rarity: 1, WeaponType: "sword", Atk: 4, Affixes: []content.Affix{{Key: "dmg", Value: 8}}}
		r.Bag = []content.Item{sw, sw}
		must(t, g, action.Action{Kind: action.Combine, Item: 0, Second: 1})
		if len(r.Bag) != 1 || r.Bag[0].Rarity != 2 || r.Bag[0].WeaponType != "sword" {
			t.Fatalf("bag %+v", r.Bag)
		}
		r.Bag = append(r.Bag, sw)
		if _, err := g.Apply(action.Action{Kind: action.Combine, Item: 0, Second: 1}); code(err) != "shop" {
			t.Fatalf("different items: %v", err)
		}
	})

	t.Run("The floor list travels to any floor reached; opened chests stay opened", func(t *testing.T) {
		g := climbGame()
		r := g.climb.run
		r.Best = 9
		if _, err := g.Apply(action.Action{Kind: action.Travel, Floor: 12}); code(err) != "no-floor" {
			t.Fatalf("an unreached floor: %v", err)
		}
		must(t, g, action.Action{Kind: action.Travel, Floor: 9})
		if g.B.Floor != 9 || len(g.View().Climb.Floors) != 9 {
			t.Fatal("not on floor 9")
		}
		r.Memory(9).Opened = []state.Tile{{X: 1, Y: 1}}
		r.Memory(9).Chests = 2
		if left := g.View().Climb.Floors[0].ChestsLeft; left != 1 {
			t.Fatalf("chests left %d", left)
		}
	})

	t.Run("Each of floors 25, 50 and 75 reached adds a hero slot, up to 8", func(t *testing.T) {
		r := climbGame().climb.run
		for best, want := range map[int]int{24: 5, 25: 6, 50: 7, 75: 8, 300: 8} {
			if r.Best = best; r.Slots() != want {
				t.Errorf("best floor %d: %d slots", best, r.Slots())
			}
		}
	})

	t.Run("Reaching every 100th floor gives a free legendary", func(t *testing.T) {
		g := climbGame()
		r := g.climb.run
		r.Floor, r.Best = 99, 99
		g.prepare()
		must(t, g, action.Action{Kind: action.Start})
		winNow(g)
		if s := g.climb.summary; s.Gift == nil || s.Gift.Rarity != "legendary" || r.Gifts != 1 {
			t.Fatalf("summary %+v", s)
		}
	})

	t.Run("A hero with an ascension can change it once per floor; the ultimate changes with it", func(t *testing.T) {
		g := climbGame()
		h := g.climb.run.Hero("Warrior")
		h.Level, h.Ascension = 30, g.C.AscensionChoices["Warrior"][0]
		other := g.C.AscensionChoices["Warrior"][1]
		g.prepare()
		must(t, g, action.Action{Kind: action.Ascend, Hero: "Warrior", Choice: other})
		if h.Ascension != other || h.Ultimate != g.C.Ascensions[other].Ultimate {
			t.Fatalf("%s, ultimate %s", h.Ascension, h.Ultimate)
		}
		if _, err := g.Apply(action.Action{Kind: action.Ascend, Hero: "Warrior", Choice: g.C.AscensionChoices["Warrior"][2]}); code(err) != "ascension" {
			t.Fatalf("a second change: %v", err)
		}
	})

	t.Run("A floor's monster level is set on the first visit and kept", func(t *testing.T) {
		g := climbGame()
		first := g.climb.fl.Level
		for _, h := range g.climb.run.Heroes {
			h.Level = 40
		}
		g.prepare()
		if g.climb.fl.Level != first {
			t.Fatalf("level %d, first %d", g.climb.fl.Level, first)
		}
	})

	t.Run("The team picker shows each hero's numbers, what it can use and its class ability", func(t *testing.T) {
		for _, h := range climbGame().View().Climb.Heroes {
			if h.HP <= 0 || h.Atk <= 0 || h.Mov <= 0 || h.Weapons == "" || h.Ability == "" {
				t.Fatalf("%+v", h)
			}
		}
	})

	t.Run("Gear changed before the battle starts doesn't use up the hero's turn", func(t *testing.T) {
		g := climbGame()
		g.B.Bag = append(g.B.Bag, content.Item{Name: "Steel Sword", Slot: "w", Rarity: 1, WeaponType: "sword", Atk: 4})
		var w *state.Unit
		for _, id := range g.B.Party {
			if u := g.B.Unit(id); u.Class == "Warrior" {
				w = u
			}
		}
		must(t, g, action.Action{Kind: action.Equip, Unit: w.ID, Item: len(g.B.Bag) - 1})
		must(t, g, action.Action{Kind: action.Unequip, Unit: w.ID, Slot: "w"})
		if w.Done {
			t.Fatal("the Warrior is finished before the battle started")
		}
	})

	t.Run("The shop lists which bag items can be combined, two copies at a time", func(t *testing.T) {
		g := climbGame()
		r := g.climb.run
		g.climb.phase = PhaseShop
		sw := content.Item{Name: "Steel Sword", Slot: "w", Rarity: 1, WeaponType: "sword", Atk: 4}
		ax := content.Item{Name: "Iron Axe", Slot: "w", Rarity: 0, WeaponType: "axe", Atk: 3}
		r.Bag = []content.Item{sw, ax, sw, sw}
		if p := g.View().Climb.Pairs; len(p) != 1 || p[0] != [2]int{0, 2} {
			t.Fatalf("pairs %v", p)
		}
	})

	t.Run("Aiming a skill asks for one pick at a time, and nothing more once it is aimed", func(t *testing.T) {
		g := climbGame()
		must(t, g, action.Action{Kind: action.Start})
		u := g.B.Unit(g.B.Party[0])
		u.Skills = append(u.Skills, g.C.Skills["Bog of Despair"])
		u.Cooldowns = append(u.Cooldowns, 0)
		slot := len(u.Skills) - 1
		_, tiles := g.SkillAim(u.ID, slot, action.Action{})
		if len(tiles) == 0 {
			t.Fatal("no tiles to aim at")
		}
		if units, more := g.SkillAim(u.ID, slot, action.Action{Tiles: [][2]int{tiles[0]}}); len(units)+len(more) != 0 {
			t.Fatalf("still asking after a tile was picked: %v %v", units, more)
		}
	})

	t.Run("The gear sheet compares each bag item with what the hero wears, and says when it can't be worn", func(t *testing.T) {
		g := climbGame()
		var w *state.Unit
		for _, id := range g.B.Party {
			if u := g.B.Unit(id); u.Class == "Warrior" {
				w = u
			}
		}
		g.B.Bag = []content.Item{
			{Name: "Steel Sword", Slot: "w", Rarity: 1, WeaponType: "sword", Atk: 4},
			{Name: "Ash Staff", Slot: "w", Rarity: 0, WeaponType: "staff", Atk: 4},
		}
		f := g.bagFits(w)
		if !f[0].OK || f[0].Diff["atk"] != 4 || f[1].OK || f[1].Why == "" {
			t.Fatalf("%+v", f)
		}
	})
}
