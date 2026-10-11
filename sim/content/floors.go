package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Floors is how floors are built: map size, how many monsters and how
// strong, bosses, elites, thrones, reinforcements, chests and biomes (old
// mapN, mobN, elv, calc, newFloor0, spawnChests and reinforce, index.html
// lines 736-760, 1000 and 1209).
type Floors struct {
	About string `json:"_about"`

	MapStart int `json:"mapStart"` // map size on floor 1 (tiles a side)
	MapEvery int `json:"mapEvery"` // one tile wider every this many floors
	MapMax   int `json:"mapMax"`

	MonstersStart  int                `json:"monstersStart"`  // monsters on floor 1
	MonstersEvery  int                `json:"monstersEvery"`  // one more every this many floors
	MonstersMax    int                `json:"monstersMax"`    //
	PartyBase      int                `json:"partyBase"`      // a team bigger than this meets more monsters
	PerExtraHero   float64            `json:"perExtraHero"`   // monsters per hero above PartyBase
	LevelPerFloor  float64            `json:"levelPerFloor"`  // monster level per floor
	LateFrom       int                `json:"lateFrom"`       // after this floor monsters gain more per floor
	LatePerFloor   float64            `json:"latePerFloor"`   //
	LevelPerPower  float64            `json:"levelPerPower"`  // monster level per point of the team's power (level + gear)
	GearPerLevel   float64            `json:"gearPerLevel"`   // gear points worth one level of power
	GearFloorShare float64            `json:"gearFloorShare"` // gear bonuses count (1 + floor / this) more on later floors
	PerExtraFoe    float64            `json:"perExtraFoe"`    // monster level lower per monster above 3
	GearWeights    map[string]float64 `json:"gearWeights"`    // power of one point of each item bonus (others 0.2)

	MonsterGrowth    float64 `json:"monsterGrowth"`    // monsters grow this many times faster than heroes
	MonsterGrowthLv  float64 `json:"monsterGrowthLv"`  // plus this much per level
	MonsterGrowthCap float64 `json:"monsterGrowthCap"` // up to this much extra

	BossEvery       int         `json:"bossEvery"`
	EliteEvery      int         `json:"eliteEvery"`
	EliteBossFloors int         `json:"eliteBossFloors"` // boss floors with an elite too (every this many floors)
	ShopEvery       int         `json:"shopEvery"`       // floors that are a shop instead of a fight (0: none)
	ThroneEvery     int         `json:"throneEvery"`     // throne floors: floor % ThroneEvery == ThroneAt, after floor 2
	ThroneAt        int         `json:"throneAt"`
	ReinforceEvery  int         `json:"reinforceEvery"` // reinforcement floors: floor % ReinforceEvery == ReinforceAt, after floor 3
	ReinforceAt     int         `json:"reinforceAt"`
	ReinforceTurns  []int       `json:"reinforceTurns"` // turns they arrive on
	ReinforceMax    int         `json:"reinforceMax"`   // monsters each time
	MaxHealers      int         `json:"maxHealers"`     // monster healers on one floor
	EnemyRows       int         `json:"enemyRows"`      // monsters start in the top rows
	AllyChance      float64     `json:"allyChance"`     // a companion spirit appears (not on boss floors or cleared floors)
	Chests          []ChestOdds `json:"chests"`
	StartRows       int         `json:"startRows"` // blue starting tiles: the bottom rows (0: the old game's fixed spots)

	Bosses   []Named  `json:"bosses"` // one per boss floor, in turn
	Elites   []Named  `json:"elites"`
	Biomes   []string `json:"biomes"`
	Elements []string `json:"elements"` // a monster with no element rolls one of these
}

// Named is a boss or elite: its class and element.
type Named struct {
	Name    string `json:"name"`
	Element string `json:"element"`
}

// ChestOdds is the chance of a chest of a kind on a floor.
type ChestOdds struct {
	Kind   string  `json:"kind"`
	Chance float64 `json:"chance"`
}

func oldFloors(tables fs.FS) (Floors, error) {
	f := Floors{
		About:    "How floors are built: size, monsters and their level, bosses, elites, thrones, reinforcements, chests and biomes. From the old game's TUNE, BOSS, ELT, BIOMES, EB and AFW tables and its floor code.",
		MapStart: 14, MapEvery: 4, MapMax: 32, // mapN, line 740
		MonstersStart: 4, MonstersMax: 16, PartyBase: 5, // mobN, line 740
		BossEvery: 10, EliteEvery: 5, EliteBossFloors: 50, ShopEvery: 7, // newFloor0, lines 744 and 752
		ThroneEvery: 3, ThroneAt: 2, ReinforceEvery: 4, ReinforceAt: 3, // line 745
		ReinforceTurns: []int{3, 5}, ReinforceMax: 2, // reinforce, line 1209
		MaxHealers: 1, EnemyRows: 3, // lines 753 and 757
		Chests: []ChestOdds{{"wood", .4}, {"silver", .18}, {"gold", .06}}, // spawnChests, line 1000
	}
	var tune map[string]float64
	var bosses, elites []struct {
		N string `json:"n"`
		E string `json:"e"`
	}
	var biomes []struct {
		N string `json:"n"`
	}
	var eb map[string]json.RawMessage
	for name, v := range map[string]any{"TUNE.json": &tune, "BOSS.json": &bosses, "ELT.json": &elites,
		"BIOMES.json": &biomes, "EB.json": &eb, "AFW.json": &f.GearWeights} {
		b, err := fs.ReadFile(tables, name)
		if err != nil {
			return f, fmt.Errorf("old table %s is missing: %w", name, err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			return f, fmt.Errorf("old table %s: %w", name, err)
		}
	}
	f.MonstersEvery, f.PerExtraHero = int(tune["mc"]), tune["ps"]
	f.LevelPerFloor, f.LateFrom, f.LatePerFloor = tune["fk"], int(tune["f2s"]), tune["fk2"]
	f.LevelPerPower, f.GearPerLevel, f.PerExtraFoe, f.GearFloorShare = tune["pk"], tune["gd"], tune["cp"], 40
	f.MonsterGrowth, f.MonsterGrowthLv, f.MonsterGrowthCap = tune["eh"], tune["eg"], tune["gc"]
	f.AllyChance = tune["cr"]
	if tune["eh"] != tune["ea"] || tune["eh"] != tune["ed"] {
		return f, fmt.Errorf("old table TUNE.json: eh, ea and ed differ; the port assumes one monster growth")
	}
	for _, b := range bosses {
		f.Bosses = append(f.Bosses, Named{b.N, b.E})
	}
	for _, e := range elites {
		f.Elites = append(f.Elites, Named{e.N, e.E})
	}
	for _, b := range biomes {
		f.Biomes = append(f.Biomes, b.N)
	}
	f.Elements = []string{"fire", "ice", "elec", "light", "dark"} // Object.keys(EB)
	if len(eb) != len(f.Elements) {
		return f, fmt.Errorf("old table EB.json: expected %d elements", len(f.Elements))
	}
	return f, nil
}
