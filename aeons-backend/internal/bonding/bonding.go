package bonding

import (
	"aeons/internal/creature"
	"aeons/internal/player"
	"log"
)

type Effect struct {
	BondingPlayer  *player.Unit
	BondedCreature *creature.Unit
}

func (ee *Effect) Apply() {
	ee.BondedCreature.Bond()
	log.Printf("[Player Phase] %s has bonded with %s (%d/%d).", ee.BondingPlayer.Name, ee.BondedCreature.NameWithID(), ee.BondedCreature.BondCounter, ee.BondedCreature.BondThreshold)
}
