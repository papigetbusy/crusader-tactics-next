package engine

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/items"
	"github.com/papigetbusy/crusader-tactics-next/sim/run"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// ClimbView is the climb as the screen shows it.
type ClimbView struct {
	Phase   string       `json:"phase"`
	Floor   int          `json:"floor"`
	Best    int          `json:"best"`
	Slots   int          `json:"slots"`
	Gold    int          `json:"gold"`
	Team    []string     `json:"team"`
	Placed  []run.Placed `json:"placed"`
	Heroes  []HeroRow    `json:"heroes"`
	Floors  []FloorRow   `json:"floors"`
	Shop    []PricedItem `json:"shop,omitempty"`
	Bag     []PricedItem `json:"bag,omitempty"`   // in the shop: what selling pays
	Pairs   [][2]int     `json:"pairs,omitempty"` // in the shop: bag items that can be combined, two by two
	Summary *Summary     `json:"summary,omitempty"`
}

// HeroRow is one hero of the roster.
type HeroRow struct {
	Class      string   `json:"class"`
	Form       string   `json:"form"`
	Level      int      `json:"level"`
	XP         int      `json:"xp"`
	XPNext     int      `json:"xpNext"`
	Ascension  string   `json:"ascension,omitempty"`
	InTeam     bool     `json:"inTeam,omitempty"`
	CanAscend  []string `json:"canAscend,omitempty"` // other ascensions it may switch to now
	SkillsUsed bool     `json:"skillsUsed,omitempty"`
	// For the team picker: the hero's picture stage, numbers with its gear,
	// what it can use, and its class ability.
	Stage       int    `json:"stage,omitempty"`
	HP          int    `json:"hp"`
	Atk         int    `json:"atk"`
	Def         int    `json:"def"`
	Mov         int    `json:"mov"`
	Weapons     string `json:"weapons"`
	Ability     string `json:"ability,omitempty"`
	AbilityText string `json:"abilityText,omitempty"`
}

// FloorRow is one line of the floor travel list.
type FloorRow struct {
	Number        int    `json:"number"`
	Biome         string `json:"biome,omitempty"`
	Here          bool   `json:"here,omitempty"`
	Top           bool   `json:"top,omitempty"`
	Cleared       bool   `json:"cleared,omitempty"`
	Boss          bool   `json:"boss,omitempty"`
	BossLegendary bool   `json:"bossLegendary,omitempty"`
	Elite         bool   `json:"elite,omitempty"`
	Shop          bool   `json:"shop,omitempty"`
	ChestsLeft    int    `json:"chestsLeft,omitempty"`
}

// PricedItem is an item with a price.
type PricedItem struct {
	ItemView
	Price int `json:"price"`
}

func (g *Game) climbView() *ClimbView {
	cl, r := g.climb, g.climb.run
	v := &ClimbView{Phase: cl.phase, Floor: r.Floor, Best: r.Best, Slots: r.Slots(), Gold: r.Gold,
		Team: cl.team, Placed: cl.placed, Summary: cl.summary}
	if cl.phase == PhaseBattle || cl.phase == PhasePrepare {
		v.Gold = g.B.Gold
	}
	for _, h := range r.Heroes {
		row := HeroRow{Class: h.Class, Form: stats.FormName(g.C, h.Class, h.Level), Level: h.Level, XP: h.XP,
			XPNext: stats.XPToNext(g.C, h.Level), Ascension: h.Ascension, InTeam: slices.Contains(cl.team, h.Class),
			SkillsUsed: h.SkillsVisit == r.Visit, Stage: stats.Form(g.C, h.Level, h.Ascension), Weapons: g.weaponsText(h.Class, h.Level)}
		n, _ := stats.Geared(g.C, h.Class, h.Level, h.Ascension, h.Gear)
		row.HP, row.Atk, row.Def, row.Mov = n.MaxHP, n.Atk, n.Def, n.Mov
		if ab, ok := g.C.AbilityFor(h.Class, h.Level, h.Ascension, false); ok {
			row.Ability, row.AbilityText = ab.Name, ab.Text
		}
		if h.Ascension != "" && h.AscensionVisit != r.Visit && cl.phase != PhaseBattle {
			for _, a := range g.C.AscensionChoices[h.Class] {
				if a != h.Ascension {
					row.CanAscend = append(row.CanAscend, a)
				}
			}
		}
		v.Heroes = append(v.Heroes, row)
	}
	for f := r.Best; f >= 1; f-- {
		m := r.Floors[f]
		row := FloorRow{Number: f, Here: f == r.Floor, Top: f == r.Best}
		boss := g.C.Floors.BossEvery > 0 && f%g.C.Floors.BossEvery == 0
		row.Boss, row.Elite = boss, !boss && f%g.C.Floors.EliteEvery == 0
		row.Shop = g.C.Floors.ShopBefore > 0 && f%g.C.Floors.ShopBefore == 0
		if m != nil {
			row.Biome, row.Cleared, row.BossLegendary = m.Biome, m.Cleared, m.BossLegendary
			row.ChestsLeft = max(0, m.Chests-len(m.Opened))
		}
		v.Floors = append(v.Floors, row)
	}
	if cl.phase == PhaseShop {
		for _, it := range r.Stock(g.C, r.Floor) {
			v.Shop = append(v.Shop, PricedItem{g.itemView(it.Slot, it), items.Price(g.C, it)})
		}
		for _, it := range r.Bag {
			v.Bag = append(v.Bag, PricedItem{g.itemView(it.Slot, it), items.SellPrice(g.C, it)})
		}
		v.Pairs = r.Pairs(g.C)
	}
	return v
}
