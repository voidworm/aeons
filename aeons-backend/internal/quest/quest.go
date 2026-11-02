package quest

import (
	"aeons/internal/card"
	"aeons/internal/effect"
)

type Status int

const (
	StatusPending Status = iota // not reached yet
	StatusActive
	StatusFinished
	StatusAbandoned // its group finished without it
)

type Quest struct {
	ID        int
	Name      string
	Flavor    string
	Mandatory bool // while it is finishable, nothing but finishing it may be done
	Goals     []*Goal
	OnStart   []effect.Effect
	OnFinish  []effect.Effect
	Status    Status
}

func (q *Quest) Finishable(w World) bool {
	if q.Status != StatusActive {
		return false
	}
	for _, g := range q.Goals {
		if !g.Met(w) {
			return false
		}
	}
	return true
}

// the goal that makes the players pay when the quest is finished, if there is one
func (q *Quest) PaymentGoal() *Goal {
	for _, g := range q.Goals {
		if g.Payment != nil {
			return g
		}
	}
	return nil
}

func (q *Quest) begin() {
	q.Status = StatusActive
	applyAll(q.OnStart)
}

func (q *Quest) Finish() {
	q.Status = StatusFinished
	applyAll(q.OnFinish)
}

func (q *Quest) activeQuests() []*Quest {
	if q.Status == StatusActive {
		return []*Quest{q}
	}
	return nil
}

func (q *Quest) started() bool {
	return q.Status != StatusPending
}

func (q *Quest) done() bool {
	return q.Status == StatusFinished
}

func (q *Quest) abandon() {
	if q.Status == StatusActive {
		q.Status = StatusAbandoned
	}
}

func (q *Quest) GetTargetType() card.TargetType {
	return card.TargetTypeQuest
}

func (q *Quest) GetID() int {
	return q.ID
}

func applyAll(effects []effect.Effect) {
	for _, e := range effects {
		e.Apply()
	}
}
