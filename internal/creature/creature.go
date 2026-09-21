package creature

import (
	"aeons/internal/combat"
	"aeons/internal/moving"
)

type Unit struct {
	moving.MovingEntity
	combat.HealthPool
	ID     int
	Name   string
	Aloof  bool
	Hunter bool
	Damage int
}

func (ee *Unit) OutgoingDamage() int {
	return ee.Damage
}
