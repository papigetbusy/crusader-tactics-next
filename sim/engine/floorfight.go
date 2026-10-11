package engine

import (
	"fmt"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
	"github.com/papigetbusy/crusader-tactics-next/sim/floor"
	"github.com/papigetbusy/crusader-tactics-next/sim/state"
)

// practiceTeam is who fights a built floor on the practice page until team
// select arrives (7c).
var practiceTeam = []string{"Warrior", "Tank", "Mage", "Healer", "Ranger"}

// FloorSetup builds floor number of a run for the practice page: the real
// floor, with a practice team of five at a level fit for the floor, standing
// spread along the bottom row (team select and placing come in 7c).
func FloorSetup(c *content.Content, run uint32, number int) Setup {
	level := max(1, min(60, (number*9+5)/10))
	team := make([]floor.Hero, len(practiceTeam))
	for i := range team {
		team[i] = floor.Hero{Level: level}
	}
	fl := floor.Build(c, run, number, team, floor.Memory{}, nil)
	kind := "Floor"
	switch {
	case fl.Boss:
		kind = "Boss floor"
	case fl.Throne != nil:
		kind = "Throne floor"
	case fl.Reinforce:
		kind = "Reinforcement floor"
	}
	s := Setup{Name: fmt.Sprintf("%s %d · %dx%d %s", kind, number, fl.N, fl.N, fl.Biome), Floor: number, N: fl.N,
		Ground: fl.Ground, Heights: fl.Heights, Throne: fl.Throne, Chests: fl.Chests,
		Biome: fl.Biome, Reinforce: fl.Reinforce, Starts: fl.Starts}
	for i, class := range practiceTeam {
		n := float64(fl.N)
		x := min(fl.N-1, int((float64(i)+.5)*n/float64(len(practiceTeam))))
		s.Units = append(s.Units, UnitSetup{Team: state.Heroes, Class: class, Level: level, X: x, Y: fl.N - 1,
			Gear: map[string]content.Item{}})
	}
	for _, e := range fl.Enemies {
		s.Units = append(s.Units, UnitSetup{Team: state.Enemies, Class: e.Class, Level: e.Level, X: e.X, Y: e.Y,
			HP: e.MaxHP, Atk: e.Atk, Def: e.Def, Mov: e.Mov, Element: e.Element, Boss: e.Boss, Elite: e.Elite,
			Kind: e.Kind, Ally: e.Ally})
	}
	return s
}
