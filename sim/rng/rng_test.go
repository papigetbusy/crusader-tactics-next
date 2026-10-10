package rng

import "testing"

func TestDice(t *testing.T) {
	t.Run("The same seed always gives the same rolls", func(t *testing.T) {
		a, b := New(123, 1), New(123, 1)
		for i := 0; i < 1000; i++ {
			if a.Roll("x") != b.Roll("x") {
				t.Fatalf("roll %d differs", i)
			}
		}
	})

	t.Run("Different seeds give different rolls", func(t *testing.T) {
		a, b := New(123, 1), New(124, 1)
		same := 0
		for i := 0; i < 100; i++ {
			if a.Roll("x") == b.Roll("x") {
				same++
			}
		}
		if same > 2 {
			t.Fatalf("%d of 100 rolls were identical", same)
		}
	})

	t.Run("Different streams from one seed give different rolls", func(t *testing.T) {
		a, b := New(7, 1), New(7, 2)
		if a.Roll("x") == b.Roll("x") && a.Roll("x") == b.Roll("x") {
			t.Fatal("streams look identical")
		}
	})

	t.Run("A roll is always at least 0 and below 1", func(t *testing.T) {
		p := New(99, 3)
		for i := 0; i < 100000; i++ {
			if r := p.Roll("x"); r < 0 || r >= 1 {
				t.Fatalf("roll %v out of range", r)
			}
		}
	})

	t.Run("A whole-number roll of N gives every value from 0 to N-1 and nothing else", func(t *testing.T) {
		p := New(5, 1)
		seen := map[int]int{}
		for i := 0; i < 60000; i++ {
			seen[IntN(p, "x", 6)]++
		}
		for v := 0; v < 6; v++ {
			if seen[v] < 9000 || seen[v] > 11000 {
				t.Fatalf("value %d came up %d times out of 60000 (expected about 10000)", v, seen[v])
			}
		}
		if len(seen) != 6 {
			t.Fatalf("unexpected values: %v", seen)
		}
	})

	// This is the published test vector for the standard PCG32 generator
	// (seed 42, stream 54). Matching it proves we built the real thing, and
	// any accidental change to the generator would break it, because changing
	// it would silently change every replay.
	t.Run("The dice generator matches the published PCG32 reference numbers", func(t *testing.T) {
		p := New(42, 54)
		want := []uint32{0xa15c02b7, 0x7b47f409, 0xba1d3330, 0x83d2f293, 0xbfa4784b, 0xcbed606e}
		for i, w := range want {
			if got := p.next(); got != w {
				t.Fatalf("roll %d: got %#x, want %#x", i, got, w)
			}
		}
	})

	t.Run("A recorded-dice feed replays the recorded rolls in order", func(t *testing.T) {
		f := NewFeed([]uint32{0, 2147483648, 4294967295})
		if f.Roll("a") != 0 || f.Roll("b") != 0.5 {
			t.Fatal("feed values wrong")
		}
		if f.Used() != 2 || f.Labels()[1] != "b" {
			t.Fatal("feed bookkeeping wrong")
		}
	})

	t.Run("Asking a feed for more rolls than the old game made is reported", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected a report")
			}
		}()
		f := NewFeed(nil)
		f.Roll("extra")
	})
}
