package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/evading"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"aeons/internal/prompt"
	"fmt"
	"log"
)

func (gs *GameState) ResolvePlayerPhaseStep() error {
	activePlayer, err := gs.PromptForNextPlayer()
	if err != nil {
		return err
	}

	cancelled, action, err := gs.PromptForPlayerAction(activePlayer)
	if err != nil {
		return err
	} else if cancelled {
		gs.ResolvePlayerPhaseStep()
	}

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
	return nil
}

func (gs *GameState) ChooseTargetForPlayerMove(player *player.Unit) (bool, *location.Unit, error) {
	targets := player.CurrentLocation.OutgoingConnections
	optionsArray := []string{}

	for _, v := range targets {
		optionsArray = append(optionsArray, v.Name)
	}

	cancelled, position, err := prompt.PromptCancellable(">>> --- Choose a location to move to --- <<<", optionsArray)

	if err != nil {
		return false, &location.Unit{}, err
	} else if cancelled {
		return true, &location.Unit{}, nil
	} else {
		return false, targets[position], nil
	}
}

func (gs *GameState) ChooseEvadeTargetForPlayer(player *player.Unit) (bool, *creature.Unit, error) {
	inRange := gs.EvadableEnemiesAtLocation(player.CurrentLocation)
	promptList := []string{}
	for _, v := range inRange {
		promptItem := fmt.Sprintf("%s (Remaining Health: %d/%d)", v.Name, v.CurrentHealth, v.MaxHealth)
		promptList = append(promptList, promptItem)
	}

	cancelled, position, err := prompt.PromptCancellable(">>> --- Chose a target to evade --- <<<", promptList)
	if err != nil {
		return false, &creature.Unit{}, err
	} else if cancelled {
		return true, &creature.Unit{}, nil
	} else {
		return false, inRange[position], nil
	}
}

func (gs *GameState) ChooseTargetForPlayerAttack(player *player.Unit) (bool, *creature.Unit, error) {
	inRange := gs.EnemiesAtLocation(player.CurrentLocation)
	promptList := []string{}
	for _, v := range inRange {
		promptItem := fmt.Sprintf("%s (Remaining Health: %d/%d)", v.Name, v.CurrentHealth, v.MaxHealth)
		promptList = append(promptList, promptItem)
	}

	cancelled, position, err := prompt.PromptCancellable(">>> --- Chose a target to attack --- <<<", promptList)
	if err != nil {
		return false, &creature.Unit{}, err
	} else if cancelled {
		return true, &creature.Unit{}, nil
	} else {
		return false, inRange[position], nil
	}
}

func (gs *GameState) ChooseHarvestableForHarvestAction(player *player.Unit) (bool, *harvesting.Unit, error) {
	inRange := gs.HarvestUnitsAtLocation(player.CurrentLocation)
	promptList := []string{}
	for _, v := range inRange {
		promptItem := fmt.Sprintf("%s (Remaining: %d Healing, %d Resources)", v.Name, v.HealYieldConfig.CurrentCapacity, v.ResourceYieldConfig.CurrentCapacity)
		promptList = append(promptList, promptItem)
	}

	cancelled, position, err := prompt.PromptCancellable(">>> --- What will you harvest? --- <<<", promptList)
	if err != nil {
		return false, &harvesting.Unit{}, err
	} else if cancelled {
		return true, &harvesting.Unit{}, nil
	} else {
		return false, inRange[position], nil
	}
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
