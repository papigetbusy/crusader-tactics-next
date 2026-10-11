package stats

import (
	"os"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

func sword(atk int, af ...content.Affix) content.Item {
	return content.Item{Name: "Steel Sword", Slot: "w", Rarity: 1, WeaponType: "sword", Atk: atk, Affixes: af}
}

func armor(slot, weight, set string, def, hp int, af ...content.Affix) content.Item {
	return content.Item{Name: "Test " + slot, Slot: slot, Weight: weight, Def: def, HP: hp, Set: set, Affixes: af}
}

// geared is a hero on the board wearing this gear.
func geared(c *content.Content, class string, level int, gear map[string]content.Item) *state.Unit {
	n, base := Geared(c, class, level, "", gear)
	u := &state.Unit{ID: 1, Team: state.Heroes, Class: class, Level: level, HP: n.MaxHP, MaxHP: n.MaxHP,
		Atk: n.Atk, Def: n.Def, Mov: n.Mov, BaseAtk: base, Gear: gear, Status: map[string]int{}}
	InitSets(c, u)
	return u
}

func TestGear(t *testing.T) {
	c := content.MustDefault()

	t.Run("Armor, helm, boots and charm add their numbers before the hero's form multiplies them", func(t *testing.T) {
		bare := Hero(c, "Warrior", 10, "")
		n, _ := Geared(c, "Warrior", 10, "", map[string]content.Item{"a": armor("a", "H", "", 6, 12)})
		p := Plain(c, "Warrior", 10)
		f := c.Progress.Forms[Form(c, 10, "")]
		if n.MaxHP != int(Round(float64(p.MaxHP+12)*f.HP)) || n.Def != bare.Def+6 {
			t.Fatalf("HP %d (bare %d), DEF %d (bare %d)", n.MaxHP, bare.MaxHP, n.Def, bare.Def)
		}
	})

	t.Run("A weapon adds its attack only while its weapon set is in hand", func(t *testing.T) {
		axe := content.Item{Name: "Battle Axe", Slot: "w", Rarity: 1, WeaponType: "axe", Atk: 9}
		u := geared(c, "Warrior", 10, map[string]content.Item{"w": sword(5), "w2": axe})
		if u.Atk != u.BaseAtk+5 || u.Weapon != "sword" {
			t.Fatalf("set 1: attack %d (base %d), weapon %q", u.Atk, u.BaseAtk, u.Weapon)
		}
		UseSet(c, u, 1)
		if u.Atk != u.BaseAtk+9 || u.Weapon != "axe" {
			t.Fatalf("set 2: attack %d (base %d), weapon %q", u.Atk, u.BaseAtk, u.Weapon)
		}
	})

	t.Run("Dual wielders hold both weapons in one set, and both add attack", func(t *testing.T) {
		k := func(atk int) content.Item {
			return content.Item{Name: "Steel Dagger", Slot: "w", Rarity: 1, WeaponType: "dagger", Atk: atk}
		}
		u := geared(c, "Assassin", 10, map[string]content.Item{"w": k(4), "w2": k(6)})
		if len(SetsOf(c, "Assassin")) != 1 || u.Atk != u.BaseAtk+10 {
			t.Fatalf("%d sets, attack %d (base %d)", len(SetsOf(c, "Assassin")), u.Atk, u.BaseAtk)
		}
	})

	t.Run("Two pieces of an armor set give its two-piece bonus, three pieces the full bonus", func(t *testing.T) {
		two := map[string]content.Item{"h": armor("h", "H", "H", 3, 0), "a": armor("a", "H", "H", 5, 0)}
		if af := AffixSum(c, two); af["dr"] != 8 || af["thorns"] != 0 {
			t.Fatalf("two pieces: %v", af)
		}
		two["b"] = armor("b", "H", "H", 2, 0)
		if af := AffixSum(c, two); af["dr"] != 8 || af["thorns"] != 25 || af["hpPct"] != 15 {
			t.Fatalf("three pieces: %v", af)
		}
	})

	t.Run("Gear's % bonuses raise max HP and DEF, and boots with +1 move add move", func(t *testing.T) {
		gear := map[string]content.Item{"h": armor("h", "M", "", 0, 0, content.Affix{Key: "hpPct", Value: 10}, content.Affix{Key: "defPct", Value: 50}),
			"b": armor("b", "M", "", 0, 0, content.Affix{Key: "movA", Value: 1})}
		bare := Hero(c, "Knight", 12, "")
		n, _ := Geared(c, "Knight", 12, "", gear)
		if n.MaxHP != int(Round(float64(bare.MaxHP)*1.1)) || n.Def != int(Round(float64(bare.Def)*1.5)) || n.Mov != bare.Mov+1 {
			t.Fatalf("%+v (bare %+v)", n, bare)
		}
	})

	t.Run("Weapon ranks rise with level: a main weapon is C at Lv 1, B at 8, A at 18 and S at 30", func(t *testing.T) {
		for lv, want := range map[int]string{1: "C", 8: "B", 18: "A", 30: "S"} {
			if got := c.Gear.Ranks[RankFor(c, "Warrior", lv, "sword")]; got != want {
				t.Errorf("Lv %d: %s", lv, got)
			}
		}
		if got := c.Gear.Ranks[RankFor(c, "Warrior", 1, "axe")]; got != "D" {
			t.Errorf("second weapon at Lv 1: %s", got)
		}
	})

	t.Run("A hero can't use a weapon its class doesn't train, or one above its weapon rank", func(t *testing.T) {
		bow := content.Item{Name: "Yew Bow", Slot: "w", Rarity: 1, WeaponType: "bow"}
		if ok, _ := CanUse(c, "Warrior", 10, bow); ok {
			t.Error("a Warrior used a bow")
		}
		legend := content.Item{Name: "Dawnbreaker", Slot: "w", Rarity: 3, WeaponType: "sword"}
		if ok, why := CanUse(c, "Warrior", 10, legend); ok || why == "" {
			t.Error("a Lv 10 Warrior (rank B) used a legendary that needs rank S")
		}
		if ok, _ := CanUse(c, "Warrior", 30, legend); !ok {
			t.Error("a Lv 30 Warrior could not use a legendary sword")
		}
	})

	t.Run("A hero can't wear armor heavier than its class allows", func(t *testing.T) {
		if ok, _ := CanUse(c, "Mage", 10, armor("a", "H", "", 9, 0)); ok {
			t.Error("a Mage wore heavy armor")
		}
		if ok, _ := CanUse(c, "Mage", 10, armor("a", "M", "", 6, 0)); !ok {
			t.Error("a Mage could not wear medium armor")
		}
	})

	t.Run("Skills need the class's own kind of weapon in hand; Strike, heals and buffs work with any", func(t *testing.T) {
		bow := content.Item{Name: "Yew Bow", Slot: "w", Rarity: 1, WeaponType: "bow", Atk: 5}
		u := geared(c, "Ranger", 12, map[string]content.Item{"w": bow, "w2": sword(5)})
		shot, strike := c.Skills["Shot"], c.Skills["Strike"]
		if !SetFits(c, u, shot, 0) || SetFits(c, u, shot, 1) || !SetFits(c, u, strike, 1) {
			t.Fatal("a Ranger's Shot should need the bow set; Strike works with the sword")
		}
	})

	t.Run("An unarmed hero counts as its class's usual weapon in the weapon triangle", func(t *testing.T) {
		u := geared(c, "Warrior", 5, nil)
		if u.Weapon != "sword" {
			t.Fatalf("unarmed Warrior counts as %q", u.Weapon)
		}
		old, err := content.FromOld(os.DirFS("../../testdata/oldgame/tables"))
		if err != nil {
			t.Fatal(err)
		}
		if w := geared(old, "Warrior", 5, nil).Weapon; w != "" {
			t.Fatalf("with the old game's rules an unarmed Warrior counts as %q", w)
		}
	})
}
