package game

import (
	"aeons/internal/creature"
	"aeons/internal/player"
	"math/rand/v2"
	"slices"
)

// CreatureAttack is the wound one creature dealt to one player.
type CreatureAttack struct {
	TargetPlayerID int
	Fainted        bool // the wound made the player faint
}

// CreatureEvent records what one creature did in the world phase.
// Creatures that neither moved nor attacked produce no event.
type CreatureEvent struct {
	CreatureID     int
	FromLocationID int
	ToLocationID   int
	Attacks        []CreatureAttack
	Exhausted      bool // the creature became exhausted during this event
	Ready          bool // the creature readied during this event, before it moves
}

// ResolveWorldPhase runs the world phase and returns what happened, in order.
// The creatures act in random order, bonded creatures and plants do not act.
func (gs *GameState) ResolveWorldPhase() []CreatureEvent {
	events := []CreatureEvent{}
	gs.Logf("-- World phase --")

	activationOrder := slices.Clone(gs.Creatures)
	rand.Shuffle(len(activationOrder), func(i, j int) {
		activationOrder[i], activationOrder[j] = activationOrder[j], activationOrder[i]
	})
	exhaustedAtStart := map[*creature.Unit]bool{}
	for _, c := range gs.Creatures {
		exhaustedAtStart[c] = c.Exhausted
	}
	heldByPredators := []*creature.Unit{}
	for _, c := range activationOrder {
		if c.Bonded || c.Role.IsPlant() || !slices.Contains(gs.Creatures, c) {
			continue
		}
		if c.Exhausted && !exhaustedAtStart[c] {
			continue // exhausted during this phase, it readies in the next one
		}
		if c.Role.IsPredator() {
			events = append(events, gs.ActivatePredator(c)...)
		} else {
			if c.Hidden && c.Distress == 0 {
				heldByPredators = append(heldByPredators, c)
			}
			events = append(events, gs.ActivatePrey(c)...)
		}
		gs.ReconcileDefeats()
		gs.ReconcileHiding()
	}
	events = append(events, gs.MoveReleasedPrey(heldByPredators)...)

	gs.ReconcileVision()
	gs.RefreshHands()
	return events
}

// Prey that could not move because a predator held it in hiding moves once if it has come out since.
func (gs *GameState) MoveReleasedPrey(heldByPredators []*creature.Unit) []CreatureEvent {
	events := []CreatureEvent{}
	for _, prey := range heldByPredators {
		if prey.Hidden || prey.Exhausted || prey.Bonded || !slices.Contains(gs.Creatures, prey) {
			continue
		}
		if event, moved := gs.movePrey(prey); moved {
			events = append(events, event)
		}
	}
	return events
}

func (gs *GameState) logCreatureHit(verb string, attacker *creature.Unit, target *creature.Unit) {
	gs.Logf("%s %s %s (%d/%d distress)", attacker.NameWithID(), verb, target.NameWithID(), target.Distress, target.MaxDistress)
}

func (gs *GameState) logCreatureAttack(attackerName string, target *player.Unit) {
	gs.Logf("%s wounds %s (%d/%d)", attackerName, target.Name, target.Wounds, player.MaxWounds)
	if target.Fainted() {
		gs.Logf("%s faints", target.Name)
	}
}
