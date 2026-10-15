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
	"slices"
)

// CreatureAttack is the wound one creature dealt to one player.
type CreatureAttack struct {
	TargetPlayerID int
	Fainted        bool // the wound made the player faint
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
	gs.Logf("-- Creature phase --")

	for _, v := range slices.Clone(gs.Creatures) {
		if v.CurrentHealth == 0 {
			continue
		}
		from := v.CurrentLocation.ID
		wasExhausted := v.Distracted
		gs.ResolveCreatureMovement(v)
		attacks, preyReactions := gs.ResolveCreatureAttack(v)
		gs.ReconcileDefeats()

		to := v.CurrentLocation.ID
		exhausted := !wasExhausted && v.Distracted
		if from != to || len(attacks) > 0 || exhausted {
			events = append(events, CreatureEvent{
				CreatureID:     v.ID,
				FromLocationID: from,
				ToLocationID:   to,
				Attacks:        attacks,
				Exhausted:      exhausted,
			})
		}
		events = append(events, preyReactions...)
	}

	for _, p := range gs.Players {
		gs.EvaluatePlayerHandPlayability(p)
	}
	return events
}

func (gs *GameState) ResolveCreatureMovement(input *creature.Unit) {
	if !input.Curious && !input.Shy {
		log.Printf("[CREATURE PHASE] %s is neither curious nor shy. It does not move.", input.NameWithID())
		return
	}

	target := gs.DetermineMovementTargetForCreature(input)
	if target == input.CurrentLocation {
		if input.Distracted {
			log.Printf("[CREATURE PHASE] %s is exhausted and can not move.", input.NameWithID())
		} else {
			log.Printf("[CREATURE PHASE] %s would move but is already at its target.", input.NameWithID())
		}
		return
	}

	moveEffect := moving.MoveEffect{Entity: input, Target: target}
	moveEffect.Apply()
	log.Printf("[CREATURE PHASE] %s is moving to %s.", input.NameWithID(), input.CurrentLocation.Name)
	gs.Logf("%s moves to %s", input.NameWithID(), input.CurrentLocation.Name)

}

func (gs *GameState) DetermineMovementTargetForCreature(creature *creature.Unit) *location.Unit {
	if creature.Shy && gs.LocationThreatensShyCreature(creature) {
		return gs.DetermineMovementTargetForShyCreature(creature)
	}
	if creature.Curious {
		return gs.DetermineMovementForCuriousCreature(creature)
	}

	return creature.CurrentLocation
}

// players and other aggressive creatures at the location make shy creatures run
func (gs *GameState) LocationThreatensShyCreature(shy *creature.Unit) bool {
	if gs.LocationHasActivePlayers(shy.CurrentLocation) {
		return true
	}
	for _, other := range gs.EnemiesAtLocation(shy.CurrentLocation) {
		if other != shy && other.Aggressive {
			return true
		}
	}
	return false
}

func (gs *GameState) DetermineMovementTargetForShyCreature(creature *creature.Unit) *location.Unit {
	creatureLocation := creature.CurrentLocation

	if !gs.LocationThreatensShyCreature(creature) {
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
		log.Printf("[CREATURE PHASE] %s is shy and wants to escape but can't. It becomes enraged.", creature.NameWithID())
		return creatureLocation
	}

	target := possibleLocations[rand.IntN(len(possibleLocations))]
	log.Printf("[CREATURE PHASE] %s is shy and runs away to %s.", creature.NameWithID(), target.Name)
	return target

}

func (gs *GameState) DetermineMovementForCuriousCreature(creature *creature.Unit) *location.Unit {

	lowestDistance := math.MaxInt
	closestLocations := []*location.Unit{}
	for _, targetLocation := range gs.CuriosityTargetLocations(creature) {
		distance := creature.CurrentLocation.DistanceTo(targetLocation)
		if distance == lowestDistance {
			closestLocations = append(closestLocations, targetLocation)
		} else if distance < lowestDistance {
			closestLocations = []*location.Unit{targetLocation}
			lowestDistance = distance
		}
	}

	if len(closestLocations) == 0 {
		return creature.CurrentLocation
	}

	//selects a random entry from the array
	//if only one entry, it's that entry
	selectedIndex := rand.IntN(len(closestLocations))

	return closestLocations[selectedIndex]
}

// hunters are curious about prey and players, other curious creatures about anything not aggressive and players
func (gs *GameState) CuriosityTargetLocations(curious *creature.Unit) []*location.Unit {
	locations := []*location.Unit{}
	for _, p := range gs.Players {
		if !p.Fainted() {
			locations = append(locations, p.CurrentLocation)
		}
	}
	for _, other := range gs.Creatures {
		if other == curious {
			continue
		}
		if (curious.Aggressive && other.Prey) || (!curious.Aggressive && !other.Aggressive) {
			locations = append(locations, other.CurrentLocation)
		}
	}
	return locations
}

// returns the wounds dealt to players and the moves of prey that fled from the attack
func (gs *GameState) ResolveCreatureAttack(creature *creature.Unit) ([]CreatureAttack, []CreatureEvent) {
	if creature.Bonded {
		log.Printf("[CREATURE PHASE] %s is bonded and doesn't attack.", creature.NameWithID())
		return nil, nil
	}
	if !creature.Aggressive {
		log.Printf("[CREATURE PHASE] %s is not aggressive and doesn't attack.", creature.NameWithID())
		return nil, nil
	}
	if creature.Distracted {
		log.Printf("[CREATURE PHASE] %s is exhausted and doesn't attack.", creature.NameWithID())
		return nil, nil
	}
	return gs.ResolveAttackForCreature(creature)
}

func (gs *GameState) ResolveAttackForCreature(attacker *creature.Unit) ([]CreatureAttack, []CreatureEvent) {
	playersInRange := []*player.Unit{}
	for _, p := range gs.PlayersAtLocation(attacker) {
		if !p.Fainted() {
			playersInRange = append(playersInRange, p)
		}
	}
	preyInRange := gs.PreyAtLocation(attacker)

	if len(playersInRange) == 0 && len(preyInRange) == 0 {
		log.Printf("[CREATURE PHASE] %s has no targets for an attack.\n", attacker.NameWithID())
		return nil, nil
	}

	attacker.Distracted = true
	attacks := gs.AttackPlayers(attacker, playersInRange)
	preyReactions := gs.AttackPrey(attacker, preyInRange)
	if attacker.CurrentHealth > 0 {
		gs.Logf("%s is exhausted", attacker.NameWithID())
	}
	return attacks, preyReactions
}

func (gs *GameState) AttackPlayers(attacker *creature.Unit, targets []*player.Unit) []CreatureAttack {
	attacks := []CreatureAttack{}
	for _, target := range targets {
		effect := combat.DamageEffect{Target: target}
		effect.Apply()
		log.Printf("[CREATURE PHASE] %s wounds %s.", attacker.NameWithID(), target.Name)
		attacks = append(attacks, CreatureAttack{TargetPlayerID: target.ID, Fainted: target.Fainted()})
		gs.logCreatureAttack(attacker.NameWithID(), target)
	}
	return attacks
}

// aggressive prey retaliates once, other prey flees. A retaliation is answered by nothing.
func (gs *GameState) AttackPrey(attacker *creature.Unit, targets []*creature.Unit) []CreatureEvent {
	fleeEvents := []CreatureEvent{}
	for _, prey := range targets {
		if attacker.CurrentHealth == 0 {
			break
		}
		effect := combat.DamageEffect{Target: prey}
		effect.Apply()
		log.Printf("[CREATURE PHASE] %s attacks %s.", attacker.NameWithID(), prey.NameWithID())
		gs.logCreatureHit("attacks", attacker, prey)

		if prey.CurrentHealth == 0 {
			continue
		}
		if prey.Aggressive {
			retaliation := combat.DamageEffect{Target: attacker}
			retaliation.Apply()
			log.Printf("[CREATURE PHASE] %s retaliates against %s.", prey.NameWithID(), attacker.NameWithID())
			gs.logCreatureHit("retaliates against", prey, attacker)
			continue
		}
		if fleeEvent, fled := gs.FleeToSaferLocation(prey); fled {
			fleeEvents = append(fleeEvents, fleeEvent)
		}
	}
	return fleeEvents
}

func (gs *GameState) PreyAtLocation(attacker *creature.Unit) []*creature.Unit {
	found := []*creature.Unit{}
	for _, other := range gs.EnemiesAtLocation(attacker.CurrentLocation) {
		if other != attacker && other.Prey && other.CurrentHealth > 0 {
			found = append(found, other)
		}
	}
	return found
}

// prefers connected locations without aggressive creatures, takes any connected location otherwise
func (gs *GameState) FleeToSaferLocation(prey *creature.Unit) (CreatureEvent, bool) {
	from := prey.CurrentLocation
	connections := from.OutgoingConnections
	if len(connections) == 0 {
		return CreatureEvent{}, false
	}

	safeLocations := []*location.Unit{}
	for _, l := range connections {
		if !gs.LocationHasAggressiveCreatures(l) {
			safeLocations = append(safeLocations, l)
		}
	}
	options := connections
	if len(safeLocations) > 0 {
		options = safeLocations
	}

	target := options[rand.IntN(len(options))]
	prey.MoveTo(target)
	log.Printf("[CREATURE PHASE] %s flees to %s.", prey.NameWithID(), target.Name)
	gs.Logf("%s flees to %s", prey.NameWithID(), target.Name)
	return CreatureEvent{CreatureID: prey.ID, FromLocationID: from.ID, ToLocationID: target.ID}, true
}

func (gs *GameState) LocationHasAggressiveCreatures(le *location.Unit) bool {
	for _, v := range gs.EnemiesAtLocation(le) {
		if v.Aggressive {
			return true
		}
	}
	return false
}

func (gs *GameState) logCreatureHit(verb string, attacker *creature.Unit, target *creature.Unit) {
	gs.Logf("%s %s %s (%d/%d)", attacker.NameWithID(), verb, target.NameWithID(), target.CurrentHealth, target.MaxHealth)
	if target.CurrentHealth == 0 {
		gs.Logf("%s is defeated", target.NameWithID())
	}
}

func (gs *GameState) logCreatureAttack(attackerName string, target *player.Unit) {
	gs.Logf("%s wounds %s (%d/%d)", attackerName, target.Name, target.Wounds, player.MaxWounds)
	if target.Fainted() {
		gs.Logf("%s faints", target.Name)
	}
}
