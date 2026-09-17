package enemy

import (
	"aeons/internal/healthpool"
	"aeons/internal/location"
	"aeons/internal/moving"
)

type EnemyEntity struct {
	moving.MovingEntity
	healthpool.HealthPoolEntity
	ID     int
	Name   string
	Aloof  bool
	Hunter bool
	Damage int
}

func (ee *EnemyEntity) GenerateMoveGoal() *location.LocationEntity {

	return nil
}
