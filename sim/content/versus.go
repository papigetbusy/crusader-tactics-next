package content

// Versus is the 1v1 match between two players (owner, 2026-10-11 and
// 2026-10-16; data/rules/versus.json).
type Versus struct {
	Team         int  `json:"team"`         // heroes each player brings
	Level        int  `json:"level"`        // every hero's level
	Size         int  `json:"size"`         // the map is Size x Size
	TurnSeconds  int  `json:"turnSeconds"`  // the most one turn may last; 0 means no per-turn limit
	BankSeconds  int  `json:"bankSeconds"`  // each player's time for the whole match
	GraceSeconds int  `json:"graceSeconds"` // online: seconds after a move that are not charged
	Fog          bool `json:"fog"`          // fog of war: a side sees only enemies near its own units
	Vision       int  `json:"vision"`       // how many steps each unit sees
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-16",
		What: "Versus (online 1v1, new): two players, 8 heroes each at Lv 30 with chosen ascensions and one fixed purple " +
			"set per class, on a random mirrored map. Each player has 5 minutes and no per-turn limit; running out " +
			"of that time loses. Online, the first 2 seconds after a move are not charged. No experience, gold or loot. " +
			"The old game had no versus.",
		Apply: func(c *Content) {
			c.Versus = Versus{Team: 8, Level: 30, Size: 12, TurnSeconds: 0, BankSeconds: 300, GraceSeconds: 2}
		},
	})
}

func init() {
	Changes = append(Changes, Change{
		Date: "2026-10-16",
		What: "Versus maps are the climb's biggest size (32×32), and versus has fog of war: each side sees only the enemy " +
			"units within 6 steps of one of its own units, and can't target what it can't see. The ground is always seen. " +
			"Watchers in an online room see everything. Single player has no fog.",
		Apply: func(c *Content) { c.Versus.Size, c.Versus.Fog, c.Versus.Vision = c.Floors.MapMax, true, 6 },
	})
}
