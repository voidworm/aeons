package game

const UpkeepEnergy = 4

// UpkeepEvent records one creature recovering from exhaustion in the upkeep phase.
type UpkeepEvent struct {
	CreatureID int
}

// ResolveUpkeepPhase runs the upkeep phase and returns what happened, in order.
func (gs *GameState) ResolveUpkeepPhase() []UpkeepEvent {
	gs.Logf("-- Upkeep phase --")
	gs.ResolveRegrowth()
	gs.ResolvePlayerRefresh()
	events := gs.ResolveCreatureRecovery()

	gs.ResolveCooldownReset()
	gs.RefreshHands()
	return events
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
		v.Energy = UpkeepEnergy
		gs.Logf("%s refreshes to %d energy", v.Name, UpkeepEnergy)
	}
}

// ResolveCreatureRecovery ends the turn state of all creatures, exhausted ones produce an event.
func (gs *GameState) ResolveCreatureRecovery() []UpkeepEvent {
	events := []UpkeepEvent{}
	for _, v := range gs.Creatures {
		if v.Distracted {
			events = append(events, UpkeepEvent{CreatureID: v.ID})
			gs.Logf("%s recovers from exhaustion", v.NameWithID())
		}
		v.Tick()
	}
	return events
}
