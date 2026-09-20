package game

import (
	"aeons/internal/player"
	"aeons/internal/prompt"
	"fmt"
)

func (gs *GameState) PromptForNextPlayer() (*player.Player, error) {

	selectable := []*player.Player{}
	for _, v := range gs.Players {
		if v.RemainingActions > 0 {
			selectable = append(selectable, v)
		}
	}

	activeArray := []string{}
	for _, v := range selectable {
		activeArray = append(activeArray, fmt.Sprintf("%s (currently at %s)", v.Name, v.CurrentLocation.Name))
	}

	label := ">>> --- Choose a player to act --- <<<"
	index, err := prompt.Prompt(label, activeArray)
	if err != nil {
		return &player.Player{}, err
	}

	nextTurn := selectable[index]
	fmt.Printf("%s selected to act!\n", nextTurn.Name)
	return nextTurn, nil
}

func (gs *GameState) PromptForPlayerAction(player *player.Player) (bool, string, error) {
	actionArray := []string{"Move", "Draw", "Resource"}

	if gs.LocationHasEnemies(player.CurrentLocation) {
		actionArray = append(actionArray, "Attack", "Evade")
	}

	label := fmt.Sprintf("<<< --- What will %s do? --- >>> ", player.Name)

	cancelled, index, err := prompt.PromptCancellable(label, actionArray)
	if err != nil {
		return false, "", err
	} else if cancelled {
		return true, "", nil
	} else {
		return false, actionArray[index], nil
	}

}
