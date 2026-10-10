// Package roll is the tiny stand-in for the future game "referee".
// It has no screen code: it takes a seed and returns dice. Same seed, same dice, always.
package roll

import "math/rand/v2"

// Dice returns n rolls of a die with the given number of sides (1..sides), decided only by seed.
func Dice(seed uint64, n, sides int) []int {
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	out := make([]int, n)
	for i := range out {
		out[i] = r.IntN(sides) + 1
	}
	return out
}
