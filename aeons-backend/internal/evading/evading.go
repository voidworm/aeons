package evading

import (
	"aeons/internal/creature"
	"aeons/internal/player"
	"log"
)

type Effect struct {
	EvadingPlayer  *player.Unit
	EvadedCreature *creature.Unit
}

func (ee *Effect) Apply() {
	ee.EvadedCreature.Exhausted = true
	log.Printf("[Player Phase] %s has evaded %s.", ee.EvadingPlayer.Name, ee.EvadedCreature.Name)
}
