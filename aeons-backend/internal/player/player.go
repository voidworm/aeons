package player

import (
	"aeons/internal/card"
	"aeons/internal/moving"
)

const (
	MaxWounds   = 3
	MaxCarrying = 10 // food and resources share this inventory
	Size        = 6  // players count as prey of this size to hunters
)

type Unit struct {
	moving.MovingEntity
	ID                 int
	Name               string
	CardsInHand        []*card.PlayableCard
	Wounds             int
	Food               int
	ResourcesAvailable int
	RemainingActions   int
	Energy             int
	Hidden             bool // hidden players are ignored by animals until they cast a card
}

func (p *Unit) TakeDamage(amount int) int {
	p.Wounds = min(p.Wounds+amount, MaxWounds)
	if p.Fainted() {
		p.Energy = 0
	}
	return p.Wounds
}

func (p *Unit) Fainted() bool {
	return p.Wounds >= MaxWounds
}

// Gather fills the shared inventory, food first. What doesn't fit is lost.
func (p *Unit) Gather(food, material int) {
	space := MaxCarrying - p.Food - p.ResourcesAvailable
	foodTaken := min(food, space)
	p.Food += foodTaken
	p.ResourcesAvailable += min(material, space-foodTaken)
}

// a wound counts as one distress, a player can take as many distress as wounds
func (p *Unit) RemainingDistress() int {
	return MaxWounds - p.Wounds
}

func (p *Unit) GetSize() int {
	return Size
}

func (p *Unit) IsHuntable() bool {
	return !p.Fainted() && !p.Hidden
}

func (u *Unit) GetTargetType() card.TargetType {
	return card.TargetTypePlayer
}

func (u *Unit) GetID() int {
	return u.ID
}
