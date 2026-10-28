package game

import "aeons/internal/location"

// ReconcileVision unveils the locations the players can see and remembers the nodes they see there.
// Veiled locations keep the nodes as they were when the players last saw them.
func (gs *GameState) ReconcileVision() {
	seen := gs.LocationsWithVision()
	for _, l := range gs.Locations {
		l.Veiled = !seen[l]
		l.Known = l.Known || !l.Veiled
	}
	for _, n := range gs.Nodes {
		if !n.StaticLocation.Veiled {
			gs.KnownNodes[n.ID] = n.Snapshot()
		}
	}
}

// players that are not fainted and their kindred give vision
func (gs *GameState) LocationsWithVision() map[*location.Unit]bool {
	seen := map[*location.Unit]bool{}
	for _, p := range gs.Players {
		if !p.Fainted() {
			seen[p.CurrentLocation] = true
		}
	}
	for _, c := range gs.Creatures {
		if c.Bonded && c.BondedTo != nil && !c.BondedTo.Fainted() {
			seen[c.CurrentLocation] = true
		}
	}
	return seen
}
