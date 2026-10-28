package game

import (
	"aeons/internal/collecting"
	"aeons/internal/player"
	"testing"
)

func newGame() *GameState {
	gs := &GameState{}
	gs.Init()
	return gs
}

func TestOnlyPlayerLocationIsUnveiledAtStart(t *testing.T) {
	gs := newGame()
	for _, l := range gs.Locations {
		isStart := l == gs.Players[0].CurrentLocation
		if l.Veiled == isStart {
			t.Errorf("%s: veiled = %v, want %v", l.Name, l.Veiled, !isStart)
		}
	}
}

func TestLeftLocationIsAMemoryWithItsLastKnownNode(t *testing.T) {
	gs := newGame()
	start := gs.Players[0].CurrentLocation
	node := collecting.GenerateBerryBush(gs.NextID(), start)
	gs.Nodes = append(gs.Nodes, node)
	gs.ReconcileVision()
	lastKnown := node.Food.Remaining

	for _, p := range gs.Players {
		p.CurrentLocation = start.OutgoingConnections[0]
	}
	node.Food.Remaining--
	gs.ReconcileVision()

	if !start.Veiled || !start.Known {
		t.Fatalf("veiled = %v, known = %v, want a veiled memory", start.Veiled, start.Known)
	}
	if got := gs.KnownNodes[node.ID].Food.Remaining; got != lastKnown {
		t.Errorf("remembered food = %d, want %d", got, lastKnown)
	}
}

func TestFaintedPlayersGiveNoVision(t *testing.T) {
	gs := newGame()
	for _, p := range gs.Players {
		p.Wounds = player.MaxWounds
	}
	gs.ReconcileVision()
	for _, l := range gs.Locations {
		if !l.Veiled {
			t.Errorf("%s is unveiled without a player able to see", l.Name)
		}
	}
}
