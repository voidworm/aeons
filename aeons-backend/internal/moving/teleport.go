package moving

import "aeons/internal/location"

type TeleportEffect struct {
	Entity Movable
	Target *location.Unit
}

func (te *TeleportEffect) Apply() {
	te.Entity.MoveTo(te.Target)
}
