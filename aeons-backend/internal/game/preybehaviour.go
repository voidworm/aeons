package game

import (
	"aeons/internal/creature"
	"aeons/internal/location"
	"math/rand/v2"
	"slices"
)

const actionsPerActivation = 2

// moveCreatureTo moves the creature one step and remembers where it came from.
func (gs *GameState) moveCreatureTo(c *creature.Unit, to *location.Unit) CreatureEvent {
	from := c.CurrentLocation
	c.LastLocation = from
	c.MoveTo(to)
	return CreatureEvent{CreatureID: c.ID, FromLocationID: from.ID, ToLocationID: to.ID}
}

// A hidden prey stays hidden and relieves distress. Otherwise an exhausted prey readies and moves once.
// A ready prey moves once, or up to twice while it is under distress.
func (gs *GameState) ActivatePrey(prey *creature.Unit) []CreatureEvent {
	events := []CreatureEvent{}
	if prey.Hidden {
		if prey.Distress > 0 {
			gs.relieveWhileHidden(prey)
		}
		return events
	}

	moves := 1
	if prey.Exhausted {
		prey.Exhausted = false
		gs.Logf("%s readies", prey.NameWithID())
		events = append(events, readyEvent(prey))
	} else if prey.Distress > 0 {
		moves = actionsPerActivation
	}
	for range moves {
		gs.hideIfDistressedAtHome(prey)
		if prey.Hidden {
			break
		}
		if event, moved := gs.movePrey(prey); moved {
			events = append(events, event)
		}
	}
	gs.hideIfDistressedAtHome(prey)
	return events
}

// prey hides from hunters, and on its own turn only while it is distressed
func (gs *GameState) hideIfDistressedAtHome(prey *creature.Unit) {
	if prey.Distress > 0 {
		gs.hideAtHome(prey)
	}
}

// ReconcileHiding brings hidden prey out of hiding once it is without distress and no predator is at its location.
func (gs *GameState) ReconcileHiding() {
	for _, c := range gs.Creatures {
		if c.Hidden && c.Distress == 0 && !gs.LocationHasPredator(c.CurrentLocation, c) {
			c.Unhide()
			gs.Logf("%s comes out of hiding", c.NameWithID())
		}
	}
}

func (gs *GameState) LocationHasPredator(l *location.Unit, except *creature.Unit) bool {
	return slices.ContainsFunc(gs.EnemiesAtLocation(l), func(c *creature.Unit) bool {
		return c != except && c.Role.IsPredator()
	})
}

func (gs *GameState) LocationHasPlant(l *location.Unit) bool {
	return slices.ContainsFunc(gs.EnemiesAtLocation(l), func(c *creature.Unit) bool { return c.Role.IsPlant() })
}

func readyEvent(c *creature.Unit) CreatureEvent {
	where := c.CurrentLocation.ID
	return CreatureEvent{CreatureID: c.ID, FromLocationID: where, ToLocationID: where, Ready: true}
}

func (gs *GameState) relieveWhileHidden(prey *creature.Unit) {
	prey.Relieve(1)
	gs.Logf("%s relieves distress while hidden (%d/%d)", prey.NameWithID(), prey.Distress, prey.MaxDistress)
}

func (gs *GameState) movePrey(prey *creature.Unit) (CreatureEvent, bool) {
	destination := gs.preyDestination(prey)
	if destination == nil {
		return CreatureEvent{}, false
	}
	event := gs.moveCreatureTo(prey, destination)
	gs.Logf("%s moves to %s", prey.NameWithID(), destination.Name)
	return event, true
}

// distressed prey heads for its burrow or, without one, away from hunters, calm prey wanders around its home
func (gs *GameState) preyDestination(prey *creature.Unit) *location.Unit {
	switch {
	case prey.Distress == 0:
		return gs.wanderDestination(prey)
	case prey.CanHide:
		return prey.CurrentLocation.NextStepTo(prey.Home)
	default:
		return randomLocation(gs.neighboursWithoutHunters(prey), nil)
	}
}

// anywhere within the radius but where it just was, unless that is the only way out.
// Prey that is no predator prefers locations with a plant.
func (gs *GameState) wanderDestination(prey *creature.Unit) *location.Unit {
	options := slices.DeleteFunc(slices.Clone(prey.CurrentLocation.OutgoingConnections), func(l *location.Unit) bool {
		return prey.Home.StepsTo(l) > prey.Radius
	})
	options = withoutLocation(options, prey.LastLocation)
	if prey.Role == creature.Prey {
		withPlants := slices.DeleteFunc(slices.Clone(options), func(l *location.Unit) bool { return !gs.LocationHasPlant(l) })
		if len(withPlants) > 0 {
			options = withPlants
		}
	}
	return randomLocation(options, nil)
}

// the options without the avoided location, or all options if there is no other
func withoutLocation(options []*location.Unit, avoided *location.Unit) []*location.Unit {
	others := slices.DeleteFunc(slices.Clone(options), func(l *location.Unit) bool { return l == avoided })
	if len(others) == 0 {
		return options
	}
	return others
}

// avoids the location when there is another option
func randomLocation(options []*location.Unit, avoided *location.Unit) *location.Unit {
	options = withoutLocation(options, avoided)
	if len(options) == 0 {
		return nil
	}
	return options[rand.IntN(len(options))]
}

func (gs *GameState) neighboursWithoutHunters(prey *creature.Unit) []*location.Unit {
	return slices.DeleteFunc(slices.Clone(prey.CurrentLocation.OutgoingConnections), func(l *location.Unit) bool {
		return gs.LocationHasHunterOf(l, prey)
	})
}

func (gs *GameState) LocationHasHunterOf(l *location.Unit, prey *creature.Unit) bool {
	for _, c := range gs.EnemiesAtLocation(l) {
		if c != prey && c.CanHunt(prey) {
			return true
		}
	}
	return false
}

// a prey that can hide and is at its burrow, ready and not yet hidden, hides
func (gs *GameState) hideAtHome(prey *creature.Unit) bool {
	if !prey.CanHide || prey.Hidden || prey.Exhausted || prey.CurrentLocation != prey.Home {
		return false
	}
	prey.Hide(!prey.CurrentLocation.Veiled)
	gs.Logf("%s hides", prey.NameWithID())
	return true
}

// ReactToArrival makes every ready prey at the hunter's location run from it.
func (gs *GameState) ReactToArrival(hunter *creature.Unit) []CreatureEvent {
	events := []CreatureEvent{}
	for _, other := range gs.EnemiesAtLocation(hunter.CurrentLocation) {
		scared := other != hunter && !other.Bonded && !other.Hidden && !other.Exhausted && hunter.CanHunt(other)
		if !scared {
			continue
		}
		if event, fled := gs.Flee(other); fled {
			other.Exhausted = true
			event.Exhausted = true
			events = append(events, event)
		}
	}
	return events
}

// Flee hides the prey if it is at its burrow, otherwise it steps away.
func (gs *GameState) Flee(prey *creature.Unit) (CreatureEvent, bool) {
	if gs.hideAtHome(prey) {
		return CreatureEvent{}, false
	}
	return gs.StepAway(prey)
}

// StepAway moves the prey one step towards its burrow. Without one, or at the burrow, it moves to a neighbour without hunters.
func (gs *GameState) StepAway(prey *creature.Unit) (CreatureEvent, bool) {
	if prey.Exhausted {
		return CreatureEvent{}, false
	}
	var destination *location.Unit
	if prey.CanHide {
		destination = prey.CurrentLocation.NextStepTo(prey.Home)
	}
	if destination == nil {
		destination = randomLocation(gs.neighboursWithoutHunters(prey), nil)
	}
	if destination == nil {
		return CreatureEvent{}, false
	}
	event := gs.moveCreatureTo(prey, destination)
	gs.Logf("%s flees to %s", prey.NameWithID(), destination.Name)
	return event, true
}
