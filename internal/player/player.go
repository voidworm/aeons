package player

import (
	"aeons/internal/combat"
	"aeons/internal/moving"
)

type Unit struct {
	moving.MovingEntity
	combat.HealthPool
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
	RemainingActions   int
	Damage             int
}

func (p *Unit) OutgoingDamage() int {
	return p.Damage
}
