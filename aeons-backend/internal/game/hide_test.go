package game

import (
	"aeons/internal/card"
	"testing"
)

func hideCardOf(gs *GameState, playerIndex int) *card.PlayableCard {
	return cardOfType(gs.Players[playerIndex].CardsInHand, card.PlayerHide)
}

func TestHideCanBeCastWithAPredatorAroundAndCostsTwoEnergy(t *testing.T) {
	gs, plains := newBoard()
	druid := gs.Players[0]
	druid.CurrentLocation = plains
	gs.place(gs.NewWolfpack(plains))
	gs.RefreshHands()

	hide := hideCardOf(gs, 0)
	if !hide.CanBeCast || hide.Cost != 2 {
		t.Errorf("castable = %v, cost = %d, want castable for 2 energy", hide.CanBeCast, hide.Cost)
	}
	energy := druid.Energy
	if _, err := gs.PlayCard(hide.ID, plains.ID, ""); err != nil {
		t.Fatal(err)
	}
	if druid.Energy != energy-2 {
		t.Errorf("energy = %d, want %d", druid.Energy, energy-2)
	}
}

func TestHiddenPlayersAreIgnoredByHuntersAndCannotHideAgain(t *testing.T) {
	gs, plains := newBoard()
	druid := gs.Players[0]
	druid.CurrentLocation = plains
	wolfpack := gs.place(gs.NewWolfpack(plains))
	wolfpack.Exhausted = true
	gs.RefreshHands()

	hide := hideCardOf(gs, 0)
	if _, err := gs.PlayCard(hide.ID, plains.ID, ""); err != nil {
		t.Fatal(err)
	}
	if !druid.Hidden || hideCardOf(gs, 0).CanBeCast {
		t.Errorf("hidden = %v, Hide castable = %v, want hidden and Hide disabled", druid.Hidden, hideCardOf(gs, 0).CanBeCast)
	}
	wolfpack.Exhausted = false
	for _, target := range gs.HuntableAt(wolfpack) {
		if target == druid {
			t.Error("a hunter targets a hidden player")
		}
	}
}

func TestCastingAnyOtherCardBreaksHiding(t *testing.T) {
	gs, plains := newBoard()
	druid := gs.Players[0]
	druid.CurrentLocation = plains
	gs.RefreshHands()
	if _, err := gs.PlayCard(hideCardOf(gs, 0).ID, plains.ID, ""); err != nil {
		t.Fatal(err)
	}

	move := cardOfType(druid.CardsInHand, card.PlayerMove)
	if _, err := gs.PlayCard(move.ID, plains.OutgoingConnections[0].ID, ""); err != nil {
		t.Fatal(err)
	}
	if druid.Hidden {
		t.Error("the player stays hidden after casting a card")
	}
	druid.Energy = MaxEnergy
	gs.RefreshHands()
	if !hideCardOf(gs, 0).CanBeCast {
		t.Error("Hide stays disabled after hiding ended")
	}
}

func TestHideOnlyTargetsTheOwnLocation(t *testing.T) {
	gs, plains := newBoard()
	gs.Players[0].CurrentLocation = plains
	gs.RefreshHands()
	if _, err := gs.PlayCard(hideCardOf(gs, 0).ID, plains.OutgoingConnections[0].ID, ""); err == nil {
		t.Error("Hide was cast onto another location")
	}
}

func TestMovingTheKindredDoesNotBreakHiding(t *testing.T) {
	gs, plains := newBoard()
	druid := gs.Players[0]
	druid.CurrentLocation = plains
	kindred := gs.place(gs.NewStag(plains))
	kindred.Bonded, kindred.BondedTo = true, druid
	gs.RefreshHands()
	if _, err := gs.PlayCard(hideCardOf(gs, 0).ID, plains.ID, ""); err != nil {
		t.Fatal(err)
	}

	moveKindred := cardOfType(druid.CardsInHand, card.PlayerMoveKindred)
	if _, err := gs.PlayCard(moveKindred.ID, plains.OutgoingConnections[0].ID, ""); err != nil {
		t.Fatal(err)
	}
	if !druid.Hidden {
		t.Error("moving the kindred broke hiding")
	}
}
