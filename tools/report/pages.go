package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const style = `<style>
:root{--bg:#0e0b0b;--panel:#191312;--line:#3a2b26;--ink:#eadfcb;--muted:#a8998a;--gold:#c9a24d;--gold2:#e6c878;--blood2:#c0394b;--ok:#7dbb77}
*{box-sizing:border-box}html,body{margin:0;background:var(--bg);color:var(--ink);font:17px/1.45 "EB Garamond",Georgia,serif}
body{padding:12px 16px 40px;max-width:720px;margin:0 auto}
h1{font:700 22px/1.2 Cinzel,serif;color:var(--gold);margin:4px 0 2px}
h2{font:700 17px Cinzel,serif;color:var(--gold2);margin:22px 0 4px}
nav{display:flex;flex-wrap:wrap;gap:4px 14px;font-size:15px;margin:6px 0 14px}
nav a{color:var(--muted);text-decoration:none;border-bottom:1px solid transparent}
nav a.on,nav a:hover{color:var(--gold2);border-color:var(--gold)}
.lead{color:var(--muted)}
.big{font:700 20px Cinzel,serif;margin:10px 0}
ul.rules{list-style:none;padding:0;margin:6px 0}
ul.rules li{padding:6px 0 6px 26px;border-bottom:1px solid #241b19;position:relative}
ul.rules li span{position:absolute;left:2px;top:6px;font-weight:700}
.ok{color:var(--ok)}.no{color:var(--blood2)}
.card{background:var(--panel);border:1px solid var(--line);border-radius:8px;padding:10px 12px;margin:10px 0}
.card p{margin:4px 0}
table{width:100%;border-collapse:collapse;font-size:15px}
td,th{padding:5px 6px;border-bottom:1px solid #241b19;text-align:left;vertical-align:top}
th{color:var(--gold2);font-family:Cinzel,serif;font-size:13px}
.small{font-size:14px;color:var(--muted)}
footer{margin-top:20px;color:var(--muted);font-size:13px}
</style>`

func page(title, active, build, body string) string {
	links := [][2]string{{"./", "Practice fight"}, {"rules.html", "Rule Book"}, {"match.html", "Old-Game Match"},
		{"house.html", "House Rules"}, {"changes.html", "Changes"}}
	var nav strings.Builder
	for _, l := range links {
		cls := ""
		if l[0] == active {
			cls = ` class="on"`
		}
		fmt.Fprintf(&nav, `<a%s href="%s">%s</a>`, cls, l[0], l[1])
	}
	return `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1"><title>` + esc(title) + `</title>
<link rel="icon" href="data:,">
<link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@600;700&family=EB+Garamond:wght@400;500;600&display=swap" rel="stylesheet">
` + style + `</head><body><h1>Crusader Tactics</h1><nav>` + nav.String() + `</nav>` + body +
		`<footer>Made by the build robot from this build's tests. Build: ` + esc(build) + `</footer></body></html>`
}

func count(rs []*result) (int, int) {
	ok, bad := 0, 0
	for _, r := range rs {
		if r.Failed {
			bad++
		} else {
			ok++
		}
	}
	return ok, bad
}

func ruleList(rs []*result) string {
	var b strings.Builder
	b.WriteString(`<ul class="rules">`)
	for _, r := range rs {
		fmt.Fprintf(&b, `<li>%s%s</li>`, mark(r), esc(r.Sentence))
	}
	b.WriteString(`</ul>`)
	return b.String()
}

func rulesPage(results []*result, build string) string {
	var body strings.Builder
	total, red := 0, 0
	for _, s := range sections {
		ok, bad := count(sortedSentences(results, s.pkg))
		total += ok + bad
		red += bad
	}
	body.WriteString(`<p class="lead">Every rule of the game the referee knows so far, written as a sentence. Each one is checked by a test after every change. A tick means the rule holds right now.</p>`)
	fmt.Fprintf(&body, `<p class="big">%s %d of %d rules hold</p>`, map[bool]string{true: `<span class="ok">✓</span>`, false: `<span class="no">✗</span>`}[red == 0], total-red, total)
	for _, s := range sections {
		rs := sortedSentences(results, s.pkg)
		if len(rs) == 0 {
			continue
		}
		fmt.Fprintf(&body, `<h2>%s</h2><div class="small">%s</div>%s`, esc(s.title), esc(s.about), ruleList(rs))
	}
	return page("Rule Book", "rules.html", build, body.String())
}

func housePage(results []*result, build string) string {
	rs := sortedSentences(results, "tools/guard")
	ok, bad := count(rs)
	body := `<p class="lead">House rules keep the code clean. They are not rules of the game; they are rules of how the game is built, so it never turns back into one tangled file.</p>` +
		fmt.Sprintf(`<p class="big">%d of %d house rules hold</p>`, ok, ok+bad) + ruleList(rs) +
		`<ul class="rules"><li><span class="ok">✓</span>The code is neatly formatted (checked before the tests run)</li></ul>`
	return page("House Rules", "house.html", build, body)
}

type fightRow struct {
	Fight   int      `json:"fight"`
	Floor   int      `json:"floor"`
	Party   []string `json:"party"`
	Enemies []string `json:"enemies"`
	Steps   int      `json:"steps"`
	Matched int      `json:"matched"`
	First   string   `json:"firstDifference"`
}

func matchPage(results []*result, matchFile, build string) string {
	var body strings.Builder
	body.WriteString(`<p class="lead">The old game (v0.26.1) is the answer key. 200 of its fights were recorded with every dice roll. Go replays them with the same rolls and must end up in exactly the same place after every step.</p>`)
	names := map[string]string{
		"TestOldGameFights":    "Whole fights, replayed from the first turn",
		"TestOldGameMoves":     "Every move is legal in Go",
		"TestOldGameHits":      "Every hit and miss gives the same result",
		"TestOldGameSkillBars": "Every skill used is on the right skill bar",
	}
	for _, key := range []string{"TestOldGameFights", "TestOldGameMoves", "TestOldGameHits", "TestOldGameSkillBars"} {
		for _, r := range results {
			if r.Pkg != "tools/oldgame" || r.Name != key {
				continue
			}
			fmt.Fprintf(&body, `<div class="card"><p>%s <b>%s</b></p>`, mark(r), esc(names[key]))
			for _, l := range r.Log {
				fmt.Fprintf(&body, `<p class="small">%s</p>`, esc(l))
			}
			body.WriteString(`</div>`)
		}
	}
	var rows []fightRow
	if b, err := os.ReadFile(matchFile); err == nil {
		_ = json.Unmarshal(b, &rows)
	}
	if len(rows) > 0 {
		body.WriteString(`<h2>Fight by fight</h2><table><tr><th></th><th>Fight</th><th>Floor</th><th>Your party</th><th>Steps matched</th></tr>`)
		for _, r := range rows {
			m := `<span class="ok">✓</span>`
			if r.First != "" {
				m = `<span class="no">✗</span>`
			}
			extra := ""
			if r.First != "" {
				extra = `<br><span class="no">First difference: ` + esc(r.First) + `</span>`
			}
			fmt.Fprintf(&body, `<tr><td>%s</td><td>%d</td><td>%d</td><td>%s<br><span class="small">vs %s</span>%s</td><td>%d of %d</td></tr>`,
				m, r.Fight, r.Floor, esc(strings.Join(r.Party, ", ")), esc(strings.Join(r.Enemies, ", ")), extra, r.Matched, r.Steps)
		}
		body.WriteString(`</table><p class="small">Three of the 200 recorded fights are left out: the old game itself crashed in them (the line-2067 auto-walk bug).</p>`)
	}
	return page("Old-Game Match", "match.html", build, body.String())
}

func changesPage(path, build string) string {
	b, err := os.ReadFile(path)
	var body strings.Builder
	if err != nil {
		body.WriteString(`<p class="no">The list of changes is missing.</p>`)
	}
	inList := false
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "- "):
			if !inList {
				body.WriteString(`<ul class="rules">`)
				inList = true
			}
			fmt.Fprintf(&body, `<li><span class="small">•</span>%s</li>`, esc(t[2:]))
			continue
		case inList:
			body.WriteString(`</ul>`)
			inList = false
		}
		switch {
		case strings.HasPrefix(t, "# "):
			// the page already has a title
		case strings.HasPrefix(t, "## "):
			fmt.Fprintf(&body, `<h2>%s</h2>`, esc(t[3:]))
		case t != "":
			fmt.Fprintf(&body, `<p class="lead">%s</p>`, esc(t))
		}
	}
	if inList {
		body.WriteString(`</ul>`)
	}
	return page("Changes from the old game", "changes.html", build, body.String())
}
