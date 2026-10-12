// Command codex writes web/codex/codex.json, everything the Codex page shows,
// straight from the game's data files, so the Codex always matches the
// game. It also writes docs/codex-checklist.md: every entry of the old
// Codex (docs/codex/old-codex-v0.22.html) and where it is now.
//
//	go run ./tools/codex
//
// The test in this folder fails when either file is out of date, or when an
// entry of the old Codex is gone without being listed as removed.
package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/run"
	"github.com/papigetbusy/crusader-tactics-next/sim/stats"
)

const (
	outJSON   = "web/codex/codex.json"
	outCheck  = "docs/codex-checklist.md"
	oldCodex  = "docs/codex/old-codex-v0.22.html"
	oldSkills = "testdata/oldgame/tables/SK.json"
	changesMD = "docs/changes-from-old-game.md"
)

// roles: each class in a few words (the old Codex's wording).
var roles = map[string]string{"Warrior": "Sword fighter", "Tank": "Heavy armor, stuns", "Mage": "Elemental magic", "Healer": "Heals, buffs, light attacks",
	"Ranger": "Long-range bow", "Assassin": "Fast dagger killer", "Gunslinger": "Pistols, multi-shots", "Dragon Knight": "Flying dragon rider",
	"Witch": "Poison and curses", "Knight": "Mounted charger, huge move"}

// heroOrder is the order of the team picker.
var heroOrder = []string{"Warrior", "Tank", "Mage", "Healer", "Ranger", "Assassin", "Gunslinger", "Dragon Knight", "Witch", "Knight"}

type Stage struct {
	Key, Label, Name, Pic string
	Lv, HP, Atk, Def, Mov int
}
type Ability struct {
	Stage, Name, Text string
	Cooldown          int
	Free              bool
}
type Asc struct{ Name, Element, Bonus, Passive, Ultimate string }
type Hero struct {
	Role       string
	Stages     []Stage
	Abilities  []Ability
	Ascensions []Asc
	Prof       [][2]string
	Dual       bool
}
type Skill struct {
	Name, Class, Kind, Ascension, Type, Text, Aim string
	Level, Min, Max, Cooldown, Area, Hits         int
	Power                                         float64
	Effects, Buffs                                map[string]int
	New, Free                                     bool
}
type Monster struct {
	Name, Class, Element string
	Schedule             [][2]any `json:",omitempty"`
	Skills               []string `json:",omitempty"`
	Floors               []int    `json:",omitempty"`
}
type Codex struct {
	Heroes    map[string]Hero
	Order     []string
	Skills    []Skill
	Monsters  []Monster
	Elites    []Monster
	Bosses    []Monster
	Gear      content.Gear
	Terrain   []content.Terrain
	Biomes    []string
	Floors    content.Floors
	Progress  content.Progress
	Slots     [][2]int // [floor, heroes]
	Changes   map[string][]string
	Abilities map[string]string // every class ability -> its class (for the effects list)
}

// skillType is the kind label of a skill (the old Codex's wording).
func skillType(s content.Skill) string {
	eln := map[string]string{"fire": "Fire", "ice": "Ice", "elec": "Thunder", "light": "Light", "dark": "Dark", content.UserElement: "Own-element"}
	switch {
	case s.Heal && s.Cleanse:
		return "Support"
	case s.Heal:
		return "Heal"
	case s.Support:
		return "Buff"
	case s.Revive:
		return "Revive"
	case s.Element != "" && s.Physical:
		return eln[s.Element] + " weapon art"
	case s.Element != "":
		return eln[s.Element] + " magic"
	case s.Magic:
		return "Arcane magic"
	}
	return "Physical"
}

// prof is what a class can use at a level: "Sword C (max S) · Axe D (max A) · 🛡 Medium & Heavy armor".
func prof(c *content.Content, class string, level int) string {
	cg := c.Gear.Classes[class]
	var p []string
	seen := map[string]bool{}
	for _, w := range cg.Weapons {
		if seen[w.Type] {
			continue
		}
		seen[w.Type] = true
		st := c.Gear.RankSteps[w.Role]
		top := min(len(c.Gear.Ranks)-1, st.Start+len(st.Levels))
		p = append(p, c.Gear.WeaponNames[w.Type]+" "+c.Gear.Ranks[stats.RankFor(c, class, level, w.Type)]+" (max "+c.Gear.Ranks[top]+")")
	}
	var a []string
	for _, w := range cg.Armor {
		a = append(a, map[string]string{"L": "Light", "M": "Medium", "H": "Heavy"}[w])
	}
	return "⚔ " + strings.Join(p, " · ") + " · 🛡 " + strings.Join(a, " & ") + " armor"
}

func build(c *content.Content) (Codex, error) {
	x := Codex{Heroes: map[string]Hero{}, Order: heroOrder, Gear: c.Gear, Biomes: c.Floors.Biomes, Floors: c.Floors, Progress: c.Progress, Abilities: map[string]string{}}
	x.Gear.Classes = c.Gear.Classes
	old := map[string]bool{}
	if b, err := os.ReadFile(oldSkills); err == nil {
		var m map[string]any
		if json.Unmarshal(b, &m) == nil {
			for n := range m {
				old[n] = true
			}
		}
	}
	for _, cls := range heroOrder {
		h := Hero{Role: roles[cls], Dual: c.Gear.Classes[cls].Dual}
		lv := append([]int{1}, c.Progress.FormLevels[1:]...)
		for i, l := range lv {
			n := stats.Hero(c, cls, l, "")
			h.Stages = append(h.Stages, Stage{Key: string(rune('0' + i)), Label: "Lv " + itoa(l), Name: stats.FormName(c, cls, l), Pic: cls + " form" + itoa(i+1), Lv: l, HP: n.MaxHP, Atk: n.Atk, Def: n.Def, Mov: n.Mov})
		}
		a := c.Abilities[cls]
		for i, t := range a.Tiers {
			h.Abilities = append(h.Abilities, Ability{Stage: "Lv " + itoa([]int{1, 5, 20}[min(i, 2)]), Name: t.Name, Text: t.Text, Cooldown: t.Cooldown, Free: t.Free})
			x.Abilities[t.Name] = cls
		}
		for _, an := range c.AscensionChoices[cls] {
			as := c.Ascensions[an]
			n := stats.Hero(c, cls, c.Progress.AscendAt, an)
			h.Stages = append(h.Stages, Stage{Key: an, Label: "Lv " + itoa(c.Progress.AscendAt) + " · " + an, Name: an, Pic: cls + " " + an, Lv: c.Progress.AscendAt, HP: n.MaxHP, Atk: n.Atk, Def: n.Def, Mov: n.Mov})
			h.Ascensions = append(h.Ascensions, Asc{Name: an, Element: as.Element, Bonus: as.Bonus, Passive: as.Passive, Ultimate: as.Ultimate})
			if t, ok := a.Ascensions[an]; ok {
				h.Abilities = append(h.Abilities, Ability{Stage: "Lv " + itoa(c.Progress.AscendAt) + " · " + an, Name: t.Name, Text: t.Text, Cooldown: t.Cooldown, Free: t.Free})
				x.Abilities[t.Name] = cls
			}
		}
		for _, l := range []int{1, 8, 18, 30} {
			h.Prof = append(h.Prof, [2]string{itoa(l), prof(c, cls, l)})
		}
		x.Heroes[cls] = h
	}
	if c.CrashDown.Name != "" {
		x.Abilities[c.CrashDown.Name] = "Dragon Knight"
	}
	names := make([]string, 0, len(c.Skills))
	for n := range c.Skills {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		s := c.Skills[n]
		x.Skills = append(x.Skills, Skill{Name: n, Class: s.Class, Kind: s.Kind, Ascension: s.Ascension, Type: skillType(s), Text: s.Text, Aim: s.Aim,
			Level: s.Level, Min: s.MinRange, Max: s.MaxRange, Cooldown: s.Cooldown, Area: s.Area, Hits: s.Hits, Power: s.Power,
			Effects: s.Effects, Buffs: s.Buffs, New: len(old) > 0 && !old[n], Free: s.Free})
	}
	for _, m := range c.Monsters {
		mo := Monster{Name: m.Name, Class: m.Class, Element: m.Element}
		for _, l := range c.Classes[m.Class].Schedule {
			mo.Schedule = append(mo.Schedule, [2]any{l.Level, l.Skill.Name})
		}
		x.Monsters = append(x.Monsters, mo)
	}
	kit := func(n string) []string {
		var k []string
		for _, s := range c.Classes[n].Skills {
			k = append(k, s.Name)
		}
		return k
	}
	for _, e := range c.Floors.Elites {
		x.Elites = append(x.Elites, Monster{Name: e.Name, Element: e.Element, Skills: kit(e.Name)})
	}
	for i, b := range c.Floors.Bosses {
		var fl []int
		for f := (i + 1) * c.Floors.BossEvery; len(fl) < 3; f += len(c.Floors.Bosses) * c.Floors.BossEvery {
			fl = append(fl, f)
		}
		x.Bosses = append(x.Bosses, Monster{Name: b.Name, Element: b.Element, Skills: kit(b.Name), Floors: fl})
	}
	for _, l := range []byte("grfhtTbwm") {
		if t, ok := c.Terrain[l]; ok {
			x.Terrain = append(x.Terrain, t)
		}
	}
	last := 0
	for f := 1; f <= 200; f++ {
		if n := (&run.Run{Best: f}).Slots(); n != last {
			x.Slots = append(x.Slots, [2]int{f, n})
			last = n
		}
	}
	ch, err := changes()
	x.Changes = ch
	return x, err
}

// changes reads the list of intentional differences, by section.
func changes() (map[string][]string, error) {
	b, err := os.ReadFile(changesMD)
	if err != nil {
		return nil, err
	}
	out, sec := map[string][]string{}, ""
	for _, l := range strings.Split(string(b), "\n") {
		switch {
		case strings.HasPrefix(l, "## "):
			sec = strings.TrimPrefix(l, "## ")
		case strings.HasPrefix(l, "- ") && sec != "":
			out[sec] = append(out[sec], strings.TrimPrefix(l, "- "))
		}
	}
	return out, nil
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func render() ([]byte, []byte, error) {
	c, err := content.Default()
	if err != nil {
		return nil, nil, err
	}
	x, err := build(c)
	if err != nil {
		return nil, nil, err
	}
	j, err := json.MarshalIndent(x, "", " ")
	if err != nil {
		return nil, nil, err
	}
	md, err := checklist(x)
	return append(j, '\n'), md, err
}

func main() {
	j, md, err := render()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll("web/codex", 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outJSON, j, 0o644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outCheck, md, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s and %s", outJSON, outCheck)
}
