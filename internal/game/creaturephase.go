package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/moving"
	"aeons/internal/player"
	"log"
	"math"
	"math/rand/v2"
)

func (gs *GameState) ResolveCreaturePhase() {

	for _, v := range gs.Creatures {
		gs.ResolveCreatureMovement(v)
		gs.ResolveCreatureAttack(v)
		gs.ReconcileDefeats()
	}
}

func (gs *GameState) ResolveCreatureMovement(input *creature.Unit) {
	if !input.Hunter {
		log.Printf("[CREATURE PHASE] %s is not a hunter and does not move.", input.Name)
		return
	}

	target := gs.DetermineMovementTargetForCreature(input)
	moveEffect := moving.MoveEffect{Entity: input, Target: target.CurrentLocation}
	moveEffect.Apply()
	log.Printf("[CREATURE PHASE] %s is moving to %s.", input.Name, input.CurrentLocation.Name)

}

func (gs *GameState) DetermineMovementTargetForCreature(enemy *creature.Unit) *player.Unit {

	lowestDistance := math.MaxInt
	lowestPlayers := []*player.Unit{}
	for _, currentPlayer := range gs.Players {
		distance := enemy.CurrentLocation.DistanceTo(currentPlayer.CurrentLocation)
		if distance == lowestDistance {
			lowestPlayers = append(lowestPlayers, currentPlayer)
		} else if distance < lowestDistance {
			lowestPlayers = []*player.Unit{currentPlayer}
			lowestDistance = distance
		}
	}

	//selects a random entry from the array
	//if only one entry, it's that entry
	selectedIndex := rand.IntN(len(lowestPlayers))
	return lowestPlayers[selectedIndex]
}

func (gs *GameState) ResolveCreatureAttack(creature *creature.Unit) {
	if !creature.Aloof {
		gs.ResolveAttackForCreature(creature)
	} else {
		log.Printf("[CREATURE PHASE] %s is aloof and doesn't attack.", creature.Name)
	}
}

func (gs *GameState) ResolveAttackForCreature(creature *creature.Unit) {
	inRange := gs.PlayersAtLocation(creature)

	if len(inRange) == 0 {
		log.Printf("[CREATURE PHASE] %s has no targets for an attack.\n", creature.Name)
		return
	}

	for _, target := range inRange {
		effect := combat.DamageEffect{Source: creature, Target: target}
		effect.Apply()
		log.Printf("[CREATURE PHASE] %s attacks %s for %d damage.", creature.Name, target.Name, creature.Damage)
	}
}
