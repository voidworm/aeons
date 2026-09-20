package game

import (
	"fmt"
	"log"

	"aeons/internal/enemy"
	"aeons/internal/location"
	"aeons/internal/player"
	"aeons/internal/prompt"
)

type GameState struct {
	Locations   []*location.LocationEntity
	Players     []*player.Player
	Enemies     []*enemy.EnemyEntity
	TurnCounter int
}

func (gs *GameState) ReconcileDefeats() {
	alivePlayers := gs.Players[:0]

	for _, v := range gs.Players {
		if v.CurrentHealth > 0 {
			alivePlayers = append(alivePlayers, v)
		}
	}

	aliveEnemies := gs.Enemies[:0]
	for _, v := range gs.Enemies {
		if v.CurrentHealth > 0 {
			aliveEnemies = append(aliveEnemies, v)
		}
	}
	gs.Enemies = aliveEnemies
	gs.Players = alivePlayers
}

func (gs *GameState) StartNewTurn() {
	for _, v := range gs.Players {
		v.RemainingActions = 3
	}
	gs.TurnCounter += 1
}

func (gs *GameState) GetLocationByName(name string) *location.LocationEntity {
	//assumes locations are named uniquely!
	for _, v := range gs.Locations {
		if v.Name == name {
			return v
		}
	}

	return &location.LocationEntity{}
}

func (gs *GameState) LocationHasEnemies(le *location.LocationEntity) bool {
	for _, v := range gs.Enemies {
		if le == v.Location {
			return true
		}
	}

	return false
}

func (gs *GameState) LocationHasPlayers(le *location.LocationEntity) bool {
	for _, v := range gs.Players {
		if le == v.Location {
			return true
		}
	}

	return false
}

func (gs *GameState) EnemiesAtLocation(le *location.LocationEntity) []*enemy.EnemyEntity {

	found := []*enemy.EnemyEntity{}
	for _, v := range gs.Enemies {
		if le == v.Location {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) PlayersAtLocation(ee *enemy.EnemyEntity) []*player.Player {

	found := []*player.Player{}
	for _, v := range gs.Players {
		if ee.Location == v.Location {
			found = append(found, v)
		}
	}

	return found
}

func (gs *GameState) ResolvePlayerPhaseStep() error {
	nextPlayer, err := gs.PromptForNextPlayer()
	if err != nil {
		return err
	}

	cancelled, action, err := gs.PromptForPlayerAction(nextPlayer)
	if err != nil {
		return err
	} else if cancelled {
		gs.ResolvePlayerPhaseStep()
	}

	switch action {
	case "Move":

		cancelled, location, err := gs.ChooseTargetForPlayerMove(nextPlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolvePlayerPhaseStep()
		} else {
			nextPlayer.RemainingActions -= 1
			pme := player.PlayerMoveEffect{TargetPlayer: nextPlayer, TargetLocation: location}
			pme.Apply()
		}

	case "Attack":
		nextPlayer.RemainingActions -= 1
		cancelled, target, err := gs.ChooseTargetForPlayerAttack(nextPlayer)
		if err != nil {
			return err
		} else if cancelled {
			gs.ResolvePlayerPhaseStep()
		} else {
			pae := player.PlayerAttackEffect{TargetPlayer: nextPlayer, TargetEnemy: target, DamageAmount: 1}
			pae.Apply()
		}
	case "Cancel":
		gs.ResolveAttackForPlayer(nextPlayer)
	default:
		fmt.Println("Selected unimplemented action")
	}

	gs.ReconcileDefeats()
	return nil
}

func (gs *GameState) PlayerHaveActionsRemaining() bool {
	for _, v := range gs.Players {
		if v.RemainingActions > 0 {
			return true
		}
	}
	return false
}

func (gs *GameState) ChooseTargetForPlayerMove(player *player.Player) (bool, *location.LocationEntity, error) {
	targets := player.Location.OutgoingConnections
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

func (gs *GameState) ChooseTargetForPlayerAttack(player *player.Player) (bool, *enemy.EnemyEntity, error) {
	inRange := gs.EnemiesAtLocation(player.Location)
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

func (gs *GameState) ResolveAttackForPlayer(player *player.Player) {

	cancelled, enemy, err := gs.ChooseTargetForPlayerAttack(player)

	if err != nil {
		log.Println("The attack failed because the input was not an enemy.")
		if cancelled {
			//wemustgoback
		}
	}

	enemy.TakeDamage(1)
	log.Printf("%s attacks %s down to %d/%d health.\n", player.Name, enemy.Name, enemy.CurrentHealth, enemy.MaxHealth)
	if enemy.CurrentHealth == 0 {
		log.Printf("%s has defeated %s!", player.Name, enemy.Name)
	}
}
