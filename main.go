package main

import (
	"aeons/internal/game"
	"log"
)

func main() {

	gs := &game.GameState{TurnCounter: 1}
	gs.Init()
	running := true

	for running {
		log.Printf("Starting Turn %d", gs.TurnCounter)
		for gs.PlayerHaveActionsRemaining() {
			err := gs.ResolvePlayerPhaseStep()
			if err != nil {
				log.Println(err)
				return
			}
		}
		gs.ResolveCreaturePhase()
		gs.ResolveUpkeepPhase()
	}
}
