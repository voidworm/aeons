package game

import (
	"aeons/internal/combat"
	"aeons/internal/enemy"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"aeons/internal/prompt"
	"fmt"
	"log"
)

func (gs *GameState) ResolvePlayerPhaseStep() error {
	currentPlayer, err := gs.PromptForNextPlayer()
	if err != nil {
		return err
	}

	cancelled, action, err := gs.PromptForPlayerAction(currentPlayer)
	if err != nil {
		return err
	} else if cancelled {
		gs.ResolvePlayerPhaseStep()
	}

	switch action {
	case "Move":

		cancelled, location, err := gs.ChooseTargetForPlayerMove(currentPlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolvePlayerPhaseStep()
		} else {
			currentPlayer.RemainingActions -= 1
			moveEffect := moving.MoveEffect{Entity: currentPlayer, Target: location}
			moveEffect.Apply()
		}

	case "Attack":

		cancelled, target, err := gs.ChooseTargetForPlayerAttack(currentPlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolvePlayerPhaseStep()
		} else {
			currentPlayer.RemainingActions -= 1
			effect := combat.DamageEffect{Source: currentPlayer, Target: target}
			effect.Apply()
		}
	case "Cancel":
		gs.ResolveAttackForPlayer(currentPlayer)
	default:
		fmt.Println("Selected unimplemented action")
	}

	gs.ReconcileDefeats()
	return nil
}

func (gs *GameState) ChooseTargetForPlayerMove(player *player.Unit) (bool, *location.LocationEntity, error) {
	targets := player.CurrentLocation.OutgoingConnections
	optionsArray := []string{}

	for _, v := range targets {
		optionsArray = append(optionsArray, v.Name)
	}

	cancelled, position, err := prompt.PromptCancellable(">>> --- Choose a location to move to --- <<<", optionsArray)

	if err != nil {
		return false, &location.LocationEntity{}, err
	} else if cancelled {
		return true, &location.LocationEntity{}, nil
	} else {
		return false, targets[position], nil
	}
}

func (gs *GameState) ChooseTargetForPlayerAttack(player *player.Unit) (bool, *enemy.EnemyEntity, error) {
	inRange := gs.EnemiesAtLocation(player.CurrentLocation)
	promptList := []string{}
	for _, v := range inRange {
		promptItem := fmt.Sprintf("%s (Remaining Health: %d/%d)", v.Name, v.CurrentHealth, v.MaxHealth)
		promptList = append(promptList, promptItem)
	}

	cancelled, position, err := prompt.PromptCancellable(">>> --- Chose a target to attack --- <<<", promptList)
	if err != nil {
		return false, &enemy.EnemyEntity{}, err
	} else if cancelled {
		return true, &enemy.EnemyEntity{}, nil
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
