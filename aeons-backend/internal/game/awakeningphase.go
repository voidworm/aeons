package game

const AwakeningEnergy = 4

// AwakeningEvent records one creature recovering from exhaustion in the awakening phase.
type AwakeningEvent struct {
	CreatureID int
}

// ResolveAwakeningPhase runs the awakening phase and returns what happened, in order.
func (gs *GameState) ResolveAwakeningPhase() []AwakeningEvent {
	gs.Logf("-- Awakening phase --")
	gs.ResolveRegrowth()
	gs.ResolvePlayerRefresh()

	gs.ResolveCooldownReset()
	gs.ReconcileVision()
	gs.RefreshHands()
	return []AwakeningEvent{}
}

func (gs *GameState) ResolveRegrowth() {
	for _, v := range gs.Nodes {
		v.Tick()
	}
}

func (gs *GameState) ResolvePlayerRefresh() {
	for _, v := range gs.Players {
		if v.Fainted() {
			continue
		}
		v.RemainingActions = 3
		v.Energy = AwakeningEnergy
		gs.Logf("%s refreshes to %d energy", v.Name, AwakeningEnergy)
	}
}
