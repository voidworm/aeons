package game

import (
	"aeons/internal/card"
	"testing"
)

func teleportCardOf(gs *GameState, playerIndex int) *card.PlayableCard {
	return cardOfType(gs.Players[playerIndex].CardsInHand, card.PlayerTeleport)
}

func TestTeleportIsOnlyInHandsWhileAdminModeIsOn(t *testing.T) {
	gs, _ := newBoard()
	if teleportCardOf(gs, 0) != nil {
		t.Fatal("teleport in hand without admin mode")
	}

	gs.SetAdminMode(true)
	teleport := teleportCardOf(gs, 0)
	if teleport == nil || !teleport.CanBeCast || teleport.Cost != 0 {
		t.Fatalf("teleport = %+v, want a castable card without cost", teleport)
	}

	gs.SetAdminMode(false)
	if teleportCardOf(gs, 0) != nil {
		t.Error("teleport still in hand after admin mode ended")
	}
}

func TestTeleportMovesThePlayerAndKindredAtTheSameLocationToAnyLocation(t *testing.T) {
	gs, plains := newBoard()
	druid := gs.Players[0]
	druid.CurrentLocation = plains
	kindred := gs.place(gs.NewStag(plains))
	kindred.Bonded, kindred.BondedTo = true, druid
	far := gs.Locations[len(gs.Locations)-1]
	for _, l := range gs.Locations {
		if l != plains && plains.StepsTo(l) > 1 {
			far = l
		}
	}
	gs.SetAdminMode(true)

	if _, err := gs.PlayCard(teleportCardOf(gs, 0).ID, far.ID, ""); err != nil {
		t.Fatal(err)
	}
	if druid.CurrentLocation != far || kindred.CurrentLocation != far {
		t.Errorf("druid at %s, kindred at %s, want both at %s", druid.CurrentLocation.Name, kindred.CurrentLocation.Name, far.Name)
	}
}
