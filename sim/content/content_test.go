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
	t.Run("The live data files equal the old game's tables plus the approved changes, and nothing else", func(t *testing.T) {
		live := MustDefault()
		old, err := FromOld(os.DirFS("../../testdata/oldgame/tables"))
		if err != nil {
			t.Fatal(err)
		}
		ApplyChanges(old)
		if !reflect.DeepEqual(live, old) {
			t.Fatalf("live data differs from the old game plus approved changes:\nlive: %+v\nwant: %+v", live, old)
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

func TestApprovedChanges(t *testing.T) {
	t.Run("Monsters use the same version of a skill as heroes", func(t *testing.T) {
		c := MustDefault()
		for _, cl := range c.Classes {
			for _, s := range cl.Skills {
				if h, ok := c.Skills[s.Name]; ok && (h.MinRange != s.MinRange || h.MaxRange != s.MaxRange || h.Power != s.Power) {
					t.Errorf("%s's %s differs from the hero version", cl.Name, s.Name)
				}
			}
		}
	})
}
