package content

// Versus is the 1v1 match between two players (owner, 2026-10-11 and
// 2026-10-16; data/rules/versus.json).
type Versus struct {
	Team        int `json:"team"`        // heroes each player brings
	Level       int `json:"level"`       // every hero's level
	Size        int `json:"size"`        // the map is Size x Size
	TurnSeconds int `json:"turnSeconds"` // the most one turn may last
	BankSeconds int `json:"bankSeconds"` // each player's time for the whole match
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-16",
		What: "Versus (online 1v1, new): two players, 8 heroes each at Lv 30 with chosen ascensions and one fixed purple " +
			"set per class, on a random mirrored 12×12 map. 30 seconds per turn and 2½ minutes per player; running out " +
			"of turn time ends the turn, running out of your total time loses. No experience, gold or loot. The old game " +
			"had no versus.",
		Apply: func(c *Content) { c.Versus = Versus{Team: 8, Level: 30, Size: 12, TurnSeconds: 30, BankSeconds: 150} },
	})
}
