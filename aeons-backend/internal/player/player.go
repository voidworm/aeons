package player

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/moving"
)

type Unit struct {
	moving.MovingEntity
	combat.HealthPool
	ID                 int
	Name               string
	CardsInHand        []*card.Card
	ResourcesAvailable int
	RemainingActions   int
	Damage             int
}

func (p *Unit) OutgoingDamage() int {
	return p.Damage
}


func (u *Unit) GetTargetType() card.TargetType {
	return card.TargetTypePlayer
}

func (u *Unit) GetID() int {
	return u.ID
}

