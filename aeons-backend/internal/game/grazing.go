package game

import (
	"aeons/internal/collecting"
	"aeons/internal/creature"
	"math/rand/v2"
)

// a creature with the creature tag eats from nodes with the node tag, one unit of each pool it feeds on
type diet struct {
	creatureTag string
	nodeTag     string
	food        bool
	material    bool
}

var diets = []diet{
	{creature.Scavenger, collecting.Carcass, true, true},
	{creature.Grazer, collecting.Leafy, false, true},
	{creature.Gatherer, collecting.Fruitful, true, false},
}

func (d diet) canFeedFrom(node *collecting.Node) bool {
	return node.HasTag(d.nodeTag) && (d.food && node.Food.Available() || d.material && node.Material.Available())
}

type meal struct {
	diet diet
	node *collecting.Node
}

// a creature eats one meal per eating, a random pick from everything its tags let it eat at its location
func (gs *GameState) eatAtLocation(c *creature.Unit) {
	meals := []meal{}
	for _, d := range diets {
		if !c.HasIdentName(d.creatureTag) {
			continue
		}
		for _, node := range gs.NodesAtLocation(c.CurrentLocation) {
			if d.canFeedFrom(node) {
				meals = append(meals, meal{d, node})
			}
		}
	}
	if len(meals) == 0 {
		return
	}
	m := meals[rand.IntN(len(meals))]
	if m.diet.food && m.node.Food.Available() {
		m.node.Food.Take()
	}
	if m.diet.material && m.node.Material.Available() {
		m.node.Material.Take()
	}
	gs.Logf("%s eats from %s", c.NameWithID(), m.node.Name)
}

func (gs *GameState) ResolveGrazing() {
	for _, c := range gs.Creatures {
		gs.eatAtLocation(c)
	}
}
