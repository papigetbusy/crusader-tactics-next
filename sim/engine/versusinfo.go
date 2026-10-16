package engine

import (
	"sort"

	"github.com/papigetbusy/crusader-tactics-next/sim/content"
)

// VersusInfo is what the screen needs to let players pick a versus team: the
// rules' numbers and each class's ascensions in words.
type VersusInfo struct {
	Team        int               `json:"team"`
	Level       int               `json:"level"`
	TurnSeconds int               `json:"turnSeconds"`
	BankSeconds int               `json:"bankSeconds"`
	Classes     []VersusClassInfo `json:"classes"`
}

// VersusClassInfo is one hero class and the ascensions it can take.
type VersusClassInfo struct {
	Class      string            `json:"class"`
	Ascensions []VersusAscension `json:"ascensions"`
}

// VersusAscension is one ascension in words.
type VersusAscension struct {
	Name     string `json:"name"`
	Element  string `json:"element,omitempty"`
	Bonus    string `json:"bonus,omitempty"`
	Passive  string `json:"passive,omitempty"`
	Ultimate string `json:"ultimate,omitempty"`
}

// VersusRules describes versus for the team picker, classes in name order.
func VersusRules(c *content.Content) VersusInfo {
	k := c.Versus
	out := VersusInfo{Team: k.Team, Level: k.Level, TurnSeconds: k.TurnSeconds, BankSeconds: k.BankSeconds}
	var names []string
	for cl := range c.AscensionChoices {
		names = append(names, cl)
	}
	sort.Strings(names)
	for _, cl := range names {
		ci := VersusClassInfo{Class: cl}
		for _, n := range c.AscensionChoices[cl] {
			a := c.Ascensions[n]
			ci.Ascensions = append(ci.Ascensions, VersusAscension{Name: n, Element: a.Element, Bonus: a.Bonus, Passive: a.Passive, Ultimate: a.Ultimate})
		}
		out.Classes = append(out.Classes, ci)
	}
	return out
}
