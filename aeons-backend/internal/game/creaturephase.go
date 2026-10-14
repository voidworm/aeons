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

// CreatureAttack is the damage one creature dealt to one player.
type CreatureAttack struct {
	TargetPlayerID int
	Damage         int
}

// CreatureEvent records what one creature did in the creature phase.
// Creatures that neither moved nor attacked produce no event.
type CreatureEvent struct {
	CreatureID     int
	FromLocationID int
	ToLocationID   int
	Attacks        []CreatureAttack
	Exhausted      bool // the creature became exhausted during this event
}

// ResolveCreaturePhase runs the creature phase and returns what happened, in order.
func (gs *GameState) ResolveCreaturePhase() []CreatureEvent {
	events := []CreatureEvent{}

	for _, v := range gs.Creatures {
		from := v.CurrentLocation.ID
		wasExhausted := v.Exhausted
		gs.ResolveCreatureMovement(v)
		attacks := gs.ResolveCreatureAttack(v)
		gs.ReconcileDefeats()

		to := v.CurrentLocation.ID
		if from == to && len(attacks) == 0 {
			continue
		}
		events = append(events, CreatureEvent{
			CreatureID:     v.ID,
			FromLocationID: from,
			ToLocationID:   to,
			Attacks:        attacks,
			Exhausted:      !wasExhausted && v.Exhausted,
		})
	}

	for _, p := range gs.Players {
		gs.EvaluatePlayerHandPlayability(p)
	}
	return events
}

func (gs *GameState) ResolveCreatureMovement(input *creature.Unit) {
	if !input.Curious && !input.Shy {
		log.Printf("[CREATURE PHASE] %s is neither curious nor shy. It does not move.", input.Name)
		return
	}

	target := gs.DetermineMovementTargetForCreature(input)
	if target == input.CurrentLocation {
		if input.Exhausted {
			log.Printf("[CREATURE PHASE] %s is exhausted and can not move.", input.Name)
		} else {
			log.Printf("[CREATURE PHASE] %s would move but is already at its target.", input.Name)
		}
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
		return gs.DetermineMovementTargetForShyCreature(creature)
	}

	return nil //cannot happen
}

func (gs *GameState) DetermineMovementTargetForShyCreature(creature *creature.Unit) *location.Unit {
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

	target := possibleLocations[rand.IntN(len(possibleLocations))]
	log.Printf("[CREATURE PHASE] %s is shy and runs away to %s.", creature.Name, target.Name)
	return target

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

func (gs *GameState) ResolveCreatureAttack(creature *creature.Unit) []CreatureAttack {
	if !creature.Aloof && !creature.Exhausted {
		return gs.ResolveAttackForCreature(creature)
	} else if creature.Aloof {
		log.Printf("[CREATURE PHASE] %s is aloof and doesn't attack.", creature.Name)
	} else if creature.Exhausted {
		log.Printf("[CREATURE PHASE] %s is exhausted and doesn't attack.", creature.Name)
	}
	return nil
}

func (gs *GameState) ResolveAttackForCreature(creature *creature.Unit) []CreatureAttack {
	inRange := gs.PlayersAtLocation(creature)

	if len(inRange) == 0 {
		log.Printf("[CREATURE PHASE] %s has no targets for an attack.\n", creature.Name)
		return nil
	}

	attacks := []CreatureAttack{}
	for _, target := range inRange {
		effect := combat.DamageEffect{Source: creature, Target: target}
		creature.Exhausted = true
		effect.Apply()
		log.Printf("[CREATURE PHASE] %s attacks %s for %d damage.", creature.Name, target.Name, creature.Damage)
		attacks = append(attacks, CreatureAttack{TargetPlayerID: target.ID, Damage: creature.Damage})
	}
	return attacks
}
