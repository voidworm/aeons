package game

import (
	"aeons/internal/creature"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/player"
)

type GameState struct {
	Locations    []*location.LocationEntity
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

func (gs *GameState) StartNewTurn() {
	for _, v := range gs.Players {
		v.RemainingActions = 3
	}
	gs.TurnCounter += 1
}

func (gs *GameState) LocationHasEnemies(le *location.LocationEntity) bool {
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			return true
		}
	}

	return false
}

func (gs *GameState) LocationHasHarvest(le *location.LocationEntity) bool {
	for _, v := range gs.HarvestUnits {
		if le == v.StaticLocation {
			return true
		}
	}

	return false
}

func (gs *GameState) EnemiesAtLocation(le *location.LocationEntity) []*creature.Unit {

	found := []*creature.Unit{}
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			found = append(found, v)
		}
	}
	return found
}
func (gs *GameState) HarvestUnitsAtLocation(le *location.LocationEntity) []*harvesting.Unit {

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
