package hiding

import (
	"aeons/internal/player"
	"log"
)

type Effect struct {
	HidingPlayer *player.Unit
}

func (e *Effect) Apply() {
	e.HidingPlayer.Hidden = true
	log.Printf("[Spirit Phase] %s hides.", e.HidingPlayer.Name)
}
