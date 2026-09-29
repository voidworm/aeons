package card

import (
	"aeons/internal/effect"
)

type  TargetType int 
const (
	TargetTypePlayer TargetType = iota
	TargetTypeCreature
	TargetTypeLocation
	TargetTypeHarvest
)

type EffectType int
const (
	PlayerMove EffectType = iota
	PlayerDistract
	PlayerAttack
	PlayerHarvest
)

type Targetable interface {
	GetTargetType() TargetType
	GetID()	 int
}

type TargetSpec struct {
	Type TargetType
	Target Targetable
	Filter func(Targetable) bool
}

func (s TargetSpec) Accepts(t Targetable) bool {
	if t.GetTargetType() != s.Type {
		return false
	}
	return s.Filter == nil || s.Filter(t)
}

type Card struct {
	ID int
	Type EffectType
	CanBeCast bool
	Name string
	EffectText string
	Target Targetable
	TargetSpec *TargetSpec
	Candidates []Targetable
	Effect effect.Effect
	Range int
}

func (c *Card) Play() {

	c.Effect.Apply()
}



