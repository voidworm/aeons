package collecting

import (
	"aeons/internal/player"
	"testing"
)

func TestPoolRegrowsOneUnitPerTimerUpToMax(t *testing.T) {
	p := NewPool(1, 2, 3)

	for range 3 {
		p.Tick()
	}
	if p.Remaining != 2 {
		t.Fatalf("remaining after one timer = %d, want 2", p.Remaining)
	}

	for range 3 {
		p.Tick()
	}
	if p.Remaining != 2 {
		t.Fatalf("remaining grew past max: %d", p.Remaining)
	}
}

func TestPoolWithoutTimerNeverRegrows(t *testing.T) {
	p := NewPool(0, 3, 0)
	for range 10 {
		p.Tick()
	}
	if p.Remaining != 0 {
		t.Fatalf("remaining = %d, want 0", p.Remaining)
	}
}

func TestHarvestTakesFoodForageTakesMaterial(t *testing.T) {
	n := GenerateBerryBush(1, nil)
	pl := &player.Unit{}

	(&Effect{Player: pl, Node: n, Mode: Harvest}).Apply()
	(&Effect{Player: pl, Node: n, Mode: Forage}).Apply()

	if pl.Food != 1 || pl.ResourcesAvailable != 1 {
		t.Fatalf("player has %d food and %d material, want 1 and 1", pl.Food, pl.ResourcesAvailable)
	}
	if n.Food.Remaining != 1 || n.Material.Remaining != 1 {
		t.Fatalf("node has %d food and %d material, want 1 and 1", n.Food.Remaining, n.Material.Remaining)
	}
}

func TestModesListOnlyWhatIsLeft(t *testing.T) {
	n := GenerateMushroomField(1, nil)
	if got := n.Modes(); len(got) != 1 || got[0] != Harvest {
		t.Fatalf("modes = %v, want [harvest]", got)
	}
}
