package game

import (
	"aeons/internal/collecting"
	"aeons/internal/creature"
	"aeons/internal/player"
	"math/rand/v2"
	"slices"
)

const (
	MaxEnergy   = 4
	restActions = 3 // the creature activations that pass while the players rest
)

// CreatureAttack is the wound one creature dealt to one player.
type CreatureAttack struct {
	TargetPlayerID int
	Fainted        bool // the wound made the player faint
}

// CreatureEvent records what one creature did in its activation.
// Creatures that neither moved nor attacked produce no event.
type CreatureEvent struct {
	CreatureID     int
	FromLocationID int
	ToLocationID   int
	Attacks        []CreatureAttack
	Exhausted      bool // the creature became exhausted during this event
	Ready          bool // the creature readied during this event, before it moves
}

// A round draws every creature that can act in random order. Each player action activates the next creature
// of that order, the creatures then activate once more in reverse order. Creatures a b c act as a b c c b a.
// The top of both stacks is at index 0.
func (gs *GameState) StartRound() {
	gs.ForwardStack = slices.DeleteFunc(slices.Clone(gs.Creatures), func(c *creature.Unit) bool { return !gs.canAct(c) })
	rand.Shuffle(len(gs.ForwardStack), func(i, j int) {
		gs.ForwardStack[i], gs.ForwardStack[j] = gs.ForwardStack[j], gs.ForwardStack[i]
	})
	gs.ReverseStack = nil
}

// the creatures in the order they are going to activate
func (gs *GameState) UpcomingActivations() []*creature.Unit {
	reversedForward := slices.Clone(gs.ForwardStack)
	slices.Reverse(reversedForward)
	return slices.Concat(gs.ForwardStack, reversedForward, gs.ReverseStack)
}

// ActivateNextCreature lets the next creature on the stacks take its turn, the events collect in the game state.
// When the last creature has acted the round ends and a new one starts.
func (gs *GameState) ActivateNextCreature() {
	if c := gs.popNextActingCreature(); c != nil {
		gs.creatureEvents = append(gs.creatureEvents, gs.activate(c)...)
	}
	if len(gs.ForwardStack) == 0 && len(gs.ReverseStack) == 0 {
		gs.EndRound()
	}
}

// creatures that were defeated or bonded since the stack was drawn are skipped
func (gs *GameState) popNextActingCreature() *creature.Unit {
	for len(gs.ForwardStack) > 0 {
		c := gs.ForwardStack[0]
		gs.ForwardStack = gs.ForwardStack[1:]
		if gs.canAct(c) {
			gs.ReverseStack = slices.Insert(gs.ReverseStack, 0, c)
			return c
		}
	}
	for len(gs.ReverseStack) > 0 {
		c := gs.ReverseStack[0]
		gs.ReverseStack = gs.ReverseStack[1:]
		if gs.canAct(c) {
			return c
		}
	}
	return nil
}

func (gs *GameState) canAct(c *creature.Unit) bool {
	return !c.Bonded && !c.Role.IsPlant() && slices.Contains(gs.Creatures, c)
}

func (gs *GameState) activate(c *creature.Unit) []CreatureEvent {
	events := []CreatureEvent{}
	if c.Exhausted {
		c.Exhausted = false
		gs.Logf("%s readies", c.NameWithID())
		events = append(events, readyEvent(c))
	}
	if c.Role.IsPredator() {
		events = append(events, gs.ActivatePredator(c)...)
	} else {
		events = append(events, gs.ActivatePrey(c)...)
	}
	gs.eatAtLocation(c)
	gs.ReconcileDefeats()
	gs.ReconcileHiding()
	gs.ReconcileVision()
	return events
}

func (gs *GameState) EndRound() {
	gs.Cycle++
	gs.ResolveGrazing()
	gs.ResolveRegrowth()
	gs.ResolveCooldownReset()
	gs.StartRound()
	gs.RefreshHands()
}

func (gs *GameState) ResolveRegrowth() {
	for _, v := range gs.Nodes {
		v.Tick()
	}
	gs.removeDecayedNodes()
}

func (gs *GameState) removeDecayedNodes() {
	gs.Nodes = slices.DeleteFunc(gs.Nodes, func(n *collecting.Node) bool {
		if !n.HasDecayedAway() {
			return false
		}
		delete(gs.KnownNodes, n.ID)
		gs.Logf("%s has decayed away", n.Name)
		return true
	})
}

// Rest refills the energy of the players. The world acts as if three actions passed, and meanwhile
// the players and their bonded creatures are out of sight of the hunters.
func (gs *GameState) Rest() {
	gs.Logf("-- The players rest --")
	for _, p := range gs.Players {
		if !p.Fainted() {
			p.Energy = MaxEnergy
		}
	}
	gs.Resting = true
	for range restActions {
		gs.ActivateNextCreature()
	}
	gs.Resting = false
	gs.reconcileAfterAction()
}

// a hunter hunts what it can by size, unless the target rests out of sight
func (gs *GameState) canHunt(hunter *creature.Unit, target creature.Quarry) bool {
	return hunter.CanHunt(target) && !(gs.Resting && isRestingHidden(target))
}

func isRestingHidden(target creature.Quarry) bool {
	switch t := target.(type) {
	case *player.Unit:
		return true
	case *creature.Unit:
		return t.Bonded
	}
	return false
}

// DrainCreatureEvents returns what the creatures did since the last call.
func (gs *GameState) DrainCreatureEvents() []CreatureEvent {
	events := gs.creatureEvents
	gs.creatureEvents = nil
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
