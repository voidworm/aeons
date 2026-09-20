package game

import (
	"aeons/internal/combat"
	"aeons/internal/enemy"
	"aeons/internal/moving"
	"aeons/internal/player"
	"log"
	"math"
	"math/rand/v2"
)

func (gs *GameState) ResolveEnemyPhase() {
	for _, v := range gs.Enemies {
		gs.ResolveEnemyMovement(v)
		gs.ResolveEnemyAttacks(v)
		gs.ReconcileDefeats()
	}
}

func (gs *GameState) ResolveEnemyMovement(input *enemy.EnemyEntity) {
	if !input.Hunter {
		log.Printf("%s is not a hunter and does not move.", input.Name)
		return
	}

	target := gs.DetermineHuntingTargetForEnemy(input)
	moveEffect := moving.MoveEffect{Entity: input, Target: target.CurrentLocation}
	moveEffect.Apply()

}

func (gs *GameState) DetermineHuntingTargetForEnemy(enemy *enemy.EnemyEntity) *player.Player {

	lowestDistance := math.MaxInt
	lowestPlayers := []*player.Player{}
	for _, currentPlayer := range gs.Players {
		distance := enemy.CurrentLocation.DistanceTo(currentPlayer.CurrentLocation)
		if distance == lowestDistance {
			lowestPlayers = append(lowestPlayers, currentPlayer)
		} else if distance < lowestDistance {
			lowestPlayers = []*player.Player{currentPlayer}
			lowestDistance = distance
		}
	}

	//selects a random entry from the array
	//if only one entry, it's that entry
	selectedIndex := rand.IntN(len(lowestPlayers))
	return lowestPlayers[selectedIndex]
}

func (gs *GameState) ResolveEnemyAttacks(enemy *enemy.EnemyEntity) {
	if !enemy.Aloof {
		gs.ResolveAttackForEnemy(enemy)
	} else {
		log.Printf("%s is aloof and doesn't attack.", enemy.Name)
	}
}

func (gs *GameState) ResolveAttackForEnemy(enemy *enemy.EnemyEntity) {
	inRange := gs.PlayersAtLocation(enemy)

	if len(inRange) == 0 {
		log.Printf("%s has no targets for an attack and doesn't attack.\n", enemy.Name)
		return
	}

	for _, target := range inRange {
		effect := combat.DamageEffect{Source: enemy, Target: target}
		effect.Apply()
	}
}
