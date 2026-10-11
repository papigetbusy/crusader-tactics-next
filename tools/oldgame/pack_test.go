package oldgame

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestPack turns raw recorder output into a slim recordings file that keeps
// only what the checks read. It runs only when asked:
//
//	OLDGAME_PACK="part-a.json,part-b.json,part-c.json:out.json.gz" go test ./tools/oldgame -run TestPack
func TestPack(t *testing.T) {
	spec := os.Getenv("OLDGAME_PACK")
	if spec == "" {
		t.Skip("set OLDGAME_PACK to pack recordings")
	}
	ins, out, _ := strings.Cut(spec, ":")
	var all []Fight
	for _, in := range strings.Split(ins, ",") {
		b, err := os.ReadFile(in)
		if err != nil {
			t.Fatal(err)
		}
		var fs []Fight
		if err := json.Unmarshal(b, &fs); err != nil {
			t.Fatal(err)
		}
		all = append(all, fs...)
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err := json.NewEncoder(z).Encode(all); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("packed %d fights into %s", len(all), out)
}
