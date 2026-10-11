// Command convert writes the live data files from the old game's saved tables.
// It was used to create data/skills.json, classes.json, elements.json and
// rules/combat.json. After that, data/ is edited by hand; the old tables stay
// as the answer key.
//
//	go run ./tools/convert
package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

func main() {
	c, err := content.FromOld(os.DirFS("testdata/oldgame/tables"))
	if err != nil {
		log.Fatal(err)
	}
	content.ApplyChanges(c) // the live data includes every approved change
	// Keep the old table order: it decides which skills a hero starts with.
	var skills []content.Skill
	for _, name := range c.SkillOrder {
		skills = append(skills, c.Skills[name])
	}
	var classes []content.Class
	for _, k := range c.Classes {
		classes = append(classes, k)
	}
	sort.Slice(classes, func(i, j int) bool { return classes[i].Name < classes[j].Name })

	write("data/skills.json", map[string]any{
		"_about": "Hero skills. Power multiplies attack; ranges are in tiles; effects are statuses put on the target (name: turns). From the old game's SK table.",
		"skills": skills})
	write("data/classes.json", map[string]any{
		"_about":  "Hero classes and monster types: base numbers and the skills monsters of each type use. Monster skills can differ from the hero skill of the same name. From the old game's CL and MSK tables.",
		"classes": classes})
	write("data/elements.json", map[string]any{
		"_about": "The element wheel: each element beats the one named after it.",
		"beats":  c.Beats})
	write("data/rules/combat.json", c.Combat)
	write("data/rules/turns.json", c.Turns)
	write("data/rules/ai.json", c.AI)
	write("data/rules/progress.json", c.Progress)
	var asc []content.Ascension
	for _, cl := range []string{"Warrior", "Tank", "Mage", "Healer", "Ranger", "Assassin", "Gunslinger", "Dragon Knight", "Witch", "Knight"} {
		for _, name := range c.AscensionChoices[cl] {
			asc = append(asc, c.Ascensions[name])
		}
	}
	write("data/ascensions.json", map[string]any{
		"_about":     "The Lv 30 ascensions: element, stat changes, ultimate, and the numbers their passive uses. From the old game's SPEC and SPECS tables.",
		"ascensions": asc, "choices": c.AscensionChoices})
}

func write(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}
