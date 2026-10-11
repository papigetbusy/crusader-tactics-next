// Package guard holds the house rules: tests that keep the project clean.
// If one of these turns red, the code has broken a rule of the design, not
// a rule of the game.
package guard

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/papigetbusy/crusader-tactics-next/"

// layers is the order of the referee's folders. A folder may only use folders
// that come before it in this list (see docs/skeleton-design.md, section 1.4).
var layers = []string{"rng", "content", "state", "action", "board", "stats", "items",
	"combat", "floor", "save", "ai", "turn", "run", "engine"}

// forbidden are things the referee must never touch: the browser, the clock,
// the network, files, or Go's own unseeded random numbers.
var forbidden = map[string]string{
	"syscall/js":    "the browser",
	"time":          "the clock",
	"math/rand":     "Go's own random numbers",
	"math/rand/v2":  "Go's own random numbers",
	"os":            "files and the computer",
	"net":           "the network",
	"net/http":      "the network",
	"crypto/rand":   "unseeded random numbers",
	"os/exec":       "other programs",
	"path/filepath": "files",
	"io/ioutil":     "files",
	"unsafe":        "unsafe memory tricks",
	"runtime/debug": "runtime internals",
	"plugin":        "plugins",
	"reflect":       "reflection (keeps rules plain and fast)",
}

const maxLines = 400

func root() string { return filepath.Join("..", "..") }

// simFiles lists every non-test Go file in sim/, by package folder.
func simFiles(t *testing.T) map[string][]string {
	out := map[string][]string{}
	err := filepath.WalkDir(filepath.Join(root(), "sim"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(filepath.Join(root(), "sim"), filepath.Dir(p))
		pkg := strings.Split(filepath.ToSlash(rel), "/")[0]
		out[pkg] = append(out[pkg], p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func imports(t *testing.T, file string) []string {
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, im := range f.Imports {
		p, _ := strconv.Unquote(im.Path.Value)
		out = append(out, p)
	}
	return out
}

func rank(name string) int {
	for i, l := range layers {
		if l == name {
			return i
		}
	}
	return -1
}

func TestHouseRules(t *testing.T) {
	files := simFiles(t)

	t.Run("The referee never touches the screen, the clock, the network or unseeded random", func(t *testing.T) {
		for _, list := range files {
			for _, f := range list {
				for _, im := range imports(t, f) {
					if why, bad := forbidden[im]; bad {
						t.Errorf("%s uses %q (%s); rules must not", f, im, why)
					}
				}
			}
		}
	})

	t.Run("Each referee folder only uses the folders below it", func(t *testing.T) {
		for pkg, list := range files {
			me := rank(pkg)
			if me < 0 {
				t.Errorf("sim/%s is not in the layer list; add it to tools/guard and docs/skeleton-design.md", pkg)
				continue
			}
			for _, f := range list {
				for _, im := range imports(t, f) {
					if !strings.HasPrefix(im, module+"sim/") {
						continue
					}
					other := strings.Split(strings.TrimPrefix(im, module+"sim/"), "/")[0]
					if rank(other) >= me {
						t.Errorf("sim/%s uses sim/%s, but %s comes after %s in the layer order", pkg, other, other, pkg)
					}
				}
			}
		}
	})

	t.Run("No code file is longer than 400 lines", func(t *testing.T) {
		err := filepath.WalkDir(root(), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && (d.Name() == ".git" || d.Name() == "testdata") {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(p, ".go") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if n := strings.Count(string(b), "\n"); n > maxLines {
				t.Errorf("%s has %d lines (limit %d): split it", p, n, maxLines)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
