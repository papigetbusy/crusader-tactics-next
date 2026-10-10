// Package rng is the referee's dice. It is the only source of gameplay randomness.
//
// We use our own small generator (PCG32) instead of Go's built-in one, so the
// numbers a seed produces can never change when Go is updated. That keeps
// replays and, later, online play reproducible forever.
package rng

// Dice is what every rule asks for a roll. Each roll is labelled with its job
// ("hit roll", "damage spread"...), so a failed old-game check can say which
// roll was out of order.
type Dice interface {
	// Roll returns a fraction from 0 up to (not including) 1, like the old
	// game's Math.random().
	Roll(label string) float64
}

// IntN rolls a whole number from 0 to n-1, the way the old game's rnd(n) did:
// floor(roll * n).
func IntN(d Dice, label string, n int) int {
	return int(d.Roll(label) * float64(n))
}

// PCG is a seeded generator. The same seed and stream always give the same rolls.
type PCG struct {
	state, inc uint64
}

const pcgMult = 6364136223846793005

// New makes a generator from a seed and a stream number. Different streams
// from the same seed never overlap, so combat, AI and loot can each have their own.
func New(seed, stream uint64) *PCG {
	p := &PCG{inc: stream<<1 | 1}
	p.next()
	p.state += seed
	p.next()
	return p
}

// next is the standard PCG-XSH-RR 32-bit output step.
func (p *PCG) next() uint32 {
	old := p.state
	p.state = old*pcgMult + p.inc
	xorshifted := uint32(((old >> 18) ^ old) >> 27)
	rot := uint32(old >> 59)
	return xorshifted>>rot | xorshifted<<((-rot)&31)
}

// Roll returns next/2^32: a fraction with 32 bits of precision, the same
// precision the Phase 1 recordings stored for the old game's dice.
func (p *PCG) Roll(string) float64 {
	return float64(p.next()) / 4294967296.0
}
