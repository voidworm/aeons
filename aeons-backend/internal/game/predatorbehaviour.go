package game

import (
	"aeons/internal/creature"
	"aeons/internal/location"
)

// An exhausted predator readies. A ready one does two actions, each attacks if there is something to hunt
// at its location and moves otherwise. Only a distract card exhausts a predator.
func (gs *GameState) ActivatePredator(hunter *creature.Unit) []CreatureEvent {
	if hunter.Exhausted {
		hunter.Exhausted = false
		gs.Logf("%s readies", hunter.NameWithID())
		return []CreatureEvent{readyEvent(hunter)}
	}

	events := []CreatureEvent{}
	for range actionsPerActivation {
		if targets := gs.HuntableAt(hunter); len(targets) > 0 {
			events = append(events, gs.attackOnce(hunter, pickTarget(targets))...)
		} else {
			events = append(events, gs.huntMove(hunter)...)
		}
	}
	return events
}

func (gs *GameState) attackOnce(hunter *creature.Unit, target creature.Quarry) []CreatureEvent {
	attack, woundedPlayer := gs.Attack(hunter, target)
	if !woundedPlayer {
		return nil
	}
	where := hunter.CurrentLocation.ID
	return []CreatureEvent{{CreatureID: hunter.ID, FromLocationID: where, ToLocationID: where, Attacks: []CreatureAttack{attack}}}
}

// The hunter locks on to its target the moment it enters a location with prey it hunts. Prey there flees.
func (gs *GameState) huntMove(hunter *creature.Unit) []CreatureEvent {
	destination := gs.huntDestination(hunter)
	if destination == nil {
		return nil
	}
	event := gs.moveCreatureTo(hunter, destination)
	gs.Logf("%s moves to %s", hunter.NameWithID(), destination.Name)

	if targets := gs.HuntableAt(hunter); len(targets) > 0 {
		hunter.Chasing = pickTarget(targets)
		gs.Logf("%s locks on to its prey", hunter.NameWithID())
	}
	return append([]CreatureEvent{event}, gs.ReactToArrival(hunter)...)
}

// towards the target it chases, anywhere but where it just was otherwise
func (gs *GameState) huntDestination(hunter *creature.Unit) *location.Unit {
	if hunter.Chasing != nil && stillChasing(hunter) {
		return hunter.CurrentLocation.NextStepTo(hunter.Chasing.Location())
	}
	hunter.Chasing = nil
	return randomLocation(hunter.CurrentLocation.OutgoingConnections, hunter.LastLocation)
}

// a chase ends when the target hides, is gone, or gets two moves away
func stillChasing(hunter *creature.Unit) bool {
	target := hunter.Chasing
	return hunter.CanHunt(target) && hunter.CurrentLocation.StepsTo(target.Location()) < 2
}
