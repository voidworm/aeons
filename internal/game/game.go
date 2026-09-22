package game

import (
	"aeons/internal/creature"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/player"
)

type GameState struct {
	Locations    []*location.Unit
	Players      []*player.Unit
	Creatures    []*creature.Unit
	HarvestUnits []*harvesting.Unit
	TurnCounter  int
}

func (gs *GameState) ReconcileDefeats() {
	alivePlayers := gs.Players[:0]

	for _, v := range gs.Players {
		if v.CurrentHealth > 0 {
			alivePlayers = append(alivePlayers, v)
		}
	}

	aliveCreatures := gs.Creatures[:0]
	for _, v := range gs.Creatures {
		if v.CurrentHealth > 0 {
			aliveCreatures = append(aliveCreatures, v)
		}
	}
	gs.Creatures = aliveCreatures
	gs.Players = alivePlayers
}

func (gs *GameState) LocationHasEnemies(le *location.Unit) bool {
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationHasPlayers(le *location.Unit) bool {
	for _, v := range gs.Players {
		if le == v.CurrentLocation {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationHaEvadableCreatures(le *location.Unit) bool {

	if !gs.LocationHasEnemies(le) {
		return false
	}

	for _, v := range gs.Creatures {
		if v.IsEvadable() {
			if v.CurrentLocation == le {
				return true
			}
		}
	}
	return false
}

func (gs *GameState) LocationCanBeLeft(le *location.Unit) bool {

	if !gs.LocationHasEnemies(le) {
		return true
	}

	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			creatureBlocksMovement := !v.Aloof && !v.Exhausted
			if creatureBlocksMovement {
				return false
			}
		}
	}

	return true
}

func (gs *GameState) LocationHasHarvest(le *location.Unit) bool {
	for _, v := range gs.HarvestUnits {
		if le == v.StaticLocation {
			return true
		}
	}

	return false
}

func (gs *GameState) EnemiesAtLocation(le *location.Unit) []*creature.Unit {

	found := []*creature.Unit{}
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			found = append(found, v)
		}
	}
	return found
}
func (gs *GameState) EvadableEnemiesAtLocation(le *location.Unit) []*creature.Unit {

	found := []*creature.Unit{}
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation && !v.Aloof && !v.Exhausted {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) HarvestUnitsAtLocation(le *location.Unit) []*harvesting.Unit {

	found := []*harvesting.Unit{}
	for _, v := range gs.HarvestUnits {
		if le == v.StaticLocation {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) PlayersAtLocation(ee *creature.Unit) []*player.Unit {

	found := []*player.Unit{}
	for _, v := range gs.Players {
		if ee.CurrentLocation == v.CurrentLocation {
			found = append(found, v)
		}
	}

	return found
}

func (gs *GameState) PlayerHaveActionsRemaining() bool {
	for _, v := range gs.Players {
		if v.RemainingActions > 0 {
			return true
		}
	}
	return false
}
