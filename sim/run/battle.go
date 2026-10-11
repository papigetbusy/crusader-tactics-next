package run

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Placed is a hero on a starting tile.
type Placed struct {
	Class string     `json:"class"`
	Tile  state.Tile `json:"tile"`
}

// startingPotions is what every hero carries into a floor (old newFloor0,
// line 743: potions refill to 3).
const startingPotions = 3

// Battle sets up the fight on a built floor with the heroes placed: full
// health, potions refilled, the floor's monsters, chests and memory.
func (r *Run) Battle(c *content.Content, fl floor.Floor, team []Placed) *state.Battle {
	m := r.Memory(fl.Number)
	b := &state.Battle{N: fl.N, Ground: fl.Ground, Heights: fl.Heights, Turn: 1, Side: state.Heroes,
		Floor: fl.Number, Reinforce: fl.Reinforce, Biome: fl.Biome, Starts: fl.Starts, Throne: fl.Throne,
		Gold: r.Gold, Bag: slices.Clone(r.Bag), Chests: slices.Clone(fl.Chests),
		BossLegendary: m.BossLegendary, Replayed: m.Cleared}
	for _, p := range team {
		h := r.Hero(p.Class)
		if h == nil {
			continue
		}
		b.NextID++
		u := &state.Unit{ID: b.NextID, Team: state.Heroes, Class: h.Class, Level: h.Level, XP: h.XP,
			X: p.Tile.X, Y: p.Tile.Y, Learned: slices.Clone(h.Learned), Equipped: slices.Clone(h.Equipped),
			Ultimate: h.Ultimate, Ascension: h.Ascension, Souls: slices.Clone(h.Souls), Bond: h.Bond,
			SkillUses: map[string]int{}, Gear: map[string]content.Item{}, Potions: startingPotions,
			Swapped: h.SkillsVisit == r.Visit, Status: map[string]int{}}
		for k, v := range h.SkillUses {
			u.SkillUses[k] = v
		}
		for k, v := range h.Gear {
			u.Gear[k] = v
		}
		b.Units = append(b.Units, u)
		stats.Refresh(c, b, u)
		if h.WeaponSet > 0 && h.WeaponSet < len(stats.SetsOf(c, u.Class)) {
			stats.UseSet(c, u, h.WeaponSet)
		}
		u.HP = u.MaxHP
		b.Party = append(b.Party, u.ID)
	}
	for _, e := range fl.Enemies {
		b.NextID++
		e.ID = b.NextID
		b.Units = append(b.Units, e)
	}
	b.NextID++
	for _, id := range b.Party {
		combat.SpawnCompanion(c, b, b.Unit(id))
	}
	b.LevelCap = r.LevelCap(c)
	return b
}

// Keep brings a finished (or left) battle back into the run: what each hero
// learned and earned, gold, the bag, and the chests opened. Win or lose,
// progress is kept (old: heroes, gold and loot persist; a defeat re-runs the
// floor).
func (r *Run) Keep(b *state.Battle, startChests []state.Chest) {
	for _, id := range b.Party {
		u := partyUnit(b, id)
		if u == nil {
			continue
		}
		h := r.Hero(u.Class)
		h.Level, h.XP, h.Ultimate, h.Ascension, h.Bond = u.Level, u.XP, u.Ultimate, u.Ascension, u.Bond
		h.Learned, h.Equipped, h.Souls = slices.Clone(u.Learned), slices.Clone(u.Equipped), slices.Clone(u.Souls)
		h.SkillUses, h.Gear, h.WeaponSet = map[string]int{}, map[string]content.Item{}, u.WeaponSet
		for k, v := range u.SkillUses {
			h.SkillUses[k] = v
		}
		for k, v := range u.Gear {
			h.Gear[k] = v
		}
		if u.Swapped {
			h.SkillsVisit = r.Visit
		}
	}
	r.Gold, r.Bag = b.Gold, slices.Clone(b.Bag)
	m := r.Memory(b.Floor)
	for _, ch := range startChests {
		if !slices.Contains(b.Chests, ch) {
			m.Opened = append(m.Opened, state.Tile{X: ch.X, Y: ch.Y})
		}
	}
	m.BossLegendary = m.BossLegendary || b.BossLegendary
}

func partyUnit(b *state.Battle, id int) *state.Unit {
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
