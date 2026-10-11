package oldgame

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/ai"
	"github.com/papigetbusy/crusader-tactics-next/sim/combat"
	"github.com/papigetbusy/crusader-tactics-next/sim/items"
)

// whereRolled maps each kind of Go roll to the old game's code line that
// made the same roll.
var whereRolled = map[string]string{
	combat.RollHit:          "hurt:1689",
	combat.RollDamage:       "hurt:1690",
	combat.RollCrit:         "hurt:1690",
	combat.RollBeastBurn:    "hurt:1702",
	combat.RollStatusChance: "hurt:1702",
	combat.RollDouble:       "act:1722",
	ai.RollPick:             "aiTurn:",
	ai.RollSpecial:          "aiTurn:",
	combat.RollShuffle:      "act:",
	items.RollDrop:          "drop:",
	items.RollBand:          "bandRoll:",
	items.RollOwner:         "lootFor:",
	items.RollSlot:          "lootPick:",
	items.RollItem:          "genItem:",
	items.RollBonus:         "rollAffixes:",
	items.RollChest:         "openChest:",
}

// notPortedYet are old-game rolls for rules that arrive later (loot, chests,
// where reinforcements appear). The feed steps over them.
var notPortedYet = []string{
	"anon:1203", "file:1203", // the old biased sort of an item's bonuses; the picks are served by PickAffixes
	"reinforce:",
	"file:1721", "act:1721"} // the old multi-target shuffle, replaced by the approved fair shuffle

func (sf *siteFeed) skippable(site string) bool {
	if sf.noChests {
		for _, p := range lootSites {
			if strings.HasPrefix(site, p) {
				return true
			}
		}
	}
	for _, p := range notPortedYet {
		if strings.HasPrefix(site, p) {
			return true
		}
	}
	return false
}

// siteFeed serves one step's recorded dice to Go. Each Go roll must match the
// old game's next roll of the same kind; rolls for rules not ported yet are
// stepped over. Anything else is a mismatch, reported in plain words.
type siteFeed struct {
	step    *Step
	newLoot [][]string // the bonuses of each new non-legendary item in this step, in order (nil: not recorded)
	// noChests: a step where a hero opened a chest in a recording that did
	// not store where chests were; that step's loot rolls are stepped over.
	noChests bool
	served   int
	picks    int // multi-target picks served so far
	vals     []uint32
	sites    []string
	pos      int
	used     int
	skipped  int
}

func newSiteFeed(f *Fight, s *Step) *siteFeed {
	sf := &siteFeed{vals: s.Dice.G, step: s}
	for _, i := range s.Dice.S {
		sf.sites = append(sf.sites, f.Sites[i])
	}
	return sf
}

func (sf *siteFeed) Roll(label string) float64 {
	want, ok := whereRolled[label]
	if !ok {
		panic(fmt.Sprintf("Go made a roll the checker does not know (%q)", label))
	}
	for sf.pos < len(sf.vals) {
		site := sf.sites[sf.pos]
		if strings.HasPrefix(site, want) {
			v := sf.vals[sf.pos]
			sf.pos++
			sf.used++
			return float64(v) / 4294967296.0
		}
		if !sf.skippable(site) {
			panic(fmt.Sprintf("Go rolled a %q, but the old game's next roll was at %s", label, site))
		}
		sf.pos++
		sf.skipped++
	}
	panic(fmt.Sprintf("Go rolled a %q, but the old game made no more rolls in this step", label))
}

// leftover reports old-game rolls Go never made.
func (sf *siteFeed) leftover() string {
	for ; sf.pos < len(sf.vals); sf.pos++ {
		if !sf.skippable(sf.sites[sf.pos]) {
			return sf.sites[sf.pos]
		}
		sf.skipped++
	}
	return ""
}

// PickExtra serves the extra targets the old game picked for the n-th
// multi-target skill use in this step: the units it hit after the main
// target, in the order it hit them.
func (sf *siteFeed) PickExtra(pool []int, n int) []int {
	inPool := map[int]bool{}
	for _, id := range pool {
		inPool[id] = true
	}
	acts := 0
	for i, e := range sf.step.Ev {
		if e.K != "act" {
			continue
		}
		var hit []int
		seen := map[int]bool{e.V: true}
		for _, h := range sf.step.Ev[i+1:] {
			if h.K == "act" {
				break
			}
			if h.K == "hit" && h.B != nil && !seen[h.B.ID] && inPool[h.B.ID] {
				seen[h.B.ID] = true
				hit = append(hit, h.B.ID)
			}
		}
		if len(hit) == 0 {
			continue
		}
		if acts == sf.picks {
			sf.picks++
			return hit[:min(n, len(hit))]
		}
		acts++
	}
	sf.picks++
	return nil
}

// PickAffixes serves the bonuses the old game gave the next new item in
// this step (read from the bag before and after). Older recordings did not
// store the bag; then any pick will do, since the bag is not compared.
func (sf *siteFeed) PickAffixes(cands []string, n int) ([]string, bool) {
	if sf.newLoot == nil {
		return cands[:n], true
	}
	if sf.served >= len(sf.newLoot) {
		return nil, false
	}
	sf.served++
	return sf.newLoot[sf.served-1], true
}

// newItems are the items in the bag after a step that were not there before,
// in bag order: the bonus keys of each one that is not a legendary.
func newItems(prev, now []RecItem) [][]string {
	had := map[string]int{}
	key := func(it RecItem) string { b, _ := json.Marshal(it); return string(b) }
	for _, it := range prev {
		had[key(it)]++
	}
	out := [][]string{}
	for _, it := range now {
		if k := key(it); had[k] > 0 {
			had[k]--
			continue
		}
		if it.R >= 3 {
			continue
		}
		keys := []string{}
		for _, a := range it.Af {
			keys = append(keys, a[0].(string))
		}
		out = append(out, keys)
	}
	return out
}

// lootSites are the old game's loot rolls.
var lootSites = []string{"openChest:", "lootFor:", "lootPick:", "genItem:", "rollAffixes:"}

// opensChest reports a recorded step in which the old game opened a chest.
func opensChest(f *Fight, s *Step) bool {
	for _, i := range s.Dice.S {
		if strings.HasPrefix(f.Sites[i], "openChest:") {
			return true
		}
	}
	return false
}
