package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/location"
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
	if !input.Curious && !input.Shy {
		log.Printf("[CREATURE PHASE] %s is neither curious nor shy. It does not move.", input.Name)
		return
	}

	target := gs.DetermineMovementTargetForCreature(input)
	if target == input.CurrentLocation {
		log.Printf("[CREATURE PHASE] %s has no need to move.", input.Name)
		return
	}

	moveEffect := moving.MoveEffect{Entity: input, Target: target}
	moveEffect.Apply()
	log.Printf("[CREATURE PHASE] %s is moving to %s.", input.Name, input.CurrentLocation.Name)

}

func (gs *GameState) DetermineMovementTargetForCreature(creature *creature.Unit) *location.Unit {
	if creature.Curious {
		return gs.DetermineMovementForCuriousCreature(creature)
	} else if creature.Shy {
		return gs.DetermineMovementTargetForCreature(creature)
	}

	return nil //cannot happen
}

func (gs *GameState) DetermineMovementForShyCreature(creature *creature.Unit) *location.Unit {
	creatureLocation := creature.CurrentLocation

	if !gs.LocationHasPlayers(creatureLocation) {
		return creatureLocation
	}

	possibleLocations := []*location.Unit{}
	for _, v := range creatureLocation.OutgoingConnections {
		if !gs.LocationHasEnemies(v) {
			possibleLocations = append(possibleLocations, v)
		}
	}

	if len(possibleLocations) == 0 {
		creature.Enraged = true
		log.Printf("[CREATURE PHASE] %s is shy and wants to escape but can't. It becomes enraged.", creature.Name)
		return creatureLocation
	}

	return possibleLocations[rand.IntN(len(possibleLocations))]
}

func (gs *GameState) DetermineMovementForCuriousCreature(creature *creature.Unit) *location.Unit {

	lowestDistance := math.MaxInt
	lowestPlayers := []*player.Unit{}
	for _, currentPlayer := range gs.Players {
		distance := creature.CurrentLocation.DistanceTo(currentPlayer.CurrentLocation)
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

	return lowestPlayers[selectedIndex].CurrentLocation
}

func (gs *GameState) ResolveCreatureAttack(creature *creature.Unit) {
	if !creature.Aloof || !creature.Exhausted {
		gs.ResolveAttackForCreature(creature)
	} else if creature.Aloof {
		log.Printf("[CREATURE PHASE] %s is aloof and doesn't attack.", creature.Name)
	} else if creature.Exhausted {
		log.Printf("[CREATURE PHASE] %s is exhausted and doesn't attack.", creature.Name)
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
		creature.Exhausted = true
		effect.Apply()
		log.Printf("[CREATURE PHASE] %s attacks %s for %d damage.", creature.Name, target.Name, creature.Damage)
	}
}
