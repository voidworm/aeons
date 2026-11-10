package game

import "aeons/internal/card"

// SetAdminMode gives every player the Teleport card while it is on, and takes it away when it is off
func (gs *GameState) SetAdminMode(on bool) {
	gs.AdminMode = on
	gs.RefreshHands()
}

func (gs *GameState) ReconcileTeleportCards() {
	for _, p := range gs.Players {
		teleport := cardOfType(p.CardsInHand, card.PlayerTeleport)
		switch {
		case gs.AdminMode && teleport == nil:
			teleport = gs.GenerateTeleportCard()
			p.CardsInHand = append(p.CardsInHand, teleport)
			gs.PlayableCards = append(gs.PlayableCards, teleport)
		case !gs.AdminMode && teleport != nil:
			gs.removeCardFromHand(p, teleport)
		}
	}
}

func (gs *GameState) GenerateTeleportCard() *card.PlayableCard {
	return &card.PlayableCard{
		ID:         gs.NextID(),
		Type:       card.PlayerTeleport,
		Name:       "Teleport",
		EffectText: "*Move* to any location.",
		Flavor:     "Go with the wind.",
		TargetSpec: &card.TargetSpec{Type: card.TargetTypeLocation},
		Candidates: []card.Targetable{},
		Cost:       0,
	}
}
