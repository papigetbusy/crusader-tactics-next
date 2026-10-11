// Command art copies the old game's pictures into web/assets as named files,
// with web/assets/manifest.json listing them. Only the pictures are taken:
// the old game's code is never used by the new game (design-lock 18).
//
//	go run ./tools/art path/to/old/index.html
//
// It was run once on the frozen old page (sha256 in
// testdata/oldgame/index.html.sha256). After that, web/assets is ours, and the
// owner's new art replaces files in it.
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	dataURI  = regexp.MustCompile(`data:image/([a-z+]+);base64,([A-Za-z0-9+/=]+)`)
	jsonKey  = regexp.MustCompile(`"([^"]+)":\s*"$`)
	listName = regexp.MustCompile(`\["([^"]+)",\s*"$`)
	monster  = regexp.MustCompile(`\["[A-Za-z]+",\s*(null|"[a-z]+"),\s*"$`)
	nsClass  = regexp.MustCompile(`"([A-Za-z ]+)":\s*\{"bw"`)
	cssRule  = regexp.MustCompile(`([#.\w][^{}]*)\{[^{}]*$`)
	mnames   = regexp.MustCompile(`MNAMES=\[([^\]]+)\]`)
	pcProp   = regexp.MustCompile(`PC\.(\w+)='$`)
	table    = regexp.MustCompile(`(?:const (\w+)=\{|Object\.assign\((\w+),\{)`)
)

// tableOf is the name of the last object table opened before a picture
// ("PC", "TSO", ...).
func tableOf(page string, at int) string {
	ms := table.FindAllStringSubmatch(page[max(0, at-400000):at], -1)
	if len(ms) == 0 {
		return ""
	}
	m := ms[len(ms)-1]
	return m[1] + m[2]
}

// picture is one image found in the old page.
type picture struct {
	ext, data string
	before    string // the code just before it
	at        int
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./tools/art path/to/old/index.html")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	page := string(raw)
	var pics []picture
	for _, m := range dataURI.FindAllStringSubmatchIndex(page, -1) {
		pics = append(pics, picture{ext: page[m[2]:m[3]], data: page[m[4]:m[5]], before: page[max(0, m[0]-400):m[0]], at: m[0]})
	}
	monsters := monsterNames(page)
	nsStart := strings.Index(page, "const NS=")
	monIndex := 0
	type named struct {
		picture
		group, name string
	}
	var all []named
	for _, p := range pics { // names first, in page order (the monster list counts in order)
		if g, n := nameOf(p, page, nsStart, monsters, &monIndex); g != "" {
			all = append(all, named{p, g, n})
		}
	}
	// A style rule's picture is usually a named one again: name it by its table name.
	sort.SliceStable(all, func(i, j int) bool {
		return !strings.HasPrefix(all[i].name, "css ") && strings.HasPrefix(all[j].name, "css ")
	})
	manifest := map[string]map[string]string{}
	byHash := map[string]string{}
	for _, p := range all {
		group, name := p.group, p.name
		bytes, err := base64.StdEncoding.DecodeString(p.data)
		if err != nil {
			log.Fatalf("%s/%s: %v", group, name, err)
		}
		ext := map[string]string{"png": ".png", "webp": ".webp", "jpeg": ".jpg"}[p.ext]
		h := fmt.Sprintf("%x", sha256.Sum256(bytes))
		file := filepath.ToSlash(filepath.Join(group, slug(name)+ext))
		if prev, ok := byHash[h]; ok && group != "monsters" {
			file = prev // the same picture again: point at the first copy
		} else {
			byHash[h] = file
			if err := os.MkdirAll(filepath.Join("web/assets", group), 0o755); err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join("web/assets", file), bytes, 0o644); err != nil {
				log.Fatal(err)
			}
		}
		if manifest[group] == nil {
			manifest[group] = map[string]string{}
		}
		manifest[group][name] = file
	}
	for name := range manifest["ui"] {
		if strings.HasPrefix(name, "css ") { // only ever a second name for a picture above
			delete(manifest["ui"], name)
		}
	}
	out, _ := json.MarshalIndent(manifest, "", " ")
	if err := os.WriteFile("web/assets/manifest.json", append(out, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	writeSheets(page)
	files := map[string]bool{}
	for _, f := range byHash {
		files[f] = true
	}
	fmt.Printf("%d pictures found, %d files written\n", len(pics), len(files))
	var keys []string
	for g := range manifest {
		keys = append(keys, fmt.Sprintf("%s %d", g, len(manifest[g])))
	}
	sort.Strings(keys)
	fmt.Println(strings.Join(keys, ", "))
}

// nameOf gives a picture's group and name from the code around it.
func nameOf(p picture, page string, nsStart int, monsters []string, monIndex *int) (string, string) {
	b := p.before
	switch {
	case nsStart >= 0 && p.at > nsStart && strings.Contains(page[nsStart:p.at], `"bw"`) && !strings.Contains(page[nsStart:p.at], "}}};"):
		cls := nsClass.FindAllStringSubmatch(page[nsStart:p.at], -1)
		k := jsonKey.FindStringSubmatch(b)
		form := map[string]string{"0": "form1", "1": "form2", "2": "form3"}[k[1]]
		if form == "" {
			form = k[1] // an ascension
		}
		return "heroes", cls[len(cls)-1][1] + " " + form
	case monster.MatchString(b):
		i := *monIndex
		*monIndex++
		return "monsters", monsters[i]
	case strings.HasSuffix(b, "_si.src='"):
		return "monsters", monsters[1] // the old game's later Carrion Strider
	case pcProp.MatchString(b):
		return "heroes-first", pcProp.FindStringSubmatch(b)[1]
	case listName.MatchString(b):
		return "bosses", listName.FindStringSubmatch(b)[1]
	case jsonKey.MatchString(b):
		k := jsonKey.FindStringSubmatch(b)[1]
		switch {
		case strings.HasPrefix(k, "fx_"):
			return "effects", strings.TrimPrefix(k, "fx_")
		case strings.HasPrefix(k, "g_"):
			return "ground", strings.TrimPrefix(k, "g_")
		case tableOf(page, p.at) == "PC":
			return "heroes-first", k
		case strings.Contains("flat water foam bridge tree deco castle house rocks", k):
			return "terrain", k
		}
		return "ui", k
	case strings.Contains(b, "#title") || strings.Contains(b, "#tart"):
		return "title", "background"
	case cssRule.MatchString(b):
		sel := strings.TrimSpace(cssRule.FindStringSubmatch(b)[1])
		return "ui", "css " + sel
	}
	log.Printf("no name for the picture at %d: %q", p.at, b[max(0, len(b)-80):])
	return "", ""
}

func monsterNames(page string) []string {
	m := mnames.FindStringSubmatch(page)
	if m == nil {
		log.Fatal("MNAMES not found")
	}
	var out []string
	for _, n := range strings.Split(m[1], ",") {
		out = append(out, strings.Trim(n, "' "))
	}
	return out
}

var notWord = regexp.MustCompile(`[^a-z0-9]+`)

func slug(s string) string {
	return strings.Trim(notWord.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

var heroBox = regexp.MustCompile(`"([A-Za-z][A-Za-z ]*)": \{"bw": ([\d.]+), "bh": ([\d.]+)`)

// writeSheets saves how the picture sheets are cut up (effect frames, the
// item icon index, each hero picture's box) as web/assets/sheets.json.
func writeSheets(page string) {
	var tsm struct {
		Ridx map[string]int             `json:"ridx"`
		Efx  map[string]json.RawMessage `json:"efx"`
	}
	if err := json.Unmarshal([]byte(literal(page, "TSM=")), &tsm); err != nil {
		log.Fatal("TSM: ", err)
	}
	var fx2 map[string]json.RawMessage
	if err := json.Unmarshal([]byte(literal(page, "FX2M=")), &fx2); err != nil {
		log.Fatal("FX2M: ", err)
	}
	for k, v := range tsm.Efx {
		fx2[k] = v
	}
	box := map[string][2]float64{}
	for _, m := range heroBox.FindAllStringSubmatch(page, -1) {
		var w, h float64
		fmt.Sscan(m[2], &w)
		fmt.Sscan(m[3], &h)
		box[m[1]] = [2]float64{w, h}
	}
	out, _ := json.MarshalIndent(map[string]any{"effects": fx2, "itemIcons": tsm.Ridx, "heroBox": box}, "", " ")
	if err := os.WriteFile("web/assets/sheets.json", append(out, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}

// literal is the {...} written right after marker in the page.
func literal(page, marker string) string {
	i := strings.Index(page, marker)
	if i < 0 {
		log.Fatalf("%s not found", marker)
	}
	i += len(marker)
	depth, inStr := 0, false
	for j := i; j < len(page); j++ {
		c := page[j]
		switch {
		case inStr:
			if c == '\\' {
				j++
			} else if c == '"' {
				inStr = false
			}
		case c == '"':
			inStr = true
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return page[i : j+1]
			}
		}
	}
	log.Fatalf("%s is not closed", marker)
	return ""
}
