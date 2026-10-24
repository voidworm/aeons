package game

import (
	"aeons/internal/card"
	"aeons/internal/creature"
	"aeons/internal/player"
	"slices"
)

func (gs *GameState) KindredOf(p *player.Unit) *creature.Unit {
	for _, c := range gs.Creatures {
		if c.Bonded && c.BondedTo == p {
			return c
		}
	}
	return nil
}

// a player with a kindred holds Move Kindred instead of Bond, the swap follows the creatures' bond state
func (gs *GameState) ReconcileKindredCards() {
	for _, p := range gs.Players {
		hasKindred := gs.KindredOf(p) != nil
		bond := cardOfType(p.CardsInHand, card.PlayerBond)
		moveKindred := cardOfType(p.CardsInHand, card.PlayerMoveKindred)
		if hasKindred && bond != nil {
			gs.replaceCardInHand(p, bond, gs.GenerateKindredMoveCard())
		}
		if !hasKindred && moveKindred != nil {
			gs.replaceCardInHand(p, moveKindred, gs.GenerateBasicBondCard())
		}
	}
}

func cardOfType(hand []*card.PlayableCard, kind card.EffectType) *card.PlayableCard {
	for _, c := range hand {
		if c.Type == kind {
			return c
		}
	}
	return nil
}

func (gs *GameState) replaceCardInHand(p *player.Unit, old, replacement *card.PlayableCard) {
	p.CardsInHand[slices.Index(p.CardsInHand, old)] = replacement
	gs.PlayableCards[slices.Index(gs.PlayableCards, old)] = replacement
}

func (gs *GameState) RefreshHands() {
	gs.ReconcileKindredCards()
	for _, p := range gs.Players {
		gs.EvaluatePlayerHandPlayability(p)
	}
}

func (gs *GameState) ResolveCooldownReset() {
	for _, p := range gs.Players {
		for _, c := range p.CardsInHand {
			c.OnCooldown = false
		}
	}
}
