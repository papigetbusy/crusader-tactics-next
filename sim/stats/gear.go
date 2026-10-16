package stats

import (
	"fmt"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// SetsOf lists a class's weapon sets as the slots in each (old setsOf, line
// 1048): one slot per set, or both weapons in one set for dual wielders.
func SetsOf(c *content.Content, class string) [][]string {
	cg, ok := c.Gear.Classes[class]
	if !ok {
		return [][]string{{"w"}}
	}
	if cg.Dual {
		return [][]string{{"w", "w2"}}
	}
	var out [][]string
	for i := range cg.Weapons {
		out = append(out, []string{content.WeaponSlots[i]})
	}
	return out
}

// WeaponSlotsOf are the weapon slots a class has (old slotsOf, line 1046).
func WeaponSlotsOf(c *content.Content, class string) []string {
	cg, ok := c.Gear.Classes[class]
	if !ok {
		return []string{"w"}
	}
	return content.WeaponSlots[:len(cg.Weapons)]
}

// SlotType is the weapon type that goes in a weapon slot (old slotType).
func SlotType(c *content.Content, class, slot string) string {
	cg := c.Gear.Classes[class]
	if i := slices.Index(content.WeaponSlots, slot); i >= 0 && i < len(cg.Weapons) {
		return cg.Weapons[i].Type
	}
	return ""
}

// SetItems are the weapons in set i (old setItems, line 1049).
func SetItems(c *content.Content, class string, gear map[string]content.Item, i int) []content.Item {
	sets := SetsOf(c, class)
	if i < 0 || i >= len(sets) {
		return nil
	}
	var out []content.Item
	for _, sl := range sets[i] {
		if it, ok := gear[sl]; ok {
			out = append(out, it)
		}
	}
	return out
}

// RankFor is a hero's rank with a weapon type at a level, or -1 if the class
// cannot use it (old profs and rankFor, lines 1044-1045): the first slot
// with that type decides the role.
func RankFor(c *content.Content, class string, level int, weaponType string) int {
	for _, w := range c.Gear.Classes[class].Weapons {
		if w.Type == weaponType {
			return c.Gear.Rank(w.Role, level)
		}
	}
	return -1
}

// CanUse says whether a hero of this class and level may use an item, and
// why not (old canUse, lines 1053-1054).
func CanUse(c *content.Content, class string, level int, it content.Item) (bool, string) {
	cg, ok := c.Gear.Classes[class]
	if !ok {
		return true, ""
	}
	g := c.Gear
	switch it.Slot {
	case content.SlotWeapon:
		rk := RankFor(c, class, level, it.WeaponType)
		if rk < 0 {
			return false, fmt.Sprintf("A %s can't use a %s.", class, g.WeaponNames[it.WeaponType])
		}
		if need := g.RarityRank[it.Rarity]; rk < need {
			return false, fmt.Sprintf("Needs %s rank %s (yours: %s).", g.WeaponNames[it.WeaponType], g.Ranks[need], g.Ranks[rk])
		}
	case content.SlotHelm, content.SlotArmor, content.SlotBoots:
		if !slices.Contains(cg.Armor, it.Weight) {
			return false, fmt.Sprintf("A %s can't wear %s armor.", class, weightName(it.Weight))
		}
	}
	return true, ""
}

func weightName(w string) string {
	return map[string]string{"L": "light", "M": "medium", "H": "heavy"}[w]
}

// AffixSum is the bonuses of everything worn except weapons, plus armor set
// bonuses (old afSum, line 1055).
func AffixSum(c *content.Content, gear map[string]content.Item) map[string]int {
	out := map[string]int{}
	count := map[string]int{}
	for _, sl := range gearSlots {
		it, ok := gear[sl]
		if !ok {
			continue
		}
		if it.Set != "" {
			count[it.Set]++
		}
		if !slices.Contains(content.WeaponSlots, sl) {
			for _, a := range it.Affixes {
				out[a.Key] += a.Value
			}
		}
	}
	for _, key := range sortedKeys(count) {
		set, ok := c.Gear.Sets[key]
		if !ok {
			continue
		}
		if count[key] >= 2 {
			for _, a := range set.Two {
				out[a.Key] += a.Value
			}
		}
		if count[key] >= 3 {
			for _, a := range set.Three {
				out[a.Key] += a.Value
			}
		}
	}
	return out
}

// gearSlots are all slots, in a fixed order (sums do not depend on order).
var gearSlots = []string{"h", "a", "b", "w", "c", "w2", "w3"}

func sortedKeys(m map[string]int) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// SetAffixes is AffixSum plus the bonuses of the weapons in set i (old
// afOfSet, line 1056).
func SetAffixes(c *content.Content, class string, gear map[string]content.Item, i int) map[string]int {
	out := AffixSum(c, gear)
	for _, it := range SetItems(c, class, gear, i) {
		for _, a := range it.Affixes {
			out[a.Key] += a.Value
		}
	}
	return out
}

// SetAtk is the attack of the weapons in set i (both, for dual wielders).
func SetAtk(c *content.Content, class string, gear map[string]content.Item, i int) int {
	n := 0
	for _, it := range SetItems(c, class, gear, i) {
		n += it.Atk
	}
	return n
}

// SetWeaponType is the type of the first weapon in set i, or "" if empty.
func SetWeaponType(c *content.Content, class string, gear map[string]content.Item, i int) string {
	if items := SetItems(c, class, gear, i); len(items) > 0 {
		return items[0].WeaponType
	}
	return ""
}

// SetRank is the hero's rank with the weapon in its hand, or -1 when the set
// is empty or the unit carries no sets (old uSetRank, line 1059).
func SetRank(c *content.Content, u *state.Unit) int {
	if !u.Armed {
		return -1
	}
	t := SetWeaponType(c, u.Class, u.Gear, u.WeaponSet)
	if t == "" {
		return -1
	}
	return RankFor(c, u.Class, u.Level, t)
}

// HandCategory is the category of the weapon in hand: m, r or g. Empty
// hands are melee (old strikeCat, line 1112).
func HandCategory(c *content.Content, u *state.Unit) string {
	if !u.Armed {
		return "m"
	}
	if t := SetWeaponType(c, u.Class, u.Gear, u.WeaponSet); t != "" {
		return c.Gear.Category(t)
	}
	return "m"
}

// SetFits reports whether a skill works with weapon set i: Strike, heals,
// buffs, revives and summons always do; other skills need a weapon of the
// class's own category, or empty hands (old setOk, line 1114).
func SetFits(c *content.Content, u *state.Unit, s content.Skill, i int) bool {
	if !u.Armed || s.Strike || s.Heal || s.Support || s.Revive || s.Summon != nil {
		return true
	}
	t := SetWeaponType(c, u.Class, u.Gear, i)
	if t == "" {
		return true
	}
	return c.Gear.Category(t) == ClassCategory(c, u.Class)
}

// ClassCategory is the category of a class's first weapon (old SCAT).
func ClassCategory(c *content.Content, class string) string {
	cg, ok := c.Gear.Classes[class]
	if !ok || len(cg.Weapons) == 0 {
		return "m"
	}
	return c.Gear.Category(cg.Weapons[0].Type)
}

// InitSets gives a hero its weapon sets and puts set WeaponSet in hand (old
// initSets, line 1058). Monsters, pets, companions and clones carry none.
func InitSets(c *content.Content, u *state.Unit) {
	_, ok := c.Gear.Classes[u.Class]
	u.Armed = ok && u.Player() && !u.Companion && !u.Summoned && !u.Decoy
	if !u.Armed {
		return
	}
	UseSet(c, u, u.WeaponSet)
}

// UseSet puts weapon set i in the hero's hand: its bonuses, its weapon for
// the triangle, and its attack (old useSet, line 1057).
func UseSet(c *content.Content, u *state.Unit, i int) {
	if !u.Armed {
		return
	}
	u.WeaponSet = max(0, min(i, len(SetsOf(c, u.Class))-1))
	u.Affix = SetAffixes(c, u.Class, u.Gear, u.WeaponSet)
	u.Weapon = HandWeapon(c, u)
	u.Atk = u.BaseAtk + SetAtk(c, u.Class, u.Gear, u.WeaponSet)
}

// HandWeapon is the weapon a hero counts as in the weapon triangle: the
// weapon in hand if it is a sword, axe or lance. With empty hands the old
// game used none; with the approved change it is the class's usual weapon.
func HandWeapon(c *content.Content, u *state.Unit) string {
	t := SetWeaponType(c, u.Class, u.Gear, u.WeaponSet)
	if t == "" && c.Combat.UnarmedClassWeapon {
		if cg := c.Gear.Classes[u.Class]; len(cg.Weapons) > 0 {
			t = cg.Weapons[0].Type
		}
	}
	if slices.Contains(c.Combat.TriangleWeapons, t) {
		return t
	}
	return ""
}
