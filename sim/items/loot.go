// Package items is loot: what drops from monsters, chests and bounties, how a
// new item is made, and gold from kills. It copies the old game's drop,
// bandRoll, lootFor, lootPick, genItem, rollAffixes, openChest and forceDrop
// (index.html lines 784, 1003-1005, 1062-1074, 1203, 1207, 1460 and 1707).
package items

import (
	"math"
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Roll labels, so a failed old-game check can say which roll was out of order.
const (
	RollDrop   = "loot drop roll"
	RollBand   = "loot rarity band"
	RollOwner  = "loot owner roll"
	RollSlot   = "loot slot roll"
	RollItem   = "item roll"
	RollBonus  = "item bonus roll"
	RollOrder  = "item bonus order"
	RollChest  = "chest roll"
	legendary  = 3
	rareRarity = 2
)

// Band is the rarity a floor's loot rolls (old bandRoll, line 1005).
func Band(c *content.Content, floor int, dice rng.Dice) int {
	r := dice.Roll(RollBand)
	for _, bd := range c.Gear.Loot.Bands {
		if bd.UpToFloor == 0 || floor <= bd.UpToFloor {
			if r < bd.LowChance {
				return bd.Low
			}
			return bd.High
		}
	}
	return 0
}

// Drop is what a fallen monster leaves (old drop, line 784): normal monsters
// sometimes drop an item of the floor's band; elites always, sometimes one
// rarity better; bosses a legendary the first time on a floor, at 25%.
func Drop(c *content.Content, b *state.Battle, dice rng.Dice, e, killer *state.Unit, emit action.Emit) {
	l := c.Gear.Loot
	r := dice.Roll(RollDrop)
	bd := Band(c, b.Floor, dice)
	k := -1
	switch {
	case e.Boss && b.BossLegendary:
		k = bd
	case e.Boss && r < l.BossLegendary:
		k = legendary
	case e.Boss:
		k = max(rareRarity, bd)
	case e.Elite:
		k = bd
		if r < l.EliteBetter {
			k++
		}
		k = min(legendary, k)
	case r < l.DropChance:
		k = bd
	}
	if k < 0 {
		return
	}
	k = allowed(c, b, k)
	if e.Boss && k == legendary {
		b.BossLegendary = true
	}
	it, ok := lootFor(c, b, dice, killer, k)
	give(b, it, ok, emit)
}

// Bounty is the item a Wanted Poster target always drops for its hunter
// (old forceDrop, index.html line 1207).
func Bounty(c *content.Content, b *state.Battle, dice rng.Dice, killer *state.Unit, emit action.Emit) {
	k := allowed(c, b, max(1, Band(c, b.Floor, dice)))
	it, ok := lootFor(c, b, dice, killer, k)
	give(b, it, ok, emit)
}

// OpenChest opens the chest a hero stops on (old openChest, line 1003):
// usually the chest's lower rarity, sometimes its higher one.
func OpenChest(c *content.Content, b *state.Battle, dice rng.Dice, u *state.Unit, emit action.Emit) {
	if u.Team != state.Heroes || u.Companion {
		return
	}
	i := slices.IndexFunc(b.Chests, func(ch state.Chest) bool { return ch.X == u.X && ch.Y == u.Y })
	if i < 0 {
		return
	}
	ch := b.Chests[i]
	b.Chests = append(b.Chests[:i:i], b.Chests[i+1:]...)
	t := c.Gear.Loot.Chests[ch.Kind]
	k := t[1]
	if dice.Roll(RollChest) < c.Gear.Loot.ChestUsual {
		k = t[0]
	}
	it, ok := lootFor(c, b, dice, u, allowed(c, b, k))
	give(b, it, ok, emit)
}

// allowed lowers a legendary to rare until a hero has reached the level
// legendaries need (approved 2026-10-11; the old game had no such limit).
func allowed(c *content.Content, b *state.Battle, k int) int {
	need := c.Gear.Loot.LegendaryFromLevel
	if k < legendary || need == 0 {
		return k
	}
	for _, id := range b.Party {
		if u := heroOf(b, id); u != nil && u.Level >= need {
			return k
		}
	}
	return rareRarity
}

func give(b *state.Battle, it content.Item, ok bool, emit action.Emit) {
	if !ok {
		return
	}
	b.Bag = append(b.Bag, it)
	emit(action.Event{Kind: action.Looted, Skill: it.Name, Amount: it.Rarity})
}

// heroOf finds a party hero on the board or among the fallen and lost: in
// the old game every party hero, alive or not, counts for who gets loot.
func heroOf(b *state.Battle, id int) *state.Unit {
	if u := b.Unit(id); u != nil {
		return u
	}
	for _, u := range append(slices.Clone(b.Fallen), b.Lost...) {
		if u.ID == id {
			return u
		}
	}
	return nil
}

// lootFor picks whose item it is (old lootFor, line 1074): half the time the
// hero who made the kill, else the party's least geared hero.
func lootFor(c *content.Content, b *state.Battle, dice rng.Dice, killer *state.Unit, k int) (content.Item, bool) {
	if killer != nil && killer.Decoy {
		killer = b.Unit(killer.Owner) // a shadow clone carries its hero's record
	}
	var p *state.Unit
	if killer != nil && slices.Contains(b.Party, killer.ID) && dice.Roll(RollOwner) < c.Gear.Loot.ForKiller {
		p = killer
	} else {
		best := math.Inf(1)
		for _, id := range b.Party {
			if u := heroOf(b, id); u != nil && gearScore(u) < best {
				p, best = u, gearScore(u)
			}
		}
	}
	if p == nil { // no party known (a bare test fight): nobody to give it to
		return content.Item{}, false
	}
	return lootPick(c, b, dice, p, k), true
}

// gearScore is how well geared a hero is: 1 per item plus 1.5 per rarity.
func gearScore(u *state.Unit) float64 {
	s := 0.0
	for _, it := range u.Gear {
		s += 1 + float64(it.Rarity)*1.5
	}
	return s
}

// lootPick makes an item for hero p (old lootPick, lines 1071-1073): mostly
// for an empty slot; a weapon of a type it uses, no rarer than its rank
// allows (a legendary weapon it can't use becomes a legendary armor piece
// or charm instead).
func lootPick(c *content.Content, b *state.Battle, dice rng.Dice, p *state.Unit, k int) content.Item {
	armor := []string{"L", "M", "H"}
	cg, ok := c.Gear.Classes[p.Class]
	if ok {
		armor = cg.Armor
	}
	slots := append([]string{"h", "a", "b", "c"}, stats.WeaponSlotsOf(c, p.Class)...)
	var empty []string
	for _, sl := range slots {
		if _, worn := p.Gear[sl]; !worn {
			empty = append(empty, sl)
		}
	}
	var sl string
	if len(empty) > 0 && dice.Roll(RollSlot) < c.Gear.Loot.EmptySlot {
		sl = empty[rng.IntN(dice, RollSlot, len(empty))]
	} else {
		sl = slots[rng.IntN(dice, RollSlot, len(slots))]
	}
	if slices.Contains(content.WeaponSlots, sl) {
		types := weaponTypes(cg)
		t := types[0]
		if dice.Roll(RollSlot) >= c.Gear.Loot.MainWeapon {
			t = types[rng.IntN(dice, RollSlot, len(types))]
		}
		rk := stats.RankFor(c, p.Class, p.Level, t)
		kk := k
		for kk > 0 && c.Gear.RarityRank[kk] > rk {
			kk--
		}
		if k == legendary && kk < legendary {
			s2 := []string{"h", "a", "b", "c"}[rng.IntN(dice, RollSlot, 4)]
			at := ""
			if s2 != "c" {
				at = armor[rng.IntN(dice, RollSlot, len(armor))]
			}
			return Make(c, b.Floor, dice, legendary, s2, "", at, "")
		}
		return Make(c, b.Floor, dice, kk, "w", t, "", p.Class)
	}
	if sl == "c" {
		return Make(c, b.Floor, dice, k, "c", "", "", p.Class)
	}
	return Make(c, b.Floor, dice, k, sl, "", armor[rng.IntN(dice, RollSlot, len(armor))], p.Class)
}

// weaponTypes are a class's weapon types, each once, in slot order (old profs).
func weaponTypes(cg content.ClassGear) []string {
	var out []string
	for _, w := range cg.Weapons {
		if !slices.Contains(out, w.Type) {
			out = append(out, w.Type)
		}
	}
	if len(out) == 0 {
		out = []string{"sword"}
	}
	return out
}

// Gold is the gold a kill pays (old hurt, line 1707): more for higher
// levels, elites and bosses, double for a Bounty Hunter, triple for a Wanted
// Poster target, half on a floor cleared before.
func Gold(c *content.Content, b *state.Battle, killer, v *state.Unit) int {
	l := c.Gear.Loot
	mul := 1.0
	if b.Replayed {
		mul = l.GoldReplayed
	}
	size := 1.0
	if v.Boss {
		size = l.GoldBoss
	} else if v.Elite {
		size = l.GoldElite
	}
	bh := 1.0
	if killer.Ascension == "Bounty Hunter" {
		bh = l.GoldBounty
	}
	g := round(float64(float64(mul*(l.GoldBase+float64(float64(v.Level)*l.GoldPerLevel))*size) * bh))
	if v.Wanted && v.WantedBy == killer.ID {
		g *= l.GoldWanted
	}
	return g
}
