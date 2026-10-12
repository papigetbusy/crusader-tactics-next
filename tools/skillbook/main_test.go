package main

import (
	"bytes"
	"os"
	"testing"
)

func TestSkillbook(t *testing.T) {
	t.Run("The screen's skill list matches the game's data", func(t *testing.T) {
		want, err := render()
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile("../../" + out)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatal("web/assets/skillbook.json is out of date: run go run ./tools/skillbook")
		}
	})
}
