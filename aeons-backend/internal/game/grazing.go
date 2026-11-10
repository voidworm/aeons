package game

import "aeons/internal/creature"

// a prey eats one unit of food from a food node at its location, if there is any
func (gs *GameState) eatFoodAtLocation(prey *creature.Unit) {
	for _, node := range gs.NodesAtLocation(prey.CurrentLocation) {
		if node.Food.Available() {
			node.Food.Take()
			gs.Logf("%s eats from %s", prey.NameWithID(), node.Name)
			return
		}
	}
}

func (gs *GameState) ResolveGrazing() {
	for _, c := range gs.Creatures {
		if c.Role.IsPrey() {
			gs.eatFoodAtLocation(c)
		}
	}
}
