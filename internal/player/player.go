package player

import (
	"aeons/internal/combat"
	"aeons/internal/location"
	"aeons/internal/moving"
)

type Player struct {
	moving.MovingEntity
	combat.HealthPool
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
	RemainingActions   int
	Damage             int
}

func (p *Player) OutgoingDamage() int {
	return p.Damage
}

type PlayerMoveEffect struct {
	TargetPlayer   *Player
	TargetLocation *location.LocationEntity
}

func (pem *PlayerMoveEffect) Apply() {
	pem.TargetPlayer.MoveTo(pem.TargetLocation)
}
