// Command skillbook writes web/assets/skillbook.json: what the game screen
// needs to know about every skill and class ability to pick its attack
// effect (its element, whether it heals, its level...). It is drawing
// information only; the rules stay in the referee.
//
//	go run ./tools/skillbook
//
// The test in this folder fails when the file is out of date with data/.
package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

const out = "web/assets/skillbook.json"

// Entry is one skill as the screen sees it.
type Entry struct {
	Class     string  `json:"class,omitempty"`
	Hero      bool    `json:"hero,omitempty"` // a hero skill (the old SK table and the new skills)
	Kind      string  `json:"kind,omitempty"` // "u" ultimate
	Level     int     `json:"level,omitempty"`
	Power     float64 `json:"power,omitempty"`
	Element   string  `json:"element,omitempty"`
	Physical  bool    `json:"physical,omitempty"`
	Magic     bool    `json:"magic,omitempty"`
	Heal      bool    `json:"heal,omitempty"`
	Cleanse   bool    `json:"cleanse,omitempty"`
	Support   bool    `json:"support,omitempty"`
	Revive    bool    `json:"revive,omitempty"`
	Charge    bool    `json:"charge,omitempty"`
	Strike    bool    `json:"strike,omitempty"`
	Ascension string  `json:"ascension,omitempty"`
	Visual    string  `json:"visual,omitempty"`
}

// Book is the whole file.
type Book struct {
	Skills    map[string]Entry  `json:"skills"`
	Abilities map[string]string `json:"abilities"` // ability name -> its class
}

func entry(s content.Skill, hero bool) Entry {
	return Entry{Class: s.Class, Hero: hero, Kind: s.Kind, Level: s.Level, Power: s.Power, Element: s.Element,
		Physical: s.Physical, Magic: s.Magic, Heal: s.Heal, Cleanse: s.Cleanse, Support: s.Support,
		Revive: s.Revive, Charge: s.Charge, Strike: s.Strike, Ascension: s.Ascension, Visual: s.Visual}
}

// build makes the book from the game's data.
func build(c *content.Content) Book {
	b := Book{Skills: map[string]Entry{}, Abilities: map[string]string{}}
	for n, s := range c.Skills {
		b.Skills[n] = entry(s, true)
	}
	names := make([]string, 0, len(c.Classes))
	for n := range c.Classes {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		cl := c.Classes[n]
		all := append([]content.Skill{}, cl.Skills...)
		for _, l := range append(append([]content.LevelSkill{}, cl.LevelSkills...), cl.Schedule...) {
			all = append(all, l.Skill)
		}
		for _, s := range all {
			if _, ok := b.Skills[s.Name]; !ok && s.Name != "" {
				b.Skills[s.Name] = entry(s, false)
			}
		}
	}
	for cls, a := range c.Abilities {
		for _, t := range a.Tiers {
			b.Abilities[t.Name] = cls
		}
		for _, t := range a.Ascensions {
			b.Abilities[t.Name] = cls
		}
	}
	if c.CrashDown.Name != "" {
		b.Abilities[c.CrashDown.Name] = "Dragon Knight"
	}
	return b
}

func render() ([]byte, error) {
	c, err := content.Default()
	if err != nil {
		return nil, err
	}
	j, err := json.MarshalIndent(build(c), "", " ")
	return append(j, '\n'), err
}

func main() {
	j, err := render()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(out, j, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", out)
}
