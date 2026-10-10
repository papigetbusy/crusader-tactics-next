package content

import (
	"os"
	"reflect"
	"testing"
)

func TestData(t *testing.T) {
	t.Run("Every data file loads and passes its checks", func(t *testing.T) {
		if _, err := Default(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("The old game's saved tables load", func(t *testing.T) {
		if _, err := FromOld(os.DirFS("../../testdata/oldgame/tables")); err != nil {
			t.Fatal(err)
		}
	})

	// This is the seed of the "Changes from the old game" page: when a number is
	// changed on purpose, this test is updated to list the change.
	t.Run("The live movement and terrain data still match the old game", func(t *testing.T) {
		live := MustDefault()
		old, err := FromOld(os.DirFS("../../testdata/oldgame/tables"))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(live, old) {
			t.Fatalf("live data differs from the old game:\nlive: %+v\nold:  %+v", live, old)
		}
	})

	t.Run("A data file with a broken number is rejected with a plain message", func(t *testing.T) {
		c := MustDefault()
		bad := c.Terrain['f']
		bad.MoveCost = 0
		c.Terrain['f'] = bad
		if err := c.Check(); err == nil {
			t.Fatal("expected the check to complain")
		}
	})
}
