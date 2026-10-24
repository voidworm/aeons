package card

import (
	"aeons/internal/effect"
)

type  TargetType int 
const (
	TargetTypePlayer TargetType = iota
	TargetTypeCreature
	TargetTypeLocation
	TargetTypeNode // a place to harvest or forage at
)

type EffectType int
const (
	PlayerMove EffectType = iota
	PlayerDistract
	PlayerAttack
	PlayerGather
	PlayerBond
	PlayerMoveKindred
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

type PlayableCard struct {
	ID int
	Type EffectType
	CanBeCast bool
	Name string
	EffectText string
	Target Targetable
	TargetSpec *TargetSpec
	Candidates []Targetable
	Effect effect.Effect
	Cost int
	Range int
	OnCooldown bool // cannot be cast again until the next upkeep phase
}

func (c *PlayableCard) Play() {

	c.Effect.Apply()
}



