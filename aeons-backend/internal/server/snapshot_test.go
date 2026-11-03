package server

import (
	"aeons/internal/game"
	"testing"
)

func TestHiddenCreatureIsOnlySeenIfThePlayersSawItHide(t *testing.T) {
	gs := &game.GameState{}
	gs.Init()
	here := gs.Players[0].CurrentLocation
	hare := gs.NewHare(here)
	gs.Creatures = append(gs.Creatures, hare)

	cases := []struct {
		name          string
		seenHiding    bool
		fogLifted     bool
		wantInSnapshot bool
	}{
		{"hid in sight", true, false, true},
		{"hid unseen", false, false, false},
		{"hid unseen, fog lifted", false, true, true},
	}
	for _, c := range cases {
		hare.Hide(c.seenHiding)
		found := false
		for _, dto := range Snapshot(gs, c.fogLifted).Creatures {
			if dto.ID == hare.ID {
				found = true
				if !dto.Hidden {
					t.Errorf("%s: the creature is not marked hidden", c.name)
				}
			}
		}
		if found != c.wantInSnapshot {
			t.Errorf("%s: in snapshot = %v, want %v", c.name, found, c.wantInSnapshot)
		}
	}
}
