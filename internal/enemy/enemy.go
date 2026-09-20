package enemy

import (
	"aeons/internal/combat"
	"aeons/internal/moving"
)

type EnemyEntity struct {
	moving.MovingEntity
	combat.HealthPool
	ID     int
	Name   string
	Aloof  bool
	Hunter bool
	Damage int
}

func (ee *EnemyEntity) OutgoingDamage() int {
	return ee.Damage
}
