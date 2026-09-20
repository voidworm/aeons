package enemy

import (
	"aeons/internal/combat"
	"aeons/internal/location"
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

func (ee *EnemyEntity) GenerateMoveGoal(target *location.LocationEntity) *location.LocationEntity {
	return ee.GetShortestPathTo(target)[1]
}

func (ee *EnemyEntity) OutgoingDamage() int {
	return ee.Damage
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
	eme.TargetEnemy.MoveTo(eme.TargetLocation)
}
