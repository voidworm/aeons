package player

import (
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

func (pe *Player) GenerateMoveGoal() *location.LocationEntity {
	return pe.PromptMoveTargetSelection()
}
