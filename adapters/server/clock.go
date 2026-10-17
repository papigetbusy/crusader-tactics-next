package main

import "time"

// ClockView is the server's clock, sent about once a second during a match.
// It is not a move. The phone only displays it.
type ClockView struct {
	Type    string `json:"type"` // "clock"
	Seq     int    `json:"seq"`
	ThinkMs int    `json:"thinkMs"`
	Side    int    `json:"side"`
	BankMs  [2]int `json:"bankMs"`
}

// chargeMs is the time to charge, after the free seconds that follow a move.
func chargeMs(last, now time.Time, grace time.Duration) int {
	d := now.Sub(last) - grace
	if d < 0 {
		return 0
	}
	return int(d / time.Millisecond)
}

func (r *Room) think(now time.Time) int {
	if !r.playing() {
		return 0
	}
	return chargeMs(r.Last, now, r.grace)
}
