package combat

import (
	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// Forecast is what the screen shows before an attack. It runs the same
// Damage function as a real hit (one math path), at the lowest, middle and
// highest damage roll.
type Forecast struct {
	HitChance  int     `json:"hitChance"`
	Low        int     `json:"low"`
	Mid        int     `json:"mid"`
	High       int     `json:"high"`
	Mult       float64 `json:"mult"`
	CritChance int     `json:"critChance,omitempty"`
	CritLow    int     `json:"critLow,omitempty"`
	CritHi     int     `json:"critHigh,omitempty"`
}

// Preview forecasts an attack without changing anything.
func Preview(c *content.Content, b *state.Battle, a Attack) Forecast {
	k := c.Combat
	lo, m := Damage(c, b, a, k.DamageRollLow)
	mid, _ := Damage(c, b, a, k.DamageRollLow+float64(k.DamageRollSpread*0.5))
	hi, _ := Damage(c, b, a, k.DamageRollLow+k.DamageRollSpread)
	f := Forecast{HitChance: HitChance(c, b, a), Low: lo, Mid: mid, High: hi, Mult: m, CritChance: a.By.Af("crit")}
	if f.CritChance > 0 {
		f.CritLow, f.CritHi = int(Round(float64(lo)*k.CritDamage)), int(Round(float64(hi)*k.CritDamage))
	}
	return f
}
