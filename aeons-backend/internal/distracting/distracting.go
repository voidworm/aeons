package distracting

import (
	"aeons/internal/creature"
	"aeons/internal/player"
	"log"
)

type Effect struct {
	DistractingPlayer  *player.Unit
	DistractedCreature *creature.Unit
}

func (ee *Effect) Apply() {
	ee.DistractedCreature.Distracted = true
	log.Printf("[Player Phase] %s has distracted %s.", ee.DistractingPlayer.Name, ee.DistractedCreature.Name)
}
