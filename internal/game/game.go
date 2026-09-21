package game

import (
	"aeons/internal/enemy"
	"aeons/internal/location"
	"aeons/internal/player"
)

type GameState struct {
	Locations   []*location.LocationEntity
	Players     []*player.Unit
	Enemies     []*enemy.EnemyEntity
	TurnCounter int
}

func (gs *GameState) ReconcileDefeats() {
	alivePlayers := gs.Players[:0]

	for _, v := range gs.Players {
		if v.CurrentHealth > 0 {
			alivePlayers = append(alivePlayers, v)
		}
	}

	aliveEnemies := gs.Enemies[:0]
	for _, v := range gs.Enemies {
		if v.CurrentHealth > 0 {
			aliveEnemies = append(aliveEnemies, v)
		}
	}
	gs.Enemies = aliveEnemies
	gs.Players = alivePlayers
}

func (gs *GameState) StartNewTurn() {
	for _, v := range gs.Players {
		v.RemainingActions = 3
	}
	gs.TurnCounter += 1
}

func (gs *GameState) LocationHasEnemies(le *location.LocationEntity) bool {
	for _, v := range gs.Enemies {
		if le == v.CurrentLocation {
			return true
		}
	}

	return false
}

func (gs *GameState) EnemiesAtLocation(le *location.LocationEntity) []*enemy.EnemyEntity {

	found := []*enemy.EnemyEntity{}
	for _, v := range gs.Enemies {
		if le == v.CurrentLocation {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) PlayersAtLocation(ee *enemy.EnemyEntity) []*player.Unit {

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
