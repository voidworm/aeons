package game

import (
	"aeons/internal/player"
	"aeons/internal/prompt"
	"fmt"
)

func (gs *GameState) PromptForNextPlayer() (*player.Unit, error) {

	selectable := []*player.Unit{}
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
		return &player.Unit{}, err
	}

	nextTurn := selectable[index]
	fmt.Printf("%s selected to act!\n", nextTurn.Name)
	return nextTurn, nil
}

func (gs *GameState) PromptForPlayerAction(player *player.Unit) (bool, string, error) {
	actionArray := []string{}

	if gs.LocationCanBeLeft(player.CurrentLocation) {
		actionArray = append(actionArray, "Move")
	}

	if gs.LocationHasEnemies(player.CurrentLocation) {
		actionArray = append(actionArray, "Attack")
	}

	if gs.LocationHaEvadableCreatures(player.CurrentLocation) {
		actionArray = append(actionArray, "Evade")
	}

	if gs.LocationHasHarvest(player.CurrentLocation) {
		actionArray = append(actionArray, "Harvest")
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
