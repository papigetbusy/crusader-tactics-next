package oldgame

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/engine"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// goUnit builds a referee unit from a recorded one, with its skill bar.
func goUnit(c *content.Content, su SnapUnit) *state.Unit {
	u := &state.Unit{
		ID: su.ID, Team: su.Team, Class: su.Class, X: su.X, Y: su.Y,
		HP: su.HP, MaxHP: su.MaxHP, Atk: su.Atk, Def: su.Def, Mov: su.Mov, Level: su.Lv,
		Element: su.Elem, Boss: su.Boss != 0, Elite: su.Elite != 0, Defending: su.Def2 != 0, Shield: su.Sh,
		Moved: su.Moved != 0, Acted: su.Acted != 0, Done: su.Done != 0, Status: map[string]int{},
		AbilityCooldown: su.Ab, Decoy: su.Dc != 0, Companion: su.Cp != 0, Summoned: su.Sm != 0,
		WeaponSet: su.Ws,
	}
	// In these recordings every companion and raised monster is undead (Grave
	// Bloom skeletons and Raise skills); bonded companions join later floors.
	u.Undead = u.Companion || u.Summoned
	if su.Sp != nil {
		u.Ascension = *su.Sp
		u.RangeBonus = c.Ascensions[u.Ascension].Range
	}
	if su.M != nil && *su.M < len(c.Monsters) {
		u.Kind = c.Monsters[*su.M].Name
	}
	u.Ally = su.Al != 0
	if su.Wt != nil {
		u.Wanted, u.WantedBy = true, *su.Wt
	}
	if len(su.Ow) == 2 {
		u.Overwatch = &state.Overwatch{DX: su.Ow[0], DY: su.Ow[1]}
	}
	for k, v := range su.St {
		u.Status[k] = v
	}
	switch {
	case u.Decoy:
	case u.Team == state.Heroes && !u.Companion && !u.Summoned:
		u.Learned, u.Equipped, u.Ultimate = c.StartingBuild(u.Class, u.Level)
		u.Skills = c.Bar(u.Class, u.Equipped, u.Ultimate)
		if su.Sk != nil {
			u.Skills = nil
			for _, n := range su.Sk {
				if s, ok := c.SkillNamed(n); ok {
					u.Skills = append(u.Skills, s)
				}
			}
		}
	default:
		u.Skills = c.MonsterSkills(u.Class, u.Level, u.Boss || u.Elite)
		if u.Team == state.Enemies {
			u.Weapon = c.Classes[u.Class].Weapon
		}
	}
	if su.Pot != nil {
		u.Potions = *su.Pot
	}
	u.Cooldowns = make([]int, len(u.Skills))
	for i, cd := range su.Cd {
		if cd != nil && i < len(u.Cooldowns) {
			u.Cooldowns[i] = *cd
		}
	}
	return u
}

// resync makes Go's board equal to a recorded moment, keeping what the
// recording does not store (tiles walked this turn, skill uses, the numbers
// active abilities set, who owns a clone or a pet).
func resync(c *content.Content, f *Fight, g *engine.Game, s Snap, party []int) {
	old := map[int]*state.Unit{}
	b := f.Battle(Snap{Turn: s.Turn})
	if g.B != nil {
		for _, u := range g.B.Units {
			old[u.ID] = u
		}
		b.Side, b.Outcome = g.B.Side, g.B.Outcome
		b.FloorKills = g.B.FloorKills
	}
	for _, su := range s.Units {
		u := goUnit(c, su)
		if su.Dc != 0 {
			// The recorder gave shadow clones their Phantom's ID; Go gives
			// them their own.
			u.ID, u.Owner = 1_000_000+len(b.Units), su.ID
			b.Units = append(b.Units, u)
			continue
		}
		if o := old[su.ID]; o != nil {
			u.TilesMoved, u.SkillUses, u.LastSkill, u.Undo = o.TilesMoved, o.SkillUses, o.LastSkill, o.Undo
			u.Params, u.Owner, u.QuickShots, u.QuickUsed = o.Params, o.Owner, o.QuickShots, o.QuickUsed
			u.HeavyBleed, u.HeavyBleedBy, u.UnbrokenUsed, u.Danced, u.SkyReady = o.HeavyBleed, o.HeavyBleedBy, o.UnbrokenUsed, o.Danced, o.SkyReady
			if u.Overwatch != nil && o.Overwatch != nil {
				u.Overwatch.Hit = o.Overwatch.Hit
			}
		}
		b.Units = append(b.Units, u)
	}
	if len(s.Seized) == 2 {
		b.Throne = &state.Tile{X: s.Seized[0], Y: s.Seized[1]}
	}
	if g.B != nil {
		fallen := map[int]*state.Unit{}
		for _, u := range g.B.Fallen {
			fallen[u.ID] = u
		}
		for _, id := range s.Fallen {
			if u := fallen[id]; u != nil {
				b.Fallen = append(b.Fallen, u)
			}
		}
	}
	for _, fa := range s.FieldsAt {
		num := func(i int) int { v, _ := fa[i].(float64); return int(v) }
		kind, _ := fa[4].(string)
		b.Fields = append(b.Fields, state.Field{X: num(0), Y: num(1), Range: num(2), Turns: num(3), Kind: kind, Square: num(5) != 0})
	}
	for _, t := range s.TrapsAt {
		b.Traps = append(b.Traps, state.Trap{X: t[0], Y: t[1]})
	}
	b.LevelCap = levelCap(c, f.Init.Snap)
	g.B = b
	for i, id := range party {
		u := heroOf(b, id)
		if u == nil {
			continue
		}
		if i < len(s.XP) {
			u.XP = s.XP[i][1]
		}
		if i < len(s.Party) {
			p := s.Party[i]
			u.Learned, u.Equipped = p.Learned, p.Eq
			u.Ultimate, u.Ascension = deref(p.Ult), deref(p.Spec)
			if len(p.Su) > 0 && u.SkillUses == nil {
				u.SkillUses = p.Su
			}
			if p.Gear != nil {
				u.Gear = map[string]content.Item{}
				for sl, it := range p.Gear {
					u.Gear[sl] = it.Item()
				}
			}
		}
		u.BaseAtk = u.Atk - stats.SetAtk(c, u.Class, u.Gear, u.WeaponSet)
		stats.InitSets(c, u)
	}
	b.Bag = bagOf(s)
}

// bagOf is the bag at a recorded moment. Loot drops are ported in 6d; until
// then the bag is taken from the recording.
func bagOf(s Snap) []content.Item {
	var out []content.Item
	for _, it := range s.Inv {
		out = append(out, it.Item())
	}
	return out
}

// actionsFor turns one recorded player input into referee actions.
func actionsFor(step Step, b *state.Battle) []action.Action {
	switch step.Kind {
	case "wait":
		return []action.Action{{Kind: action.Wait, Unit: step.ID}}
	case "endturn":
		return []action.Action{{Kind: action.EndTurn}}
	case "defend":
		return []action.Action{{Kind: action.Defend, Unit: step.ID}}
	case "potion":
		return []action.Action{{Kind: action.Potion, Unit: step.ID}}
	case "undo":
		return []action.Action{{Kind: action.Undo, Unit: step.ID}}
	case "ability":
		a := action.Action{Kind: action.Ability, Unit: step.ID}
		var one int
		var tiles [][2]int
		if json.Unmarshal(step.V, &one) == nil {
			a.Target = one
		} else if json.Unmarshal(step.V, &tiles) == nil {
			a.Tiles = tiles
		}
		return []action.Action{a}
	case "equip":
		if step.Item == nil {
			return nil
		}
		want := step.Item.Item()
		a := action.Action{Kind: action.Equip, Unit: step.ID, Item: -1}
		for i, it := range b.Bag {
			if reflect.DeepEqual(it, want) {
				a.Item = i
				break
			}
		}
		if want.Slot == content.SlotWeapon {
			a.Slot = step.Slot
		}
		return []action.Action{a}
	case "move", "attack", "skill":
		var out []action.Action
		for _, e := range step.Ev {
			if e.K == "mv" && e.ID == step.ID {
				out = append(out, action.Action{Kind: action.Move, Unit: e.ID, X: e.To[0], Y: e.To[1]})
			}
			if e.K == "act" && e.ID == step.ID {
				u := b.Unit(e.ID)
				slot := -1
				if u != nil {
					slot = slices.IndexFunc(u.Skills, func(s content.Skill) bool { return s.Name == e.S })
				}
				a := action.Action{Kind: action.Attack, Unit: e.ID, Skill: slot, Target: e.V}
				if step.Ws != nil {
					a.Set = *step.Ws + 1
				}
				out = append(out, a)
			}
		}
		return out
	}
	return nil
}

// decision turns a recorded popup answer into Go's Decide action, reading
// the player's choice from the button pressed (and, for an ascension, from
// the party's builds right after). ok is false when Go has nothing waiting
// that matches (the old game also asked about things ported later, such as
// bonding a companion).
func decision(c *content.Content, g *engine.Game, step Step, party []int) (action.Action, bool) {
	if len(g.B.Decisions) == 0 {
		return action.Action{}, false
	}
	d := g.B.Decisions[0]
	a := action.Action{Kind: action.Decide, Choice: stats.ChooseOK}
	switch d.Kind {
	case state.DecideAscend:
		if i := slices.Index(party, d.Unit); i >= 0 && i < len(step.Snap.Party) {
			a.Choice = deref(step.Snap.Party[i].Spec)
		}
	case state.DecideLearn:
		switch {
		case strings.HasPrefix(step.Choice, "Keep"):
			a.Choice = stats.ChooseKeep
		case strings.HasPrefix(step.Choice, "Replace "):
			a.Choice = stats.ChooseReplace
			if u := heroOf(g.B, d.Unit); u != nil {
				a.Skill = max(0, slices.Index(u.Equipped, strings.TrimPrefix(step.Choice, "Replace ")))
			}
		}
	}
	return a, true
}

// arrivals adds reinforcements before the enemy phase. Where they appear is a
// floor rule ported later, so the recording says where.
func arrivals(c *content.Content, g *engine.Game, prev Snap, step Step) int {
	had := map[int]bool{}
	for _, u := range prev.Units {
		had[u.ID] = true
	}
	n := 0
	for _, su := range step.Snap.Units {
		if had[su.ID] || g.B.Unit(su.ID) != nil || su.Team != state.Enemies {
			continue
		}
		u := goUnit(c, su)
		u.HP, u.Status, u.Done, u.Moved = su.MaxHP, map[string]int{}, false, false
		for i := range u.Cooldowns {
			u.Cooldowns[i] = 0
		}
		for _, e := range step.Ev {
			if e.K == "mv" && e.ID == su.ID {
				u.X, u.Y = e.From[0], e.From[1]
				break
			}
		}
		g.B.Units = append(g.B.Units, u)
		n++
	}
	return n
}
