package game

import "fmt"

type Phase string

const (
	PhaseSpirit    Phase = "spirit"
	PhaseWorld     Phase = "world"
	PhaseAwakening Phase = "awakening"
)

// EndSpiritPhase moves from the spirit to the world phase and resolves it.
// The game then waits for StartAwakeningPhase.
func (gs *GameState) EndSpiritPhase() ([]CreatureEvent, error) {
	if gs.Phase != PhaseSpirit {
		return nil, fmt.Errorf("cannot end the spirit phase during the %s phase", gs.Phase)
	}
	gs.Phase = PhaseWorld
	return gs.ResolveWorldPhase(), nil
}

// StartAwakeningPhase moves from the world to the awakening phase and resolves it.
// The game then waits for StartSpiritPhase.
func (gs *GameState) StartAwakeningPhase() ([]AwakeningEvent, error) {
	if gs.Phase != PhaseWorld {
		return nil, fmt.Errorf("cannot start the awakening phase during the %s phase", gs.Phase)
	}
	gs.Phase = PhaseAwakening
	gs.Cycle++
	return gs.ResolveAwakeningPhase(), nil
}

func (gs *GameState) StartSpiritPhase() error {
	if gs.Phase != PhaseAwakening {
		return fmt.Errorf("cannot start the spirit phase during the %s phase", gs.Phase)
	}
	gs.Phase = PhaseSpirit
	gs.Logf("-- Spirit phase --")
	gs.RefreshHands()
	return nil
}
