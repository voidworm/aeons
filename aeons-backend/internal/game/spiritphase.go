package game

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/collecting"
	"aeons/internal/location"
	"aeons/internal/player"
	"fmt"
	"log"
	"slices"
)

// PlayCard casts a card from a player's hand on the target with the given id.
// Move, distract, attack, bond and gather cards are implemented so far.
// Gather cards return the collecting effect, all others return nil. The mode only matters to gather cards.
func (gs *GameState) PlayCard(cardID, targetID int, mode collecting.Mode) (*collecting.Effect, error) {
	if gs.PendingPayment != nil {
		return nil, fmt.Errorf("%s waits for a payment", gs.PendingPayment.Name)
	}
	c, owner := gs.findCardInHands(cardID)
	if c == nil {
		return nil, fmt.Errorf("card %d is not in any hand", cardID)
	}
	if !c.CanBeCast {
		return nil, fmt.Errorf("%s cannot be cast right now", c.Name)
	}
	if owner.Energy < c.Cost {
		return nil, fmt.Errorf("%s costs %d energy, %s has %d", c.Name, c.Cost, owner.Name, owner.Energy)
	}

	if c.Type == card.PlayerFinishQuest {
		return nil, gs.playFinishQuest(targetID)
	}
	if c.Type == card.PlayerRest {
		gs.Rest()
		return nil, nil
	}

	switch c.Type {
	case card.PlayerMove:
		target, ok := getEntryWithID(gs.Locations, targetID).(*location.Unit)
		if !ok {
			return nil, fmt.Errorf("%d is not a location", targetID)
		}
		if !slices.Contains(owner.CurrentLocation.OutgoingConnections, target) {
			return nil, fmt.Errorf("%s is not connected to %s", owner.CurrentLocation.Name, target.Name)
		}
		c.Target = target
	case card.PlayerHide:
		if targetID != owner.CurrentLocation.ID {
			return nil, fmt.Errorf("%s can only hide at their own location", owner.Name)
		}
		c.Target = owner.CurrentLocation
	case card.PlayerDistract, card.PlayerAttack, card.PlayerBond:
		target, ok := getEntryWithID(gs.Creatures, targetID).(*creature.Unit)
		if !ok {
			return nil, fmt.Errorf("%d is not a creature", targetID)
		}
		if target.CurrentLocation != owner.CurrentLocation {
			return nil, fmt.Errorf("%s is not at %s", target.Name, owner.CurrentLocation.Name)
		}
		if !slices.Contains(c.Candidates, card.Targetable(target)) {
			return nil, fmt.Errorf("%s cannot target %s", c.Name, target.NameWithID())
		}
		c.Target = target
	case card.PlayerMoveKindred:
		kindred := gs.KindredOf(owner)
		if kindred == nil {
			return nil, fmt.Errorf("%s has no kindred", owner.Name)
		}
		target, ok := getEntryWithID(gs.Locations, targetID).(*location.Unit)
		if !ok {
			return nil, fmt.Errorf("%d is not a location", targetID)
		}
		if !slices.Contains(kindred.CurrentLocation.OutgoingConnections, target) {
			return nil, fmt.Errorf("%s is not connected to %s", kindred.CurrentLocation.Name, target.Name)
		}
		c.Target = target
	case card.PlayerGather:
		target, ok := getEntryWithID(gs.Nodes, targetID).(*collecting.Node)
		if !ok {
			return nil, fmt.Errorf("%d is not a node", targetID)
		}
		if target.StaticLocation != owner.CurrentLocation {
			return nil, fmt.Errorf("%s is not at %s", target.Name, owner.CurrentLocation.Name)
		}
		if !target.CanCollect(mode) {
			return nil, fmt.Errorf("%s has nothing to %s", target.Name, mode)
		}
		c.Target = target
	default:
		return nil, fmt.Errorf("%s cannot be played yet", c.Name)
	}

	owner.Energy -= c.Cost
	gs.fillEffectForCard(c, owner, mode)
	follower := gs.KindredFollowing(owner, c)
	c.Play()
	if kindred := follower; kindred != nil {
		kindred.MoveTo(owner.CurrentLocation)
		gs.Logf("%s follows %s to %s", kindred.NameWithID(), owner.Name, owner.CurrentLocation.Name)
	}
	if c.Type == card.PlayerMoveKindred {
		c.OnCooldown = true
	}
	if kindred := gs.KindredOf(owner); kindred != nil && (follower != nil || c.Type == card.PlayerMoveKindred) {
		gs.ReactToArrival(kindred)
	}
	gs.logPlayedCard(c, owner)
	if breaksHiding(c) {
		gs.breakHiding(owner)
	}
	gs.reconcileAfterAction()
	gs.ActivateNextCreature()
	collected, _ := c.Effect.(*collecting.Effect)
	return collected, nil
}

// casting a card ends hiding, except for Hide itself, finishing a quest, moving the kindred and resting
func breaksHiding(c *card.PlayableCard) bool {
	return c.Type != card.PlayerHide && c.Type != card.PlayerFinishQuest && c.Type != card.PlayerMoveKindred && c.Type != card.PlayerRest
}

func (gs *GameState) breakHiding(p *player.Unit) {
	if p.Hidden {
		p.Hidden = false
		gs.Logf("%s is no longer hidden", p.Name)
	}
}

func (gs *GameState) reconcileAfterAction() {
	gs.ReconcileDefeats()
	gs.ReconcileHiding()
	gs.ReconcileVision()
	gs.RefreshHands()
}

func (gs *GameState) findCardInHands(cardID int) (*card.PlayableCard, *player.Unit) {
	for _, p := range gs.Players {
		for _, c := range p.CardsInHand {
			if c.ID == cardID {
				return c, p
			}
		}
	}
	return nil, nil
}

func (gs *GameState) ResolveSpiritPhaseStep() error {
	

	/*
	switch action {
	case "Move":

		cancelled, location, err := gs.ChooseTargetForPlayerMove(activePlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolveSpiritPhaseStep()
		} else {
			activePlayer.RemainingActions -= 1
			moveEffect := moving.MoveEffect{Entity: activePlayer, Target: location}
			moveEffect.Apply()
		}

	case "Attack":

		cancelled, target, err := gs.ChooseTargetForPlayerAttack(activePlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolveSpiritPhaseStep()
		} else {
			activePlayer.RemainingActions -= 1
			effect := combat.DamageEffect{Target: target}
			effect.Apply()
		}
	case "Harvest":

		cancelled, target, err := gs.ChooseHarvestableForHarvestAction(activePlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolveSpiritPhaseStep()
		} else {
			activePlayer.RemainingActions -= 1
			effect := harvesting.Effect{TargetHarvestable: target, Player: activePlayer}
			effect.Apply()
		}
	case "Evade":
		cancelled, target, err := gs.ChooseEvadeTargetForPlayer(activePlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolveSpiritPhaseStep()
		} else {
			activePlayer.RemainingActions -= 1
			effect := evading.Effect{EvadingPlayer: activePlayer, EvadedCreature: target}
			effect.Apply()
		}
	case "Yield":
		activePlayer.RemainingActions -= 1
	case "Cancel":
		gs.ResolveSpiritPhaseStep()
	default:
		fmt.Println("Selected unimplemented action")
	}

	gs.ReconcileDefeats()
	*/
	return nil
}

func (gs *GameState) ChooseTargetForPlayerMove(player *player.Unit) (bool, *location.Unit, error) {
	//old ui trash has been removed
	return false, nil, nil
}

func (gs *GameState) ChooseEvadeTargetForPlayer(player *player.Unit) (bool, *creature.Unit, error) {
	//old ui trash has been removed
	return false, nil, nil
}

func (gs *GameState) ChooseTargetForPlayerAttack(player *player.Unit) (bool, *creature.Unit, error) {
	//old ui trash has been removed
	return false, nil, nil
}

func (gs *GameState) ResolveAttackForPlayer(player *player.Unit) {

	cancelled, enemy, err := gs.ChooseTargetForPlayerAttack(player)

	if err != nil {
		log.Println("The attack failed because the input was not an enemy.")
		if cancelled {
			//wemustgoback
		}
	}

	log.Printf("%s attacks %s up to %d/%d distress.\n", player.Name, enemy.NameWithID(), enemy.Distress, enemy.MaxDistress)
	if enemy.IsDefeated() {
		log.Printf("%s has defeated %s!", player.Name, enemy.NameWithID())
	}

	effect := combat.DamageEffect{Target: enemy}
	effect.Apply()
}
