package player

import (
	"aeons/internal/enemy"
	"aeons/internal/healthpool"
	"aeons/internal/location"
	"aeons/internal/moving"
)

type Player struct {
	moving.MovingEntity
	healthpool.HealthPoolEntity
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
	RemainingActions   int
}

type PlayerMoveEffect struct {
	TargetPlayer   *Player
	TargetLocation *location.LocationEntity
}

func (pem *PlayerMoveEffect) Apply() {
	pem.TargetPlayer.MoveTo(pem.TargetLocation)
}

type PlayerAttackEffect struct {
	TargetPlayer *Player
	TargetEnemy  *enemy.EnemyEntity
	DamageAmount int
}

func (pae *PlayerAttackEffect) Apply() {
	pae.TargetEnemy.TakeDamage(pae.DamageAmount)
}
