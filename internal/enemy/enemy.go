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

func (me *EnemyEntity) GenerateMoveGoal(target *location.LocationEntity) *location.LocationEntity {
	return me.MovementPathTowards(target)[1]
}

// setter for location
// overwrite if you need to do stuff before or after moving
func (ee *EnemyEntity) MoveTo(target *location.LocationEntity) {
	ee.Location = target
}

type EnemyMoveEffect struct {
	TargetEnemy    *EnemyEntity
	TargetLocation *location.LocationEntity
}

func (eme *EnemyMoveEffect) Apply() {
	//path := eme.TargetEnemy.MovementPathTowards(eme.TargetLocation)
	eme.TargetEnemy.MoveTo(eme.TargetLocation)
}
