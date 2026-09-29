package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/player"
	"log"
)

func (gs *GameState) ResolvePlayerPhaseStep() error {
	

	/*
	switch action {
	case "Move":

		cancelled, location, err := gs.ChooseTargetForPlayerMove(activePlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolvePlayerPhaseStep()
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
			gs.ResolvePlayerPhaseStep()
		} else {
			activePlayer.RemainingActions -= 1
			effect := combat.DamageEffect{Source: activePlayer, Target: target}
			effect.Apply()
		}
	case "Harvest":

		cancelled, target, err := gs.ChooseHarvestableForHarvestAction(activePlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolvePlayerPhaseStep()
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
			gs.ResolvePlayerPhaseStep()
		} else {
			activePlayer.RemainingActions -= 1
			effect := evading.Effect{EvadingPlayer: activePlayer, EvadedCreature: target}
			effect.Apply()
		}
	case "Yield":
		activePlayer.RemainingActions -= 1
	case "Cancel":
		gs.ResolvePlayerPhaseStep()
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

func (gs *GameState) ChooseHarvestableForHarvestAction(player *player.Unit) (bool, *harvesting.Unit, error) {
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

	log.Printf("%s attacks %s down to %d/%d health.\n", player.Name, enemy.Name, enemy.CurrentHealth, enemy.MaxHealth)
	if enemy.CurrentHealth == 0 {
		log.Printf("%s has defeated %s!", player.Name, enemy.Name)
	}

	effect := combat.DamageEffect{Source: player, Target: enemy}
	effect.Apply()
}
