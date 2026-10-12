package combat

import (
	"testing"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// arm gives a hero this gear and puts weapon set 1 in hand.
func arm(c *content.Content, u *state.Unit, gear map[string]content.Item) {
	u.Gear, u.BaseAtk = gear, u.Atk
	stats.InitSets(c, u)
}

func weapon(name, typ string, rarity, atk int) content.Item {
	return content.Item{Name: name, Slot: "w", Rarity: rarity, WeaponType: typ, Atk: atk}
}

func TestGearInCombat(t *testing.T) {
	t.Run("Strike reaches 1 tile with a melee weapon and 2-6 tiles with a bow, gun, staff or wand", func(t *testing.T) {
		c, _, h, _ := arena('g')
		h.Class = "Ranger"
		arm(c, h, map[string]content.Item{"w": weapon("Yew Bow", "bow", 1, 4), "w2": weapon("Steel Sword", "sword", 1, 4)})
		if lo, hi := strikeRange(c, h); lo != 2 || hi != 6 {
			t.Fatalf("bow: %d-%d", lo, hi)
		}
		stats.UseSet(c, h, 1)
		if lo, hi := strikeRange(c, h); lo != 1 || hi != 1 {
			t.Fatalf("sword: %d-%d", lo, hi)
		}
	})

	t.Run("Strike with a staff or wand in hand is magic: it ignores half of the target's defense", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.Class, f.X, f.Def = "Mage", 2, 10
		arm(c, h, map[string]content.Item{"w": weapon("Twig Wand", "wand", 0, 0)})
		strike := c.Skills["Strike"]
		if !isMagic(c, h, strike) {
			t.Fatal("a wand Strike is not magic")
		}
		dm, _ := Damage(c, b, Attack{By: h, Skill: strike, Target: f}, 1)
		r := stats.SetRank(c, h)
		want := int(Round(float64(h.Atk)*strike.Power*c.Gear.RankDamage[r] - 5))
		if dm != want {
			t.Fatalf("damage %d, want %d", dm, want)
		}
	})

	t.Run("Weapon rank changes damage and hit chance: rank S hits 20% harder and 15% more often", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Def = 0
		h.Level = 30
		arm(c, h, map[string]content.Item{"w": weapon("Rusty Blade", "sword", 0, 0)})
		dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1)
		if dm != 24 {
			t.Fatalf("rank S damage %d, want 24 (20 x 1.2)", dm)
		}
		bare := HitChance(c, nil, Attack{By: &state.Unit{Team: state.Heroes, Status: map[string]int{}}, Skill: plain(1), Target: f})
		if hc := HitChance(c, nil, Attack{By: h, Skill: plain(1), Target: f}); hc != min(100, bare+15) {
			t.Fatalf("hit %d, bare %d", hc, bare)
		}
	})

	t.Run("Gear's damage bonus raises damage, and damage reduction gear lowers it, at most by half", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Def = 0
		h.Affix = map[string]int{"dmg": 10}
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 22 {
			t.Fatalf("with +10%% damage: %d", dm)
		}
		h.Affix = nil
		f.Affix = map[string]int{"dr": 80}
		if dm, _ := Damage(c, b, Attack{By: h, Skill: plain(1), Target: f}, 1); dm != 10 {
			t.Fatalf("with 80%% damage reduction (capped at 50%%): %d", dm)
		}
	})

	t.Run("Evade gear lowers the attacker's hit chance; hit gear raises it", func(t *testing.T) {
		c, _, h, f := arena('g')
		base := HitChance(c, nil, Attack{By: f, Skill: plain(1), Target: h})
		h.Affix = map[string]int{"evade": 20}
		f.Affix = map[string]int{"hit": 5}
		if hc := HitChance(c, nil, Attack{By: f, Skill: plain(1), Target: h}); hc != base-15 {
			t.Fatalf("hit %d, base %d", hc, base)
		}
	})

	t.Run("Crit chance from gear can make a hit critical, and crit damage gear makes it hit harder", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 200, 200
		h.Affix = map[string]int{"crit": 50, "critDmg": 50}
		r := Hurt(c, b, rolls(0, 0.5, 0.1), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if !r.Crit || r.Damage != 30 {
			t.Fatalf("crit %v damage %d, want a crit for 30 (15 x 2)", r.Crit, r.Damage)
		}
	})

	t.Run("Life steal heals the attacker for its share of the damage", func(t *testing.T) {
		c, b, h, f := arena('g')
		h.HP = 10
		h.Affix = map[string]int{"leech": 20}
		r := Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if h.HP != 10+int(Round(float64(r.Damage)*0.2)) {
			t.Fatalf("HP %d after dealing %d", h.HP, r.Damage)
		}
	})

	t.Run("Thorns hurt the attacker for a share of its damage, but never kill it", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.Affix = map[string]int{"thorns": 100}
		h.HP = 5
		Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if h.HP != 1 {
			t.Fatalf("attacker HP %d", h.HP)
		}
	})

	t.Run("Added fire damage hits once more for its share, as fire", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 200, 200
		h.Affix = map[string]int{"addFire": 20}
		var extra int
		Hurt(c, b, rolls(0, 0.5), Attack{By: h, Skill: plain(1), Target: f}, func(e action.Event) {
			if e.Kind == action.Hit && e.Skill == "fire" {
				extra = e.Amount
			}
		})
		if extra != 3 || f.HP != 200-15-3 {
			t.Fatalf("extra %d, HP %d", extra, f.HP)
		}
	})

	t.Run("Gear can bleed, poison or stun on hit", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 200, 200
		h.Affix = map[string]int{"bleedCh": 50, "poisonCh": 50, "stunCh": 50}
		Hurt(c, b, rolls(0, 0.5, 0.1, 0.9, 0.1), Attack{By: h, Skill: plain(1), Target: f}, nop)
		if f.Status["bleed"] != 3 || f.Status["poison"] != 0 || f.Status["stun"] != 1 {
			t.Fatalf("statuses %v", f.Status)
		}
	})

	t.Run("Strike twice: gear can hit the main target once more for 60%", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 200, 200
		h.Affix = map[string]int{"double": 30}
		h.Skills, h.Cooldowns = []content.Skill{plain(1)}, []int{0}
		var hits []int
		_ = UseSkill(c, b, rolls(0, 0.5, 0.1, 0, 0.5), h, 0, f, func(e action.Event) {
			if e.Kind == action.Hit {
				hits = append(hits, e.Amount)
			}
		})
		if len(hits) != 2 || hits[1] != 9 {
			t.Fatalf("hits %v, want 15 then 9", hits)
		}
	})

	t.Run("Cooldown reduction gear shortens a skill's cooldown, never below 1 turn", func(t *testing.T) {
		c, b, h, f := arena('g')
		f.HP, f.MaxHP = 200, 200
		h.Affix = map[string]int{"cdr": 5}
		s := plain(1)
		s.Cooldown = 3
		h.Skills, h.Cooldowns = []content.Skill{s}, []int{0}
		_ = UseSkill(c, b, hits(), h, 0, f, nop)
		if h.Cooldowns[0] != 1 {
			t.Fatalf("cooldown %d", h.Cooldowns[0])
		}
	})

	t.Run("A staff or wand in hand adds half its attack to heals, and healing gear heals more", func(t *testing.T) {
		c, _, h, _ := arena('g')
		h.Class, h.Level = "Healer", 1
		heal := c.Skills["Heal Ally"]
		bare := HealAmount(c, h, heal)
		arm(c, h, map[string]content.Item{"w": weapon("Oak Staff", "staff", 0, 10)})
		h.Affix["healPct"] = 50
		want := 5 + c.Gear.RankDamage[stats.SetRank(c, h)]*1.5*bare
		if got := HealAmount(c, h, heal); got < want-1e-9 || got > want+1e-9 {
			t.Fatalf("heal %.3f, want %.3f (bare %.3f)", got, want, bare)
		}
	})
}
