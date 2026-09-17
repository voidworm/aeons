package effect

import (
	"aeons/internal/moving"
)

type Effect interface {
	Apply()
}

type MoveEffectContext struct {
	TargetEntity moving.Movable
	MoveAmount   int
}

type MoveEffect struct {
	Ctx *MoveEffectContext
}

func (me *MoveEffect) Apply() {
	for i := 1; i <= me.Ctx.MoveAmount; i++ {
		goalLocation := me.Ctx.TargetEntity.GenerateMoveGoal()
		me.Ctx.TargetEntity.MoveTo(goalLocation)
	}
}
