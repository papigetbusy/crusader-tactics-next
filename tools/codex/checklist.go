package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// removed lists every old Codex entry that is no longer in the game, with
// the reason. An old entry missing from the new game and not listed here
// fails the test, so nothing disappears silently.
var removed = map[string]string{
	"Dawnbreaker":        "replaced by one legendary weapon per class (owner, 2026-10-17)",
	"Heavensfall":        "replaced by one legendary weapon per class (owner, 2026-10-17)",
	"Gale Piercer":       "replaced by one legendary weapon per class (owner, 2026-10-17)",
	"Thunderhowl":        "replaced by one legendary weapon per class (owner, 2026-10-17)",
	"Emberheart Scepter": "replaced by one legendary weapon per class (owner, 2026-10-17)",
}

// oldRules are the old Codex's floor and rule lines, each with what the new
// game does (checked by hand against docs/changes-from-old-game.md).
var oldRules = [][2]string{
	{"Floors never end; maps start at 14×14 and grow by 1 every 4 floors, up to 32×32.", "Same."},
	{"Enemies: 4 on floor 1, one more every few floors, more for a party above 5, up to 16.", "Same rule; the team counted is the heroes you bring (5 to 8 slots)."},
	{"Monster level rises with every floor and with the party's strength.", "Changed: monster level is set when the map is made, from your strongest heroes for the number of slots."},
	{"Every 5th floor has an elite; every 10th a boss on its throne.", "Same."},
	{"Every 7th floor is a shop.", "Changed: the shop comes up before every 7th floor; you shop, then fight that floor."},
	{"Recruit a new hero at floors 25, 50 and 75 (party up to 8).", "Changed: all 10 heroes are yours from the start (no recruiting). You bring 5, and one more slot opens at floors 25, 50 and 75 (8 at most)."},
	{"A free legendary every 100 floors.", "Same."},
	{"Seize floors: move a hero onto the golden throne to win.", "Same."},
	{"Reinforcement floors: 2 extra monsters arrive on turns 3 and 5.", "Same."},
	{"Seven biomes.", "Same."},
	{"A floor you return to has the same map, monsters and chest spots.", "Same."},
	{"Opened chests stay opened; shops keep their stock.", "Same."},
	{"Replaying a cleared floor gives half gold and full XP.", "Same."},
	{"A boss that already dropped a legendary falls back to normal loot.", "Same."},
	{"Level cap: at most 5 levels above the monsters of your highest floor.", "Same."},
	{"Loot rarities: common, uncommon, rare, legendary.", "Changed: a purple (epic) tier sits between rare and legendary; legendaries only drop once a hero reaches Lv 30."},
	{"Skills level up with use.", "Same: +5% power per skill level; switching skills counts double."},
	{"Each class learns its own number of skills at its own levels.", "Changed: every class learns on one even schedule (Lv 1, 3, 6 ... 27; ultimates at Lv 1, 20 and 30), with 53 new skills and Last Light."},
	{"Change ascension once per floor from the Gear sheet.", "Same, between battles."},
	{"Settings → Saves: copy, download or import a save into any slot.", "Same: Save → Export copies and downloads; Import takes a file or pasted code into a chosen slot and asks before replacing one."},
	{"Defeat: you stay on the floor; level up, gear up, try again.", "Same, on purpose: no defeat penalty (owner, 2026-10-11)."},
}

// old is the part of the old Codex's data this list compares.
type old struct {
	Heroes map[string]struct {
		Stages []struct{ Name string }
		Specs  []struct{ N string }
		Ab     []struct{ N string }
	}
	Skills []struct {
		N, Cls, T string
		Lv, Min   int
		Max, Cd   int
		Pw        float64
	}
	Fx       map[string]string
	Monsters []struct {
		N, Cls string
		El     *string
	}
	Elites []struct{ N string }
	Bosses []struct{ N string }
	Items  struct {
		WPN  [][]any
		ARN  map[string]map[string][]string
		UNQ  map[string]any
		SETS map[string]struct{ N string }
		AFX  []struct{ K string }
		NMc  []string
	}
	Status []struct{ K string }
	TINFO  map[string]struct{ N string }
}

func readOld() (old, error) {
	var o old
	b, err := os.ReadFile(oldCodex)
	if err != nil {
		return o, err
	}
	s := string(b)
	i := strings.Index(s, "const D=")
	if i < 0 {
		return o, fmt.Errorf("%s: no data found", oldCodex)
	}
	err = json.NewDecoder(strings.NewReader(s[i+8:])).Decode(&o)
	return o, err
}

// row is one line of the checklist.
type row struct{ Name, Mark, Note string }

func checklist(x Codex) ([]byte, error) {
	o, err := readOld()
	if err != nil {
		return nil, err
	}
	var missing []string
	have := func(set map[string]bool, n string) row {
		if set[n] {
			return row{n, "✅", ""}
		}
		if why, ok := removed[n]; ok {
			return row{n, "❌", why}
		}
		missing = append(missing, n)
		return row{n, "❓", "missing and not listed as removed"}
	}
	var b strings.Builder
	b.WriteString("# Codex checklist: old entries in the new game\n\nEvery entry of the old Codex (v0.22, `docs/codex/old-codex-v0.22.html`) and where it is in the new game. ✅ in the new Codex · ✏️ changed (what changed) · ❌ removed (why). Made by `go run ./tools/codex`; a test fails if an old entry disappears without being listed.\n")
	section := func(title string, rows []row) {
		fmt.Fprintf(&b, "\n## %s (%d)\n\n", title, len(rows))
		for _, r := range rows {
			fmt.Fprintf(&b, "- %s %s", r.Mark, r.Name)
			if r.Note != "" {
				fmt.Fprintf(&b, ": %s", r.Note)
			}
			b.WriteString("\n")
		}
	}

	// heroes, forms, ascensions and abilities
	hs, forms, ab := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for n, h := range x.Heroes {
		hs[n] = true
		for _, s := range h.Stages {
			forms[s.Name] = true
		}
		for _, a := range h.Abilities {
			ab[a.Name] = true
		}
	}
	var rows []row
	for _, n := range sortedKeys(o.Heroes) {
		r := have(hs, n)
		h := o.Heroes[n]
		var miss []string
		for _, s := range h.Stages {
			if !forms[s.Name] {
				miss = append(miss, s.Name)
			}
		}
		for _, a := range h.Ab {
			if !ab[a.N] {
				miss = append(miss, a.N)
			}
		}
		if len(miss) > 0 {
			r.Mark, r.Note = "❓", "missing: "+strings.Join(miss, ", ")
			missing = append(missing, miss...)
		} else {
			r.Note = fmt.Sprintf("all %d forms and ascensions, all %d class abilities; new v0.26 sprites", len(h.Stages), len(h.Ab))
		}
		rows = append(rows, r)
	}
	section("Heroes", rows)

	// skills
	ns := map[string]Skill{}
	for _, s := range x.Skills {
		ns[s.Name] = s
	}
	rows = nil
	for _, s := range o.Skills {
		n, ok := ns[s.N]
		if !ok {
			rows = append(rows, have(map[string]bool{}, s.N))
			continue
		}
		var d []string
		if n.Level != s.Lv {
			d = append(d, fmt.Sprintf("learned at Lv %d (was %d)", n.Level, s.Lv))
		}
		if (n.Kind == "u") != (s.T == "u") {
			d = append(d, map[bool]string{true: "now an ultimate", false: "now a normal skill"}[n.Kind == "u"])
		}
		if n.Min != s.Min || n.Max != s.Max {
			d = append(d, fmt.Sprintf("range %d–%d (was %d–%d)", n.Min, n.Max, s.Min, s.Max))
		}
		if n.Cooldown != s.Cd {
			d = append(d, fmt.Sprintf("cooldown %d (was %d)", n.Cooldown, s.Cd))
		}
		if fmt.Sprintf("%.2f", n.Power) != fmt.Sprintf("%.2f", s.Pw) {
			d = append(d, fmt.Sprintf("power %g (was %g)", n.Power, s.Pw))
		}
		if len(d) > 0 {
			rows = append(rows, row{s.N, "✏️", strings.Join(d, "; ")})
		} else {
			rows = append(rows, row{s.N, "✅", ""})
		}
	}
	section("Hero skills", rows)
	var news []string
	for _, s := range x.Skills {
		if s.New {
			news = append(news, s.Name+" ("+s.Class+")")
		}
	}
	fmt.Fprintf(&b, "\nNew in this game, not in the old Codex (%d): %s.\n", len(news), strings.Join(news, ", "))

	// attack effects
	fx := map[string]bool{"Strike_m": true, "Strike_r": true, "Strike_gun": true, "Strike_g": true}
	for n := range x.Abilities {
		fx[n] = true
	}
	bk, err := os.ReadFile("web/assets/skillbook.json")
	if err != nil {
		return nil, err
	}
	var book struct{ Skills map[string]any }
	if err := json.Unmarshal(bk, &book); err != nil {
		return nil, err
	}
	for n := range book.Skills {
		fx[n] = true
	}
	rows = nil
	for _, n := range sortedKeys(o.Fx) {
		r := have(fx, n)
		if r.Mark == "✅" {
			r.Note = "plays in the new Codex, animated with the game's own effect code"
		}
		rows = append(rows, r)
	}
	section("Attack effects", rows)

	mons := func(title string, names []string, set map[string]bool) {
		rows = nil
		for _, n := range names {
			rows = append(rows, have(set, n))
		}
		section(title, rows)
	}
	set := func(ms []Monster) map[string]bool {
		m := map[string]bool{}
		for _, x := range ms {
			m[x.Name] = true
		}
		return m
	}
	var on, oe, ob []string
	for _, m := range o.Monsters {
		on = append(on, m.N)
	}
	for _, m := range o.Elites {
		oe = append(oe, m.N)
	}
	for _, m := range o.Bosses {
		ob = append(ob, m.N)
	}
	mons("Monsters", on, set(x.Monsters))
	fmt.Fprintf(&b, "\nMonsters now learn on the same schedule as heroes (see each monster in the new Codex); the old Codex listed a fixed kit.\n")
	mons("Elites", oe, set(x.Elites))
	mons("Bosses", ob, set(x.Bosses))

	// items
	g := x.Gear
	items := map[string]bool{}
	for _, w := range g.Weapons {
		items[w.Name] = true
	}
	for _, sl := range g.ArmorNames {
		for _, ns := range sl {
			for _, n := range ns {
				items[n] = true
			}
		}
	}
	for _, u := range g.Uniques {
		items[u.Name] = true
	}
	for _, n := range append(append([]string{}, g.CharmNames...), g.EpicCharms...) {
		items[n] = true
	}
	for _, s := range g.Sets {
		items[s.Name] = true
	}
	for _, a := range g.Affixes {
		items["affix "+a.Key] = true
	}
	var oi []string
	for _, w := range o.Items.WPN {
		oi = append(oi, fmt.Sprint(w[0]))
	}
	for _, w := range sortedKeys(o.Items.ARN) {
		for _, sl := range sortedKeys(o.Items.ARN[w]) {
			oi = append(oi, o.Items.ARN[w][sl]...)
		}
	}
	oi = append(oi, sortedKeys(o.Items.UNQ)...)
	oi = append(oi, o.Items.NMc...)
	for _, k := range sortedKeys(o.Items.SETS) {
		oi = append(oi, o.Items.SETS[k].N)
	}
	for _, a := range o.Items.AFX {
		oi = append(oi, "affix "+a.K)
	}
	mons("Items, sets and bonuses", oi, items)
	fmt.Fprintf(&b, "\nNew in this game: the purple (epic) tier (%d purple bonuses, purple charms %s).\n", len(g.Prefixes), strings.Join(g.EpicCharms, ", "))

	// statuses
	st, err := os.ReadFile("web/game/js/statuses.js")
	if err != nil {
		return nil, err
	}
	ss := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\s+(\w+): S\(`).FindAllStringSubmatch(string(st), -1) {
		ss[m[1]] = true
	}
	var os2 []string
	for _, s := range o.Status {
		os2 = append(os2, s.K)
	}
	mons("Statuses", os2, ss)

	// terrain
	ts := map[string]bool{}
	for _, t := range x.Terrain {
		ts[t.Name] = true
	}
	var ot []string
	for _, k := range sortedKeys(o.TINFO) {
		ot = append(ot, o.TINFO[k].N)
	}
	mons("Terrain", ot, ts)

	fmt.Fprintf(&b, "\n## Floors and rules (%d)\n\n", len(oldRules))
	for _, r := range oldRules {
		mark := "✅"
		if !strings.HasPrefix(r[1], "Same") {
			mark = "✏️"
		}
		fmt.Fprintf(&b, "- %s %s → %s\n", mark, r[0], r[1])
	}
	b.WriteString("\nThe old Codex's Mechanics, Controls and Credits tabs are rewritten for the new screen in the new Codex.\n")
	if len(missing) > 0 {
		return []byte(b.String()), fmt.Errorf("old Codex entries missing and not listed as removed: %s", strings.Join(missing, ", "))
	}
	return []byte(b.String()), nil
}

func sortedKeys[V any](m map[string]V) []string {
	k := make([]string, 0, len(m))
	for n := range m {
		k = append(k, n)
	}
	sort.Strings(k)
	return k
}
