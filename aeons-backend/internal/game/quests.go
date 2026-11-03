package game

import (
	"aeons/internal/card"
	"aeons/internal/player"
	"aeons/internal/quest"
	"fmt"
	"slices"
)

func (gs *GameState) QuestWorld() quest.World {
	return quest.World{Players: gs.Players, Creatures: gs.Creatures, Locations: gs.Locations}
}

func (gs *GameState) ActiveQuests() []*quest.Quest {
	if gs.Scenario == nil {
		return nil
	}
	return gs.Scenario.ActiveQuests()
}

func (gs *GameState) FinishableQuests() []*quest.Quest {
	finishable := []*quest.Quest{}
	for _, q := range gs.ActiveQuests() {
		if q.Finishable(gs.QuestWorld()) {
			finishable = append(finishable, q)
		}
	}
	return finishable
}

// a finishable mandatory quest keeps the players from doing anything else
func (gs *GameState) MandatoryFinishPending() bool {
	return slices.ContainsFunc(gs.FinishableQuests(), func(q *quest.Quest) bool { return q.Mandatory })
}

// every player holds Finish Quest while there is something to finish, targeting what can be finished
func (gs *GameState) ReconcileQuestCards() {
	finishable := toTargetables(gs.FinishableQuests())
	for _, p := range gs.Players {
		finishCard := cardOfType(p.CardsInHand, card.PlayerFinishQuest)
		switch {
		case len(finishable) > 0 && finishCard == nil:
			finishCard = gs.GenerateFinishQuestCard()
			p.CardsInHand = append(p.CardsInHand, finishCard)
			gs.PlayableCards = append(gs.PlayableCards, finishCard)
		case len(finishable) == 0 && finishCard != nil:
			gs.removeCardFromHand(p, finishCard)
		}
		if finishCard != nil {
			finishCard.Candidates = finishable
		}
	}
}

func (gs *GameState) removeCardFromHand(p *player.Unit, c *card.PlayableCard) {
	p.CardsInHand = slices.DeleteFunc(p.CardsInHand, func(inHand *card.PlayableCard) bool { return inHand == c })
	gs.PlayableCards = slices.DeleteFunc(gs.PlayableCards, func(known *card.PlayableCard) bool { return known == c })
}

func (gs *GameState) GenerateFinishQuestCard() *card.PlayableCard {
	return &card.PlayableCard{
		ID:         gs.NextID(),
		Type:       card.PlayerFinishQuest,
		Name:       "Finish Quest",
		TargetSpec: &card.TargetSpec{Type: card.TargetTypeQuest},
		Candidates: []card.Targetable{},
		Range:      1,
		Cost:       0,
	}
}

// a quest that makes the players pay waits for their payment, otherwise it is finished right away
func (gs *GameState) playFinishQuest(questID int) error {
	q := gs.finishableQuestWithID(questID)
	if q == nil {
		return fmt.Errorf("%d is not a quest that can be finished", questID)
	}
	if q.PaymentGoal() != nil {
		gs.PendingPayment = q
		return nil
	}
	gs.finishQuest(q)
	gs.reconcileAfterAction()
	return nil
}

func (gs *GameState) finishableQuestWithID(id int) *quest.Quest {
	for _, q := range gs.FinishableQuests() {
		if q.ID == id {
			return q
		}
	}
	return nil
}

// payments maps a player id to the resources that player pays
func (gs *GameState) PayForPendingQuest(questID int, payments map[int]int) error {
	q := gs.PendingPayment
	if q == nil || q.ID != questID {
		return fmt.Errorf("quest %d is not waiting for a payment", questID)
	}
	if gs.finishableQuestWithID(questID) == nil {
		return fmt.Errorf("%s cannot be finished any more", q.Name)
	}
	payment := q.PaymentGoal().Payment
	payers := payment.Payers(gs.QuestWorld())

	total := 0
	for id, amount := range payments {
		i := slices.IndexFunc(payers, func(p *player.Unit) bool { return p.ID == id })
		if i == -1 {
			return fmt.Errorf("player %d cannot pay for %s", id, q.Name)
		}
		if amount < 0 || amount > payers[i].ResourcesAvailable {
			return fmt.Errorf("%s cannot pay %d resources", payers[i].Name, amount)
		}
		total += amount
	}
	if total != payment.Amount {
		return fmt.Errorf("%s costs %d resources, %d were offered", q.Name, payment.Amount, total)
	}

	for _, p := range payers {
		p.ResourcesAvailable -= payments[p.ID]
		if payments[p.ID] > 0 {
			gs.Logf("%s pays %d resources", p.Name, payments[p.ID])
		}
	}
	gs.PendingPayment = nil
	gs.finishQuest(q)
	gs.reconcileAfterAction()
	return nil
}

func (gs *GameState) CancelPayment() {
	gs.PendingPayment = nil
}

func (gs *GameState) finishQuest(q *quest.Quest) {
	gs.Logf("Quest finished: %s", q.Name)
	q.Finish()
	gs.Scenario.Advance()
}
