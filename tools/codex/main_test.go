package main

import (
	"bytes"
	"os"
	"testing"
)

func TestCodex(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	j, md, err := render()
	t.Run("No entry of the old Codex disappears without being listed as removed", func(t *testing.T) {
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("The Codex matches the game's data", func(t *testing.T) {
		got, _ := os.ReadFile(outJSON)
		if !bytes.Equal(got, j) {
			t.Fatal("web/codex/codex.json is out of date: run go run ./tools/codex")
		}
		got, _ = os.ReadFile(outCheck)
		if !bytes.Equal(got, md) {
			t.Fatal("docs/codex-checklist.md is out of date: run go run ./tools/codex")
		}
	})
}
