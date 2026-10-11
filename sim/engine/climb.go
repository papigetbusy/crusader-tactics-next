package engine

import (
	"slices"

	"github.com/papigetbusy/crusader-tactics-next/sim/action"
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/rng"
	"github.com/papigetbusy/crusader-tactics-next/sim/run"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

// Phases of the climb.
const (
	PhasePrepare = "prepare" // looking at the map: choose and place the team
	PhaseBattle  = "battle"
	PhaseWon     = "won"  // the floor-clear summary
	PhaseLost    = "lost" // defeated: the floor is fought again
	PhaseShop    = "shop" // the shop before every 7th floor
)

// climb is the state of a climb around the battle.
type climb struct {
	run         *run.Run
	phase       string
	fl          floor.Floor
	team        []string     // chosen hero classes
	placed      []run.Placed // where they start
	startChests []state.Chest
	before      map[string][2]int // level and total XP of each hero when the battle started
	goldBefore  int
	bagBefore   int
	summary     *Summary
	battleSeed  uint64
	battles     uint64
}

// Summary is the floor-clear (or defeat) summary (old showSummary).
type Summary struct {
	Floor    int           `json:"floor"`
	Won      bool          `json:"won"`
	Replayed bool          `json:"replayed,omitempty"` // a floor cleared before: half gold
	NewBest  bool          `json:"newBest,omitempty"`
	Heroes   []HeroSummary `json:"heroes"`
	Loot     []ItemView    `json:"loot,omitempty"`
	Gold     int           `json:"gold"`
	Gift     *ItemView     `json:"gift,omitempty"` // the legendary gift every 100 floors
	Turns    int           `json:"turns"`
}

// HeroSummary is one hero's line in the summary.
type HeroSummary struct {
	Class  string `json:"class"`
	From   int    `json:"from"`
	To     int    `json:"to"`
	XP     int    `json:"xp"`
	Capped bool   `json:"capped,omitempty"`
}

// NewClimb starts a climb: all ten heroes at level 1, floor 1, looking at
// the map. The run seed decides every floor; the battle seed the dice.
func NewClimb(c *content.Content, runSeed uint32, battleSeed uint64) *Game {
	g := &Game{C: c, climb: &climb{run: run.New(c, runSeed), battleSeed: battleSeed}}
	g.prepare()
	return g
}

// prepare builds the current floor and shows it with the team placed.
func (g *Game) prepare() {
	cl := g.climb
	cl.fl = cl.run.Build(g.C, cl.run.Floor)
	if len(cl.team) != cl.run.Slots() {
		cl.team = cl.run.Strongest()
	}
	cl.placed = spread(cl.fl, cl.team)
	cl.phase = PhasePrepare
	g.showFloor()
}

// showFloor rebuilds the battle shown while preparing (call keepPreview
// first to keep changes made on the old one).
func (g *Game) showFloor() {
	cl := g.climb
	cl.fl = cl.run.Build(g.C, cl.run.Floor)
	g.B = cl.run.Battle(g.C, cl.fl, cl.placed)
	cl.startChests = slices.Clone(g.B.Chests)
	cl.battles++
	g.Dice = rng.New(cl.battleSeed^cl.battles, streamCombat)
}

// spread puts heroes on starting tiles spread along the bottom row (the
// old game's spots), or the first free starting tiles.
func spread(fl floor.Floor, team []string) []run.Placed {
	var out []run.Placed
	used := map[state.Tile]bool{}
	n := fl.N
	for i, cl := range team {
		t := state.Tile{X: min(n-1, int((float64(i)+.5)*float64(n)/float64(len(team)))), Y: n - 1}
		if !slices.Contains(fl.Starts, t) || used[t] {
			for _, s := range fl.Starts {
				if !used[s] {
					t = s
					break
				}
			}
		}
		used[t] = true
		out = append(out, run.Placed{Class: cl, Tile: t})
	}
	return out
}

// climbAction handles the actions between battles.
func (g *Game) climbAction(a action.Action, emit action.Emit) error {
	cl, r := g.climb, g.climb.run
	switch a.Kind {
	case action.Team:
		if cl.phase != PhasePrepare {
			return no("not-now", "Choose the team while looking at the map.")
		}
		if len(a.Heroes) != r.Slots() {
			return no("slots", "Fill every slot: choose %d heroes.", r.Slots())
		}
		for i, h := range a.Heroes {
			if r.Hero(h) == nil || slices.Contains(a.Heroes[:i], h) {
				return no("bad-team", "Choose %d different heroes.", r.Slots())
			}
		}
		g.keepPreview()
		cl.team = slices.Clone(a.Heroes)
		cl.placed = spread(cl.fl, cl.team)
		g.showFloor()
	case action.Place:
		if cl.phase != PhasePrepare {
			return no("not-now", "Heroes are placed before the battle.")
		}
		i := slices.IndexFunc(cl.placed, func(p run.Placed) bool { return p.Class == a.Hero })
		t := state.Tile{X: a.X, Y: a.Y}
		if i < 0 {
			return no("not-in-team", "%s is not in the team.", a.Hero)
		}
		if !slices.Contains(cl.fl.Starts, t) {
			return no("not-start", "Heroes start on the blue tiles.")
		}
		g.keepPreview()
		if j := slices.IndexFunc(cl.placed, func(p run.Placed) bool { return p.Tile == t }); j >= 0 {
			cl.placed[j].Tile = cl.placed[i].Tile
		}
		cl.placed[i].Tile = t
		g.showFloor()
	case action.Start:
		if cl.phase != PhasePrepare {
			return no("not-now", "There is no battle to start.")
		}
		g.keepPreview()
		g.showFloor()
		cl.phase = PhaseBattle
		cl.before, cl.goldBefore, cl.bagBefore = map[string][2]int{}, g.B.Gold, len(g.B.Bag)
		for _, id := range g.B.Party {
			u := g.B.Unit(id)
			cl.before[u.Class] = [2]int{u.Level, totalXP(g.C, u.Level, u.XP)}
		}
	case action.Continue:
		switch cl.phase {
		case PhaseWon, PhaseLost:
			g.next()
		case PhaseShop:
			r.ShopVisit = r.Floor
			g.prepare()
		default:
			return no("not-now", "Finish the battle first.")
		}
	case action.Travel:
		if cl.phase == PhaseBattle {
			r.Keep(g.B, cl.startChests) // leaving keeps what was earned
		} else {
			g.keepPreview()
		}
		if err := r.Travel(a.Floor); err != nil {
			return no("no-floor", "%v", err)
		}
		g.next()
	case action.Buy, action.Sell, action.Combine:
		if cl.phase != PhaseShop {
			return no("no-shop", "The shop is only open before every 7th floor.")
		}
		var err error
		switch a.Kind {
		case action.Buy:
			err = r.Buy(g.C, r.Floor, a.Item)
		case action.Sell:
			err = r.Sell(g.C, a.Item)
		default:
			var it content.Item
			if it, err = r.Combine(g.C, a.Item, a.Second); err == nil {
				emit(action.Event{Kind: action.Combined, Skill: it.Name, Amount: it.Rarity})
			}
		}
		if err != nil {
			return no("shop", "%v", err)
		}
	case action.Ascend:
		if cl.phase == PhaseBattle {
			return no("not-now", "Ascension can't change in the middle of a battle.")
		}
		g.keepPreview()
		if err := r.ChangeAscension(g.C, a.Hero, a.Choice); err != nil {
			return no("ascension", "%v", err)
		}
		if cl.phase == PhasePrepare {
			g.showFloor()
		}
	}
	return nil
}

// keepPreview keeps changes made while preparing (gear, skills).
func (g *Game) keepPreview() {
	if g.climb.phase == PhasePrepare && g.B != nil {
		g.climb.run.Keep(g.B, nil)
	}
}

// next goes on after a battle or a trip: the shop first on every 7th floor,
// else the map.
func (g *Game) next() {
	if g.climb.run.ShopFirst(g.C) {
		g.climb.run.Stock(g.C, g.climb.run.Floor)
		g.climb.phase = PhaseShop
		return
	}
	g.prepare()
}

// finish ends a battle of the climb once its last choice is answered.
func (g *Game) finish() {
	cl, b := g.climb, g.B
	if cl.phase != PhaseBattle || b.Outcome == state.Ongoing || len(b.Decisions) > 0 {
		return
	}
	s := &Summary{Floor: b.Floor, Won: b.Outcome == state.Won, Replayed: b.Replayed, Turns: b.Turn}
	for _, id := range b.Party {
		u := heroOfParty(b, id)
		if u == nil {
			continue
		}
		bf := cl.before[u.Class]
		s.Heroes = append(s.Heroes, HeroSummary{Class: u.Class, From: bf[0], To: u.Level,
			XP: totalXP(g.C, u.Level, u.XP) - bf[1], Capped: u.Level >= b.LevelCap})
	}
	for _, it := range b.Bag[min(cl.bagBefore, len(b.Bag)):] {
		s.Loot = append(s.Loot, g.itemView(it.Slot, it))
	}
	s.Gold = b.Gold - cl.goldBefore
	if s.Won {
		best := cl.run.Best
		if gift := cl.run.Won(g.C, b, cl.startChests); gift != nil {
			iv := g.itemView(gift.Slot, *gift)
			s.Gift = &iv
		}
		s.NewBest = cl.run.Best > best
		cl.phase = PhaseWon
	} else {
		cl.run.Lost(b, cl.startChests)
		cl.phase = PhaseLost
	}
	cl.summary = s
}

// totalXP is all the XP a hero has earned to reach a level.
func totalXP(c *content.Content, level, xp int) int {
	t := xp
	for l := 1; l < level; l++ {
		t += stats.XPToNext(c, l)
	}
	return t
}

func heroOfParty(b *state.Battle, id int) *state.Unit {
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
