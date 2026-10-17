package stats

import (
	"slices"
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

// legendary makes the named legendary as it drops.
func legendary(c *content.Content, name string) content.Item {
	for _, u := range c.Gear.Uniques {
		if u.Name == name {
			return content.Item{Name: u.Name, Slot: u.Slot, Rarity: c.Gear.Legendary(), Unique: true, WeaponType: u.WeaponType,
				Weight: u.Weight, Set: u.Set, Class: u.Class, Affixes: slices.Clone(u.Affixes), Full: slices.Clone(u.Full)}
		}
	}
	panic("no legendary " + name)
}

func TestClassLegendaries(t *testing.T) {
	c := content.MustDefault()

	t.Run("Only its own class can use a class legendary", func(t *testing.T) {
		if ok, _ := CanUse(c, "Healer", 40, legendary(c, "Staff of Eternity")); !ok {
			t.Fatal("the Healer can't use her own staff")
		}
		if ok, why := CanUse(c, "Witch", 40, legendary(c, "Staff of Eternity")); ok || why != "Only a Healer can use Staff of Eternity." {
			t.Fatalf("the Witch can use the Healer's staff (%q)", why)
		}
	})

	t.Run("A class legendary's 4-piece bonus needs the weapon in hand and all 3 pieces of one legendary armor set", func(t *testing.T) {
		gear := map[string]content.Item{"w": legendary(c, "Bloodcrescent"),
			"h": legendary(c, "Dawnguard Helm"), "a": legendary(c, "Aegis of Dawn")}
		if af := SetAffixes(c, "Warrior", gear, 0); af["killReset"] != 0 {
			t.Fatal("2 armor pieces gave the set of 4")
		}
		gear["b"] = legendary(c, "Dawnstride Greaves")
		if af := SetAffixes(c, "Warrior", gear, 0); af["killReset"] != 1 {
			t.Fatalf("weapon and 3 pieces: %v", af)
		}
		gear["b"] = legendary(c, "Stormweave Treads")
		if af := SetAffixes(c, "Warrior", gear, 0); af["killReset"] != 0 {
			t.Fatal("pieces of two different sets gave the set of 4")
		}
	})

	t.Run("Two of the same class legendary in hand give its powers once and its other bonuses twice", func(t *testing.T) {
		gear := map[string]content.Item{"w": legendary(c, "Nightfang"), "w2": legendary(c, "Nightfang")}
		af := SetAffixes(c, "Assassin", gear, 0)
		if af["ambush"] != 30 || af["shadowstep"] != 1 || af["poisonCh"] != 60 {
			t.Fatalf("%v", af)
		}
	})

	t.Run("Two pieces of an epic set give its two-piece bonus, three pieces the full bonus", func(t *testing.T) {
		two := map[string]content.Item{"h": armor("h", "H", "eH", 3, 0), "a": armor("a", "H", "eH", 5, 0)}
		if af := AffixSum(c, two); af["dr"] != 5 || af["thorns"] != 0 {
			t.Fatalf("two pieces: %v", af)
		}
		two["b"] = armor("b", "H", "eH", 2, 0)
		if af := AffixSum(c, two); af["dr"] != 5 || af["thorns"] != 10 || af["hpPct"] != 10 {
			t.Fatalf("three pieces: %v", af)
		}
	})

	t.Run("Valor gives the Knight 1 more move", func(t *testing.T) {
		bare, _ := Geared(c, "Knight", 40, "", nil)
		n, _ := Geared(c, "Knight", 40, "", map[string]content.Item{"w": legendary(c, "Frostspire")})
		if n.Mov != bare.Mov+1 {
			t.Fatalf("move %d, bare %d", n.Mov, bare.Mov)
		}
	})
}
