// Command report turns test results into the four plain-language pages the
// owner reads: Rule Book, Old-Game Match, House Rules and Changes.
//
//	go test -json ./... > test.json
//	go run ./tools/report -tests test.json -match match -changes docs/changes-from-old-game.md -out site
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// event is one line of `go test -json` output.
type event struct {
	Action, Package, Test, Output string
}

// result is one test or rule sentence.
type result struct {
	Pkg, Name, Sentence string
	Passed, Failed      bool
	Log                 []string
}

const module = "github.com/papigetbusy/crusader-tactics-next/"

// sections names each part of the Rule Book, in reading order.
var sections = []struct{ pkg, title, about string }{
	{"sim/engine", "Actions", "What you may do, and what the referee refuses."},
	{"sim/board", "Movement", "Who can walk where."},
	{"sim/combat", "Hitting and damage", "Hit chance, damage, statuses from hits."},
	{"sim/stats", "Levels, skills, ascensions and gear", "How heroes earn experience, level up, evolve, learn skills and ascend, and what their gear does to their numbers."},
	{"sim/items", "Loot and gold", "What monsters, chests and bounties drop, how new items are made, and gold from kills."},
	{"sim/floor", "Floors", "How each floor of the climb is built: the map, its monsters and their level, bosses, elites, thrones, reinforcements, chests and the blue starting tiles."},
	{"sim/turn", "Turns", "What happens when a turn starts and ends."},
	{"sim/ai", "Enemy decisions", "How monsters choose what to do."},
	{"sim/rng", "Dice", "The dice every roll comes from."},
	{"sim/content", "Data files", "The game numbers you can edit."},
}

func main() {
	tests := flag.String("tests", "test.json", "go test -json output")
	match := flag.String("match", "", "folder with the per-fight results (match-<set>.json) from the old-game replay")
	changes := flag.String("changes", "docs/changes-from-old-game.md", "the list of intentional differences")
	build := flag.String("build", "dev", "build label")
	out := flag.String("out", "site", "folder to write the pages to")
	flag.Parse()

	results := readTests(*tests)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	write(filepath.Join(*out, "rules.html"), rulesPage(results, *build))
	write(filepath.Join(*out, "match.html"), matchPage(results, *match, *build))
	write(filepath.Join(*out, "house.html"), housePage(results, *build))
	write(filepath.Join(*out, "changes.html"), changesPage(*changes, *build))

	// A plain summary for the build robot's log.
	pass, fail := 0, 0
	for _, r := range results {
		if r.Sentence == "" {
			continue
		}
		if r.Failed {
			fail++
			fmt.Printf("RED: %s: %s\n", r.Pkg, r.Sentence)
		} else if r.Passed {
			pass++
		}
	}
	fmt.Printf("%d rules green, %d red\n", pass, fail)
}

func readTests(path string) []*result {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	byKey := map[string]*result{}
	var order []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		var e event
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.Test == "" {
			continue
		}
		key := e.Package + " " + e.Test
		r := byKey[key]
		if r == nil {
			r = &result{Pkg: strings.TrimPrefix(e.Package, module), Name: e.Test}
			if i := strings.Index(e.Test, "/"); i >= 0 {
				r.Sentence = strings.ReplaceAll(e.Test[i+1:], "_", " ")
			}
			byKey[key] = r
			order = append(order, key)
		}
		switch e.Action {
		case "pass":
			r.Passed = true
		case "fail":
			r.Failed = true
		case "output":
			line := strings.TrimSpace(e.Output)
			if i := strings.Index(line, ".go:"); i >= 0 && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "===") {
				if j := strings.Index(line[i:], ": "); j >= 0 {
					r.Log = append(r.Log, line[i+j+2:])
				}
			}
		}
	}
	var out []*result
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out
}

func write(path, html string) {
	if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
		log.Fatal(err)
	}
}

func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

func mark(r *result) string {
	if r.Failed {
		return `<span class="no">✗</span>`
	}
	return `<span class="ok">✓</span>`
}

func sortedSentences(results []*result, pkg string) []*result {
	var out []*result
	for _, r := range results {
		if r.Pkg == pkg && r.Sentence != "" {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Failed && !out[j].Failed })
	return out
}
