package rng

import "fmt"

// Feed serves pre-recorded rolls in order. It is used only by the old-game
// match checks: the rules cannot tell a Feed from real dice, so if Go gives
// the same results as the old game with the same rolls, the rules match.
type Feed struct {
	values []uint32 // each roll stored as fraction * 2^32, as Phase 1 recorded them
	used   []string // the label of every roll taken, in order
}

// NewFeed makes a feed from recorded values.
func NewFeed(values []uint32) *Feed { return &Feed{values: values} }

// Roll returns the next recorded value. Asking for more rolls than were
// recorded is a mismatch with the old game, so it panics with a plain message.
func (f *Feed) Roll(label string) float64 {
	if len(f.used) >= len(f.values) {
		panic(fmt.Sprintf("Go asked for an extra roll (%q) the old game never made", label))
	}
	v := f.values[len(f.used)]
	f.used = append(f.used, label)
	return float64(v) / 4294967296.0
}

// Used is how many rolls have been taken so far.
func (f *Feed) Used() int { return len(f.used) }

// Labels lists the job of each roll taken, in order.
func (f *Feed) Labels() []string { return append([]string(nil), f.used...) }
