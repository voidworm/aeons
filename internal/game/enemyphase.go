package game

import (
	"aeons/internal/enemy"
	"aeons/internal/location"
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

func (gs *GameState) ResolveEnemyAttacks(enemy *enemy.EnemyEntity) {
	if !enemy.Aloof {
		gs.ResolveAttackForEnemy(enemy)
	} else {
		log.Printf("%s is aloof and doesn't attack.", enemy.Name)
	}
}

func (gs *GameState) ResolveEnemyMovement(input *enemy.EnemyEntity) {

	if input.Hunter {

		target, steps := gs.DetermineHuntingTargetForEnemy(input)

		if len(steps) == 1 {
			log.Printf("%s is already at its prey location and does not need to move.\n", input.Name)
		} else {
			targetLocation := steps[1]
			eme := enemy.EnemyMoveEffect{TargetEnemy: input, TargetLocation: targetLocation}
			eme.Apply()
			log.Printf("%s has %s as target and will move to %s to hunt its prey.\n", input.Name, target.Name, steps[1])
		}

	} else {
		log.Printf("%s is not a hunter and does not move.", input.Name)
	}
}

func (gs *GameState) ResolveAttackForEnemy(enemy *enemy.EnemyEntity) {
	inRange := gs.PlayersAtLocation(enemy)

	if len(inRange) == 0 {
		log.Printf("%s has no targets for an attack and doesn't attack.\n", enemy.Name)
		return
	}

	for _, target := range inRange {
		target.TakeDamage(enemy.Damage)
		log.Printf("%s attacks %s for %d damage down to %d/%d health.\n", enemy.Name, target.Name, enemy.Damage, target.CurrentHealth, target.MaxHealth)
		if target.CurrentHealth == 0 {
			log.Printf("%s has defeated %s!\n", target.Name, target.Name)
		}
	}
}

func (gs *GameState) DetermineHuntingTargetForEnemy(enemy *enemy.EnemyEntity) (*player.Player, []*location.LocationEntity) {
	CurrentTargets := []*player.Player{}
	CurrentSteps := [][]*location.LocationEntity{}
	CurrentMinDistance := math.MaxInt

	for _, v := range gs.Players {
		steps := enemy.Location.GetShortestPathTo(v.Location)

		if len(steps) == CurrentMinDistance {
			CurrentTargets = append(CurrentTargets, v)
			CurrentSteps = append(CurrentSteps, steps)
		}

		if len(steps) < CurrentMinDistance {
			CurrentMinDistance = len(steps)
			CurrentTargets = []*player.Player{v}
			CurrentSteps = [][]*location.LocationEntity{steps}
		}
	}

	n := rand.IntN(len(CurrentTargets))
	finalTarget := CurrentTargets[n]
	finalSteps := CurrentSteps[n]
	return finalTarget, finalSteps
}
